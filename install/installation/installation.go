package installation

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/kyma-incubator/hydroform/install/scheme"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/kyma-incubator/hydroform/install/k8s"

	"k8s.io/client-go/discovery"

	"k8s.io/apimachinery/pkg/watch"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"

	installationClientset "github.com/kyma-project/kyma/components/kyma-operator/pkg/client/clientset/versioned"

	"errors"

	"github.com/kyma-project/kyma/components/kyma-operator/pkg/apis/installer/v1alpha1"
	installationTyped "github.com/kyma-project/kyma/components/kyma-operator/pkg/client/clientset/versioned/typed/installer/v1alpha1"
	"k8s.io/client-go/rest"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/restmapper"
)

const (
	ComponentOverridesLabelKey = "component"

	installerOverridesLabelKey = "installer"
	installerOverridesLabelVal = "overrides"

	installationActionLabel              = "action"
	defaultInstallationResourceNamespace = "default"

	kymaInstallerNamespace = "kyma-installer"
	kymaInstallationName   = "kyma-installation"

	kubeSystemNamespace      = "kube-system"
	tillerLabelSelector      = "name=tiller"
	defaultTillerWaitTimeout = 2 * time.Minute
	tillerCheckInterval      = 2 * time.Second

	defaultWatcherTimeoutSeconds = 3600
)

type Installer interface {
	PrepareInstallation(installation Installation) error
	PrepareUpgrade(artifacts Installation) error
	StartInstallation(context context.Context) (<-chan InstallationState, <-chan error, error)
}

type Logger interface {
	Infof(format string, a ...interface{})
}

// Installation provides configuration for Kyma installation
type Installation struct {
	// TillerYaml is a content of yaml file with all resources related to Tiller which are required by Kyma
	TillerYaml string
	// InstallerYaml is a content of yaml file with all resources related to and required by Installer
	InstallerYaml string
	// Configuration specifies the configuration to be used for the installation
	Configuration Configuration
}

const (
	NoInstallationState = "NoInstallation"
)

type InstallationState struct {
	State       string
	Description string
}

func CheckInstallationState(kubeconfig *rest.Config) (InstallationState, error) {
	installationClient, err := installationClientset.NewForConfig(kubeconfig)
	if err != nil {
		return InstallationState{}, err
	}

	installationCR, err := installationClient.
		InstallerV1alpha1().
		Installations(defaultInstallationResourceNamespace).
		Get(kymaInstallationName, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return InstallationState{
				State:       NoInstallationState,
				Description: "Kyma Installation CR not found on the cluster",
			}, nil
		}

		return InstallationState{}, err
	}

	return getInstallationState(*installationCR)
}

func TriggerUninstall(kubeconfig *rest.Config) error {
	installationClient, err := installationClientset.NewForConfig(kubeconfig)
	if err != nil {
		return fmt.Errorf("error creating Installation client: %s", err.Error())
	}

	installationCR, err := installationClient.
		InstallerV1alpha1().
		Installations(defaultInstallationResourceNamespace).
		Get(kymaInstallationName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("error getting Installation client: %s", err.Error())
	}

	if installationCR.Labels == nil {
		installationCR.Labels = map[string]string{}
	}

	installationCR.Labels[installationActionLabel] = "uninstall"

	_, err = installationClient.InstallerV1alpha1().
		Installations(defaultInstallationResourceNamespace).
		Update(installationCR)
	if err != nil {
		return fmt.Errorf("error labeling Installation CR with action=uninstall label: %s", err.Error())
	}

	return nil
}

// NewKymaInstaller initializes new KymaInstaller configured to work with the cluster from provided Kubeconfig
func NewKymaInstaller(kubeconfig *rest.Config, opts ...InstallationOption) (*KymaInstaller, error) {
	options := &installationOptions{
		installationCRModificationFunc: func(installation *v1alpha1.Installation) {},
		tillerWaitTime:                 defaultTillerWaitTimeout,
	}

	for _, o := range opts {
		o.apply(options)
	}

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(kubeconfig)
	if err != nil {
		return nil, err
	}

	groupResources, err := restmapper.GetAPIGroupResources(discoveryClient)
	if err != nil {
		return nil, err
	}

	restMapper := restmapper.NewDiscoveryRESTMapper(groupResources)

	dynamicClient, err := dynamic.NewForConfig(kubeconfig)
	if err != nil {
		return nil, err
	}

	installationClient, err := installationClientset.NewForConfig(kubeconfig)
	if err != nil {
		return nil, err
	}

	coreClient, err := kubernetes.NewForConfig(kubeconfig)
	if err != nil {
		return nil, err
	}

	decoder, err := scheme.DefaultDecoder()
	if err != nil {
		return nil, err
	}

	return &KymaInstaller{
		installationOptions:               options,
		installationWatcherTimeoutSeconds: defaultWatcherTimeoutSeconds,
		decoder:                           decoder,
		k8sGenericClient:                  k8s.NewGenericClient(restMapper, dynamicClient, coreClient),
		installationClient:                installationClient.InstallerV1alpha1().Installations(defaultInstallationResourceNamespace),
	}, nil
}

type KymaInstaller struct {
	*installationOptions

	installationWatcherTimeoutSeconds int64
	decoder                           runtime.Decoder
	k8sGenericClient                  *k8s.GenericClient
	installationClient                installationTyped.InstallationInterface
}

// PrepareInstallation creates all the required resources for Kyma Installation.
// It installs Tiller and deploys the Kyma Installer to the cluster, together with provided configuration
// It does not start the installation.
func (k KymaInstaller) PrepareInstallation(artifacts Installation) error {
	k.infof("Preparing Kyma Installation...")

	err := k.installTiller(artifacts.TillerYaml, k.k8sGenericClient.CreateResources)
	if err != nil {
		return err
	}

	err = k.deployInstallerForIstallation(artifacts.InstallerYaml)
	if err != nil {
		return err
	}

	err = k.applyConfiguration(artifacts.Configuration)
	if err != nil {
		return err
	}

	k.infof("Ready to start installation.")
	return nil
}

func (k KymaInstaller) PrepareUpgrade(artifacts Installation) error {
	k.infof("Preparing Kyma Upgrade...")

	err := k.installTiller(artifacts.TillerYaml, k.k8sGenericClient.ApplyResources)
	if err != nil {
		return err
	}

	err = k.deployInstallerForUpgrade(artifacts.InstallerYaml)
	if err != nil {
		return err
	}

	err = k.applyConfiguration(artifacts.Configuration)
	if err != nil {
		return err
	}

	k.infof("Ready to start upgrade.")
	return nil
}

// StartInstallation triggers Kyma installation to start.
// It expects that the cluster is already prepared meaning the Tiller is installed and the Kyma Installer is deployed
// and Installation CR exists
func (k KymaInstaller) StartInstallation(context context.Context) (<-chan InstallationState, <-chan error, error) {
	err := checkContextNotCanceled(context)
	if err != nil {
		return nil, nil, fmt.Errorf("context already canceled: %w", err)
	}

	err = k.triggerInstallation()
	if err != nil {
		return nil, nil, err
	}

	stateChannel := make(chan InstallationState)
	errorChannel := make(chan error)

	go k.waitForInstallation(context, stateChannel, errorChannel)

	return stateChannel, errorChannel, nil
}

func checkContextNotCanceled(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

func (k KymaInstaller) installTiller(tillerYaml string, createFunction func([]k8s.K8sObject) error) error {
	k.infof("Preparing Tiller installation...")
	k8sTillerObjects, err := k8s.ParseYamlToK8sObjects(k.decoder, tillerYaml)
	if err != nil {
		return fmt.Errorf("failed to parse Tiller yaml file to Kubernetes dynamicClientObjects: %w", err)
	}

	k.infof("Deploying Tiller...")
	err = createFunction(k8sTillerObjects)
	if err != nil {
		return fmt.Errorf("failed to apply Tiller resources: %w", err)
	}
	k.infof("Tiller installed successfully")

	k.infof("Waiting for Tiller to start...")
	err = k.k8sGenericClient.WaitForPodByLabel(kubeSystemNamespace, tillerLabelSelector, corev1.PodRunning, k.installationOptions.tillerWaitTime, tillerCheckInterval)
	if err != nil {
		return fmt.Errorf("timeout waiting for Tiller to start running: %w", err)
	}
	k.infof("Tiller is running")

	return nil
}

func (k KymaInstaller) deployInstallerForIstallation(installerYaml string) error {
	return k.deployInstaller(installerYaml, k.k8sGenericClient.CreateResources)
}

func (k KymaInstaller) deployInstallerForUpgrade(installerYaml string) error {
	return k.deployInstaller(installerYaml, k.k8sGenericClient.ApplyResources)
}

func (k KymaInstaller) deployInstaller(installerYaml string, createResourcesFunc func(resources []k8s.K8sObject) error) error {
	k.infof("Deploying Installer...")

	k8sInstallerObjects, err := k8s.ParseYamlToK8sObjects(k.decoder, installerYaml)
	if err != nil {
		return fmt.Errorf("failed to parse Installer yaml file to Kubernetes dynamicClientObjects: %w", err)
	}

	var installationCR *v1alpha1.Installation
	installationCR, k8sInstallerObjects, err = k.extractInstallationCR(k8sInstallerObjects)
	if err != nil {
		return fmt.Errorf("failed to get Installation CR: %w", err)
	}

	_, found := installationCR.Labels[installationActionLabel]
	if found {
		delete(installationCR.Labels, installationActionLabel)
	}

	err = createResourcesFunc(k8sInstallerObjects)
	if err != nil {
		return fmt.Errorf("failed to apply Installer resources: %w", err)
	}

	k.infof("Applying Installation CR modifications...")
	k.installationOptions.installationCRModificationFunc(installationCR)
	k.infof("Applying Installation CR...")
	err = k.applyInstallationCR(installationCR)
	if err != nil {
		return fmt.Errorf("failed to apply Installation resources: %w", err)
	}
	k.infof("Installer deployed.")
	return nil
}

func (k KymaInstaller) applyConfiguration(configuration Configuration) error {
	configMaps, secrets := configurationToK8sResources(configuration)

	err := k.k8sGenericClient.ApplyConfigMaps(configMaps, kymaInstallerNamespace)
	if err != nil {
		return fmt.Errorf("failed to create configuration config maps: %s", err.Error())
	}

	err = k.k8sGenericClient.ApplySecrets(secrets, kymaInstallerNamespace)
	if err != nil {
		return fmt.Errorf("failed to create configuration secrets: %s", err.Error())
	}

	return nil
}

func (k KymaInstaller) applyInstallationCR(installationCR *v1alpha1.Installation) error {
	if installationCR.Namespace == "" {
		installationCR.Namespace = defaultInstallationResourceNamespace
	}

	_, err := k.installationClient.Create(installationCR)
	if err != nil {
		if k8serrors.IsAlreadyExists(err) {
			k.infof("installation %s already exists, trying to update...", installationCR.Name)
			get, err := k.installationClient.Get(installationCR.Name, metav1.GetOptions{})
			if err != nil {
				return fmt.Errorf("installation CR already exists, failed to get installation CR: %w", err)
			}

			installationCR.ResourceVersion = get.ResourceVersion

			_, err = k.installationClient.Update(installationCR)
			if err != nil {
				return fmt.Errorf("installation CR already exists, failed to updated installation CR: %w", err)
			}
			return nil
		}
		return fmt.Errorf("failed to apply Installation CR: %w", err)
	}

	return nil
}

// extractInstallationCR finds and removes first Installation CR from the slice of K8sObjects and returns it
func (k KymaInstaller) extractInstallationCR(k8sObjects []k8s.K8sObject) (*v1alpha1.Installation, []k8s.K8sObject, error) {
	for i, k8sObject := range k8sObjects {
		if k8sObject.GVK.Group == v1alpha1.Group && k8sObject.GVK.Kind == "Installation" {
			installationCR, ok := k8sObject.Object.(*v1alpha1.Installation)
			if !ok {
				return nil, nil, fmt.Errorf("unexpected type of Installation object: %T, failed to cast to *Installation", k8sObject.Object)
			}

			k8sObjects = append(k8sObjects[:i], k8sObjects[i+1:]...)
			return installationCR, k8sObjects, nil
		}
	}

	return nil, k8sObjects, fmt.Errorf("installation object not found in the objects slice")
}

func (k KymaInstaller) triggerInstallation() error {
	installation, err := k.installationClient.Get(kymaInstallationName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if installation.Status.State == v1alpha1.StateInProgress {
		return fmt.Errorf("failed to trigger installation, installation already in progress")
	}

	if installation.Labels == nil {
		installation.Labels = map[string]string{}
	}

	installation.Labels[installationActionLabel] = "install"

	_, err = k.installationClient.Update(installation)
	if err != nil {
		return fmt.Errorf("failed label Installation CR: %w", err)
	}

	return nil
}

func (k KymaInstaller) waitForInstallation(context context.Context, stateChannel chan<- InstallationState, errorChannel chan<- error) {
	defer close(errorChannel)
	defer close(stateChannel)

	installationWatcher, err := k.newInstallationWatcher(k.installationWatcherTimeoutSeconds)
	if err != nil {
		errorChannel <- fmt.Errorf("failed to setup installation watcher: %w", err)
		return
	}

	installationWatchChan := installationWatcher.ResultChan()

	for {
		select {
		case <-context.Done():
			errorChannel <- fmt.Errorf("context canceled, waiting for installation interrupted")
			return
		case event, ok := <-installationWatchChan:
			if !ok {
				installationWatcher, err := k.newInstallationWatcher(k.installationWatcherTimeoutSeconds)
				if err != nil {
					errorChannel <- fmt.Errorf("failed to renew installation watcher: %w", err)
					return
				}

				installationWatchChan = installationWatcher.ResultChan()
				break
			}

			installationStatus, err := handleInstallationEvent(event)
			if err != nil {
				if errors.Is(err, installationObjectDeleted) {
					errorChannel <- fmt.Errorf("installation CR deleted unexpectedly")
					return
				}
				errorChannel <- err
			} else {
				stateChannel <- installationStatus
				if installationStatus.State == string(v1alpha1.StateInstalled) {
					return
				}
			}
		default:
			time.Sleep(1 * time.Second)
		}
	}
}

func (k KymaInstaller) newInstallationWatcher(timeout int64) (watch.Interface, error) {
	return k.installationClient.Watch(metav1.ListOptions{FieldSelector: fmt.Sprintf("%s=%s", "metadata.name", kymaInstallationName), TimeoutSeconds: &timeout})
}

var installationObjectDeleted error = fmt.Errorf("installation object deleted")

func handleInstallationEvent(event watch.Event) (InstallationState, error) {
	switch event.Type {
	case watch.Added:
		return InstallationState{
			State:       "Starting",
			Description: "Installation starting",
		}, nil
	case watch.Modified:
		installation, ok := event.Object.(*v1alpha1.Installation)
		if !ok {
			return InstallationState{}, fmt.Errorf("installation watcher returned invalid type %T, expected Installation", event.Object)
		}

		return getInstallationState(*installation)
	case watch.Error:
		return InstallationState{}, fmt.Errorf("installation watch error occured: %s", tryToExtractErrorStatus(event.Object))
	case watch.Deleted:
		return InstallationState{}, installationObjectDeleted
	default:
		return InstallationState{}, fmt.Errorf("received watch event of unexpected type: %s", event.Type)
	}
}

func getInstallationState(installation v1alpha1.Installation) (InstallationState, error) {
	switch installation.Status.State {
	case v1alpha1.StateEmpty:
		return InstallationState{
			State:       "",
			Description: installation.Status.Description,
		}, nil
	case v1alpha1.StateInstalled:
		return InstallationState{
			State:       string(v1alpha1.StateInstalled),
			Description: installation.Status.Description,
		}, nil
	case v1alpha1.StateError:
		return InstallationState{}, newInstallationError(installation)
	case v1alpha1.StateInProgress:
		return InstallationState{
			State:       string(v1alpha1.StateInProgress),
			Description: installation.Status.Description,
		}, nil
	default:
		return InstallationState{}, fmt.Errorf("invalid installation state: %s", installation.Status.State)
	}
}

func newInstallationError(installation v1alpha1.Installation) InstallationError {
	installationError := InstallationError{
		ShortMessage: fmt.Sprintf("installation error occurred: %s", installation.Status.Description),
		ErrorEntries: make([]ErrorEntry, 0, len(installation.Status.ErrorLog)),
	}

	for _, errLog := range installation.Status.ErrorLog {
		installationError.ErrorEntries = append(installationError.ErrorEntries, ErrorEntry{
			Component:   errLog.Component,
			Log:         errLog.Log,
			Occurrences: errLog.Occurrences,
		})
	}

	return installationError
}

func tryToExtractErrorStatus(object runtime.Object) string {
	status, ok := object.(*metav1.Status)
	if !ok {
		return "unable to extract watch status"
	}

	errorStatusCauses := ""
	for _, c := range status.Details.Causes {
		errorStatusCauses = fmt.Sprintf("%sType: %s, Message: %s\n", errorStatusCauses, c.Type, c.Message)
	}

	return strings.TrimSuffix(errorStatusCauses, "\n")
}

func (k KymaInstaller) infof(format string, a ...interface{}) {
	if k.logger != nil {
		k.logger.Infof(format, a...)
	}
}

package installation

import (
	"context"
	"fmt"
	"time"

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

// TODO - comments

const (
	installationActionLabel              = "action"
	defaultInstallationResourceNamespace = "default"

	kymaInstallerNamespace = "kyma-installer"
	kymaInstallationName   = "kyma-installation"

	tillerNamespace          = "kube-system"
	tillerLabelSelector      = "name=tiller"
	defaultTillerWaitTimeout = 2 * time.Minute
	tillerCheckInterval      = 2 * time.Second

	defaultWatcherTimeoutSeconds = 3600

	installerOverridesLabelKey = "installer"
	installerOverridesLabelVal = "overrides"
	ComponentOverridesLabelKey = "component"
)

// TODO - consider adding CleanupInstallation method

type Installer interface {
	PrepareInstallation(artifacts Installation) error
	StartInstallation(context context.Context) (<-chan InstallationState, <-chan error, error)
}

type Logger interface {
	Infof(format string, a ...interface{})
}

type Installation struct {
	TillerYaml    string
	InstallerYaml string
	Configuration Configuration
}

type InstallationState struct {
	State       string
	Description string
}

// TODO - installation error

func NewKymaInstaller(kubeconfig *rest.Config, opts ...InstallationOption) (*KymaInstaller, error) {
	options := &installationOptions{
		installationCRModificationFunc: func(installation *v1alpha1.Installation) {},
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

	installtionClient, err := installationClientset.NewForConfig(kubeconfig)
	if err != nil {
		return nil, err
	}

	coreClient, err := kubernetes.NewForConfig(kubeconfig)
	if err != nil {
		return nil, err
	}

	decoder, err := DefaultDecoder()
	if err != nil {
		return nil, err
	}

	return &KymaInstaller{
		installationOptions:               options,
		installationWatcherTimeoutSeconds: defaultWatcherTimeoutSeconds,
		tillerWaitTimeout:                 defaultTillerWaitTimeout,
		yamlParser:                        k8s.NewK8sYamlParser(decoder),
		k8sGenericClient:                  k8s.NewGenericClient(restMapper, dynamicClient, coreClient, installtionClient),
		installationClient:                installtionClient.InstallerV1alpha1().Installations(defaultInstallationResourceNamespace),
	}, nil
}

type KymaInstaller struct {
	*installationOptions

	installationWatcherTimeoutSeconds int64
	tillerWaitTimeout                 time.Duration
	yamlParser                        *k8s.YamlParser
	k8sGenericClient                  *k8s.GenericClient
	installationClient                installationTyped.InstallationInterface
}

func (k KymaInstaller) PrepareInstallation(artifacts Installation) error {
	k.infof("Preparing Kyma Installation...")

	err := k.installTiller(artifacts.TillerYaml)
	if err != nil {
		return err
	}

	err = k.deployInstaller(artifacts.InstallerYaml)
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

func (k KymaInstaller) installTiller(tillerYaml string) error {
	k.infof("Preparing Tiller installation...")
	k8sTillerObjects, err := k.yamlParser.ParseYamlToK8sObjects(tillerYaml)
	if err != nil {
		return fmt.Errorf("failed to parse Tiller yaml file to Kubernetes dynamicClientObjects: %w", err)
	}

	k.infof("Deploying Tiller...")
	err = k.k8sGenericClient.ApplyResources(k8sTillerObjects)
	if err != nil {
		return fmt.Errorf("failed to apply Tiller resources: %w", err)
	}
	k.infof("Tiller installed successfully")

	k.infof("Waiting for Tiller to start...")
	err = k.k8sGenericClient.WaitForPodByLabel(tillerNamespace, tillerLabelSelector, corev1.PodRunning, k.tillerWaitTimeout, tillerCheckInterval)
	if err != nil {
		return fmt.Errorf("timeout waiting for Tiller to start running: %w", err)
	}
	k.infof("Tiller is running")

	return nil
}

func (k KymaInstaller) deployInstaller(installerYaml string) error {
	k.infof("Deploying Installer...")

	k8sInstallerObjects, err := k.yamlParser.ParseYamlToK8sObjects(installerYaml)
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

	err = k.k8sGenericClient.ApplyResources(k8sInstallerObjects)
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
	configMaps, secrets := ConfigurationToK8sResources(configuration)

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
			_, err := k.installationClient.Update(installationCR)
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
			errorChannel <- errors.New("context canceled, waiting for installation interrupted")
			return
		case event, ok := <-installationWatchChan:
			if !ok {
				// TODO - with retries?
				installationWatcher, err := k.newInstallationWatcher(k.installationWatcherTimeoutSeconds)
				if err != nil {
					errorChannel <- fmt.Errorf("failed to update installation watcher: %w", err)
					return
				}

				installationWatchChan = installationWatcher.ResultChan()
				break
			}
			fmt.Println("Received event: ", event.Type)

			installationStatus, err := handleInstallationEvent(event)
			if err != nil {
				if errors.Is(err, installationObjectDeleted) {
					errorChannel <- fmt.Errorf("installation CR deleted unexpectedly")
					return
				}
				errorChannel <- err
			} else {
				stateChannel <- installationStatus
			}
		default:
			fmt.Println("Waiting for watcher events")
			time.Sleep(2 * time.Second)
		}
	}
}

func (k KymaInstaller) newInstallationWatcher(timeout int64) (watch.Interface, error) {
	return k.installationClient.Watch(metav1.ListOptions{FieldSelector: fmt.Sprintf("%s=%s", "metadata.name", kymaInstallationName), TimeoutSeconds: &timeout})
}

var installationObjectDeleted error = fmt.Errorf("installation object deleted")

func handleInstallationEvent(event watch.Event) (InstallationState, error) {
	switch event.Type {
	case watch.Modified:
		installation, ok := event.Object.(*v1alpha1.Installation)
		if !ok {
			return InstallationState{}, fmt.Errorf("installation watcher returned invalid type %T, expected Installation", event.Object)
		}

		switch installation.Status.State {
		case v1alpha1.StateInstalled:
			fmt.Println("INSTALLED")
			return InstallationState{
				State:       string(v1alpha1.StateInstalled),
				Description: installation.Status.Description,
			}, nil
		case v1alpha1.StateError:
			fmt.Println("ERROR")
			return InstallationState{}, fmt.Errorf("installation error occured, current errors: %s", "") // TODO
		case v1alpha1.StateInProgress:
			fmt.Println("IN PROGRESS")
			return InstallationState{
				State:       string(v1alpha1.StateInProgress),
				Description: installation.Status.Description,
			}, nil
		default:
			fmt.Println("INVALID")
			return InstallationState{}, fmt.Errorf("invalid installation state: %s", installation.Status.State)
		}
	case watch.Error:
		//err, ok := event.Object.(*metav1.Status) // TODO - consider extracting error info
		return InstallationState{}, fmt.Errorf("installation watch error occured")
	case watch.Deleted:
		return InstallationState{}, installationObjectDeleted
	default:
		time.Sleep(2 * time.Second)
		return InstallationState{}, fmt.Errorf("received watch event of unexpected type: %s", event.Type)
	}
}

func (k KymaInstaller) infof(format string, a ...interface{}) {
	if k.logger != nil {
		k.logger.Infof(format, a...)
	}
}

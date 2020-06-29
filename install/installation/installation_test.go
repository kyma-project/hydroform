package installation

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/kyma-incubator/hydroform/install/k8s"

	"k8s.io/apimachinery/pkg/runtime"

	v12 "k8s.io/api/core/v1"

	v1alpha12 "github.com/kyma-project/kyma/components/kyma-operator/pkg/client/clientset/versioned/typed/installer/v1alpha1"

	"k8s.io/apimachinery/pkg/api/meta"

	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/stretchr/testify/assert"

	"k8s.io/client-go/kubernetes"

	"github.com/kyma-project/kyma/components/kyma-operator/pkg/apis/installer/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"

	installationClientset "github.com/kyma-project/kyma/components/kyma-operator/pkg/client/clientset/versioned"
	installationFake "github.com/kyma-project/kyma/components/kyma-operator/pkg/client/clientset/versioned/fake"
	"github.com/stretchr/testify/require"
	dynamicFake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/fake"
)

var (
	deploymentGVR           = schema.GroupVersionResource{Group: "extensions", Version: "v1beta1", Resource: "deployments"}
	clusterRoleGVR          = schema.GroupVersionResource{Group: "rbac.authorization.k8s.io", Version: "v1beta1", Resource: "clusterroles"}
	clusterRoleV1GVR        = schema.GroupVersionResource{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterroles"}
	clusterRoleBindingGVR   = schema.GroupVersionResource{Group: "rbac.authorization.k8s.io", Version: "v1beta1", Resource: "clusterrolebindings"}
	clusterRoleBindingV1GVR = schema.GroupVersionResource{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterrolebindings"}
	serviceAccountGVR       = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "serviceaccounts"}
	serviceGVR              = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "services"}
	namespaceGVR            = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "namespaces"}
	limitRangeGVR           = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "limitranges"}
	roleBindingGVR          = schema.GroupVersionResource{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "rolebindings"}
	roleGVR                 = schema.GroupVersionResource{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "roles"}
	jobGVR                  = schema.GroupVersionResource{Group: "batch", Version: "v1", Resource: "jobs"}
	crdGVR                  = schema.GroupVersionResource{Group: "apiextensions.k8s.io", Version: "v1beta1", Resource: "customresourcedefinitions"}
)

func TestKymaInstaller_PrepareInstallation(t *testing.T) {

	runningTillerPod := &v12.Pod{
		ObjectMeta: v1.ObjectMeta{Name: "tiller-pod", Namespace: kubeSystemNamespace, Labels: map[string]string{"name": "tiller"}},
		Status:     v12.PodStatus{Phase: v12.PodRunning},
	}

	t.Run("should prepare Kyma Installation", func(t *testing.T) {
		// given
		dynamicClient := dynamicFake.NewSimpleDynamicClient(resourcesSchema)
		k8sClientSet := fake.NewSimpleClientset(runningTillerPod)
		installationClientSet := installationFake.NewSimpleClientset()

		mapper := dummyRestMapper{}

		kymaInstaller := newKymaInstaller(mapper, dynamicClient, k8sClientSet, installationClientSet)

		installationComponents := []v1alpha1.KymaComponent{
			{Name: "application-connector", ReleaseName: "application-connector", Namespace: "kyma-integration"},
		}

		kymaInstaller.installationCRModificationFunc = func(installation *v1alpha1.Installation) {
			installation.Spec.Components = installationComponents
		}

		configuration := Configuration{
			Configuration: []ConfigEntry{
				{
					Key:   "global.test.key",
					Value: "global-value",
				},
				{
					Key:    "global.test.secret.key",
					Value:  "global-secret-value",
					Secret: true,
				},
			},
			ComponentConfiguration: []ComponentConfiguration{
				{
					Component: "application-connector",
					Configuration: []ConfigEntry{
						{
							Key:   "component.test.key",
							Value: "component-value",
						},
						{
							Key:    "component.test.secret.key",
							Value:  "component-secret-value",
							Secret: true,
						},
					},
				},
			},
		}

		installation := Installation{
			TillerYaml:      tillerYamlContent,
			InstallerYaml:   installerYamlContent,
			InstallerCRYaml: installerCRYamlContent,
			Configuration:   configuration,
		}

		// when
		err := kymaInstaller.PrepareInstallation(installation)

		// then
		require.NoError(t, err)

		assertInstallation(t, installationClientSet, installationComponents)
		assertConfiguration(t, k8sClientSet, configuration.Configuration, "global", "")

		for _, componentConfig := range configuration.ComponentConfiguration {
			assertConfiguration(t, k8sClientSet, componentConfig.Configuration, componentConfig.Component, componentConfig.Component)
		}

		assertInstallerResources(t, dynamicClient)
	})

	t.Run("should prepare Kyma Installation with Tiller artifacts passed", func(t *testing.T) {
		// given
		dynamicClient := dynamicFake.NewSimpleDynamicClient(resourcesSchema)
		k8sClientSet := fake.NewSimpleClientset(runningTillerPod)
		installationClientSet := installationFake.NewSimpleClientset()

		mapper := dummyRestMapper{}

		kymaInstaller := newKymaInstaller(mapper, dynamicClient, k8sClientSet, installationClientSet)

		installationComponents := []v1alpha1.KymaComponent{
			{Name: "application-connector", ReleaseName: "application-connector", Namespace: "kyma-integration"},
		}

		kymaInstaller.installationCRModificationFunc = func(installation *v1alpha1.Installation) {
			installation.Spec.Components = installationComponents
		}

		configuration := Configuration{
			Configuration: []ConfigEntry{
				{
					Key:   "global.test.key",
					Value: "global-value",
				},
				{
					Key:    "global.test.secret.key",
					Value:  "global-secret-value",
					Secret: true,
				},
			},
			ComponentConfiguration: []ComponentConfiguration{
				{
					Component: "application-connector",
					Configuration: []ConfigEntry{
						{
							Key:   "component.test.key",
							Value: "component-value",
						},
						{
							Key:    "component.test.secret.key",
							Value:  "component-secret-value",
							Secret: true,
						},
					},
				},
			},
		}

		installation := Installation{
			TillerYaml:    tillerYamlContent,
			InstallerYaml: installerYamlContent,
			Configuration: configuration,
		}

		// when
		err := kymaInstaller.PrepareInstallation(installation)

		// then
		require.NoError(t, err)

		assertInstallation(t, installationClientSet, installationComponents)
		assertConfiguration(t, k8sClientSet, configuration.Configuration, "global", "")

		for _, componentConfig := range configuration.ComponentConfiguration {
			assertConfiguration(t, k8sClientSet, componentConfig.Configuration, componentConfig.Component, componentConfig.Component)
		}

		assertTillerResources(t, dynamicClient)
		assertInstallerResources(t, dynamicClient)
	})

	t.Run("should return error", func(t *testing.T) {

		for _, testCase := range []struct {
			description          string
			dynamicClientObjects []runtime.Object
			k8sClientsetObjects  []runtime.Object
			installation         Installation
			errorContains        string
		}{
			{
				description: "when invalid Tiller yaml content",
				dynamicClientObjects: []runtime.Object{&v12.ServiceAccount{
					ObjectMeta: v1.ObjectMeta{
						Name:      "tiller",
						Namespace: kubeSystemNamespace,
					},
				}},
				installation:  Installation{TillerYaml: "invalid ", InstallerYaml: installerYamlContent, InstallerCRYaml: installerCRYamlContent, Configuration: Configuration{}},
				errorContains: "failed to parse Tiller yaml",
			},
			{
				description: "when one of Tiller resources already exists",
				dynamicClientObjects: []runtime.Object{&v12.ServiceAccount{
					ObjectMeta: v1.ObjectMeta{
						Name:      "tiller",
						Namespace: kubeSystemNamespace,
					},
				}},
				installation:  Installation{TillerYaml: tillerYamlContent, InstallerYaml: installerYamlContent, InstallerCRYaml: installerCRYamlContent, Configuration: Configuration{}},
				errorContains: "failed to apply Tiller resources",
			},
			{
				description:   "when Tiller pod is not running",
				installation:  Installation{TillerYaml: tillerYamlContent, InstallerYaml: installerYamlContent, InstallerCRYaml: installerCRYamlContent, Configuration: Configuration{}},
				errorContains: "timeout waiting for Tiller to start",
			},
			{
				description:         "when invalid Installer yaml content",
				k8sClientsetObjects: []runtime.Object{runningTillerPod},
				installation:        Installation{TillerYaml: tillerYamlContent, InstallerYaml: "invalid yaml", InstallerCRYaml: installerCRYamlContent, Configuration: Configuration{}},
				errorContains:       "failed to parse yaml",
			},
			{
				description:         "when Installation CR not present in installer YAML",
				k8sClientsetObjects: []runtime.Object{runningTillerPod},
				installation:        Installation{TillerYaml: tillerYamlContent, InstallerYaml: installerYamlContent, InstallerCRYaml: "invalid.yaml", Configuration: Configuration{}},
				errorContains:       "failed to parse yaml",
			},
			{
				description: "when one of Installer resources already exists",
				dynamicClientObjects: []runtime.Object{&v12.ServiceAccount{
					ObjectMeta: v1.ObjectMeta{
						Name:      "helm-certs-job-sa",
						Namespace: kymaInstallerNamespace,
					},
				}},
				k8sClientsetObjects: []runtime.Object{runningTillerPod},
				installation:        Installation{TillerYaml: tillerYamlContent, InstallerYaml: installerYamlContent, Configuration: Configuration{}},
				errorContains:       "failed to apply Installer resources",
			},
		} {
			t.Run(testCase.description, func(t *testing.T) {
				// given
				dynamicClient := dynamicFake.NewSimpleDynamicClient(resourcesSchema, testCase.dynamicClientObjects...)
				k8sClientSet := fake.NewSimpleClientset(testCase.k8sClientsetObjects...)
				installationClientSet := installationFake.NewSimpleClientset()

				mapper := dummyRestMapper{}

				kymaInstaller := newKymaInstaller(mapper, dynamicClient, k8sClientSet, installationClientSet)

				// when
				err := kymaInstaller.PrepareInstallation(testCase.installation)

				// then
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.errorContains)
			})
		}

	})

	t.Run("should update installation CR if already exists", func(t *testing.T) {
		// given
		dynamicClient := dynamicFake.NewSimpleDynamicClient(resourcesSchema)
		k8sClientSet := fake.NewSimpleClientset(runningTillerPod)
		installationClientSet := installationFake.NewSimpleClientset(&v1alpha1.Installation{
			ObjectMeta: v1.ObjectMeta{
				Name:      kymaInstallationName,
				Namespace: defaultInstallationResourceNamespace,
			},
		})

		mapper := dummyRestMapper{}

		kymaInstaller := newKymaInstaller(mapper, dynamicClient, k8sClientSet, installationClientSet)

		installation := Installation{
			TillerYaml:    tillerYamlContent,
			InstallerYaml: installerYamlContent,
			Configuration: Configuration{},
		}

		// when
		err := kymaInstaller.PrepareInstallation(installation)

		// then
		require.NoError(t, err)
	})
}

func TestKymaInstaller_PrepareUpgrade(t *testing.T) {
	runningTillerPod := &v12.Pod{
		ObjectMeta: v1.ObjectMeta{Name: "tiller-pod", Namespace: kubeSystemNamespace, Labels: map[string]string{"name": "tiller"}},
		Status:     v12.PodStatus{Phase: v12.PodRunning},
	}

	t.Run("should prepare upgrade", func(t *testing.T) {
		// given
		dynamicClient := dynamicFake.NewSimpleDynamicClient(resourcesSchema)
		k8sClientSet := fake.NewSimpleClientset(runningTillerPod)
		installationClientSet := installationFake.NewSimpleClientset()

		mapper := dummyRestMapper{}

		kymaInstaller := newKymaInstaller(mapper, dynamicClient, k8sClientSet, installationClientSet)

		installationComponents := []v1alpha1.KymaComponent{
			{Name: "application-connector", ReleaseName: "application-connector", Namespace: "kyma-integration"},
		}

		kymaInstaller.installationCRModificationFunc = func(installation *v1alpha1.Installation) {
			installation.Spec.Components = installationComponents
		}

		configuration := Configuration{
			Configuration: []ConfigEntry{
				{
					Key:   "global.test.key",
					Value: "global-value",
				},
				{
					Key:    "global.test.secret.key",
					Value:  "global-secret-value",
					Secret: true,
				},
			},
			ComponentConfiguration: []ComponentConfiguration{
				{
					Component: "application-connector",
					Configuration: []ConfigEntry{
						{
							Key:   "component.test.key",
							Value: "component-value",
						},
						{
							Key:    "component.test.secret.key",
							Value:  "component-secret-value",
							Secret: true,
						},
					},
				},
			},
		}

		installation := Installation{
			InstallerYaml: installerYamlContent,
			Configuration: configuration,
		}

		// when
		err := kymaInstaller.PrepareInstallation(installation)

		// then
		require.NoError(t, err)

		//given
		upgrade := Installation{
			InstallerYaml: upgradeInstallerYamlContent,
			Configuration: configuration,
		}

		//when
		err = kymaInstaller.PrepareUpgrade(upgrade)

		//then
		require.NoError(t, err)

		assertInstallerResources(t, dynamicClient)

		assertDynamicResource(t, dynamicClient, serviceGVR, "kyma-upgrade-check", kymaInstallerNamespace)
	})

	t.Run("should prepare upgrade with Tiller", func(t *testing.T) {
		// given
		dynamicClient := dynamicFake.NewSimpleDynamicClient(resourcesSchema)
		k8sClientSet := fake.NewSimpleClientset(runningTillerPod)
		installationClientSet := installationFake.NewSimpleClientset()

		mapper := dummyRestMapper{}

		kymaInstaller := newKymaInstaller(mapper, dynamicClient, k8sClientSet, installationClientSet)

		installationComponents := []v1alpha1.KymaComponent{
			{Name: "application-connector", ReleaseName: "application-connector", Namespace: "kyma-integration"},
		}

		kymaInstaller.installationCRModificationFunc = func(installation *v1alpha1.Installation) {
			installation.Spec.Components = installationComponents
		}

		configuration := Configuration{
			Configuration: []ConfigEntry{
				{
					Key:   "global.test.key",
					Value: "global-value",
				},
				{
					Key:    "global.test.secret.key",
					Value:  "global-secret-value",
					Secret: true,
				},
			},
			ComponentConfiguration: []ComponentConfiguration{
				{
					Component: "application-connector",
					Configuration: []ConfigEntry{
						{
							Key:   "component.test.key",
							Value: "component-value",
						},
						{
							Key:    "component.test.secret.key",
							Value:  "component-secret-value",
							Secret: true,
						},
					},
				},
			},
		}

		installation := Installation{
			TillerYaml:    tillerYamlContent,
			InstallerYaml: installerYamlContent,
			Configuration: configuration,
		}

		// when
		err := kymaInstaller.PrepareInstallation(installation)

		// then
		require.NoError(t, err)

		//given
		upgrade := Installation{
			TillerYaml:      upgradeTillerYamlContent,
			InstallerYaml:   upgradeInstallerYamlContent,
			InstallerCRYaml: upgradeInstallerCRYamlContent,
			Configuration:   configuration,
		}

		//when
		err = kymaInstaller.PrepareUpgrade(upgrade)

		//then
		require.NoError(t, err)

		assertTillerResources(t, dynamicClient)
		assertInstallerResources(t, dynamicClient)

		assertDynamicResource(t, dynamicClient, serviceGVR, "tiller-upgrade-check", kubeSystemNamespace)
		assertDynamicResource(t, dynamicClient, serviceGVR, "kyma-upgrade-check", kymaInstallerNamespace)

	})

	t.Run("should return error", func(t *testing.T) {

		for _, testCase := range []struct {
			description          string
			dynamicClientObjects []runtime.Object
			k8sClientsetObjects  []runtime.Object
			installation         Installation
			errorContains        string
		}{
			{
				description: "when invalid Tiller yaml content",
				dynamicClientObjects: []runtime.Object{&v12.ServiceAccount{
					ObjectMeta: v1.ObjectMeta{
						Name:      "tiller",
						Namespace: kubeSystemNamespace,
					},
				}},
				installation:  Installation{TillerYaml: "invalid ", InstallerYaml: installerYamlContent, InstallerCRYaml: upgradeInstallerCRYamlContent, Configuration: Configuration{}},
				errorContains: "failed to parse Tiller yaml",
			},
			{
				description:   "when Tiller pod is not running",
				installation:  Installation{TillerYaml: tillerYamlContent, InstallerYaml: installerYamlContent, InstallerCRYaml: upgradeInstallerCRYamlContent, Configuration: Configuration{}},
				errorContains: "timeout waiting for Tiller to start",
			},
			{
				description:         "when invalid Installer yaml content",
				k8sClientsetObjects: []runtime.Object{runningTillerPod},
				installation:        Installation{TillerYaml: tillerYamlContent, InstallerYaml: "invalid yaml", InstallerCRYaml: upgradeInstallerCRYamlContent, Configuration: Configuration{}},
				errorContains:       "failed to parse yaml",
			},
			{
				description:         "when Installation CR not present in installer YAML",
				k8sClientsetObjects: []runtime.Object{runningTillerPod},
				installation:        Installation{TillerYaml: tillerYamlContent, InstallerYaml: installerYamlContent, InstallerCRYaml: "invalid yaml", Configuration: Configuration{}},
				errorContains:       "failed to parse yaml",
			},
		} {
			t.Run(testCase.description, func(t *testing.T) {
				// given
				dynamicClient := dynamicFake.NewSimpleDynamicClient(resourcesSchema, testCase.dynamicClientObjects...)
				k8sClientSet := fake.NewSimpleClientset(testCase.k8sClientsetObjects...)
				installationClientSet := installationFake.NewSimpleClientset()

				mapper := dummyRestMapper{}

				kymaInstaller := newKymaInstaller(mapper, dynamicClient, k8sClientSet, installationClientSet)

				// when
				err := kymaInstaller.PrepareUpgrade(testCase.installation)

				// then
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.errorContains)
			})
		}

	})
}

func TestKymaInstaller_StartInstallation(t *testing.T) {

	installation := &v1alpha1.Installation{
		TypeMeta: v1.TypeMeta{},
		ObjectMeta: v1.ObjectMeta{
			Name:      kymaInstallationName,
			Namespace: defaultInstallationResourceNamespace,
		},
	}

	t.Run("should report installation status", func(t *testing.T) {
		// given
		k8sClientSet := fake.NewSimpleClientset()
		installationClientSet := installationFake.NewSimpleClientset(installation)
		installationClient := installationClientSet.InstallerV1alpha1().Installations(defaultInstallationResourceNamespace)

		kymaInstaller := newKymaInstaller(nil, nil, k8sClientSet, installationClientSet)

		expectedStates := []InstallationState{
			{State: string(v1alpha1.StateEmpty), Description: ""},
			{State: string(v1alpha1.StateInProgress), Description: "In progress"},
			{State: string(v1alpha1.StateInProgress), Description: "Still in progress"},
			{State: string(v1alpha1.StateInstalled), Description: "Kyma installed"},
		}

		// when
		installationStateChan, errorChan, err := kymaInstaller.StartInstallation(context.Background())
		require.NoError(t, err)

		updateErrChan := make(chan error)

		go updateInstallationPeriodically(updateErrChan, installationClient,
			updateInstallationStatusFunc(&v1alpha1.InstallationStatus{State: v1alpha1.StateEmpty, Description: ""}),
			updateInstallationStatusFunc(&v1alpha1.InstallationStatus{State: v1alpha1.StateInProgress, Description: "In progress"}),
			updateInstallationStatusFunc(&v1alpha1.InstallationStatus{State: v1alpha1.StateInProgress, Description: "Still in progress"}),
			updateInstallationStatusFunc(&v1alpha1.InstallationStatus{State: v1alpha1.StateInstalled, Description: "Kyma installed"}))

		iter := 0
		finished := false
		// then
		for {
			select {
			case state := <-installationStateChan:
				assert.Equal(t, expectedStates[iter], state)
				iter++
				if iter == len(expectedStates) {
					finished = true
					break
				}
			case err := <-errorChan:
				t.Fatalf("Received error: %s", err.Error())
			case updateErr := <-updateErrChan:
				t.Fatalf("Received update error: %s", updateErr.Error())
			}

			if finished == true {
				break
			}
		}

		_, opened := <-installationStateChan
		assert.False(t, opened)
		_, opened = <-errorChan
		assert.False(t, opened)
	})

	t.Run("should send to error channel if installation error occurs", func(t *testing.T) {
		// given
		k8sClientSet := fake.NewSimpleClientset()
		installationClientSet := installationFake.NewSimpleClientset(installation)
		installationClient := installationClientSet.InstallerV1alpha1().Installations(defaultInstallationResourceNamespace)

		kymaInstaller := newKymaInstaller(nil, nil, k8sClientSet, installationClientSet)

		expectedStates := []InstallationState{
			{State: string(v1alpha1.StateInProgress), Description: "In progress"},
			{State: string(v1alpha1.StateError), Description: "Installation Error"},
		}

		errorLog := []v1alpha1.ErrorLogEntry{{
			Component:   "Istio",
			Log:         "error",
			Occurrences: 1,
		}}

		// when
		installationStateChan, errorChan, err := kymaInstaller.StartInstallation(context.Background())
		require.NoError(t, err)

		updateErrChan := make(chan error)

		go updateInstallationPeriodically(updateErrChan, installationClient,
			updateInstallationStatusFunc(&v1alpha1.InstallationStatus{State: v1alpha1.StateInProgress, Description: "In progress"}),
			updateInstallationStatusFunc(&v1alpha1.InstallationStatus{State: v1alpha1.StateError, Description: "Installation Error", ErrorLog: errorLog}))

		finished := false
		// then
		for {
			select {
			case state := <-installationStateChan:
				assert.Equal(t, expectedStates[0], state)
			case err := <-errorChan:
				assert.Error(t, err)
				var installationError InstallationError
				ok := errors.As(err, &installationError)
				require.True(t, ok)
				assert.Equal(t, 1, len(installationError.ErrorEntries))
				finished = true
			case updateErr := <-updateErrChan:
				t.Fatalf("Received update error: %s", updateErr.Error())
			}

			if finished == true {
				break
			}
		}
	})

	t.Run("should close channels if Installation CR deleted", func(t *testing.T) {
		// given
		k8sClientSet := fake.NewSimpleClientset()
		installationClientSet := installationFake.NewSimpleClientset(installation)
		installationClient := installationClientSet.InstallerV1alpha1().Installations(defaultInstallationResourceNamespace)

		kymaInstaller := newKymaInstaller(nil, nil, k8sClientSet, installationClientSet)

		// when
		installationStateChan, errorChan, err := kymaInstaller.StartInstallation(context.Background())
		require.NoError(t, err)

		go func() {
			time.Sleep(3 * time.Second)
			err := installationClient.Delete(kymaInstallationName, &v1.DeleteOptions{})
			assert.NoError(t, err)
		}()

		finished := false
		// then
		for {
			select {
			case _ = <-installationStateChan:
				t.Fatalf("unexpected message in state channel")
			case err := <-errorChan:
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "deleted")
				finished = true
			}

			if finished == true {
				break
			}
		}

		_, opened := <-installationStateChan
		assert.False(t, opened)
		_, opened = <-errorChan
		assert.False(t, opened)
	})

	t.Run("should stop if context canceled", func(t *testing.T) {
		// given
		k8sClientSet := fake.NewSimpleClientset()
		installationClientSet := installationFake.NewSimpleClientset(installation)

		kymaInstaller := newKymaInstaller(nil, nil, k8sClientSet, installationClientSet)
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		// when
		installationStateChan, errorChan, err := kymaInstaller.StartInstallation(ctx)
		require.NoError(t, err)

		finished := false
		// then
		for {
			select {
			case state := <-installationStateChan:
				assert.Equal(t, "InProgress", state.State)
			case err := <-errorChan:
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "context canceled")
				finished = true
			}

			if finished == true {
				break
			}
		}
	})

	t.Run("should return error if installation already in progress", func(t *testing.T) {
		// given
		installation := &v1alpha1.Installation{
			TypeMeta: v1.TypeMeta{},
			ObjectMeta: v1.ObjectMeta{
				Name:      kymaInstallationName,
				Namespace: defaultInstallationResourceNamespace,
			},
			Status: v1alpha1.InstallationStatus{State: v1alpha1.StateInProgress},
		}

		k8sClientSet := fake.NewSimpleClientset()
		installationClientSet := installationFake.NewSimpleClientset(installation)

		kymaInstaller := newKymaInstaller(nil, nil, k8sClientSet, installationClientSet)

		// when
		_, _, err := kymaInstaller.StartInstallation(context.Background())

		// then
		require.Error(t, err)
	})

	t.Run("should return error when installation does not exist", func(t *testing.T) {
		// given
		k8sClientSet := fake.NewSimpleClientset()
		installationClientSet := installationFake.NewSimpleClientset()

		kymaInstaller := newKymaInstaller(nil, nil, k8sClientSet, installationClientSet)

		// when
		_, _, err := kymaInstaller.StartInstallation(context.Background())

		// then
		require.Error(t, err)
	})

	t.Run("should return error when context already canceled", func(t *testing.T) {
		// given
		k8sClientSet := fake.NewSimpleClientset()
		installationClientSet := installationFake.NewSimpleClientset()

		kymaInstaller := newKymaInstaller(nil, nil, k8sClientSet, installationClientSet)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		// when
		_, _, err := kymaInstaller.StartInstallation(ctx)

		// then
		require.Error(t, err)
	})

}

type installationStatusUpdateFunc func(installationStatus *v1alpha1.InstallationStatus)

func updateInstallationStatusFunc(status *v1alpha1.InstallationStatus) installationStatusUpdateFunc {
	return func(installationStatus *v1alpha1.InstallationStatus) {
		status.DeepCopyInto(installationStatus)
	}
}

func updateInstallationPeriodically(errChan chan<- error, installationClient v1alpha12.InstallationInterface, updateFunctions ...installationStatusUpdateFunc) {
	for _, updateFunc := range updateFunctions {
		time.Sleep(time.Second * 2)
		installation, err := installationClient.Get(kymaInstallationName, v1.GetOptions{})
		if err != nil {
			errChan <- err
			return
		}

		actionLabel, ok := installation.Labels[installationActionLabel]
		if !ok {
			errChan <- fmt.Errorf("action label not found on Installation CR")
			return
		}
		if actionLabel != "install" {
			errChan <- fmt.Errorf("unexpected value of action label, expected: %s, actual: %s", "install", actionLabel)
			return
		}

		updateFunc(&installation.Status)

		_, err = installationClient.Update(installation)
		if err != nil {
			errChan <- err
			return
		}
	}
}

func assertInstallation(t *testing.T, installationClientSet *installationFake.Clientset, components []v1alpha1.KymaComponent) {
	kymaInstallation, err := installationClientSet.InstallerV1alpha1().Installations(defaultInstallationResourceNamespace).Get(kymaInstallationName, v1.GetOptions{})
	require.NoError(t, err)
	assert.Equal(t, components, kymaInstallation.Spec.Components)
}

func assertConfiguration(t *testing.T, clientSet *fake.Clientset, configuration []ConfigEntry, namePrefix, component string) {
	cmClient := clientSet.CoreV1().ConfigMaps(kymaInstallerNamespace)
	secretsClient := clientSet.CoreV1().Secrets(kymaInstallerNamespace)

	configMap, err := cmClient.Get(namePrefix+"-installer-config", v1.GetOptions{})
	require.NoError(t, err)

	secret, err := secretsClient.Get(namePrefix+"-installer-config", v1.GetOptions{})
	require.NoError(t, err)

	if component != "" {
		assert.Equal(t, component, configMap.Labels[ComponentOverridesLabelKey])
		assert.Equal(t, component, secret.Labels[ComponentOverridesLabelKey])
	}

	for _, config := range configuration {
		if config.Secret {
			assert.Equal(t, config.Value, string(secret.Data[config.Key]))
		} else {
			assert.Equal(t, config.Value, configMap.Data[config.Key])
		}
	}
}

func assertTillerResources(t *testing.T, dynamicClient *dynamicFake.FakeDynamicClient) {
	assertDynamicResource(t, dynamicClient, serviceAccountGVR, "tiller", kubeSystemNamespace)
	assertDynamicResource(t, dynamicClient, clusterRoleBindingGVR, "tiller-cluster-admin", "")
	assertDynamicResource(t, dynamicClient, deploymentGVR, "tiller-deploy", kubeSystemNamespace)
	assertDynamicResource(t, dynamicClient, serviceGVR, "tiller-deploy", kubeSystemNamespace)
	assertDynamicResource(t, dynamicClient, serviceAccountGVR, "tiller-certs-sa", kubeSystemNamespace)
	assertDynamicResource(t, dynamicClient, roleBindingGVR, "tiller-certs", kubeSystemNamespace)
	assertDynamicResource(t, dynamicClient, roleGVR, "tiller-certs-installer", kubeSystemNamespace)
	assertDynamicResource(t, dynamicClient, jobGVR, "tiller-certs-job", kubeSystemNamespace)
}

func assertInstallerResources(t *testing.T, dynamicClient *dynamicFake.FakeDynamicClient) {
	assertDynamicResource(t, dynamicClient, namespaceGVR, "kyma-installer", "")
	assertDynamicResource(t, dynamicClient, limitRangeGVR, "kyma-default", kymaInstallerNamespace)
	assertDynamicResource(t, dynamicClient, crdGVR, "installations.installer.kyma-project.io", "")
	assertDynamicResource(t, dynamicClient, crdGVR, "releases.release.kyma-project.io", "")
	assertDynamicResource(t, dynamicClient, serviceAccountGVR, "helm-certs-job-sa", kymaInstallerNamespace)
	assertDynamicResource(t, dynamicClient, roleBindingGVR, "helm-certs-rolebinding", kubeSystemNamespace)
	assertDynamicResource(t, dynamicClient, roleBindingGVR, "helm-certs-rolebinding", kymaInstallerNamespace)
	assertDynamicResource(t, dynamicClient, roleGVR, "helm-certs-getter", kubeSystemNamespace)
	assertDynamicResource(t, dynamicClient, roleGVR, "helm-certs-setter", kymaInstallerNamespace)
	assertDynamicResource(t, dynamicClient, clusterRoleV1GVR, "all-psp", "")
	assertDynamicResource(t, dynamicClient, clusterRoleBindingV1GVR, "all-psp", "")
	assertDynamicResource(t, dynamicClient, jobGVR, "helm-certs-job", kymaInstallerNamespace)
	assertDynamicResource(t, dynamicClient, serviceAccountGVR, "kyma-installer", kymaInstallerNamespace)
	assertDynamicResource(t, dynamicClient, deploymentGVR, "kyma-installer", kymaInstallerNamespace)
	assertDynamicResource(t, dynamicClient, clusterRoleGVR, "kyma-installer-reader", "")
	assertDynamicResource(t, dynamicClient, clusterRoleBindingGVR, "kyma-installer", "")
}

func assertDynamicResource(t *testing.T, dynamicClient *dynamicFake.FakeDynamicClient, gvr schema.GroupVersionResource, name, namespace string) {
	var client dynamic.ResourceInterface

	if namespace != "" {
		client = dynamicClient.Resource(gvr).Namespace(namespace)
	} else {
		client = dynamicClient.Resource(gvr)
	}

	object, err := client.Get(name, v1.GetOptions{})
	require.NoError(t, err)

	assert.Equal(t, name, object.GetName())
}

func newKymaInstaller(mapper k8s.RESTMapper, dynamicInterface dynamic.Interface, k8sClientSet kubernetes.Interface, installationClientSet installationClientset.Interface) KymaInstaller {
	kymaInstaller := KymaInstaller{
		installationWatcherTimeoutSeconds: defaultWatcherTimeoutSeconds,
		installationOptions: &installationOptions{
			installationCRModificationFunc: func(installation *v1alpha1.Installation) {},
		},
		decoder:            decoder,
		k8sGenericClient:   k8s.NewGenericClient(mapper, dynamicInterface, k8sClientSet),
		installationClient: installationClientSet.InstallerV1alpha1().Installations(defaultInstallationResourceNamespace),
	}

	return kymaInstaller
}

type dummyRestMapper struct{}

func (d dummyRestMapper) RESTMapping(gk schema.GroupKind, versions ...string) (*meta.RESTMapping, error) {
	if len(versions) < 1 {
		return nil, fmt.Errorf("no version provided")
	}

	return &meta.RESTMapping{
		Resource: schema.GroupVersionResource{
			Group:    gk.Group,
			Version:  versions[0],
			Resource: kindToResource(gk.Kind),
		},
		GroupVersionKind: schema.GroupVersionKind{
			Group:   gk.Group,
			Version: versions[0],
			Kind:    gk.Kind,
		},
		Scope: nil,
	}, nil
}

func kindToResource(kind string) string {
	return fmt.Sprintf("%ss", strings.ToLower(kind))
}

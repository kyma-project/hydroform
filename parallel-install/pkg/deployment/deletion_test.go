package deployment

import (
	"context"
	"fmt"

	"github.com/avast/retry-go"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/config"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/deployment/mocks"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/engine"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/helm"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/logger"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/overrides"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes"
	k8sfake "k8s.io/client-go/kubernetes/fake"

	"testing"
	"time"
)

func TestDeployment_StartKymaUninstallation(t *testing.T) {

	kubeClient := k8sfake.NewSimpleClientset(&v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "kyma-installer",
			Labels: map[string]string{"istio-injection": "disabled", "kyma-project.io/installation": ""},
		},
	})
	dynamicClient := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(), fixCrdGvrMap())
	manager := &mocks.ResourceManager{}
	manager.On("DeleteCollectionOfResources", mock.AnythingOfType("schema.GroupVersionKind"), mock.AnythingOfType("v1.DeleteOptions"), mock.AnythingOfType("v1.ListOptions")).Return(nil)
	i := newDeletion(t, nil, kubeClient, dynamicClient, manager, nil)

	t.Run("should uninstall Kyma", func(t *testing.T) {
		hc := &mockHelmClient{}
		provider := &mockProvider{
			hc: hc,
		}
		overridesProvider := &mockOverridesProvider{}
		prerequisitesEng := engine.NewEngine(overridesProvider, provider, engine.Config{
			WorkersCount: 1,
			Log:          logger.NewLogger(true),
		})
		componentsEng := engine.NewEngine(overridesProvider, provider, engine.Config{
			WorkersCount: 2,
			Log:          logger.NewLogger(true),
		})

		err := i.startKymaUninstallation(prerequisitesEng, componentsEng)

		assert.NoError(t, err)
	})

	t.Run("should fail to uninstall Kyma components", func(t *testing.T) {
		t.Run("due to cancel timeout", func(t *testing.T) {
			hc := &mockHelmClient{
				componentProcessingTime: 200,
			}
			provider := &mockProvider{
				hc: hc,
			}
			overridesProvider := &mockOverridesProvider{}
			prerequisitesEng := engine.NewEngine(overridesProvider, provider, engine.Config{
				WorkersCount: 1,
				Log:          logger.NewLogger(true),
			})
			componentsEng := engine.NewEngine(overridesProvider, provider, engine.Config{
				WorkersCount: 2,
				Log:          logger.NewLogger(true),
			})

			start := time.Now()
			err := i.startKymaUninstallation(prerequisitesEng, componentsEng)
			end := time.Now()

			elapsed := end.Sub(start)

			assert.Error(t, err)
			assert.EqualError(t, err, "Kyma uninstallation failed due to the timeout")

			t.Logf("Elapsed time: %v", elapsed.Seconds())
			// Cancel timeout occurs at 150 ms
			// Quit timeout occurs at 250 ms
			// Blocking process (single component deployment) takes about 201[ms]
			// Quit condition should be detected before processing next component.
			// Check if program quits as expected after cancel timeout
			assert.GreaterOrEqual(t, elapsed.Milliseconds(), int64(150))
			assert.Less(t, elapsed.Milliseconds(), int64(220))
		})
		t.Run("due to quit timeout", func(t *testing.T) {
			hc := &mockHelmClient{
				componentProcessingTime: 300,
			}
			provider := &mockProvider{
				hc: hc,
			}
			overridesProvider := &mockOverridesProvider{}
			prerequisitesEng := engine.NewEngine(overridesProvider, provider, engine.Config{
				WorkersCount: 1,
				Log:          logger.NewLogger(true),
			})
			componentsEng := engine.NewEngine(overridesProvider, provider, engine.Config{
				WorkersCount: 2,
				Log:          logger.NewLogger(true),
			})

			start := time.Now()
			err := i.startKymaUninstallation(prerequisitesEng, componentsEng)
			end := time.Now()

			elapsed := end.Sub(start)

			assert.Error(t, err)
			assert.EqualError(t, err, "Force quit: Kyma uninstallation failed due to the timeout")

			t.Logf("Elapsed time: %v", elapsed.Seconds())
			// One component deployment lasts 300 ms
			// Quit timeout occurs at 250 ms
			// Check if program ends just after quit timeout
			assert.GreaterOrEqual(t, elapsed.Milliseconds(), int64(250))
			assert.Less(t, elapsed.Milliseconds(), int64(260))
		})
	})

	t.Run("should uninstall components and fail to uninstall Kyma prerequisites", func(t *testing.T) {
		t.Run("due to cancel timeout", func(t *testing.T) {
			hc := &mockHelmClient{
				componentProcessingTime: 40,
			}
			provider := &mockProvider{
				hc: hc,
			}
			overridesProvider := &mockOverridesProvider{}
			prerequisitesEng := engine.NewEngine(overridesProvider, provider, engine.Config{
				WorkersCount: 1,
				Log:          logger.NewLogger(true),
			})
			componentsEng := engine.NewEngine(overridesProvider, provider, engine.Config{
				WorkersCount: 2,
				Log:          logger.NewLogger(true),
			})

			start := time.Now()
			err := i.startKymaUninstallation(prerequisitesEng, componentsEng)
			end := time.Now()

			elapsed := end.Sub(start)

			assert.Error(t, err)
			assert.EqualError(t, err, "Kyma uninstallation failed due to the timeout")

			t.Logf("Elapsed time: %v", elapsed.Seconds())
			// Cancel timeout occurs at 150 ms
			// Quit timeout occurs at 250 ms
			// Blocking process (component deployment) ends in the meantime (it's a multiple of 41[ms])
			// Check if program quits as expected after cancel timeout and before quit timeout
			assert.GreaterOrEqual(t, elapsed.Milliseconds(), int64(150))
			assert.Less(t, elapsed.Milliseconds(), int64(200))
		})
		t.Run("due to quit timeout", func(t *testing.T) {
			kubeClient := k8sfake.NewSimpleClientset()

			inst := newDeletion(t, nil, kubeClient, nil, nil, nil)

			// Changing it to higher amounts to minimize difference between cancel and quit timeout
			// and give program enough time to process
			inst.cfg.CancelTimeout = 240 * time.Millisecond
			inst.cfg.QuitTimeout = 260 * time.Millisecond

			hc := &mockHelmClient{
				componentProcessingTime: 70,
			}
			provider := &mockProvider{
				hc: hc,
			}
			overridesProvider := &mockOverridesProvider{}
			prerequisitesEng := engine.NewEngine(overridesProvider, provider, engine.Config{
				WorkersCount: 1,
				Log:          logger.NewLogger(true),
			})
			componentsEng := engine.NewEngine(overridesProvider, provider, engine.Config{
				WorkersCount: 2,
				Log:          logger.NewLogger(true),
			})

			start := time.Now()
			err := inst.startKymaUninstallation(prerequisitesEng, componentsEng)
			end := time.Now()

			elapsed := end.Sub(start)

			assert.Error(t, err)
			assert.EqualError(t, err, "Force quit: Kyma uninstallation failed due to the timeout")

			t.Logf("Elapsed time: %v", elapsed.Seconds())
			// Prerequisites and two components deployment lasts over 280 ms (multiple of 71[ms], 2 workers uninstalling components in parallel)
			// Quit timeout occurs at 260 ms
			// Check if program ends just after quit timeout
			assert.GreaterOrEqual(t, elapsed.Milliseconds(), int64(260))
			assert.Less(t, elapsed.Milliseconds(), int64(270))
		})
	})
}

func TestPostUninstaller_UninstallCRDs(t *testing.T) {

	scheme := runtime.NewScheme()

	t.Run("should not delete CRDs", func(t *testing.T) {
		t.Run("when no resources of any kind are present on a cluster", func(t *testing.T) {
			// given
			client := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(scheme, fixCrdGvrMap())
			requireNoCrdsOnTheCluster(t, client)
			manager := &mocks.ResourceManager{}
			manager.On("DeleteCollectionOfResources", mock.AnythingOfType("schema.GroupVersionKind"), mock.AnythingOfType("v1.DeleteOptions"), mock.AnythingOfType("v1.ListOptions")).Return(nil)
			deletion := newDeletion(t, nil, nil, client, manager, nil)

			// when
			err := deletion.deleteKymaCrds()

			// then
			require.NoError(t, err, "should not return any error")
			manager.AssertNumberOfCalls(t, "DeleteCollectionOfResources", 2)
			requireNoCrdsOnTheCluster(t, client)
		})

		t.Run("when CRDs not labeled by Kyma are present on a cluster", func(t *testing.T) {
			// given
			client := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(scheme, fixCrdGvrMap())
			crds := createThreeCrdsUsingGivenNamesAndApply(t, client, fixCrdGvrV1Beta1(),
				"apiextensions.k8s.io", "v1beta1", "label", "unknown", "crd1", "crd2", "crd3")
			requireNoKymaCrdsOnTheCluster(t, client)
			manager := &mocks.ResourceManager{}
			manager.On("DeleteCollectionOfResources", mock.AnythingOfType("schema.GroupVersionKind"), mock.AnythingOfType("v1.DeleteOptions"), mock.AnythingOfType("v1.ListOptions")).Return(nil)
			deletion := newDeletion(t, nil, nil, client, manager, nil)

			// when
			err := deletion.deleteKymaCrds()

			// then
			require.NoError(t, err, "should not return any error")
			requireAllObjsExistAndUnchanged(t, client, crds, fixCrdGvrV1Beta1())
			manager.AssertNumberOfCalls(t, "DeleteCollectionOfResources", 2)
		})

		t.Run("when CRDs labeled with incorrect value are present on a cluster", func(t *testing.T) {
			// given
			client := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(scheme, fixCrdGvrMap())
			crds := createThreeCrdsUsingGivenNamesAndApply(t, client, fixCrdGvrV1Beta1(),
				"apiextensions.k8s.io", "v1beta1", "origin", "unknown", "crd1", "crd2", "crd3")
			requireNoKymaCrdsOnTheCluster(t, client)
			manager := &mocks.ResourceManager{}
			manager.On("DeleteCollectionOfResources", mock.AnythingOfType("schema.GroupVersionKind"), mock.AnythingOfType("v1.DeleteOptions"), mock.AnythingOfType("v1.ListOptions")).Return(nil)
			deletion := newDeletion(t, nil, nil, client, manager, nil)

			// when
			err := deletion.deleteKymaCrds()

			// then
			require.NoError(t, err, "should not return any error")
			requireAllObjsExistAndUnchanged(t, client, crds, fixCrdGvrV1Beta1())
			manager.AssertNumberOfCalls(t, "DeleteCollectionOfResources", 2)
		})

		t.Run("when CRDs labeled by Kyma are present on a cluster but CRD api does not match", func(t *testing.T) {
			// given
			client := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(scheme, fixCrdGvrMap())
			crds := createThreeCrdsUsingGivenNamesAndApply(t, client, fixCrdGvrOtherGroup(),
				"otherapi", "v1beta1", "origin", "kyma", "crd1", "crd2", "crd3")
			requireNoGenericCrdsOnTheCluster(t, client)
			manager := &mocks.ResourceManager{}
			manager.On("DeleteCollectionOfResources", mock.AnythingOfType("schema.GroupVersionKind"), mock.AnythingOfType("v1.DeleteOptions"), mock.AnythingOfType("v1.ListOptions")).Return(nil)
			deletion := newDeletion(t, nil, nil, client, manager, nil)

			// when
			err := deletion.deleteKymaCrds()

			// then
			require.NoError(t, err, "should not return any error")
			requireAllObjsExistAndUnchanged(t, client, crds, fixCrdGvrOtherGroup())
			manager.AssertNumberOfCalls(t, "DeleteCollectionOfResources", 2)
		})

		t.Run("when CRDs labeled by Kyma are present on a cluster but CRD version does not match", func(t *testing.T) {
			// given
			client := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(scheme, fixCrdGvrMap())
			crds := createThreeCrdsUsingGivenNamesAndApply(t, client, fixCrdGvrOtherVersion(),
				"apiextensions.k8s.io", "otherversion", "origin", "kyma", "crd1", "crd2", "crd3")
			requireNoGenericCrdsOnTheCluster(t, client)
			manager := &mocks.ResourceManager{}
			manager.On("DeleteCollectionOfResources", mock.AnythingOfType("schema.GroupVersionKind"), mock.AnythingOfType("v1.DeleteOptions"), mock.AnythingOfType("v1.ListOptions")).Return(nil)
			deletion := newDeletion(t, nil, nil, client, manager, nil)

			// when
			err := deletion.deleteKymaCrds()

			// then
			require.NoError(t, err, "should not return any error")
			requireAllObjsExistAndUnchanged(t, client, crds, fixCrdGvrOtherVersion())
			manager.AssertNumberOfCalls(t, "DeleteCollectionOfResources", 2)
		})

		t.Run("when objects of type different than CRD labeled by Kyma are present on a cluster", func(t *testing.T) {
			// given
			client := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(scheme, fixCrdGvrMap())
			namespaces := createThreeNamespacesAndApply(t, client, fixNamespaceGvr(), "origin", "kyma")
			requireNoCrdsOnTheCluster(t, client)
			manager := &mocks.ResourceManager{}
			manager.On("DeleteCollectionOfResources", mock.AnythingOfType("schema.GroupVersionKind"), mock.AnythingOfType("v1.DeleteOptions"), mock.AnythingOfType("v1.ListOptions")).Return(nil)
			deletion := newDeletion(t, nil, nil, client, manager, nil)

			// when
			err := deletion.deleteKymaCrds()

			// then
			require.NoError(t, err, "should not return any error")
			requireAllObjsExistAndUnchanged(t, client, namespaces, fixNamespaceGvr())
			manager.AssertNumberOfCalls(t, "DeleteCollectionOfResources", 2)
		})

		t.Run("when all of them were correct but errors occurred when deleting them", func(t *testing.T) {
			// given
			client := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(scheme, fixCrdGvrMap())
			crdsV1Beta1 := createThreeCrdsUsingGivenNamesAndApply(t, client, fixCrdGvrV1Beta1(),
				"apiextensions.k8s.io", "v1beta1", "origin", "kyma", "crd1", "crd2", "crd3")
			crdsV1 := createThreeCrdsUsingGivenNamesAndApply(t, client, fixCrdGvrV1(),
				"apiextensions.k8s.io", "v1", "origin", "kyma", "crd4", "crd5", "crd6")
			manager := &mocks.ResourceManager{}
			manager.On("DeleteCollectionOfResources", fixCrdGvkV1Beta1(), mock.AnythingOfType("v1.DeleteOptions"), mock.AnythingOfType("v1.ListOptions")).Return(
				func(gvk schema.GroupVersionKind, opts metav1.DeleteOptions, listOps metav1.ListOptions) error {
					return deleteAllMockObjs(t, client, crdsV1Beta1, fixCrdGvrV1Beta1())
				})
			manager.On("DeleteCollectionOfResources", fixCrdGvkV1(), mock.AnythingOfType("v1.DeleteOptions"), mock.AnythingOfType("v1.ListOptions")).Return(
				func(gvk schema.GroupVersionKind, opts metav1.DeleteOptions, listOps metav1.ListOptions) error {
					return deleteAllMockObjs(t, client, crdsV1, fixCrdGvrV1())
				})
			deletion := newDeletion(t, nil, nil, client, manager, nil)

			// when
			err := deletion.deleteKymaCrds()

			// then
			require.NoError(t, err, "should not return any error")
			requireAllObjsNotExist(t, client, crdsV1Beta1, fixCrdGvrV1Beta1())
			requireAllObjsNotExist(t, client, crdsV1, fixCrdGvrV1Beta1())
			manager.AssertNumberOfCalls(t, "DeleteCollectionOfResources", 2)
		})
	})

	t.Run("should delete CRDs", func(t *testing.T) {
		t.Run("when only CRDs labeled by Kyma are present on a cluster", func(t *testing.T) {
			// given
			client := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(scheme, fixCrdGvrMap())
			crdsV1Beta1 := createThreeCrdsUsingGivenNamesAndApply(t, client, fixCrdGvrV1Beta1(),
				"apiextensions.k8s.io", "v1beta1", "origin", "kyma", "crd1", "crd2", "crd3")
			crdsV1 := createThreeCrdsUsingGivenNamesAndApply(t, client, fixCrdGvrV1(),
				"apiextensions.k8s.io", "v1", "origin", "kyma", "crd4", "crd5", "crd6")
			manager := &mocks.ResourceManager{}
			manager.On("DeleteCollectionOfResources", fixCrdGvkV1Beta1(), mock.AnythingOfType("v1.DeleteOptions"), mock.AnythingOfType("v1.ListOptions")).Return(
				func(gvk schema.GroupVersionKind, opts metav1.DeleteOptions, listOps metav1.ListOptions) error {
					return deleteAllMockObjs(t, client, crdsV1Beta1, fixCrdGvrV1Beta1())
				})
			manager.On("DeleteCollectionOfResources", fixCrdGvkV1(), mock.AnythingOfType("v1.DeleteOptions"), mock.AnythingOfType("v1.ListOptions")).Return(
				func(gvk schema.GroupVersionKind, opts metav1.DeleteOptions, listOps metav1.ListOptions) error {
					return deleteAllMockObjs(t, client, crdsV1, fixCrdGvrV1())
				})
			deletion := newDeletion(t, nil, nil, client, manager, nil)

			// when
			err := deletion.deleteKymaCrds()

			// then
			require.NoError(t, err, "should not return any error")
			requireAllObjsNotExist(t, client, crdsV1Beta1, fixCrdGvrV1Beta1())
			requireAllObjsNotExist(t, client, crdsV1, fixCrdGvrV1Beta1())
			manager.AssertNumberOfCalls(t, "DeleteCollectionOfResources", 2)
		})

		t.Run("labeled by Kyma and leave other", func(t *testing.T) {
			// given
			client := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(scheme, fixCrdGvrMap())
			crdsLabeledByKyma := createThreeCrdsUsingGivenNamesAndApply(t, client, fixCrdGvrV1Beta1(),
				"apiextensions.k8s.io", "v1beta1", "origin", "kyma", "crd1", "crd2", "crd3")
			crdsNotLabeledByKyma := createThreeCrdsUsingGivenNamesAndApply(t, client, fixCrdGvrV1(),
				"apiextensions.k8s.io", "v1", "origin", "unknown", "crd4", "crd5", "crd6")
			manager := &mocks.ResourceManager{}
			manager.On("DeleteCollectionOfResources", fixCrdGvkV1Beta1(), mock.AnythingOfType("v1.DeleteOptions"), mock.AnythingOfType("v1.ListOptions")).Return(
				func(gvk schema.GroupVersionKind, opts metav1.DeleteOptions, listOps metav1.ListOptions) error {
					return deleteAllMockObjs(t, client, crdsLabeledByKyma, fixCrdGvrV1Beta1())
				})
			manager.On("DeleteCollectionOfResources", fixCrdGvkV1(), mock.AnythingOfType("v1.DeleteOptions"), mock.AnythingOfType("v1.ListOptions")).Return(nil)
			deletion := newDeletion(t, nil, nil, client, manager, nil)

			// when
			err := deletion.deleteKymaCrds()

			// then
			require.NoError(t, err, "should not return any error")
			requireAllObjsNotExist(t, client, crdsLabeledByKyma, fixCrdGvrV1Beta1())
			requireAllObjsExistAndUnchanged(t, client, crdsNotLabeledByKyma, fixCrdGvrV1())
			manager.AssertNumberOfCalls(t, "DeleteCollectionOfResources", 2)
		})

		t.Run("but only partially when error occurred for first CRD and no error occurred for the rest", func(t *testing.T) {
			// given
			client := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(scheme, fixCrdGvrMap())
			crdsV1Beta1 := createThreeCrdsUsingGivenNamesAndApply(t, client, fixCrdGvrV1Beta1(),
				"apiextensions.k8s.io", "v1beta1", "origin", "kyma", "crd1", "crd2", "crd3")
			crdsV1 := createThreeCrdsUsingGivenNamesAndApply(t, client, fixCrdGvrV1(),
				"apiextensions.k8s.io", "v1", "origin", "kyma", "crd4", "crd5", "crd6")
			manager := &mocks.ResourceManager{}
			manager.On("DeleteCollectionOfResources", fixCrdGvkV1Beta1(), mock.AnythingOfType("v1.DeleteOptions"), mock.AnythingOfType("v1.ListOptions")).Return(
				func(gvk schema.GroupVersionKind, opts metav1.DeleteOptions, listOps metav1.ListOptions) error {
					return deleteAllMockObjs(t, client, crdsV1Beta1, fixCrdGvrV1Beta1())
				})
			manager.On("DeleteCollectionOfResources", fixCrdGvkV1(), mock.AnythingOfType("v1.DeleteOptions"), mock.AnythingOfType("v1.ListOptions")).Return(errors.New("Error"))
			deletion := newDeletion(t, nil, nil, client, manager, nil)

			// when
			err := deletion.deleteKymaCrds()

			// then
			require.NoError(t, err, "should not return any error")
			requireAllObjsNotExist(t, client, crdsV1Beta1, fixCrdGvrV1Beta1())
			requireAllObjsExistAndUnchanged(t, client, crdsV1, fixCrdGvrV1())
			manager.AssertNumberOfCalls(t, "DeleteCollectionOfResources", 2)
		})
	})

}

func TestDeployment_DeleteNamespaces(t *testing.T) {
	kymaLabelPrefix := "kyma-project.io/install."
	kubeClient := k8sfake.NewSimpleClientset(&v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "kyma-test",
			Labels: map[string]string{"kyma-project.io/installation": ""},
		}},
		&v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "sh.helm.release.v1.test1.v1",
				Namespace: "kyma-test",
				Labels: map[string]string{
					kymaLabelPrefix + "name":      "test1",
					kymaLabelPrefix + "namespace": "kyma-test",
					kymaLabelPrefix + "component": "true",
				},
			},
		})
	dynamicClient := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(), fixCrdGvrMap())
	manager := &mocks.ResourceManager{}
	manager.On("DeleteCollectionOfResources", mock.AnythingOfType("schema.GroupVersionKind"), mock.AnythingOfType("v1.DeleteOptions"), mock.AnythingOfType("v1.ListOptions")).Return(nil)
	i := newDeletion(t, nil, kubeClient, dynamicClient, manager, nil)

	t.Run("should uninstall components and Kyma namespaces", func(t *testing.T) {
		t.Run("without errors", func(t *testing.T) {
			hc := &mockHelmClient{}
			provider := &mockProvider{
				hc: hc,
			}
			overridesProvider := &mockOverridesProvider{}
			prerequisitesEng := engine.NewEngine(overridesProvider, provider, engine.Config{
				WorkersCount: 1,
				Log:          logger.NewLogger(true),
			})
			componentsEng := engine.NewEngine(overridesProvider, provider, engine.Config{
				WorkersCount: 2,
				Log:          logger.NewLogger(true),
			})

			err := i.startKymaUninstallation(prerequisitesEng, componentsEng)
			assert.NoError(t, err)

			ns, err := kubeClient.CoreV1().Namespaces().List(nil, metav1.ListOptions{})
			assert.NoError(t, err)
			assert.Equal(t, 0, len(ns.Items))
		})
	})

	kubeClientWithPod := k8sfake.NewSimpleClientset(&v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "kyma-test",
			Labels: map[string]string{"kyma-project.io/installation": ""},
		}},
		&v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "sh.helm.release.v1.test1.v1",
				Namespace: "kyma-test",
				Labels: map[string]string{
					kymaLabelPrefix + "name":      "test1",
					kymaLabelPrefix + "namespace": "kyma-test",
					kymaLabelPrefix + "component": "true",
				},
			},
		},
		&v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pod",
				Namespace: "kyma-test",
			},
			Status: v1.PodStatus{
				Phase: v1.PodRunning,
			},
		})
	retryOpts := []retry.Option{
		retry.Delay(10 * time.Millisecond),
		retry.Attempts(1),
		retry.DelayType(retry.FixedDelay),
	}
	i = newDeletion(t, nil, kubeClientWithPod, dynamicClient, manager, retryOpts)

	t.Run("should uninstall components and fail to uninstall Kyma namespaces", func(t *testing.T) {
		t.Run("due to running Pods", func(t *testing.T) {
			hc := &mockHelmClient{}
			provider := &mockProvider{
				hc: hc,
			}
			overridesProvider := &mockOverridesProvider{}
			prerequisitesEng := engine.NewEngine(overridesProvider, provider, engine.Config{
				WorkersCount: 1,
				Log:          logger.NewLogger(true),
			})
			componentsEng := engine.NewEngine(overridesProvider, provider, engine.Config{
				WorkersCount: 2,
				Log:          logger.NewLogger(true),
			})

			err := i.startKymaUninstallation(prerequisitesEng, componentsEng)
			assert.NoError(t, err)

			ns, err := kubeClientWithPod.CoreV1().Namespaces().List(nil, metav1.ListOptions{})
			assert.NoError(t, err)
			assert.Equal(t, 1, len(ns.Items))
		})
	})
}

// Pass optionally an receiver-channel to get progress updates
func newDeletion(t *testing.T, procUpdates func(ProcessUpdate), kubeClient kubernetes.Interface, dynamicClient dynamic.Interface, manager ResourceManager, retryOptions []retry.Option) *Deletion {
	compList, err := config.NewComponentList("../test/data/componentlist.yaml")
	assert.NoError(t, err)
	cfg := &config.Config{
		CancelTimeout:                 cancelTimeout,
		QuitTimeout:                   quitTimeout,
		BackoffInitialIntervalSeconds: 1,
		BackoffMaxElapsedTimeSeconds:  1,
		Log:                           logger.NewLogger(true),
		ComponentList:                 compList,
	}
	core := newCore(cfg, &overrides.Builder{}, kubeClient, procUpdates)
	metaProv := helm.GetKymaMetadataProvider(kubeClient)
	return &Deletion{core, metaProv, nil, dynamicClient, manager, retryOptions}
}

func createThreeCrdsUsingGivenNamesAndApply(t *testing.T, client *dynamicfake.FakeDynamicClient, gvr schema.GroupVersionResource, api, version, label, value, name1, name2, name3 string) []unstructured.Unstructured {
	crds := createThreeCrdsUsing(api, version, label, value, name1, name2, name3)
	for _, crd := range crds {
		applyMockObj(t, client, &crd, gvr)
	}
	return crds
}

func createThreeCrdsUsing(api, version, label, value, name1, name2, name3 string) []unstructured.Unstructured {
	crd1 := fixCrdResourceWith(name1, api, version, label, value)
	crd2 := fixCrdResourceWith(name2, api, version, label, value)
	crd3 := fixCrdResourceWith(name3, api, version, label, value)
	return []unstructured.Unstructured{*crd1, *crd2, *crd3}
}

func createThreeNamespacesUsing(label, value, name1, name2, name3 string) []unstructured.Unstructured {
	ns1 := fixResourceWithGiven(name1, label, value)
	ns2 := fixResourceWithGiven(name2, label, value)
	ns3 := fixResourceWithGiven(name3, label, value)
	return []unstructured.Unstructured{*ns1, *ns2, *ns3}
}

func createThreeNamespacesAndApply(t *testing.T, client *dynamicfake.FakeDynamicClient, gvr schema.GroupVersionResource, label, value string) []unstructured.Unstructured {
	namespaces := createThreeNamespacesUsing(label, value, "ns1", "ns2", "ns3")
	for _, ns := range namespaces {
		applyMockObj(t, client, &ns, gvr)
	}
	return namespaces
}

func applyMockObj(t *testing.T, client *dynamicfake.FakeDynamicClient, obj *unstructured.Unstructured, gvr schema.GroupVersionResource) {
	resultObj, err := client.Resource(gvr).Create(context.TODO(), obj, metav1.CreateOptions{})
	require.NoError(t, err, "object should be correctly created by fake client")
	require.NotNil(t, resultObj, "object returned by fake client should exist")
	require.Equal(t, obj, resultObj, "object returned by fake client should be equal to the created one")
}

func deleteMockObj(t *testing.T, client *dynamicfake.FakeDynamicClient, obj *unstructured.Unstructured, gvr schema.GroupVersionResource) error {
	err := client.Resource(gvr).Delete(context.TODO(), obj.GetName(), metav1.DeleteOptions{})
	require.NoError(t, err, "object should be correctly deleted by fake client")
	return err
}

func deleteAllMockObjs(t *testing.T, client *dynamicfake.FakeDynamicClient, objs []unstructured.Unstructured, gvr schema.GroupVersionResource) error {
	for _, obj := range objs {
		_ = deleteMockObj(t, client, &obj, gvr)
	}
	return nil
}

func requireNoCrdsOnTheCluster(t *testing.T, client *dynamicfake.FakeDynamicClient) {
	requireNoGenericCrdsOnTheCluster(t, client)
	requireNoKymaCrdsOnTheCluster(t, client)
}

func requireNoGenericCrdsOnTheCluster(t *testing.T, client *dynamicfake.FakeDynamicClient) {
	requireNoGivenCrdsOnTheCluster(t, client, fixCrdGvrV1Beta1(), metav1.ListOptions{})
	requireNoGivenCrdsOnTheCluster(t, client, fixCrdGvrV1(), metav1.ListOptions{})
}

func requireNoKymaCrdsOnTheCluster(t *testing.T, client *dynamicfake.FakeDynamicClient) {
	requireNoGivenCrdsOnTheCluster(t, client, fixCrdGvrV1Beta1(), metav1.ListOptions{LabelSelector: "origin=kyma"})
	requireNoGivenCrdsOnTheCluster(t, client, fixCrdGvrV1(), metav1.ListOptions{LabelSelector: "origin=kyma"})
}

func requireNoGivenCrdsOnTheCluster(t *testing.T, client *dynamicfake.FakeDynamicClient, gvr schema.GroupVersionResource, listOpts metav1.ListOptions) {
	resourcesList, err := client.Resource(gvr).List(context.TODO(), listOpts)
	require.NoError(t, err)
	require.Empty(t, resourcesList.Items)
}

func requireObjExistsAndUnchanged(t *testing.T, client *dynamicfake.FakeDynamicClient, obj *unstructured.Unstructured, gvr schema.GroupVersionResource) {
	resultObj, err := client.Resource(gvr).Get(context.TODO(), obj.GetName(), metav1.GetOptions{})
	require.NoError(t, err, "object should be correctly returned by fake client")
	require.NotNil(t, resultObj, "object returned by fake client should exist")
	require.Equal(t, obj, resultObj, "object returned by fake client should be equal to the created one")
}

func requireAllObjsExistAndUnchanged(t *testing.T, client *dynamicfake.FakeDynamicClient, objs []unstructured.Unstructured, gvr schema.GroupVersionResource) {
	for _, obj := range objs {
		requireObjExistsAndUnchanged(t, client, &obj, gvr)
	}
}

func requireObjNotExists(t *testing.T, client *dynamicfake.FakeDynamicClient, obj *unstructured.Unstructured, gvr schema.GroupVersionResource) {
	resultObj, err := client.Resource(gvr).Get(context.TODO(), obj.GetName(), metav1.GetOptions{})
	require.Error(t, err, "object should be not found")
	require.Nil(t, resultObj, "object returned by fake client should not exist")
}

func requireAllObjsNotExist(t *testing.T, client *dynamicfake.FakeDynamicClient, objs []unstructured.Unstructured, gvr schema.GroupVersionResource) {
	for _, obj := range objs {
		requireObjNotExists(t, client, &obj, gvr)
	}
}

func fixCrdGvrMap() map[schema.GroupVersionResource]string {
	return map[schema.GroupVersionResource]string{
		fixCrdGvrV1():            "CrdList",
		fixCrdGvrV1Beta1():       "CrdList",
		fixCrdGvrOtherGroup():    "CrdList",
		fixCrdGvrOtherVersion():  "CrdList",
		fixCrdGvrOtherResource(): "CrdList",
	}
}

func fixCrdGvrV1() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "apiextensions.k8s.io",
		Version:  "v1",
		Resource: "customresourcedefinitions",
	}
}

func fixCrdGvrV1Beta1() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "apiextensions.k8s.io",
		Version:  "v1beta1",
		Resource: "customresourcedefinitions",
	}
}

func fixCrdGvkV1() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   "apiextensions.k8s.io",
		Version: "v1",
		Kind:    "customresourcedefinition",
	}
}

func fixCrdGvkV1Beta1() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   "apiextensions.k8s.io",
		Version: "v1beta1",
		Kind:    "customresourcedefinition",
	}
}

func fixCrdGvrOtherGroup() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "othergroup",
		Version:  "v1beta1",
		Resource: "customresourcedefinitions",
	}
}

func fixCrdGvrOtherVersion() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "apiextensions.k8s.io",
		Version:  "otherversion",
		Resource: "customresourcedefinitions",
	}
}

func fixCrdGvrOtherResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "apiextensions.k8s.io",
		Version:  "v1beta1",
		Resource: "otherresource",
	}
}

func fixNamespaceGvr() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "Namespace",
	}
}

func fixCrdResourceWith(name, api, version, label, value string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": fmt.Sprintf("%s/%s", api, version),
			"kind":       "CustomResourceDefinition",
			"metadata": map[string]interface{}{
				"name": name,
				"labels": map[string]interface{}{
					label: value,
				},
			},
			"spec": map[string]interface{}{
				"group": "group",
			},
		},
	}
}

func fixResourceWithGiven(name, label, value string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Namespace",
			"metadata": map[string]interface{}{
				"name": name,
				"labels": map[string]interface{}{
					label: value,
				},
			},
		},
	}
}

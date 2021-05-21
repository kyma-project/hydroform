package postuninstaller

import (
	"context"
	"fmt"
	"github.com/avast/retry-go"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/config"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/logger"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/preinstaller"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/preinstaller/mocks"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/fake"
	"testing"
)

func TestPostUninstaller_UninstallCRDs(t *testing.T) {

	scheme := runtime.NewScheme()
	cfg := getTestingConfig()
	retryOptions := getTestingRetryOptions()

	t.Run("should not uninstall CRDs", func(t *testing.T) {
		t.Run("when no resources of any kind are present on a cluster", func(t *testing.T) {
			// given
			client := fake.NewSimpleDynamicClientWithCustomListKinds(scheme, fixCrdGvrMap())
			requireNoCrdsOnTheCluster(t, client)
			resourceManager := &mocks.ResourceManager{}
			resourceManager.On("DeleteCollectionOfResources", mock.AnythingOfType("schema.GroupVersionKind"), mock.AnythingOfType("v1.DeleteOptions"), mock.AnythingOfType("v1.ListOptions")).Return(nil)
			uninstaller := getPostUninstaller(cfg, resourceManager, client, retryOptions)

			// when
			_, err := uninstaller.UninstallCRDs()

			// then
			require.NoError(t, err, "should not return any error")
			resourceManager.AssertNumberOfCalls(t, "DeleteCollectionOfResources", 2)
			requireNoCrdsOnTheCluster(t, client)
		})

		t.Run("when CRDs not labeled by Kyma are present on a cluster", func(t *testing.T) {
			// given
			client := fake.NewSimpleDynamicClientWithCustomListKinds(scheme, fixCrdGvrMap())
			crds := createThreeCrdsUsingGivenNamesAndApply(t, client, fixCrdGvrV1Beta1(),
				"apiextensions.k8s.io", "v1beta1", "label", "unknown", "crd1", "crd2", "crd3")
			requireNoKymaCrdsOnTheCluster(t, client)
			resourceManager := &mocks.ResourceManager{}
			resourceManager.On("DeleteCollectionOfResources", mock.AnythingOfType("schema.GroupVersionKind"), mock.AnythingOfType("v1.DeleteOptions"), mock.AnythingOfType("v1.ListOptions")).Return(nil)
			uninstaller := getPostUninstaller(cfg, resourceManager, client, retryOptions)

			// when
			_, err := uninstaller.UninstallCRDs()

			// then
			require.NoError(t, err, "should not return any error")
			requireAllObjsExistAndUnchanged(t, client, crds, fixCrdGvrV1Beta1())
			resourceManager.AssertNumberOfCalls(t, "DeleteCollectionOfResources", 2)
		})

		t.Run("when CRDs labeled with incorrect value are present on a cluster", func(t *testing.T) {
			// given
			client := fake.NewSimpleDynamicClientWithCustomListKinds(scheme, fixCrdGvrMap())
			crds := createThreeCrdsUsingGivenNamesAndApply(t, client, fixCrdGvrV1Beta1(),
				"apiextensions.k8s.io", "v1beta1", "origin", "unknown", "crd1", "crd2", "crd3")
			requireNoKymaCrdsOnTheCluster(t, client)
			resourceManager := &mocks.ResourceManager{}
			resourceManager.On("DeleteCollectionOfResources", mock.AnythingOfType("schema.GroupVersionKind"), mock.AnythingOfType("v1.DeleteOptions"), mock.AnythingOfType("v1.ListOptions")).Return(nil)
			uninstaller := getPostUninstaller(cfg, resourceManager, client, retryOptions)

			// when
			_, err := uninstaller.UninstallCRDs()

			// then
			require.NoError(t, err, "should not return any error")
			requireAllObjsExistAndUnchanged(t, client, crds, fixCrdGvrV1Beta1())
			resourceManager.AssertNumberOfCalls(t, "DeleteCollectionOfResources", 2)
		})

		t.Run("when CRDs labeled by Kyma are present on a cluster but CRD api does not match", func(t *testing.T) {
			// given
			client := fake.NewSimpleDynamicClientWithCustomListKinds(scheme, fixCrdGvrMap())
			crds := createThreeCrdsUsingGivenNamesAndApply(t, client, fixCrdGvrOtherGroup(),
				"otherapi", "v1beta1", "origin", "kyma", "crd1", "crd2", "crd3")
			requireNoGenericCrdsOnTheCluster(t, client)
			resourceManager := &mocks.ResourceManager{}
			resourceManager.On("DeleteCollectionOfResources", mock.AnythingOfType("schema.GroupVersionKind"), mock.AnythingOfType("v1.DeleteOptions"), mock.AnythingOfType("v1.ListOptions")).Return(nil)
			uninstaller := getPostUninstaller(cfg, resourceManager, client, retryOptions)

			// when
			_, err := uninstaller.UninstallCRDs()

			// then
			require.NoError(t, err, "should not return any error")
			requireAllObjsExistAndUnchanged(t, client, crds, fixCrdGvrOtherGroup())
			resourceManager.AssertNumberOfCalls(t, "DeleteCollectionOfResources", 2)
		})

		t.Run("when CRDs labeled by Kyma are present on a cluster but CRD version does not match", func(t *testing.T) {
			// given
			client := fake.NewSimpleDynamicClientWithCustomListKinds(scheme, fixCrdGvrMap())
			crds := createThreeCrdsUsingGivenNamesAndApply(t, client, fixCrdGvrOtherVersion(),
				"apiextensions.k8s.io", "otherversion", "origin", "kyma", "crd1", "crd2", "crd3")
			requireNoGenericCrdsOnTheCluster(t, client)
			resourceManager := &mocks.ResourceManager{}
			resourceManager.On("DeleteCollectionOfResources", mock.AnythingOfType("schema.GroupVersionKind"), mock.AnythingOfType("v1.DeleteOptions"), mock.AnythingOfType("v1.ListOptions")).Return(nil)
			uninstaller := getPostUninstaller(cfg, resourceManager, client, retryOptions)

			// when
			_, err := uninstaller.UninstallCRDs()

			// then
			require.NoError(t, err, "should not return any error")
			requireAllObjsExistAndUnchanged(t, client, crds, fixCrdGvrOtherVersion())
			resourceManager.AssertNumberOfCalls(t, "DeleteCollectionOfResources", 2)
		})

		t.Run("when objects of type different than CRD labeled by Kyma are present on a cluster", func(t *testing.T) {
			// given
			client := fake.NewSimpleDynamicClientWithCustomListKinds(scheme, fixCrdGvrMap())
			namespaces := createThreeNamespacesAndApply(t, client, fixNamespaceGvr(), "origin", "kyma")
			requireNoCrdsOnTheCluster(t, client)
			resourceManager := &mocks.ResourceManager{}
			resourceManager.On("DeleteCollectionOfResources", mock.AnythingOfType("schema.GroupVersionKind"), mock.AnythingOfType("v1.DeleteOptions"), mock.AnythingOfType("v1.ListOptions")).Return(nil)
			uninstaller := getPostUninstaller(cfg, resourceManager, client, retryOptions)

			// when
			_, err := uninstaller.UninstallCRDs()

			// then
			require.NoError(t, err, "should not return any error")
			requireAllObjsExistAndUnchanged(t, client, namespaces, fixNamespaceGvr())
			resourceManager.AssertNumberOfCalls(t, "DeleteCollectionOfResources", 2)
		})

		t.Run("when all of them were correct but errors occurred when deleting them", func(t *testing.T) {
			// given
			client := fake.NewSimpleDynamicClientWithCustomListKinds(scheme, fixCrdGvrMap())
			crdsV1Beta1 := createThreeCrdsUsingGivenNamesAndApply(t, client, fixCrdGvrV1Beta1(),
				"apiextensions.k8s.io", "v1beta1", "origin", "kyma", "crd1", "crd2", "crd3")
			crdsV1 := createThreeCrdsUsingGivenNamesAndApply(t, client, fixCrdGvrV1(),
				"apiextensions.k8s.io", "v1", "origin", "kyma", "crd4", "crd5", "crd6")
			resourceManager := &mocks.ResourceManager{}
			resourceManager.On("DeleteCollectionOfResources", fixCrdGvkV1Beta1(), mock.AnythingOfType("v1.DeleteOptions"), mock.AnythingOfType("v1.ListOptions")).Return(
				func(gvk schema.GroupVersionKind, opts metav1.DeleteOptions, listOps metav1.ListOptions) error {
					return deleteAllMockObjs(t, client, crdsV1Beta1, fixCrdGvrV1Beta1())
				})
			resourceManager.On("DeleteCollectionOfResources", fixCrdGvkV1(), mock.AnythingOfType("v1.DeleteOptions"), mock.AnythingOfType("v1.ListOptions")).Return(
				func(gvk schema.GroupVersionKind, opts metav1.DeleteOptions, listOps metav1.ListOptions) error {
					return deleteAllMockObjs(t, client, crdsV1, fixCrdGvrV1())
				})
			uninstaller := getPostUninstaller(cfg, resourceManager, client, retryOptions)

			// when
			_, err := uninstaller.UninstallCRDs()

			// then
			require.NoError(t, err, "should not return any error")
			requireAllObjsNotExist(t, client, crdsV1Beta1, fixCrdGvrV1Beta1())
			requireAllObjsNotExist(t, client, crdsV1, fixCrdGvrV1Beta1())
			resourceManager.AssertNumberOfCalls(t, "DeleteCollectionOfResources", 2)
		})
	})

	t.Run("should uninstall CRDs", func(t *testing.T) {
		t.Run("when only CRDs labeled by Kyma are present on a cluster", func(t *testing.T) {
			// given
			client := fake.NewSimpleDynamicClientWithCustomListKinds(scheme, fixCrdGvrMap())
			crdsV1Beta1 := createThreeCrdsUsingGivenNamesAndApply(t, client, fixCrdGvrV1Beta1(),
				"apiextensions.k8s.io", "v1beta1", "origin", "kyma", "crd1", "crd2", "crd3")
			crdsV1 := createThreeCrdsUsingGivenNamesAndApply(t, client, fixCrdGvrV1(),
				"apiextensions.k8s.io", "v1", "origin", "kyma", "crd4", "crd5", "crd6")
			resourceManager := &mocks.ResourceManager{}
			resourceManager.On("DeleteCollectionOfResources", fixCrdGvkV1Beta1(), mock.AnythingOfType("v1.DeleteOptions"), mock.AnythingOfType("v1.ListOptions")).Return(
				func(gvk schema.GroupVersionKind, opts metav1.DeleteOptions, listOps metav1.ListOptions) error {
					return deleteAllMockObjs(t, client, crdsV1Beta1, fixCrdGvrV1Beta1())
				})
			resourceManager.On("DeleteCollectionOfResources", fixCrdGvkV1(), mock.AnythingOfType("v1.DeleteOptions"), mock.AnythingOfType("v1.ListOptions")).Return(
				func(gvk schema.GroupVersionKind, opts metav1.DeleteOptions, listOps metav1.ListOptions) error {
					return deleteAllMockObjs(t, client, crdsV1, fixCrdGvrV1())
				})
			uninstaller := getPostUninstaller(cfg, resourceManager, client, retryOptions)

			// when
			_, err := uninstaller.UninstallCRDs()

			// then
			require.NoError(t, err, "should not return any error")
			requireAllObjsNotExist(t, client, crdsV1Beta1, fixCrdGvrV1Beta1())
			requireAllObjsNotExist(t, client, crdsV1, fixCrdGvrV1Beta1())
			resourceManager.AssertNumberOfCalls(t, "DeleteCollectionOfResources", 2)
		})

		t.Run("labeled by Kyma and leave other", func(t *testing.T) {
			// given
			client := fake.NewSimpleDynamicClientWithCustomListKinds(scheme, fixCrdGvrMap())
			crdsLabeledByKyma := createThreeCrdsUsingGivenNamesAndApply(t, client, fixCrdGvrV1Beta1(),
				"apiextensions.k8s.io", "v1beta1", "origin", "kyma", "crd1", "crd2", "crd3")
			crdsNotLabeledByKyma := createThreeCrdsUsingGivenNamesAndApply(t, client, fixCrdGvrV1(),
				"apiextensions.k8s.io", "v1", "origin", "unknown", "crd4", "crd5", "crd6")
			resourceManager := &mocks.ResourceManager{}
			resourceManager.On("DeleteCollectionOfResources", fixCrdGvkV1Beta1(), mock.AnythingOfType("v1.DeleteOptions"), mock.AnythingOfType("v1.ListOptions")).Return(
				func(gvk schema.GroupVersionKind, opts metav1.DeleteOptions, listOps metav1.ListOptions) error {
					return deleteAllMockObjs(t, client, crdsLabeledByKyma, fixCrdGvrV1Beta1())
				})
			resourceManager.On("DeleteCollectionOfResources", fixCrdGvkV1(), mock.AnythingOfType("v1.DeleteOptions"), mock.AnythingOfType("v1.ListOptions")).Return(nil)
			uninstaller := getPostUninstaller(cfg, resourceManager, client, retryOptions)

			// when
			_, err := uninstaller.UninstallCRDs()

			// then
			require.NoError(t, err, "should not return any error")
			requireAllObjsNotExist(t, client, crdsLabeledByKyma, fixCrdGvrV1Beta1())
			requireAllObjsExistAndUnchanged(t, client, crdsNotLabeledByKyma, fixCrdGvrV1())
			resourceManager.AssertNumberOfCalls(t, "DeleteCollectionOfResources", 2)
		})

		t.Run("but only partially when error occurred for first CRD and no error occurred for the rest", func(t *testing.T) {
			// given
			client := fake.NewSimpleDynamicClientWithCustomListKinds(scheme, fixCrdGvrMap())
			crdsV1Beta1 := createThreeCrdsUsingGivenNamesAndApply(t, client, fixCrdGvrV1Beta1(),
				"apiextensions.k8s.io", "v1beta1", "origin", "kyma", "crd1", "crd2", "crd3")
			crdsV1 := createThreeCrdsUsingGivenNamesAndApply(t, client, fixCrdGvrV1(),
				"apiextensions.k8s.io", "v1", "origin", "kyma", "crd4", "crd5", "crd6")
			resourceManager := &mocks.ResourceManager{}
			resourceManager.On("DeleteCollectionOfResources", fixCrdGvkV1Beta1(), mock.AnythingOfType("v1.DeleteOptions"), mock.AnythingOfType("v1.ListOptions")).Return(
				func(gvk schema.GroupVersionKind, opts metav1.DeleteOptions, listOps metav1.ListOptions) error {
					return deleteAllMockObjs(t, client, crdsV1Beta1, fixCrdGvrV1Beta1())
				})
			resourceManager.On("DeleteCollectionOfResources", fixCrdGvkV1(), mock.AnythingOfType("v1.DeleteOptions"), mock.AnythingOfType("v1.ListOptions")).Return(errors.New("Error"))
			uninstaller := getPostUninstaller(cfg, resourceManager, client, retryOptions)

			// when
			_, err := uninstaller.UninstallCRDs()

			// then
			require.NoError(t, err, "should not return any error")
			requireAllObjsNotExist(t, client, crdsV1Beta1, fixCrdGvrV1Beta1())
			requireAllObjsExistAndUnchanged(t, client, crdsV1, fixCrdGvrV1())
			resourceManager.AssertNumberOfCalls(t, "DeleteCollectionOfResources", 2)
		})
	})

}

func createThreeCrdsUsingGivenNamesAndApply(t *testing.T, client *fake.FakeDynamicClient, gvr schema.GroupVersionResource, api, version, label, value, name1, name2, name3 string) []unstructured.Unstructured {
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
	ns1 := fixResourceWith(name1, label, value)
	ns2 := fixResourceWith(name2, label, value)
	ns3 := fixResourceWith(name3, label, value)
	return []unstructured.Unstructured{*ns1, *ns2, *ns3}
}

func createThreeNamespacesAndApply(t *testing.T, client *fake.FakeDynamicClient, gvr schema.GroupVersionResource, label, value string) []unstructured.Unstructured {
	namespaces := createThreeNamespacesUsing(label, value, "ns1", "ns2", "ns3")
	for _, ns := range namespaces {
		applyMockObj(t, client, &ns, gvr)
	}
	return namespaces
}

func applyMockObj(t *testing.T, client *fake.FakeDynamicClient, obj *unstructured.Unstructured, gvr schema.GroupVersionResource) {
	resultObj, err := client.Resource(gvr).Create(context.TODO(), obj, metav1.CreateOptions{})
	require.NoError(t, err, "object should be correctly created by fake client")
	require.NotNil(t, resultObj, "object returned by fake client should exist")
	require.Equal(t, obj, resultObj, "object returned by fake client should be equal to the created one")
}

func deleteMockObj(t *testing.T, client *fake.FakeDynamicClient, obj *unstructured.Unstructured, gvr schema.GroupVersionResource) error {
	err := client.Resource(gvr).Delete(context.TODO(), obj.GetName(), metav1.DeleteOptions{})
	require.NoError(t, err, "object should be correctly deleted by fake client")
	return err
}

func deleteAllMockObjs(t *testing.T, client *fake.FakeDynamicClient, objs []unstructured.Unstructured, gvr schema.GroupVersionResource) error {
	for _, obj := range objs {
		_ = deleteMockObj(t, client, &obj, gvr)
	}
	return nil
}

func requireNoCrdsOnTheCluster(t *testing.T, client *fake.FakeDynamicClient) {
	requireNoGenericCrdsOnTheCluster(t, client)
	requireNoKymaCrdsOnTheCluster(t, client)
}

func requireNoGenericCrdsOnTheCluster(t *testing.T, client *fake.FakeDynamicClient) {
	requireNoGivenCrdsOnTheCluster(t, client, fixCrdGvrV1Beta1(), metav1.ListOptions{})
	requireNoGivenCrdsOnTheCluster(t, client, fixCrdGvrV1(), metav1.ListOptions{})
}

func requireNoKymaCrdsOnTheCluster(t *testing.T, client *fake.FakeDynamicClient) {
	requireNoGivenCrdsOnTheCluster(t, client, fixCrdGvrV1Beta1(), metav1.ListOptions{LabelSelector: "origin=kyma"})
	requireNoGivenCrdsOnTheCluster(t, client, fixCrdGvrV1(), metav1.ListOptions{LabelSelector: "origin=kyma"})
}

func requireNoGivenCrdsOnTheCluster(t *testing.T, client *fake.FakeDynamicClient, gvr schema.GroupVersionResource, listOpts metav1.ListOptions) {
	resourcesList, err := client.Resource(gvr).List(context.TODO(), listOpts)
	require.NoError(t, err)
	require.Empty(t, resourcesList.Items)
}

func requireObjExistsAndUnchanged(t *testing.T, client *fake.FakeDynamicClient, obj *unstructured.Unstructured, gvr schema.GroupVersionResource) {
	resultObj, err := client.Resource(gvr).Get(context.TODO(), obj.GetName(), metav1.GetOptions{})
	require.NoError(t, err, "object should be correctly returned by fake client")
	require.NotNil(t, resultObj, "object returned by fake client should exist")
	require.Equal(t, obj, resultObj, "object returned by fake client should be equal to the created one")
}

func requireAllObjsExistAndUnchanged(t *testing.T, client *fake.FakeDynamicClient, objs []unstructured.Unstructured, gvr schema.GroupVersionResource) {
	for _, obj := range objs {
		requireObjExistsAndUnchanged(t, client, &obj, gvr)
	}
}

func requireObjNotExists(t *testing.T, client *fake.FakeDynamicClient, obj *unstructured.Unstructured, gvr schema.GroupVersionResource) {
	resultObj, err := client.Resource(gvr).Get(context.TODO(), obj.GetName(), metav1.GetOptions{})
	require.Error(t, err, "object should be not found")
	require.Nil(t, resultObj, "object returned by fake client should not exist")
}

func requireAllObjsNotExist(t *testing.T, client *fake.FakeDynamicClient, objs []unstructured.Unstructured, gvr schema.GroupVersionResource) {
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

func fixResourceWith(name, label, value string) *unstructured.Unstructured {
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

func getTestingConfig() Config {
	return Config{
		Log:                      logger.NewLogger(true),
		InstallationResourcePath: "installationResourcePath",
		KubeconfigSource: config.KubeconfigSource{
			Path:    "path",
			Content: "",
		},
	}
}

func getTestingRetryOptions() []retry.Option {
	return []retry.Option{
		retry.Delay(0),
		retry.Attempts(1),
		retry.DelayType(retry.FixedDelay),
	}
}

func getPostUninstaller(cfg Config, resourceManager preinstaller.ResourceManager, dynamicClient dynamic.Interface, retryOptions []retry.Option) *PostUninstaller {
	return &PostUninstaller{
		cfg:             cfg,
		dynamicClient:   dynamicClient,
		resourceManager: resourceManager,
		retryOptions:    retryOptions,
	}
}

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
			resourceManager := &mocks.ResourceManager{}
			uninstaller := getPostUninstaller(cfg, resourceManager, client, retryOptions)

			// when
			output, err := uninstaller.UninstallCRDs()

			// then
			requireNoErrorAndNoOutput(t, err, output)
			resourceManager.AssertNumberOfCalls(t, "DeleteResource", 0)
		})

		t.Run("when CRDs not labeled by Kyma are present on a cluster", func(t *testing.T) {
			// given
			crd1 := fixCrdResourceWith("crd1", "apiextensions.k8s.io", "v1beta1", "not-kyma-label", "true")
			crd2 := fixCrdResourceWith("crd2", "apiextensions.k8s.io", "v1beta1", "not-kyma-label", "true")
			crd3 := fixCrdResourceWith("crd3", "apiextensions.k8s.io", "v1beta1", "not-kyma-label", "true")
			client := fake.NewSimpleDynamicClientWithCustomListKinds(scheme, fixCrdGvrMap())
			applyMockObj(t, client, crd1, fixCrdGvrV1Beta1())
			applyMockObj(t, client, crd2, fixCrdGvrV1Beta1())
			applyMockObj(t, client, crd3, fixCrdGvrV1Beta1())
			resourceManager := &mocks.ResourceManager{}
			uninstaller := getPostUninstaller(cfg, resourceManager, client, retryOptions)

			// when
			output, err := uninstaller.UninstallCRDs()

			// then
			requireNoErrorAndNoOutput(t, err, output)
			requireObjExistsAndUnchanged(t, client, crd1, fixCrdGvrV1Beta1())
			requireObjExistsAndUnchanged(t, client, crd2, fixCrdGvrV1Beta1())
			requireObjExistsAndUnchanged(t, client, crd3, fixCrdGvrV1Beta1())
			resourceManager.AssertNumberOfCalls(t, "DeleteResource", 0)
		})

		t.Run("when CRDs labeled by Kyma but with incorrect value are present on a cluster", func(t *testing.T) {
			// given
			crd1 := fixCrdResourceWith("crd1", "apiextensions.k8s.io", "v1beta1", "not-kyma-label", "value")
			crd2 := fixCrdResourceWith("crd2", "apiextensions.k8s.io", "v1beta1", "not-kyma-label", "value")
			crd3 := fixCrdResourceWith("crd3", "apiextensions.k8s.io", "v1beta1", "not-kyma-label", "value")
			client := fake.NewSimpleDynamicClientWithCustomListKinds(scheme, fixCrdGvrMap())
			applyMockObj(t, client, crd1, fixCrdGvrV1Beta1())
			applyMockObj(t, client, crd2, fixCrdGvrV1Beta1())
			applyMockObj(t, client, crd3, fixCrdGvrV1Beta1())
			resourceManager := &mocks.ResourceManager{}
			uninstaller := getPostUninstaller(cfg, resourceManager, client, retryOptions)

			// when
			output, err := uninstaller.UninstallCRDs()

			// then
			requireNoErrorAndNoOutput(t, err, output)
			requireObjExistsAndUnchanged(t, client, crd1, fixCrdGvrV1Beta1())
			requireObjExistsAndUnchanged(t, client, crd2, fixCrdGvrV1Beta1())
			requireObjExistsAndUnchanged(t, client, crd3, fixCrdGvrV1Beta1())
			resourceManager.AssertNumberOfCalls(t, "DeleteResource", 0)
		})

		t.Run("when CRDs labeled by Kyma are present on a cluster but CRD api does not match", func(t *testing.T) {
			// given
			crd1 := fixCrdResourceWith("crd1", "otherapi", "v1beta1", "kyma-crd", "true")
			crd2 := fixCrdResourceWith("crd2", "otherapi", "v1beta1", "kyma-crd", "true")
			crd3 := fixCrdResourceWith("crd3", "otherapi", "v1beta1", "kyma-crd", "true")
			client := fake.NewSimpleDynamicClientWithCustomListKinds(scheme, fixCrdGvrMap())
			resourceManager := &mocks.ResourceManager{}
			uninstaller := getPostUninstaller(cfg, resourceManager, client, retryOptions)
			applyMockObj(t, client, crd1, fixCrdGvrOtherGroup())
			applyMockObj(t, client, crd2, fixCrdGvrOtherGroup())
			applyMockObj(t, client, crd3, fixCrdGvrOtherGroup())

			// when
			output, err := uninstaller.UninstallCRDs()

			// then
			requireNoErrorAndNoOutput(t, err, output)
			requireObjExistsAndUnchanged(t, client, crd1, fixCrdGvrOtherGroup())
			requireObjExistsAndUnchanged(t, client, crd2, fixCrdGvrOtherGroup())
			requireObjExistsAndUnchanged(t, client, crd3, fixCrdGvrOtherGroup())
			resourceManager.AssertNumberOfCalls(t, "DeleteResource", 0)
		})

		t.Run("when CRDs labeled by Kyma are present on a cluster but CRD version does not match", func(t *testing.T) {
			// given
			crd1 := fixCrdResourceWith("crd1", "apiextensions.k8s.io", "otherversion", "kyma-crd", "true")
			crd2 := fixCrdResourceWith("crd2", "apiextensions.k8s.io", "otherversion", "kyma-crd", "true")
			crd3 := fixCrdResourceWith("crd3", "apiextensions.k8s.io", "otherversion", "kyma-crd", "true")
			client := fake.NewSimpleDynamicClientWithCustomListKinds(scheme, fixCrdGvrMap())
			applyMockObj(t, client, crd1, fixCrdGvrOtherVersion())
			applyMockObj(t, client, crd2, fixCrdGvrOtherVersion())
			applyMockObj(t, client, crd3, fixCrdGvrOtherVersion())
			resourceManager := &mocks.ResourceManager{}
			uninstaller := getPostUninstaller(cfg, resourceManager, client, retryOptions)

			// when
			output, err := uninstaller.UninstallCRDs()

			// then
			requireNoErrorAndNoOutput(t, err, output)
			requireObjExistsAndUnchanged(t, client, crd1, fixCrdGvrOtherVersion())
			requireObjExistsAndUnchanged(t, client, crd2, fixCrdGvrOtherVersion())
			requireObjExistsAndUnchanged(t, client, crd3, fixCrdGvrOtherVersion())
			resourceManager.AssertNumberOfCalls(t, "DeleteResource", 0)
		})

		t.Run("when objects of type different than CRD labeled by Kyma are present on a cluster", func(t *testing.T) {
			// given
			obj1 := fixResourceWith("obj1", "kyma-crd", "true")
			obj2 := fixResourceWith("obj2", "kyma-crd", "true")
			obj3 := fixResourceWith("obj3", "kyma-crd", "true")
			client := fake.NewSimpleDynamicClientWithCustomListKinds(scheme, fixCrdGvrMap())
			applyMockObj(t, client, obj1, fixNamespaceGvr())
			applyMockObj(t, client, obj2, fixNamespaceGvr())
			applyMockObj(t, client, obj3, fixNamespaceGvr())
			resourceManager := &mocks.ResourceManager{}
			uninstaller := getPostUninstaller(cfg, resourceManager, client, retryOptions)

			// when
			output, err := uninstaller.UninstallCRDs()

			// then
			requireNoErrorAndNoOutput(t, err, output)
			requireObjExistsAndUnchanged(t, client, obj1, fixNamespaceGvr())
			requireObjExistsAndUnchanged(t, client, obj2, fixNamespaceGvr())
			requireObjExistsAndUnchanged(t, client, obj3, fixNamespaceGvr())
			resourceManager.AssertNumberOfCalls(t, "DeleteResource", 0)
		})

		t.Run("when all of them were correct but errors occurred when deleting them", func(t *testing.T) {
			// given
			crd1 := fixCrdResourceWith("crd1", "apiextensions.k8s.io", "v1beta1", "kyma-crd", "true")
			crd2 := fixCrdResourceWith("crd2", "apiextensions.k8s.io", "v1beta1", "kyma-crd", "true")
			crd3 := fixCrdResourceWith("crd3", "apiextensions.k8s.io", "v1beta1", "kyma-crd", "true")
			client := fake.NewSimpleDynamicClientWithCustomListKinds(scheme, fixCrdGvrMap())
			applyMockObj(t, client, crd1, fixCrdGvrV1Beta1())
			applyMockObj(t, client, crd2, fixCrdGvrV1Beta1())
			applyMockObj(t, client, crd3, fixCrdGvrV1Beta1())
			resourceManager := &mocks.ResourceManager{}
			resourceManager.On("DeleteResource", mock.AnythingOfType("string"), mock.AnythingOfType("schema.GroupVersionKind")).Return(
				func(name string, gvk schema.GroupVersionKind) error {
					return errors.New("Could not delete " + name)
				})
			uninstaller := getPostUninstaller(cfg, resourceManager, client, retryOptions)

			// when
			output, err := uninstaller.UninstallCRDs()

			// then
			require.NoError(t, err, "should not return any error")
			require.Empty(t, output.Deleted, "should delete crd1, crd2")
			require.Equal(t,3, len(output.NotDeleted), "should not delete crd3")
			requireObjExistsAndUnchanged(t, client, crd1, fixCrdGvrV1Beta1())
			requireObjExistsAndUnchanged(t, client, crd2, fixCrdGvrV1Beta1())
			requireObjExistsAndUnchanged(t, client, crd3, fixCrdGvrV1Beta1())
			resourceManager.AssertNumberOfCalls(t, "DeleteResource", 3)
		})
	})

	t.Run("should uninstall CRDs", func(t *testing.T) {
		t.Run("when only CRDs labeled by Kyma are present on a cluster", func(t *testing.T) {
			// given
			crd1 := fixCrdResourceWith("crd1", "apiextensions.k8s.io", "v1beta1", "kyma-crd", "true")
			crd2 := fixCrdResourceWith("crd2", "apiextensions.k8s.io", "v1beta1", "kyma-crd", "true")
			crd3 := fixCrdResourceWith("crd3", "apiextensions.k8s.io", "v1beta1", "kyma-crd", "true")
			client := fake.NewSimpleDynamicClientWithCustomListKinds(scheme, fixCrdGvrMap())
			applyMockObj(t, client, crd1, fixCrdGvrV1Beta1())
			applyMockObj(t, client, crd2, fixCrdGvrV1Beta1())
			applyMockObj(t, client, crd3, fixCrdGvrV1Beta1())
			resourceManager := &mocks.ResourceManager{}
			resourceManager.On("DeleteResource", mock.AnythingOfType("string"), mock.AnythingOfType("schema.GroupVersionKind")).Return(
				func(name string, gvk schema.GroupVersionKind) error {
					return client.Resource(fixCrdGvrV1Beta1()).Delete(context.TODO(), name, metav1.DeleteOptions{})
				})
			uninstaller := getPostUninstaller(cfg, resourceManager, client, retryOptions)

			// when
			output, err := uninstaller.UninstallCRDs()

			// then
			require.NoError(t, err, "should not return any error")
			require.Equal(t,3, len(output.Deleted), "should delete crd1, crd2 and crd3")
			require.Empty(t, output.NotDeleted, "should leave all other resources")
			requireObjNotExists(t, client, crd1, fixCrdGvrV1Beta1())
			requireObjNotExists(t, client, crd2, fixCrdGvrV1Beta1())
			requireObjNotExists(t, client, crd3, fixCrdGvrV1Beta1())
			resourceManager.AssertNumberOfCalls(t, "DeleteResource", 3)
		})

		t.Run("labeled by Kyma and leave other", func(t *testing.T) {
			// given
			crd1 := fixCrdResourceWith("crd1", "apiextensions.k8s.io", "v1beta1", "kyma-crd", "true")
			crd2 := fixCrdResourceWith("crd2", "apiextensions.k8s.io", "v1beta1", "kyma-crd", "true")
			crd3 := fixCrdResourceWith("crd3", "apiextensions.k8s.io", "v1beta1", "not-kyma-label", "true")
			client := fake.NewSimpleDynamicClientWithCustomListKinds(scheme, fixCrdGvrMap())
			applyMockObj(t, client, crd1, fixCrdGvrV1Beta1())
			applyMockObj(t, client, crd2, fixCrdGvrV1Beta1())
			applyMockObj(t, client, crd3, fixCrdGvrV1Beta1())
			resourceManager := &mocks.ResourceManager{}
			resourceManager.On("DeleteResource", mock.AnythingOfType("string"), mock.AnythingOfType("schema.GroupVersionKind")).Return(
				func(name string, gvk schema.GroupVersionKind) error {
					return client.Resource(fixCrdGvrV1Beta1()).Delete(context.TODO(), name, metav1.DeleteOptions{})
				})
			uninstaller := getPostUninstaller(cfg, resourceManager, client, retryOptions)

			// when
			output, err := uninstaller.UninstallCRDs()

			// then
			require.NoError(t, err, "should not return any error")
			require.Equal(t,2, len(output.Deleted), "should delete crd1, crd2 and crd3")
			require.Empty(t, output.NotDeleted, "should leave all other resources")
			requireObjNotExists(t, client, crd1, fixCrdGvrV1Beta1())
			requireObjNotExists(t, client, crd2, fixCrdGvrV1Beta1())
			requireObjExistsAndUnchanged(t, client, crd3, fixCrdGvrV1Beta1())
			resourceManager.AssertNumberOfCalls(t, "DeleteResource", 2)
		})
	})

	t.Run("but only partially when error occurred for first CRD and no error occurred for the rest", func(t *testing.T) {
		// given
		crd1 := fixCrdResourceWith("crd1", "apiextensions.k8s.io", "v1beta1", "kyma-crd", "true")
		crd2 := fixCrdResourceWith("crd2", "apiextensions.k8s.io", "v1beta1", "kyma-crd", "true")
		crd3 := fixCrdResourceWith("crd3", "apiextensions.k8s.io", "v1beta1", "kyma-crd", "true")
		client := fake.NewSimpleDynamicClientWithCustomListKinds(scheme, fixCrdGvrMap())
		applyMockObj(t, client, crd1, fixCrdGvrV1Beta1())
		applyMockObj(t, client, crd2, fixCrdGvrV1Beta1())
		applyMockObj(t, client, crd3, fixCrdGvrV1Beta1())
		resourceManager := &mocks.ResourceManager{}
		resourceManager.On("DeleteResource", mock.AnythingOfType("string"), mock.AnythingOfType("schema.GroupVersionKind")).Twice().Return(
			func(name string, gvk schema.GroupVersionKind) error {
				return client.Resource(fixCrdGvrV1Beta1()).Delete(context.TODO(), name, metav1.DeleteOptions{})
			})
		resourceManager.On("DeleteResource", mock.AnythingOfType("string"), mock.AnythingOfType("schema.GroupVersionKind")).Once().Return(
			func(name string, gvk schema.GroupVersionKind) error {
				return errors.New("Could not delete " + name)
			})
		uninstaller := getPostUninstaller(cfg, resourceManager, client, retryOptions)

		// when
		output, err := uninstaller.UninstallCRDs()

		// then
		require.NoError(t, err, "should not return any error")
		require.Equal(t,2, len(output.Deleted), "should delete crd1, crd2")
		require.Equal(t,1, len(output.NotDeleted), "should not delete crd3")
		requireObjNotExists(t, client, crd1, fixCrdGvrV1Beta1())
		requireObjNotExists(t, client, crd2, fixCrdGvrV1Beta1())
		requireObjExistsAndUnchanged(t, client, crd3, fixCrdGvrV1Beta1())
		resourceManager.AssertNumberOfCalls(t, "DeleteResource", 3)
	})
}

func applyMockObj(t *testing.T, client *fake.FakeDynamicClient, obj *unstructured.Unstructured, gvr schema.GroupVersionResource) {
	resultObj, err := client.Resource(gvr).Create(context.TODO(), obj, metav1.CreateOptions{})
	require.NoError(t, err, "object should be correctly created by fake client")
	require.NotNil(t, resultObj, "object returned by fake client should exist")
	require.Equal(t, obj, resultObj, "object returned by fake client should be equal to the created one")
}

func requireNoErrorAndNoOutput(t *testing.T,  err error, output Output) {
	require.NoError(t, err, "should not return any error")
	require.Empty(t, output.Deleted, "should not delete any resource")
	require.Empty(t, output.NotDeleted, "should leave all other resources")
}

func requireObjExistsAndUnchanged(t *testing.T, client *fake.FakeDynamicClient, obj *unstructured.Unstructured, gvr schema.GroupVersionResource) {
	resultObj, err := client.Resource(gvr).Get(context.TODO(), obj.GetName(), metav1.GetOptions{})
	require.NoError(t, err, "object should be correctly returned by fake client")
	require.NotNil(t, resultObj, "object returned by fake client should exist")
	require.Equal(t, obj, resultObj, "object returned by fake client should be equal to the created one")
}

func requireObjNotExists(t *testing.T, client *fake.FakeDynamicClient, obj *unstructured.Unstructured, gvr schema.GroupVersionResource) {
	resultObj, err := client.Resource(gvr).Get(context.TODO(), obj.GetName(), metav1.GetOptions{})
	require.Error(t, err, "object should be not found")
	require.Nil(t, resultObj, "object returned by fake client should not exist")
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

func fixCrdResourceWith(name string, api string, version string, label string, value string) *unstructured.Unstructured {
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

func fixResourceWith(name string, label string, value string) *unstructured.Unstructured {
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

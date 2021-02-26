package preinstaller

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/fake"
	"reflect"
	"regexp"
	"testing"
)

func TestResourceManager_CreateResource(t *testing.T) {

	scheme := runtime.NewScheme()
	dynamicClient := fake.NewSimpleDynamicClient(scheme)
	cfg := getTestingConfig()
	retryOptions := getTestingRetryOptions(cfg)

	t.Run("should create resource", func(t *testing.T) {
		// given
		manager := NewDefaultResourceManager(dynamicClient, retryOptions)
		resourceName := "namespace"
		resource := fixNamespaceResourceWith(resourceName)
		resourceSchema := prepareSchemaFor(resource)

		// when
		err := manager.CreateResource(resource, resourceSchema)

		// then
		assert.NoError(t, err)
	})
}

func TestResourceManager_GetResource(t *testing.T) {

	scheme := runtime.NewScheme()
	dynamicClient := fake.NewSimpleDynamicClient(scheme)
	cfg := getTestingConfig()
	retryOptions := getTestingRetryOptions(cfg)

	t.Run("should not get not found resource", func(t *testing.T) {
		// given
		manager := NewDefaultResourceManager(dynamicClient, retryOptions)
		resourceName := "resourceName"
		resourceSchema := schema.GroupVersionResource{}

		// when
		obj, err := manager.GetResource(resourceName, resourceSchema)

		// then
		expectedError := "not found"
		receivedError := err.Error()
		matched, err := regexp.MatchString(expectedError, receivedError)
		assert.True(t, matched, fmt.Sprintf("Expected error message: %s but got: %s", expectedError, receivedError))
		assert.Nil(t, obj)
	})

	t.Run("should get pre-created resource", func(t *testing.T) {
		// given
		resourceName := "namespace"
		resource := fixNamespaceResourceWith(resourceName)
		customDynamicClient := fake.NewSimpleDynamicClient(scheme, resource)
		manager := NewDefaultResourceManager(customDynamicClient, retryOptions)
		resourceSchema := prepareSchemaFor(resource)

		// when
		retrievedResource, err := manager.GetResource(resourceName, resourceSchema)

		// then
		assert.NoError(t, err)
		assert.NotNil(t, retrievedResource)
		receivedResourceName := retrievedResource.GetName()
		matched, err := regexp.MatchString(resourceName, receivedResourceName)
		assert.True(t, matched, fmt.Sprintf("Expected error message: %s but got: %s", resourceName, receivedResourceName))
	})
}

func TestResourceManager_UpdateRefreshableResource(t *testing.T) {

	scheme := runtime.NewScheme()
	cfg := getTestingConfig()
	retryOptions := getTestingRetryOptions(cfg)

	t.Run("should update resource", func(t *testing.T) {
		// given
		resourceName := "namespace"
		resource := fixNamespaceResourceWith(resourceName)
		customDynamicClient := fake.NewSimpleDynamicClient(scheme, resource)
		manager := NewDefaultResourceManager(customDynamicClient, retryOptions)
		resourceSchema := prepareSchemaFor(resource)
		labels := map[string]string{
			"key": "value",
		}
		resource.SetLabels(labels)

		// when
		newResource, err := manager.UpdateResource(resource, resourceSchema)

		// then
		assert.NoError(t, err)
		assert.NotNil(t, newResource)
		assert.True(t, reflect.DeepEqual(newResource.GetLabels(), labels))
	})
}

func fixNamespaceResourceWith(name string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Namespace",
			"metadata": map[string]interface{}{
				"name": name,
			},
		},
	}
}

func prepareSchemaFor(resource *unstructured.Unstructured) schema.GroupVersionResource {
	gvk := resource.GroupVersionKind()
	return schema.GroupVersionResource{
		Group:    gvk.Group,
		Version:  gvk.Version,
		Resource: pluralForm(gvk.Kind),
	}
}

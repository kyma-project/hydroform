package preinstaller

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/fake"
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
		resource := unstructured.Unstructured{}
		resourceSchema := schema.GroupVersionResource{}

		// when
		err := manager.CreateResource(&resource, resourceSchema)

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

	t.Run("should get created resource", func(t *testing.T) {
		// given
		manager := NewDefaultResourceManager(dynamicClient, retryOptions)
		resourceName := "resourceName"
		testObj := unstructured.Unstructured{Object: map[string]interface{}{
			"name":      "resourceName",
		}}
		resourceSchema := schema.GroupVersionResource{}
		err := manager.CreateResource(&testObj, resourceSchema)
		assert.NoError(t, err)

		// when
		obj, err := manager.GetResource(resourceName, resourceSchema)

		// then
		expectedError := "not found"
		receivedError := err.Error()
		matched, err := regexp.MatchString(expectedError, receivedError)
		assert.True(t, matched, fmt.Sprintf("Expected error message: %s but got: %s", expectedError, receivedError))
		assert.Nil(t, obj)
	})
}

func TestResourceManager_UpdateRefreshableResource(t *testing.T) {

	scheme := runtime.NewScheme()
	dynamicClient := fake.NewSimpleDynamicClient(scheme)
	cfg := getTestingConfig()
	retryOptions := getTestingRetryOptions(cfg)

	t.Run("should update resource", func(t *testing.T) {
		// given
		manager := NewDefaultResourceManager(dynamicClient, retryOptions)
		resource := unstructured.Unstructured{}
		resourceSchema := schema.GroupVersionResource{}

		// when
		err := manager.UpdateRefreshableResource(&resource, resourceSchema)

		// then
		assert.NoError(t, err)
	})
}

package deployment

import (
	"fmt"
	"github.com/avast/retry-go"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/logger"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/fake"
	//"k8s.io/client-go/testing"
	"reflect"
	"regexp"
	"testing"
)

func TestResourceManager_CreateResource(t *testing.T) {

	scheme := runtime.NewScheme()
	dynamicClient := fake.NewSimpleDynamicClient(scheme)
	retryOptions := getTestingRetryOptions()
	log := logger.NewLogger(true)

	t.Run("should create resource", func(t *testing.T) {
		// given
		manager := getDefaultResourceManager(dynamicClient, log, retryOptions)
		resourceName := "namespace"
		resource := fixResourceWith(resourceName)
		resourceSchema := fixResourceGvkSchema()

		// when
		err := manager.CreateResource(resource, resourceSchema, metav1.CreateOptions{})

		// then
		assert.NoError(t, err)
	})
}

func TestResourceManager_GetResource(t *testing.T) {

	scheme := runtime.NewScheme()
	dynamicClient := fake.NewSimpleDynamicClient(scheme)
	retryOptions := getTestingRetryOptions()
	log := logger.NewLogger(true)

	t.Run("should proceed without error when resource is not found", func(t *testing.T) {
		// given
		manager := getDefaultResourceManager(dynamicClient, log, retryOptions)
		resourceName := "resourceName"
		resourceSchema := schema.GroupVersionKind{}

		// when
		obj, err := manager.GetResource(resourceName, resourceSchema, metav1.GetOptions{})

		// then
		assert.NoError(t, err)
		assert.Nil(t, obj)
	})

	t.Run("should get pre-created resource", func(t *testing.T) {
		// given
		resourceName := "namespace"
		resource := fixResourceWith(resourceName)
		customDynamicClient := fake.NewSimpleDynamicClient(scheme, resource)
		manager := getDefaultResourceManager(customDynamicClient, log, retryOptions)
		resourceSchema := fixResourceGvkSchema()

		// when
		retrievedResource, err := manager.GetResource(resourceName, resourceSchema, metav1.GetOptions{})

		// then
		assert.NoError(t, err)
		assert.NotNil(t, retrievedResource)
		receivedResourceName := retrievedResource.GetName()
		matched, err := regexp.MatchString(resourceName, receivedResourceName)
		assert.True(t, matched, fmt.Sprintf("Expected error message: %s but got: %s", resourceName, receivedResourceName))
	})
}

func TestResourceManager_UpdateResource(t *testing.T) {

	scheme := runtime.NewScheme()
	retryOptions := getTestingRetryOptions()
	log := logger.NewLogger(true)

	t.Run("should update resource", func(t *testing.T) {
		// given
		resourceName := "namespace"
		resource := fixResourceWith(resourceName)
		resourceSchema := fixResourceGvkSchema()
		customDynamicClient := fake.NewSimpleDynamicClient(scheme, resource)
		manager := getDefaultResourceManager(customDynamicClient, log, retryOptions)
		labels := map[string]string{
			"key": "value",
		}
		resource.SetLabels(labels)

		// when
		newResource, err := manager.UpdateResource(resource, resourceSchema, metav1.UpdateOptions{})

		// then
		assert.NoError(t, err)
		assert.NotNil(t, newResource)
		assert.True(t, reflect.DeepEqual(newResource.GetLabels(), labels))
	})
}

func TestResourceManager_DeleteCollectionOfResources(t *testing.T) {

	scheme := runtime.NewScheme()
	retryOptions := getTestingRetryOptions()
	log := logger.NewLogger(true)

	t.Run("should call delete-collection action when delecting collection of resources", func(t *testing.T) {
		// given
		client := fake.NewSimpleDynamicClient(scheme)
		manager := getDefaultResourceManager(client, log, retryOptions)
		//gvk := fixResourceGvkSchema()

		// when
		err := manager.DeleteCollectionOfResources(fixResourceGvkSchema(), metav1.DeleteOptions{}, metav1.ListOptions{})

		// then
		assert.NoError(t, err)

		// and then
		actions := client.Actions()
		assert.NotNil(t, actions)
		assert.Equal(t, 1, len(actions))
		assert.Equal(t, "delete-collection", actions[0].GetVerb())
		assert.Equal(t, fixResourceGvrSchema(), actions[0].GetResource())
	})
}

func fixResourceWith(name string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "group/v1",
			"kind":       "kind",
			"metadata": map[string]interface{}{
				"name": name,
				"labels": map[string]interface{}{
					"key": "value",
				},
			},
		},
	}
}

func fixResourceGvkSchema() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   "group",
		Version: "v1",
		Kind:    "kind",
	}
}

func fixResourceGvrSchema() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "group",
		Version:  "v1",
		Resource: "kinds",
	}
}

func getDefaultResourceManager(dynamicClient dynamic.Interface, log logger.Interface, retryOptions []retry.Option) *DefaultResourceManager {
	return &DefaultResourceManager{
		dynamicClient: dynamicClient,
		log:           log,
		retryOptions:  retryOptions,
	}
}

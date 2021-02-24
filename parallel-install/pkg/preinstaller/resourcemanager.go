package preinstaller

import (
	"context"
	"github.com/avast/retry-go"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// ResourceManager manages resources on a k8s cluster.
type ResourceManager interface {
	// CreateResource of any type that matches the schema on k8s cluster.
	CreateResource(resource *unstructured.Unstructured, resourceSchema schema.GroupVersionResource) error

	// GetResource of a given name from a k8s cluster, that matches the schema.
	GetResource(resourceName string, resourceSchema schema.GroupVersionResource) (*unstructured.Unstructured, error)
}

// DefaultResourceManager provides a default implementation of ResourceManager.
type DefaultResourceManager struct {
	dynamicClient dynamic.Interface
	retryOptions  []retry.Option
}

// NewResourceManager creates a new instance of ResourceManager.
func NewDefaultResourceManager(dynamicClient dynamic.Interface, retryOptions []retry.Option) *DefaultResourceManager {
	return &DefaultResourceManager{
		dynamicClient: dynamicClient,
		retryOptions:  retryOptions,
	}
}

func (c *DefaultResourceManager) CreateResource(resource *unstructured.Unstructured, resourceSchema schema.GroupVersionResource) error {
	var err error
	err = retry.Do(func() error {
		if _, err = c.dynamicClient.Resource(resourceSchema).Create(context.TODO(), resource, metav1.CreateOptions{}); err != nil {
			return err
		}

		return nil
	}, c.retryOptions...)

	return err
}

func (c *DefaultResourceManager) GetResource(resourceName string, resourceSchema schema.GroupVersionResource) (*unstructured.Unstructured, error) {
	var obj *unstructured.Unstructured
	err := retry.Do(func() error {
		var err error

		if obj, err = c.dynamicClient.Resource(resourceSchema).Get(context.TODO(), resourceName, metav1.GetOptions{}); err != nil {
			return err
		}

		return nil
	}, c.retryOptions...)

	return obj, err
}

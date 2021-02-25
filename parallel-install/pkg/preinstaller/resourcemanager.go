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
	// Performs retries on unsuccessful resource creation action.
	CreateResource(resource *unstructured.Unstructured, resourceSchema schema.GroupVersionResource) error

	// GetResource of a given name from a k8s cluster, that matches the schema.
	// Performs retries on unsuccessful resource retrieval action.
	GetResource(resourceName string, resourceSchema schema.GroupVersionResource) (*unstructured.Unstructured, error)

	// UpdateResourceNoRetries of a given name from a k8s cluster, that matches the schema.
	// Performs only one update attempt.
	UpdateResourceNoRetries(resource *unstructured.Unstructured, resourceSchema schema.GroupVersionResource) error

	// UpdateRefreshableResource of a given name from a k8s cluster, that matches the schema.
	// Performs retries on unsuccessful resource update action. Before each update the latest resource version is
	// retrieved from the k8s cluster.
	UpdateRefreshableResource(resource *unstructured.Unstructured, resourceSchema schema.GroupVersionResource) error
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

func (c *DefaultResourceManager) UpdateResourceNoRetries(resource *unstructured.Unstructured, resourceSchema schema.GroupVersionResource) error {
	return c.updateResource(resource, resourceSchema)
}

func (c *DefaultResourceManager) UpdateRefreshableResource(resource *unstructured.Unstructured, resourceSchema schema.GroupVersionResource) error {
	var err error
	err = retry.Do(func() error {
		refreshResource, err := c.GetResource(resource.GetName(), resourceSchema)
		if err != nil {
			return nil
		}

		err = c.updateResource(refreshResource, resourceSchema)
		if err != nil {
			return err
		}

		return nil
	}, c.retryOptions...)

	return err
}

func (c *DefaultResourceManager) updateResource(resource *unstructured.Unstructured, resourceSchema schema.GroupVersionResource) error {
	var err error
	if _, err = c.dynamicClient.Resource(resourceSchema).Update(context.TODO(), resource, metav1.UpdateOptions{}); err != nil {
		return err
	}

	return nil
}


package mocks

import (
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// RetrievalErrorResourceManager is a mock type for the ResourceManager type.
type RetrievalErrorResourceManager struct{}

// GetResource provides a mock function wih given fields: resourceName, resourceSchema.
func (c *RetrievalErrorResourceManager) GetResource(resourceName string, resourceSchema schema.GroupVersionResource) (*unstructured.Unstructured, error) {
	return nil, errors.New("Get resource error")
}

// CreateResource provides a mock function wih given fields: resource, resourceSchema.
func (c *RetrievalErrorResourceManager) CreateResource(resource *unstructured.Unstructured, resourceSchema schema.GroupVersionResource) error {
	return errors.New("Create resource error")
}

// ResourceExistedResourceManagerMock is a mock type for the ResourceManager type.
type ResourceExistedResourceManagerMock struct{}

// GetResource provides a mock function wih given fields: resourceName, resourceSchema.
func (c *ResourceExistedResourceManagerMock) GetResource(resourceName string, resourceSchema schema.GroupVersionResource) (*unstructured.Unstructured, error) {
	return &unstructured.Unstructured{}, nil
}

// CreateResource provides a mock function wih given fields: resource, resourceSchema.
func (c *ResourceExistedResourceManagerMock) CreateResource(resource *unstructured.Unstructured, resourceSchema schema.GroupVersionResource) error {
	return nil
}

// CreationErrorResourceManager is a mock type for the ResourceManager type.
type CreationErrorResourceManager struct{}

// GetResource provides a mock function wih given fields: resourceName, resourceSchema.
func (c *CreationErrorResourceManager) GetResource(resourceName string, resourceSchema schema.GroupVersionResource) (*unstructured.Unstructured, error) {
	return nil, nil
}

// CreateResource provides a mock function wih given fields: resource, resourceSchema.
func (c *CreationErrorResourceManager) CreateResource(resource *unstructured.Unstructured, resourceSchema schema.GroupVersionResource) error {
	return errors.New("Create resource error")
}

// ValidResourceManager is a mock type for the ResourceManager type.
type ValidResourceManager struct{}

// GetResource provides a mock function wih given fields: resourceName, resourceSchema.
func (c *ValidResourceManager) GetResource(resourceName string, resourceSchema schema.GroupVersionResource) (*unstructured.Unstructured, error) {
	return nil, nil
}

// CreateResource provides a mock function wih given fields: resource, resourceSchema.
func (c *ValidResourceManager) CreateResource(resource *unstructured.Unstructured, resourceSchema schema.GroupVersionResource) error {
	return nil
}

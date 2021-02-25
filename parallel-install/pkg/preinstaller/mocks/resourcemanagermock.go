package mocks

import (
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// GetErrorResourceManager is a mock type for the ResourceManager type.
type GetErrorResourceManager struct{}

// GetResource provides a mock function wih given fields: resourceName, resourceSchema.
func (c *GetErrorResourceManager) GetResource(resourceName string, resourceSchema schema.GroupVersionResource) (*unstructured.Unstructured, error) {
	return nil, errors.New("Get resource error")
}

// CreateResource provides a mock function wih given fields: resource, resourceSchema.
func (c *GetErrorResourceManager) CreateResource(resource *unstructured.Unstructured, resourceSchema schema.GroupVersionResource) error {
	return errors.New("Create resource error")
}

// UpdateRefreshableResource provides a mock function wih given fields: resource, resourceSchema.
func (c *GetErrorResourceManager) UpdateRefreshableResource(resource *unstructured.Unstructured, resourceSchema schema.GroupVersionResource) error {
	return errors.New("Update resource error")
}

// UpdateErrorResourceManager is a mock type for the ResourceManager type.
type UpdateErrorResourceManager struct{}

// GetResource provides a mock function wih given fields: resourceName, resourceSchema.
func (c *UpdateErrorResourceManager) GetResource(resourceName string, resourceSchema schema.GroupVersionResource) (*unstructured.Unstructured, error) {
	return &unstructured.Unstructured{}, nil
}

// CreateResource provides a mock function wih given fields: resource, resourceSchema.
func (c *UpdateErrorResourceManager) CreateResource(resource *unstructured.Unstructured, resourceSchema schema.GroupVersionResource) error {
	return nil
}

// UpdateRefreshableResource provides a mock function wih given fields: resource, resourceSchema.
func (c *UpdateErrorResourceManager) UpdateRefreshableResource(resource *unstructured.Unstructured, resourceSchema schema.GroupVersionResource) error {
	return errors.New("Update resource error")
}

// CreateErrorResourceManager is a mock type for the ResourceManager type.
type CreateErrorResourceManager struct{}

// GetResource provides a mock function wih given fields: resourceName, resourceSchema.
func (c *CreateErrorResourceManager) GetResource(resourceName string, resourceSchema schema.GroupVersionResource) (*unstructured.Unstructured, error) {
	return nil, nil
}

// CreateResource provides a mock function wih given fields: resource, resourceSchema.
func (c *CreateErrorResourceManager) CreateResource(resource *unstructured.Unstructured, resourceSchema schema.GroupVersionResource) error {
	return errors.New("Create resource error")
}

// UpdateRefreshableResource provides a mock function wih given fields: resource, resourceSchema.
func (c *CreateErrorResourceManager) UpdateRefreshableResource(resource *unstructured.Unstructured, resourceSchema schema.GroupVersionResource) error {
	return nil
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

// UpdateRefreshableResource provides a mock function wih given fields: resource, resourceSchema.
func (c *ValidResourceManager) UpdateRefreshableResource(resource *unstructured.Unstructured, resourceSchema schema.GroupVersionResource) error {
	return nil
}

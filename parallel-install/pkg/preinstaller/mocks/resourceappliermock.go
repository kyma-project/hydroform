package mocks

import (
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// AllowResourceApplierMock is a mock type for the ResourceApplier type.
type AllowResourceApplierMock struct{}

// Apply provides a mock function wih given fields: resource.
func (c *AllowResourceApplierMock) Apply(resource *unstructured.Unstructured) error {
	return nil
}

// MixedResourceApplierMock is a mock type for the ResourceApplier type.
type MixedResourceApplierMock struct{}

// Apply provides a mock function wih given fields: resource.
func (c *MixedResourceApplierMock) Apply(resource *unstructured.Unstructured) error {
	if resource == nil {
		return errors.New("Applier error")
	}

	return nil
}

// DenyResourceApplierMock is a mock type for the ResourceApplier type.
type DenyResourceApplierMock struct{}

// Apply provides a mock function wih given fields: resource.
func (c *DenyResourceApplierMock) Apply(resource *unstructured.Unstructured) error {
	return errors.New("Applier error")
}

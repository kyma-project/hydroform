package mocks

import "github.com/pkg/errors"

// AllowResourceApplierMock is a mock type for the ResourceApplier type.
type AllowResourceApplierMock struct{}

// Apply provides a mock function wih given fields: manifest.
func (c *AllowResourceApplierMock) Apply(manifest string) (bool, error) {
	return true, nil
}

// DenyResourceApplierMock is a mock type for the ResourceApplier type.
type DenyResourceApplierMock struct{}

// Apply provides a mock function wih given fields: manifest.
func (c *DenyResourceApplierMock) Apply(manifest string) (bool, error) {
	return false, errors.New("Applier error")
}

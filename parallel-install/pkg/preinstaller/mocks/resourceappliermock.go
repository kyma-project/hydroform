package mocks

import (
	"github.com/pkg/errors"
	"regexp"
)

// AllowResourceApplierMock is a mock type for the ResourceApplier type.
type AllowResourceApplierMock struct{}

// Apply provides a mock function wih given fields: manifest.
func (c *AllowResourceApplierMock) Apply(path string) error {
	return nil
}

// MixedResourceApplierMock is a mock type for the ResourceApplier type.
type MixedResourceApplierMock struct{}

// Apply provides a mock function wih given fields: manifest.
func (c *MixedResourceApplierMock) Apply(path string) error {
	matched, err := regexp.MatchString("incorrect", path)
	applierError := errors.New("Applier error")
	if err != nil || matched {
		return applierError
	}

	return nil
}

// DenyResourceApplierMock is a mock type for the ResourceApplier type.
type DenyResourceApplierMock struct{}

// Apply provides a mock function wih given fields: manifest.
func (c *DenyResourceApplierMock) Apply(path string) error {
	return errors.New("Applier error")
}

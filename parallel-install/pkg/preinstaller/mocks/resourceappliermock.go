package mocks

import "github.com/pkg/errors"

type AllowResourceApplierMock struct{}

func (c *AllowResourceApplierMock) Apply(manifest string) error {
	return nil
}

type DenyResourceApplierMock struct{}

func (c *DenyResourceApplierMock) Apply(manifest string) error {
	return errors.New("Applier error")
}

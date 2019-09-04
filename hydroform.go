package hydroform

import (
	gcp "github.com/kyma-incubator/hydroform/internal/gcp"
)

type Provider interface {
	Provision() error
	Status() error
	Credentials() error
	Deprovision() error
}

func NewGoogleProvider() Provider {
	return gcp.New()
}

func NewAWSProvider() Provider {
	return nil
}

func NewAzureProvider() Provider {
	return nil
}

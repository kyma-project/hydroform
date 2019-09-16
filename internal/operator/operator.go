package operator

import "github.com/kyma-incubator/hydroform/types"

type Operator interface {
	Create(providerType types.ProviderType, configuration map[string]interface{}) error
	Delete(providerType types.ProviderType, configuration map[string]interface{}) error
}

func NewTerraform() Operator {
	return &Terraform{}
}

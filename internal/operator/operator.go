package operator

import "github.com/kyma-incubator/hydroform/types"

//go:generate mockery -name=Operator -case=snake

// Operator allows switching easily between different types of provisioning operators.
type Operator interface {
	Create(providerType types.ProviderType, configuration map[string]interface{}) (*types.ClusterInfo, error)
	Delete(state *types.InternalState, providerType types.ProviderType, configuration map[string]interface{}) error
	Status(providerType types.ProviderType, configuration map[string]interface{}) (*types.ClusterInfo, error)
}

// Type points out the type of the operator.
type Type string

const (
	// TerraformOperator indicates the type of the operator is Terraform.
	TerraformOperator Type = "terraform"
)

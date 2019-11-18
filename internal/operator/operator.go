package operator

import "github.com/kyma-incubator/hydroform/types"

//go:generate mockery -name=Operator -case=snake

// Operator allows switching easily between different types of provisioning operators.
type Operator interface {
	Create(p types.ProviderType, cfg map[string]interface{}) (*types.ClusterInfo, error)
	Status(p types.ProviderType, cfg map[string]interface{}) (*types.ClusterStatus, error)
	Delete(p types.ProviderType, cfg map[string]interface{}) error
}

// Type points out the type of the operator.
type Type string

const (
	// TerraformOperator indicates the type of the operator is Terraform.
	TerraformOperator Type = "terraform"
)

package operator

import (
	"github.com/hashicorp/terraform/states/statefile"
	"github.com/kyma-incubator/hydroform/types"
)

//go:generate mockery -name=Operator -case=snake

// Operator allows switching easily between different types of provisioning operators.
type Operator interface {
	// Create creates a new cluster on the given provider based on the configuration and returns the same cluster enriched with its current state.
	Create(p types.ProviderType, cfg map[string]interface{}) (*types.ClusterInfo, error)
	// Status checks the cluster status based on the given state.
	// If the state is empty or nil, Status will attempt to load the state from the file system.
	Status(state *statefile.File, p types.ProviderType, cfg map[string]interface{}) (*types.ClusterStatus, error)
	// Delete removes a cluster. For this operation a valid state is necessary.
	// If the state is empty or nil, Delete will attempt to load the state from the file system.
	Delete(state *statefile.File, p types.ProviderType, cfg map[string]interface{}) error
}

// Type points out the type of the operator.
type Type string

const (
	// TerraformOperator indicates the type of the operator is Terraform.
	TerraformOperator Type = "terraform"
)

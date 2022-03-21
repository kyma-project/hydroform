package native

import (
	"github.com/kyma-incubator/hydroform/provision/internal/operator/native/gardener"
	"github.com/kyma-incubator/hydroform/provision/types"
	"github.com/pkg/errors"
)

// Operator implements the native operator for provisioning clusters.
// A native operator communicates directly with the Cloud Provider in their native API.
type Operator struct {
	ops *types.Options
}

// New creates a new native operator with the given options
func New(ops *types.Options) *Operator {
	return &Operator{
		ops: ops,
	}
}

// Create creates a new cluster on the given provider based on the configuration and returns the same cluster enriched with its current state.
func (o *Operator) Create(p types.ProviderType, cfg map[string]interface{}) (*types.ClusterInfo, error) {
	switch p {
	case types.Gardener:
		return gardener.Create(o.ops, cfg)
	default:
		return nil, errors.Errorf("Provider %s is not supported by the native operator", p)
	}
}

// Status checks the cluster status based on the given state.
// If the state is empty or nil, Status will attempt to load the state from the file system.
func (o *Operator) Status(info *types.ClusterInfo, p types.ProviderType, cfg map[string]interface{}) (*types.ClusterStatus, error) {
	switch p {
	case types.Gardener:
		return gardener.Status(o.ops, info, cfg)
	default:
		return nil, errors.Errorf("Provider %s is not supported by the native operator", p)
	}
}

// Delete removes a cluster. For this operation a valid state is necessary.
// If the state is empty or nil, Delete will attempt to load the state from the file system.
func (o *Operator) Delete(info *types.ClusterInfo, p types.ProviderType, cfg map[string]interface{}) error {
	switch p {
	case types.Gardener:
		return gardener.Delete(o.ops, info, cfg)
	default:
		return errors.Errorf("Provider %s is not supported by the native operator", p)
	}
}

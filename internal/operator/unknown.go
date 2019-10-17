package operator

import (
	"errors"

	"github.com/kyma-incubator/hydroform/types"
)

// Unknown represents an invalid operator.
type Unknown struct {
}

// Create returns an error if the operator is unknown.
func (u *Unknown) Create(providerType types.ProviderType, configuration map[string]interface{}) (*types.ClusterInfo, error) {
	return nil, errors.New("unknown operator")
}

// Delete returns an error if the operator is unknown.
func (u *Unknown) Delete(state *types.InternalState, providerType types.ProviderType, configuration map[string]interface{}) error {
	return errors.New("unknown operator")
}

func (t *Unknown) Status(state *types.InternalState, configuration map[string]interface{}) (*types.ClusterStatus, error) {
	return nil, errors.New("unknown operator")
}

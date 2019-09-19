package operator

import (
	"errors"

	"github.com/kyma-incubator/hydroform/types"
)

// Unknown implements Operator
type Unknown struct {
}

// Create returns error in case the operator is unknown.
func (u *Unknown) Create(providerType types.ProviderType, configuration map[string]interface{}) (*types.ClusterInfo, error) {
	return nil, errors.New("unknown operator")
}

// Delete returns error in case the operator is unknown.
func (u *Unknown) Delete(state *types.InternalState, providerType types.ProviderType, configuration map[string]interface{}) error {
	return errors.New("unknown operator")
}

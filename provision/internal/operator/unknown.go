package operator

import (
	"errors"

	"github.com/kyma-incubator/hydroform/provision/types"
)

// Unknown represents an invalid operator.
type Unknown struct {
}

// Create returns an error if the operator is unknown.
func (u *Unknown) Create(p types.ProviderType, cfg map[string]interface{}) (*types.ClusterInfo, error) {
	return nil, errors.New("unknown operator")
}

func (u *Unknown) Status(info *types.ClusterInfo, p types.ProviderType, cfg map[string]interface{}) (*types.ClusterStatus, error) {
	return nil, errors.New("unknown operator")
}

// Delete returns an error if the operator is unknown.
func (u *Unknown) Delete(info *types.ClusterInfo, p types.ProviderType, cfg map[string]interface{}) error {
	return errors.New("unknown operator")
}

package types

import (
	"context"
	. "github.com/kyma-incubator/hydroform/install/util"

	. "k8s.io/api/core/v1"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SecretClient interface {
	Update(ctx context.Context, configMap *Secret, opts v1.UpdateOptions) (*Secret, error)
	Get(ctx context.Context, name string, opts v1.GetOptions) (*Secret, error)
}

type SecretOverride struct {
	NewItem *Secret
	Client  SecretClient
}

func (s SecretOverride) Labels() *map[string]string {
	return &s.NewItem.Labels
}

func (s SecretOverride) Update() error {
	_, err := s.Client.Update(context.Background(), s.NewItem, v1.UpdateOptions{})
	return err
}

func (s SecretOverride) Merge() error {
	old, err := s.Client.Get(context.Background(), s.NewItem.Name, v1.GetOptions{})
	if err == nil {
		MergeByteMaps(old.Data, s.NewItem.Data)
	}

	return err
}

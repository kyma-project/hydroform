package types

import (
	"context"
	"github.com/kyma-incubator/hydroform/install/util"

	. "k8s.io/api/core/v1"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ConfigMapClient interface {
	Update(ctx context.Context, configMap *ConfigMap, opts v1.UpdateOptions) (*ConfigMap, error)
	Get(ctx context.Context, name string, opts v1.GetOptions) (*ConfigMap, error)
}

type ConfigMapOverride struct {
	NewItem *ConfigMap
	Client  ConfigMapClient
}

func (s ConfigMapOverride) Labels() *map[string]string {
	return &s.NewItem.Labels
}

func (s ConfigMapOverride) Update() error {
	_, err := s.Client.Update(context.Background(), s.NewItem, v1.UpdateOptions{})
	return err
}

func (s ConfigMapOverride) Merge() error {
	old, err := s.Client.Get(context.Background(), s.NewItem.Name, v1.GetOptions{})
	if err == nil {
		util.MergeStringMaps(old.Data, s.NewItem.Data)
	}

	return err
}

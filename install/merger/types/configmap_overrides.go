package types

import (
	"context"
	"errors"
	"github.com/kyma-incubator/hydroform/install/merger"
	. "github.com/kyma-incubator/hydroform/install/util"

	. "k8s.io/api/core/v1"
	. "k8s.io/client-go/kubernetes/typed/core/v1"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ConfigMapOverride struct {
	NewItem *ConfigMap
	Client  ConfigMapInterface
}

func (s ConfigMapOverride) Name() string {
	return s.NewItem.Name
}

func (s ConfigMapOverride) Labels() map[string]string {
	return s.NewItem.Labels
}

func (s ConfigMapOverride) LoadOld() (merger.Data, error) {
	old, err := s.Client.Get(context.Background(), s.NewItem.Name, v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return s.New(old)
}

func (s ConfigMapOverride) New(item *ConfigMap) (merger.Data, error) {
	return ConfigMapOverride{
		item,
		s.Client,
	}, nil
}

func (s ConfigMapOverride) Update() error {
	_, err := s.Client.Update(context.Background(), s.NewItem, v1.UpdateOptions{})
	return err
}

func (s ConfigMapOverride) Merge(old merger.Data) error {
	if casted, ok := old.(ConfigMapOverride); ok {
		MergeStringMaps(casted.NewItem.Data, s.NewItem.Data)
		return nil
	}

	return errors.New("casting error")
}

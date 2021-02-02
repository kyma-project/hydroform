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

type SecretOverride struct {
	NewItem *Secret
	Client  SecretInterface
}

func (s SecretOverride) Name() string {
	return s.NewItem.Name
}

func (s SecretOverride) Labels() map[string]string {
	return s.NewItem.Labels
}

func (s SecretOverride) LoadOld() (merger.Data, error) {
	item, err := s.Client.Get(context.Background(), s.NewItem.Name, v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return SecretOverride{
		item,
		s.Client,
	}, nil
}

func (s SecretOverride) Update() error {
	_, err := s.Client.Update(context.Background(), s.NewItem, v1.UpdateOptions{})
	return err
}

func (s SecretOverride) Merge(old merger.Data) error {
	if casted, ok := old.(SecretOverride); ok {
		MergeByteMaps(casted.NewItem.Data, s.NewItem.Data)
		return nil
	}

	return errors.New("casting error")
}

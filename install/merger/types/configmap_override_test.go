package types

import (
	"context"
	"errors"
	"github.com/kyma-incubator/hydroform/install/merger/types/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	v1 "k8s.io/api/core/v1"
	. "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

const (
	name = "item name"
)

func TestConfigMaps(t *testing.T) {

	t.Run("should merge string maps without error", func(t *testing.T) {
		// given
		client := &mocks.ConfigMapClient{}
		old := oldConfig()
		client.On("Get", context.Background(), name, GetOptions{}).Return(old, nil)
		override := newConfig(client)

		// when
		_ = override.Merge()

		// then
		assert.Contains(t, old.Data, "anotherKey")
		assert.Contains(t, old.Data, "key")
		mock.AssertExpectationsForObjects(t, client)
	})

	t.Run("should ignore merge if error", func(t *testing.T) {
		// given
		client := &mocks.ConfigMapClient{}
		old := oldConfig()
		client.On("Get", context.Background(), name, GetOptions{}).
			Return(nil, errors.New(""))
		override := newConfig(client)

		// when
		err := override.Merge()

		// then
		assert.NotNil(t, err)
		assert.NotContains(t, old.Data, "anotherKey")
		assert.Contains(t, old.Data, "key")
		mock.AssertExpectationsForObjects(t, client)
	})

	t.Run("should delegate to client update", func(t *testing.T) {
		// given
		client := &mocks.ConfigMapClient{}
		override := newConfig(client)
		item := override.NewItem
		client.On("Update", context.Background(), item, UpdateOptions{}).Return(item, nil)

		// when
		_ = override.Update()

		// then
		mock.AssertExpectationsForObjects(t, client)
	})

	t.Run("should return config maps labels if queried", func(t *testing.T) {
		// given
		client := &mocks.ConfigMapClient{}

		// when
		override := newConfig(client)

		// then
		assert.Contains(t, *override.Labels(), "anotherKey")
	})
}

func oldConfig() *v1.ConfigMap {
	return &v1.ConfigMap{
		Data: map[string]string{
			"key": "value",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

func newConfig(client *mocks.ConfigMapClient) *ConfigMapOverride {
	return &ConfigMapOverride{
		NewItem: &v1.ConfigMap{
			Data: map[string]string{
				"anotherKey": "anotherValue",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
				Labels: map[string]string{
					"anotherKey": "anotherValue",
				},
			},
		},
		Client: client,
	}
}

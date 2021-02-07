package types

import (
	"context"
	"github.com/kyma-incubator/hydroform/install/merger/types/mocks"
	"github.com/stretchr/testify/assert"
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
		client := &mocks.ConfigMapClient{}

		old := oldConfig()

		client.On("Get", context.Background(), name, GetOptions{}).Return(old, nil)

		override := newConfig(client)
		_ = override.Merge()
		assert.Contains(t, old.Data, "anotherKey")
		assert.Contains(t, old.Data, "key")
	})

	t.Run("should ignore merge if error", func(t *testing.T) {
		client := &mocks.ConfigMapClient{}

		old := oldConfig()

		client.On("Get", context.Background(), name, GetOptions{}).Return(nil, MockError{})

		override := newConfig(client)
		err := override.Merge()
		assert.NotNil(t, err)
		assert.NotContains(t, old.Data, "anotherKey")
		assert.Contains(t, old.Data, "key")
	})

	t.Run("should delegate to client update", func(t *testing.T) {
		client := &mocks.ConfigMapClient{}

		override := newConfig(client)

		item := override.NewItem
		client.On("Update", context.Background(), item, UpdateOptions{}).Return(item, nil)

		_ = override.Update()
	})

	t.Run("should return config maps labels if queried", func(t *testing.T) {
		client := &mocks.ConfigMapClient{}

		override := newConfig(client)

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

type MockError struct {
}

func (MockError) Error() string {
	return "Mock Error"
}

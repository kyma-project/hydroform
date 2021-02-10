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

func TestSecrets(t *testing.T) {

	t.Run("should merge string maps without error", func(t *testing.T) {
		// given
		client := &mocks.SecretClient{}
		old := oldSecret()
		client.On("Get", context.Background(), name, GetOptions{}).Return(old, nil)
		override := newSecret(client)

		// when
		_ = override.Merge()

		// then
		assert.Contains(t, old.Data, "anotherKey")
		assert.Contains(t, old.Data, "key")
		mock.AssertExpectationsForObjects(t, client)
	})

	t.Run("should ignore merge if error", func(t *testing.T) {
		// given
		client := &mocks.SecretClient{}
		old := oldSecret()
		client.On("Get", context.Background(), name, GetOptions{}).
			Return(nil, errors.New(""))
		override := newSecret(client)

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
		client := &mocks.SecretClient{}
		override := newSecret(client)
		item := override.NewItem
		client.On("Update", context.Background(), item, UpdateOptions{}).Return(item, nil)

		// when
		_ = override.Update()

		// then
		mock.AssertExpectationsForObjects(t, client)
	})

	t.Run("should return config maps labels if queried", func(t *testing.T) {
		// given
		client := &mocks.SecretClient{}

		// when
		override := newSecret(client)

		// then
		assert.Contains(t, *override.Labels(), "anotherKey")
	})
}

func oldSecret() *v1.Secret {
	return &v1.Secret{
		Data: map[string][]byte{
			"key": []byte("value"),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

func newSecret(client *mocks.SecretClient) *SecretOverride {
	return &SecretOverride{
		NewItem: &v1.Secret{
			Data: map[string][]byte{
				"anotherKey": []byte("anotherValue"),
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

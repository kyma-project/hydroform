package k8s

import (
	"testing"

	"github.com/stretchr/testify/require"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

const (
	namespace = "test"
)

func TestGenericClient_ApplyConfigMaps(t *testing.T) {

	t.Run("should apply config maps", func(t *testing.T) {
		// given
		existingCMs := []runtime.Object{
			&v1.ConfigMap{
				ObjectMeta: v12.ObjectMeta{Name: "test1", Namespace: namespace},
				Data:       map[string]string{"key1": "value1"},
			},
		}

		cmsToApply := []*v1.ConfigMap{
			{
				ObjectMeta: v12.ObjectMeta{Name: "test1", Namespace: namespace},
				Data:       map[string]string{"key1": "value1", "key2": "value2"},
			},
			{
				ObjectMeta: v12.ObjectMeta{Name: "test2", Namespace: namespace},
				Data:       map[string]string{"key1": "value1"},
			},
		}

		k8sClientSet := fake.NewSimpleClientset(existingCMs...)

		client := NewGenericClient(nil, nil, k8sClientSet, nil)

		// when
		err := client.ApplyConfigMaps(cmsToApply, namespace)

		// then
		require.NoError(t, err)

		cmClient := k8sClientSet.CoreV1().ConfigMaps(namespace)
		cm, err := cmClient.Get("test1", v12.GetOptions{})
		require.NoError(t, err)
		assert.Equal(t, cmsToApply[0].Data, cm.Data)
		cm2, err := cmClient.Get("test2", v12.GetOptions{})
		require.NoError(t, err)
		assert.Equal(t, cmsToApply[1].Data, cm2.Data)
	})
}

func TestGenericClient_ApplySecrets(t *testing.T) {

	t.Run("should apply config maps", func(t *testing.T) {
		// given
		existingSecretss := []runtime.Object{
			&v1.Secret{
				ObjectMeta: v12.ObjectMeta{Name: "test1", Namespace: namespace},
				Data:       map[string][]byte{"key1": []byte("value1")},
			},
		}

		secretsToApply := []*v1.Secret{
			{
				ObjectMeta: v12.ObjectMeta{Name: "test1", Namespace: namespace},
				Data:       map[string][]byte{"key1": []byte("value1"), "key2": []byte("value2")},
			},
			{
				ObjectMeta: v12.ObjectMeta{Name: "test2", Namespace: namespace},
				Data:       map[string][]byte{"key1": []byte("value1")},
			},
		}

		k8sClientSet := fake.NewSimpleClientset(existingSecretss...)

		client := NewGenericClient(nil, nil, k8sClientSet, nil)

		// when
		err := client.ApplySecrets(secretsToApply, namespace)

		// then
		require.NoError(t, err)

		secretClient := k8sClientSet.CoreV1().Secrets(namespace)
		secret, err := secretClient.Get("test1", v12.GetOptions{})
		require.NoError(t, err)
		assert.Equal(t, secretsToApply[0].Data, secret.Data)
		secret2, err := secretClient.Get("test2", v12.GetOptions{})
		require.NoError(t, err)
		assert.Equal(t, secretsToApply[1].Data, secret2.Data)
	})
}

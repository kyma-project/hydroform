package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_KubeConfigManager_New(t *testing.T) {

	t.Run("should not create a new instance of KubeConfigManager", func(t *testing.T) {

		t.Run("when path and content are empty", func(t *testing.T) {
			// given
			kubeconfigSource := KubeconfigSource{
				Path:    "",
				Content: "",
			}

			// when
			manager, err := NewKubeConfigManager(kubeconfigSource)

			// then
			assert.Nil(t, manager)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "property has to be set")
		})

	})

	t.Run("should create a new instance of KubeConfigManager", func(t *testing.T) {

		t.Run("when path exists and content does not exist", func(t *testing.T) {
			// given
			kubeconfigSource := KubeconfigSource{
				Path:    "path",
				Content: "",
			}

			// when
			manager, err := NewKubeConfigManager(kubeconfigSource)

			// then
			assert.NotNil(t, manager)
			assert.NoError(t, err)
			assert.Equal(t, manager.path, kubeconfigSource.Path)
		})

		t.Run("when path does not exist and content exists", func(t *testing.T) {
			// given
			tempDir := os.TempDir()
			kubeconfigSource := KubeconfigSource{
				Path:    "",
				Content: "content",
			}

			// when
			manager, err := NewKubeConfigManager(kubeconfigSource)

			// then
			assert.NotNil(t, manager)
			assert.NoError(t, err)
			assert.Contains(t, manager.path, tempDir)
			assert.Contains(t, manager.path, "kubeconfig")
			assert.Contains(t, manager.path, ".yaml")
		})

		t.Run("when path and content exists", func(t *testing.T) {
			// given
			kubeconfigSource := KubeconfigSource{
				Path:    "path",
				Content: "content",
			}

			// when
			manager, err := NewKubeConfigManager(kubeconfigSource)

			// then
			assert.NotNil(t, manager)
			assert.NoError(t, err)
			assert.Equal(t, manager.path, kubeconfigSource.Path)
			assert.Empty(t, manager.content)
		})

	})

}

func getKubeConfig() string {
	return `apiVersion: v1
kind: Config
clusters:
  - name: test
    cluster:
      server: 'https://test.example.com'
      certificate-authority-data: >-
        somerandomcert
contexts:
  - name: test
    context:
      cluster: test
      user: test-token
current-context: test
users:
  - name: test-token
    user:
      token: >-
        somerandomtoken
`
}

package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_KubeConfigManager_New(t *testing.T) {

	t.Run("should not create a new instance of KubeConfigManager", func(t *testing.T) {

		t.Run("when path and content do not exist", func(t *testing.T) {
			// when
			manager, err := NewKubeConfigManager(nil, nil)

			// then
			assert.Nil(t, manager)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "property has to be set")
		})

		t.Run("when path and content are empty", func(t *testing.T) {
			// given
			path := ""
			content := ""

			// when
			manager, err := NewKubeConfigManager(&path, &content)

			// then
			assert.Nil(t, manager)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "property has to be set")
		})

	})

	t.Run("should create a new instance of KubeConfigManager", func(t *testing.T) {

		t.Run("when path exists and content does not exist", func(t *testing.T) {
			// given
			path := "path"
			content := ""

			// when
			manager, err := NewKubeConfigManager(&path, &content)

			// then
			assert.NotNil(t, manager)
			assert.NoError(t, err)
			assert.Equal(t, manager.path, path)
		})

		t.Run("when path does not exist and content exists", func(t *testing.T) {
			// given
			path := ""
			content := "content"

			// when
			manager, err := NewKubeConfigManager(&path, &content)

			// then
			assert.NotNil(t, manager)
			assert.NoError(t, err)
			assert.Equal(t, manager.content, content)
		})

		t.Run("When path and content exists", func(t *testing.T) {
			// given
			path := "path"
			content := "content"

			// when
			manager, err := NewKubeConfigManager(&path, &content)

			// then
			assert.NotNil(t, manager)
			assert.NoError(t, err)
			assert.Equal(t, manager.path, path)
			assert.Empty(t, manager.content)
		})

	})

}

func Test_KubeConfigManager_Path(t *testing.T) {

	t.Run("should return a path to kubeconfig file", func(t *testing.T) {

		t.Run("when path exists and content does not exist", func(t *testing.T) {
			// given
			path := "path"
			content := ""

			// when
			manager, err := NewKubeConfigManager(&path, &content)

			// then
			assert.NotNil(t, manager)
			assert.NoError(t, err)
			assert.Equal(t, manager.Path(), path)
		})

		t.Run("when path does not exist and content exists", func(t *testing.T) {
			// given
			path := ""
			content := getKubeConfig()
			tempDir := os.TempDir()
			manager, err := NewKubeConfigManager(&path, &content)

			// when
			returnedPath := manager.Path()

			// then
			assert.NotNil(t, returnedPath)
			assert.NoError(t, err)
			assert.Contains(t, returnedPath, tempDir)
			assert.Contains(t, returnedPath, "kubeconfig")
			assert.Contains(t, returnedPath, ".yaml")
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

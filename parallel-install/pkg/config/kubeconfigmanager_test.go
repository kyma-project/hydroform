package config

import (
	"os"
	"testing"

	"errors"

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
			assert.Contains(t, err.Error(), "property must be set")
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
			assert.Equal(t, getPath(t, manager), kubeconfigSource.Path)
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

			path := getPath(t, manager)
			path2 := getPath(t, manager)
			path3 := getPath(t, manager)

			// all paths are the same (only single file is created)
			assert.Equal(t, path, path2)
			assert.Equal(t, path, path3)

			assert.NotNil(t, manager)
			assert.NoError(t, err)
			assert.Contains(t, path, tempDir)
			assert.Contains(t, path, "kubeconfig")
			assert.Contains(t, path, ".yaml")
		})

		t.Run("when path and content exists", func(t *testing.T) {
			// given

			kubeconfigSource := KubeconfigSource{
				Path:    "/a/path/to/a/file.yaml",
				Content: "content",
			}

			// when
			manager, err := NewKubeConfigManager(kubeconfigSource)

			// then
			assert.NotNil(t, manager)
			assert.NoError(t, err)
			assert.Equal(t, getPath(t, manager), kubeconfigSource.Path)
		})

	})

	t.Run("should properly handle a temporary kubeconfig file", func(t *testing.T) {

		t.Run("when Path().cleanup is invoked multiple times", func(t *testing.T) {
			// given
			kubeconfigSource := KubeconfigSource{
				Content: "abcd",
			}
			manager, err := NewKubeConfigManager(kubeconfigSource)
			assert.Nil(t, err)

			// when
			path, cleanup, err := manager.Path()
			assert.Nil(t, err)

			// then the file should exist
			_, err = os.Stat(path)
			assert.Nil(t, err) //File exists

			// then first cleanup invocation removes the file
			err = cleanup()
			assert.Nil(t, err)
			_, err = os.Stat(path)
			assert.True(t, errors.Is(err, os.ErrNotExist))

			// then subsequent cleanup invocation does nothing
			err = cleanup()
			assert.Nil(t, err)
			_, err = os.Stat(path)
			assert.True(t, errors.Is(err, os.ErrNotExist))

			err = cleanup()
			assert.Nil(t, err)
			_, err = os.Stat(path)
			assert.True(t, errors.Is(err, os.ErrNotExist))
		})

		t.Run("when Path() is invoked multiple times", func(t *testing.T) {
			// given
			kubeconfigSource := KubeconfigSource{
				Content: "abcd",
			}
			manager, err := NewKubeConfigManager(kubeconfigSource)
			assert.Nil(t, err)

			// when invoked multiple times
			path1, cleanup1, err1 := manager.Path()
			assert.Nil(t, err1)

			path2, cleanup2, err2 := manager.Path()
			assert.Nil(t, err2)

			path3, cleanup3, err3 := manager.Path()
			assert.Nil(t, err3)

			// then paths should be the same
			assert.Equal(t, path1, path2)
			assert.Equal(t, path1, path3)

			// then the file should exist
			_, err = os.Stat(path1)
			assert.Nil(t, err) //File exists

			// then first cleanup invocation removes the file
			err = cleanup1()
			assert.Nil(t, err)
			_, err = os.Stat(path1)
			assert.True(t, errors.Is(err, os.ErrNotExist))

			// then subsequent cleanup invocation does nothing
			err = cleanup2()
			assert.Nil(t, err)
			_, err = os.Stat(path1)
			assert.True(t, errors.Is(err, os.ErrNotExist))

			err = cleanup3()
			assert.Nil(t, err)
			_, err = os.Stat(path1)
			assert.True(t, errors.Is(err, os.ErrNotExist))
		})
	})

}

func getPath(t *testing.T, manager *kubeConfigManager) string {
	path, cleanup, err := manager.Path()
	assert.NotNil(t, cleanup)
	assert.Nil(t, err)
	return path
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

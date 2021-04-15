package config

import (
	"os"
	"path"
	"testing"

	"errors"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/test"
	"github.com/stretchr/testify/assert"
)

func Test_RestConfig_Function(t *testing.T) {

	testKubeconfigFile := path.Join(test.GetTestDataDirectory(), "test-kubeconfig.yaml")

	t.Run("should return an error", func(t *testing.T) {

		t.Run("when path and content are empty", func(t *testing.T) {
			// given
			kubeconfigSource := KubeconfigSource{
				Path:    "",
				Content: "",
			}

			// when
			res, err := RestConfig(kubeconfigSource)

			// then
			assert.Nil(t, res)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "property must be set")
		})

		t.Run("when used with incorrect kubeconfig content", func(t *testing.T) {
			// given
			kubeconfigSource := KubeconfigSource{
				Path:    "",
				Content: incorrectKubeConfig(),
			}

			// when
			res, err := RestConfig(kubeconfigSource)

			// then
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "illegal base64 data")
			assert.Nil(t, res)
		})
	})

	t.Run("should succeed", func(t *testing.T) {

		t.Run("when path exists and content does not exist", func(t *testing.T) {
			// given
			kubeconfigSource := KubeconfigSource{
				Path:    testKubeconfigFile,
				Content: "",
			}

			// when
			res, err := RestConfig(kubeconfigSource)

			// then
			assert.NoError(t, err)
			assert.NotNil(t, res)
			assert.Equal(t, "https://from.file.example.com", res.Host)
		})

		t.Run("when content exists and path does not exist", func(t *testing.T) {
			// given
			kubeconfigSource := KubeconfigSource{
				Path:    "",
				Content: correctKubeConfig(),
			}

			// when
			res, err := RestConfig(kubeconfigSource)

			// then
			assert.NoError(t, err)
			assert.NotNil(t, res)
			assert.Equal(t, "https://from.content.example.com", res.Host)
		})

		t.Run("when content exists and path exists", func(t *testing.T) {
			// given
			kubeconfigSource := KubeconfigSource{
				Path:    testKubeconfigFile,
				Content: correctKubeConfig(),
			}

			// when
			res, err := RestConfig(kubeconfigSource)

			// then
			assert.NoError(t, err)
			assert.NotNil(t, res)
			assert.Equal(t, "https://from.file.example.com", res.Host)
		})
	})
}

func Test_Path_Function(t *testing.T) {

	testKubeconfigFile := path.Join(test.GetTestDataDirectory(), "test-kubeconfig.yaml")

	t.Run("when path and content are empty", func(t *testing.T) {
		t.Run("should return an error", func(t *testing.T) {
			// given
			kubeconfigSource := KubeconfigSource{
				Path:    "",
				Content: "",
			}

			// when
			res, cleanup, err := Path(kubeconfigSource)

			// then
			assert.Empty(t, res)
			assert.Nil(t, cleanup)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "property must be set")
		})

	})

	t.Run("when given a path to an existing file and no content", func(t *testing.T) {

		t.Run("returns a cleanup function that doesn't delete the file", func(t *testing.T) {

			// given
			kubeconfigSource := KubeconfigSource{
				Path:    testKubeconfigFile,
				Content: "",
			}

			// the file should exist
			_, err := os.Stat(testKubeconfigFile)
			assert.Nil(t, err) //File exists

			// when
			path, cleanup, err := Path(kubeconfigSource)

			// then
			assert.Nil(t, err)
			assert.NotNil(t, cleanup)
			assert.Equal(t, testKubeconfigFile, path)

			// then first cleanup invocation does not remove the file
			err = cleanup()
			assert.Nil(t, err)

			// then the file should exist
			_, err = os.Stat(path)
			assert.Nil(t, err) //File exists

			// then subsequent cleanup invocation does not remove the file
			err = cleanup()
			assert.Nil(t, err)

			// then the file should exist
			_, err = os.Stat(path)
			assert.Nil(t, err) //File exists
		})
	})

	t.Run("when given a content of a kubeconfig file and no path", func(t *testing.T) {

		t.Run("returns a path to a temporary file and a cleanup function that deletes the file", func(t *testing.T) {

			// given
			tmpDir := os.TempDir()

			kubeconfigSource := KubeconfigSource{
				Path:    "",
				Content: correctKubeConfig(),
			}

			// when
			path, cleanup, err := Path(kubeconfigSource)

			// then
			assert.Nil(t, err)
			assert.NotNil(t, cleanup)
			assert.Contains(t, path, tmpDir)
			assert.Contains(t, path, "kubeconfig")
			assert.Contains(t, path, ".yaml")

			// then the file should exist
			_, err = os.Stat(path)
			assert.Nil(t, err) //File exists

			// then first cleanup invocation does remove the file
			err = cleanup()
			assert.Nil(t, err)

			// then the file should not exist
			_, err = os.Stat(path)
			assert.True(t, errors.Is(err, os.ErrNotExist))

			// then subsequent cleanup invocation should not fails
			err = cleanup()
			assert.Nil(t, err)

			// then the file should not exist
			_, err = os.Stat(path)
			assert.True(t, errors.Is(err, os.ErrNotExist))
		})
	})

	t.Run("when given a path to an existing file and a kubeconfigcontent", func(t *testing.T) {

		t.Run("returns an existing path and a cleanup function that doesn't delete the file", func(t *testing.T) {

			// given
			kubeconfigSource := KubeconfigSource{
				Path:    testKubeconfigFile,
				Content: correctKubeConfig(),
			}

			// the file should exist
			_, err := os.Stat(testKubeconfigFile)
			assert.Nil(t, err) //File exists

			// when
			path, cleanup, err := Path(kubeconfigSource)

			// then
			assert.Nil(t, err)
			assert.NotNil(t, cleanup)
			assert.Equal(t, testKubeconfigFile, path)

			// then first cleanup invocation does not remove the file
			err = cleanup()
			assert.Nil(t, err)

			// then the file should exist
			_, err = os.Stat(path)
			assert.Nil(t, err) //File exists

			// then subsequent cleanup invocation does not remove the file
			err = cleanup()
			assert.Nil(t, err)

			// then the file should exist
			_, err = os.Stat(path)
			assert.Nil(t, err) //File exists
		})
	})
}

func correctKubeConfig() string {
	return `apiVersion: v1
kind: Config
clusters:
  - name: test
    cluster:
      server: 'https://from.content.example.com'
      certificate-authority-data: >-
        c29tZXJhbmRvbWNlcnQ=
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
        c29tZXJhbmRvbXRva2Vu
`
}

func incorrectKubeConfig() string {
	return `apiVersion: v1
kind: Config
clusters:
  - name: test
    cluster:
      server: 'https://from.content.example.com'
      certificate-authority-data: >-
        somerandomdata
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

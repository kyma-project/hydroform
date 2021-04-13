package config

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_ValidateDeletion(t *testing.T) {
	var config Config
	var err error

	t.Run("Check workers count", func(t *testing.T) {
		config = Config{
			WorkersCount: 0,
		}
		err = config.ValidateDeletion()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Workers count cannot be")
	})

	t.Run("Components file not found", func(t *testing.T) {
		_, err := NewComponentList("/a/file/which/doesnot/exist.json")
		require.Error(t, err)
	})

	t.Run("Happy path", func(t *testing.T) {
		fpath := filePath(t)
		config = Config{
			WorkersCount:   1,
			ComponentList:  newComponentList(t),
			KubeconfigPath: filepath.Dir(fpath),
		}
		err = config.ValidateDeletion()
		assert.NoError(t, err)
	})

	t.Run("KubeconfigPath path not found", func(t *testing.T) {
		config = Config{
			WorkersCount:   1,
			ComponentList:  newComponentList(t),
			KubeconfigPath: "/a/dir/which/doesnot/exist",
		}
		err = config.ValidateDeletion()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func Test_ValidateDeployment(t *testing.T) {
	var config Config
	t.Run("Resource path not found", func(t *testing.T) {
		config = Config{
			WorkersCount:  1,
			ComponentList: newComponentList(t),
			ResourcePath:  "/a/dir/which/doesnot/exist",
		}
		err := config.ValidateDeployment()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("InstallationResourcePath path not found", func(t *testing.T) {
		fpath := filePath(t)
		config = Config{
			WorkersCount:             1,
			ComponentList:            newComponentList(t),
			ResourcePath:             filepath.Dir(fpath),
			InstallationResourcePath: "/a/dir/which/doesnot/exist",
		}
		err := config.ValidateDeployment()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("KubeconfigPath path not found", func(t *testing.T) {
		fpath := filePath(t)
		config = Config{
			WorkersCount:             1,
			ComponentList:            newComponentList(t),
			ResourcePath:             filepath.Dir(fpath),
			InstallationResourcePath: filepath.Dir(fpath),
			KubeconfigPath:           "/a/dir/which/doesnot/exist",
		}
		err := config.ValidateDeployment()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("Version empty", func(t *testing.T) {
		fpath := filePath(t)
		config = Config{
			WorkersCount:             1,
			ComponentList:            newComponentList(t),
			ResourcePath:             filepath.Dir(fpath),
			InstallationResourcePath: filepath.Dir(fpath),
			KubeconfigPath:           filepath.Dir(fpath),
		}
		err := config.ValidateDeployment()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Version is empty")
	})

	t.Run("Happy path", func(t *testing.T) {
		fpath := filePath(t)
		config = Config{
			WorkersCount:             1,
			ComponentList:            newComponentList(t),
			ResourcePath:             filepath.Dir(fpath),
			InstallationResourcePath: filepath.Dir(fpath),
			KubeconfigPath:           filepath.Dir(fpath),
			Version:                  "abc",
		}
		err := config.ValidateDeployment()
		assert.NoError(t, err)
	})
}

func newComponentList(t *testing.T) *ComponentList {
	compList, err := NewComponentList("../test/data/componentlist.yaml")
	require.NoError(t, err)
	return compList
}

func filePath(t *testing.T) string {
	_, fpath, _, ok := runtime.Caller(0)
	assert.True(t, ok)
	return fpath
}

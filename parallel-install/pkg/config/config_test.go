package config

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ConfigValidation(t *testing.T) {
	var config Config
	var err error

	t.Run("Check workers count", func(t *testing.T) {
		config = Config{
			WorkersCount: 0,
		}
		err = config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Workers count cannot be")
	})

	t.Run("Components file not found", func(t *testing.T) {
		config = Config{
			WorkersCount:       1,
			ComponentsListFile: "/a/file/which/doesnot/exist.json",
		}
		err = config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("Resource path not found", func(t *testing.T) {
		_, fpath, _, ok := runtime.Caller(0)
		assert.True(t, ok)
		config = Config{
			WorkersCount:       1,
			ComponentsListFile: fpath,
			ResourcePath:       "/a/dir/which/doesnot/exist",
		}
		err = config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("Resource path not found", func(t *testing.T) {
		_, fpath, _, ok := runtime.Caller(0)
		assert.True(t, ok)
		config = Config{
			WorkersCount:       1,
			ComponentsListFile: fpath,
			ResourcePath:       filepath.Dir(fpath),
			CrdPath:            "/a/dir/which/doesnot/exist",
		}
		err = config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("Version empty", func(t *testing.T) {
		_, fpath, _, ok := runtime.Caller(0)
		assert.True(t, ok)
		config = Config{
			WorkersCount:       1,
			ComponentsListFile: fpath,
			ResourcePath:       filepath.Dir(fpath),
			CrdPath:            filepath.Dir(fpath),
		}
		err = config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Version is empty")
	})

	t.Run("Happy path", func(t *testing.T) {
		_, fpath, _, ok := runtime.Caller(0)
		assert.True(t, ok)
		config = Config{
			WorkersCount:       1,
			ComponentsListFile: fpath,
			ResourcePath:       filepath.Dir(fpath),
			CrdPath:            filepath.Dir(fpath),
			Version:            "abc",
		}
		err = config.Validate()
		assert.NoError(t, err)
	})
}

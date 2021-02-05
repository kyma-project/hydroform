package docker

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestInline(t *testing.T) {
	path, err := ioutil.TempDir(os.TempDir(), "test-")
	require.Equal(t, nil, err)
	defer os.RemoveAll(path)
	f1, err := os.Create(fmt.Sprintf("%s/%s", path, "test1"))
	require.Equal(t, nil, err)
	defer f1.Close()
	f2, err := os.Create(fmt.Sprintf("%s/%s", path, "test2"))
	require.Equal(t, nil, err)
	defer f2.Close()
	f3, err := os.Create(fmt.Sprintf("%s/%s", path, "test3"))
	require.Equal(t, nil, err)
	defer f3.Close()

	t.Run("should create new context and end without error", func(t *testing.T) {
		counter := 0
		got, err := Inline(ContextOpts{
			DirPrefix:  "test-context-",
			Dockerfile: "test-dockerfile",
			SrcDir:     path,
			SrcFiles:   []string{"test1", "test2", "test3"},
		}, func(_ string, _ ...interface{}) {
			counter++
		})

		require.Equal(t, nil, err)
		require.DirExists(t, got)
		defer os.RemoveAll(got)
		require.FileExists(t, fmt.Sprintf("%s/%s", got, dockerfileFilename))
		require.FileExists(t, fmt.Sprintf("%s/%s/%s", got, codeDir, "test1"))
		require.FileExists(t, fmt.Sprintf("%s/%s/%s", got, codeDir, "test2"))
		require.FileExists(t, fmt.Sprintf("%s/%s/%s", got, codeDir, "test3"))
		require.Equal(t, 5, counter)
	})

	t.Run("should return error while creating new tmp dir", func(t *testing.T) {
		counter := 0
		got, err := Inline(ContextOpts{
			DirPrefix:  "test-context/-",
			Dockerfile: "test-dockerfile",
			SrcDir:     path,
			SrcFiles:   []string{"test1", "test2", "test3"},
		}, func(_ string, _ ...interface{}) {
			counter++
		})

		require.Error(t, err)
		require.NoDirExists(t, got)
		require.NoFileExists(t, fmt.Sprintf("%s/%s", got, dockerfileFilename))
		require.NoFileExists(t, fmt.Sprintf("%s/%s/%s", got, codeDir, "test1"))
		require.NoFileExists(t, fmt.Sprintf("%s/%s/%s", got, codeDir, "test2"))
		require.NoFileExists(t, fmt.Sprintf("%s/%s/%s", got, codeDir, "test3"))
		require.Equal(t, 0, counter)
	})

	t.Run("should return error while creating new src dir", func(t *testing.T) {
		counter := 0
		got, err := Inline(ContextOpts{
			DirPrefix:  "test-context-",
			Dockerfile: "test-dockerfile",
			SrcDir:     path,
			SrcFiles:   []string{"test1", "test2", "test3"},
		}, func(_ string, path ...interface{}) {
			if counter == 0 {
				os.Remove(path[0].(string))
			}
			counter++
		})

		require.Error(t, err)
		require.NoDirExists(t, got)
		require.Equal(t, 1, counter)
	})

	t.Run("should return error while creating file in context", func(t *testing.T) {
		counter := 0
		got, err := Inline(ContextOpts{
			DirPrefix:  "test-context-",
			Dockerfile: "test-dockerfile",
			SrcDir:     path,
			SrcFiles:   []string{"test1"},
		}, func(_ string, path ...interface{}) {
			if counter == 1 {
				os.Remove(filepath.Dir(path[1].(string)))
			}
			counter++
		})

		require.Error(t, err)
		require.DirExists(t, got)
		require.NoDirExists(t, fmt.Sprintf("%s/%s", got, "src"))
		require.Equal(t, 2, counter)
	})

	t.Run("should return error while opening file in dir", func(t *testing.T) {
		f4, err := os.Create(fmt.Sprintf("%s/%s", path, "test4"))
		require.NoError(t, err)
		defer f4.Close()
		counter := 0
		got, err := Inline(ContextOpts{
			DirPrefix:  "test-context-",
			Dockerfile: "test-dockerfile",
			SrcDir:     path,
			SrcFiles:   []string{"test4"},
		}, func(_ string, _ ...interface{}) {
			if counter == 1 {
				os.Remove(fmt.Sprintf("%s/%s", path, "test4"))
			}
			counter++
		})

		require.Error(t, err)
		require.DirExists(t, got)
		require.DirExists(t, fmt.Sprintf("%s/%s", got, "src"))
		require.FileExists(t, fmt.Sprintf("%s/%s/%s", got, codeDir, "test4"))
		require.Equal(t, 2, counter)
	})

	t.Run("should return error while creating dockerfile", func(t *testing.T) {
		counter := 0
		got, err := Inline(ContextOpts{
			DirPrefix:  "test-context-",
			Dockerfile: "test-dockerfile",
			SrcDir:     path,
			SrcFiles:   []string{"test1"},
		}, func(_ string, path ...interface{}) {
			if counter == 2 {
				t.Log(path[0])
				os.RemoveAll(filepath.Dir(path[0].(string)))
			}
			counter++
		})

		require.Error(t, err)
		require.NoDirExists(t, got)
		require.Equal(t, 3, counter)
	})
}

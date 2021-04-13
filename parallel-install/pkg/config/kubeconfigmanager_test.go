package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
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
			assert.NotEmpty(t, manager.Path())
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
			assert.NotEmpty(t, manager.Path())
		})

	})

}

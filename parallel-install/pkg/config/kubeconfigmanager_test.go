package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_KubeConfigManager_New(t *testing.T) {

	t.Run("should create a new instance of KubeConfigManager", func(t *testing.T) {

		t.Run("when path and content exist", func(t *testing.T) {
			// given
			path := "aaa"
			content := "bbb"

			// when
			manager, err := NewKubeConfigManager(&path, &content)

			// then
			assert.NotNil(t, manager)
			assert.NoError(t, err)
		})

	})

}

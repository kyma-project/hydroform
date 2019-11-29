package installation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigEntries_Get(t *testing.T) {

	t.Run("should get config entry", func(t *testing.T) {
		expectedEntry := ConfigEntry{Key: "testKey2", Value: "testVal2", Secret: false}

		var configEntries ConfigEntries = []ConfigEntry{
			{Key: "testKey", Value: "testVal", Secret: false},
			expectedEntry,
			{Key: "testKey3", Value: "testVal3", Secret: true},
		}

		configEntry, found := configEntries.Get("testKey2")
		assert.True(t, found)
		assert.Equal(t, expectedEntry, configEntry)
	})

	t.Run("should return false if entry not found", func(t *testing.T) {
		var configEntries ConfigEntries = []ConfigEntry{
			{Key: "testKey", Value: "testVal", Secret: false},
			{Key: "testKey2", Value: "testVal2", Secret: true},
		}

		configEntry, found := configEntries.Get("non-existent")
		assert.False(t, found)
		assert.Empty(t, configEntry)
	})
}

func TestConfigEntries_Set(t *testing.T) {

	for _, testCase := range []struct {
		description     string
		configEntries   ConfigEntries
		setEntry        ConfigEntry
		expectedEntries ConfigEntries
	}{
		{
			description: "set config entry when it does not exist",
			configEntries: []ConfigEntry{
				{Key: "testKey", Value: "testVal", Secret: false},
			},
			setEntry: ConfigEntry{Key: "testKey2", Value: "testVal2", Secret: false},
			expectedEntries: []ConfigEntry{
				{Key: "testKey", Value: "testVal", Secret: false},
				{Key: "testKey2", Value: "testVal2", Secret: false},
			},
		},
		{
			description: "override config entry",
			configEntries: []ConfigEntry{
				{Key: "testKey", Value: "testVal", Secret: false},
				{Key: "testKey2", Value: "some value", Secret: true},
			},
			setEntry: ConfigEntry{Key: "testKey2", Value: "testVal2", Secret: false},
			expectedEntries: []ConfigEntry{
				{Key: "testKey", Value: "testVal", Secret: false},
				{Key: "testKey2", Value: "testVal2", Secret: false},
			},
		},
	} {
		t.Run("should "+testCase.description, func(t *testing.T) {
			testCase.configEntries.Set(testCase.setEntry.Key, testCase.setEntry.Value, testCase.setEntry.Secret)
			assert.Equal(t, testCase.expectedEntries, testCase.configEntries)
		})
	}

}

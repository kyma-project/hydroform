package config

import (
	"testing"
	"time"

	"github.com/kyma-incubator/hydroform/types"
	"github.com/stretchr/testify/assert"
)

func Test_LoadTimeoutConfiguration(t *testing.T) {

	for _, testCase := range []struct {
		description    string
		timeouts       types.Timeouts
		expectedConfig map[string]interface{}
	}{
		{
			description: "timeouts provided",
			timeouts: types.Timeouts{
				Create: 100 * time.Minute,
				Update: 200 * time.Minute,
				Delete: 300 * time.Minute,
			},
			expectedConfig: map[string]interface{}{
				"create_timeout": 100 * time.Minute,
				"update_timeout": 200 * time.Minute,
				"delete_timeout": 300 * time.Minute,
			},
		},
		{
			description: "timeouts not provided",
			timeouts:    types.Timeouts{},
			expectedConfig: map[string]interface{}{
				"create_timeout": defaultCreateTimeout,
				"update_timeout": defaultUpdateTimeout,
				"delete_timeout": defaultDeleteTimeout,
			},
		},
	} {
		t.Run("should load timeouts configuration when "+testCase.description, func(t *testing.T) {
			// when
			config := LoadTimeoutConfiguration(testCase.timeouts)

			// then
			assert.Equal(t, testCase.expectedConfig, config)
		})
	}

}

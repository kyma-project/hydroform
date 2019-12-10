package config

import (
	"time"

	"github.com/kyma-incubator/hydroform/types"
)

const (
	defaultCreateTimeout = 30 * time.Minute
	defaultUpdateTimeout = 20 * time.Minute
	defaultDeleteTimeout = 15 * time.Minute
)

func LoadTimeoutConfiguration(timeouts types.Timeouts) map[string]interface{} {
	if timeouts.Create == 0 {
		timeouts.Create = defaultCreateTimeout
	}
	if timeouts.Update == 0 {
		timeouts.Update = defaultUpdateTimeout
	}
	if timeouts.Delete == 0 {
		timeouts.Delete = defaultDeleteTimeout
	}

	return map[string]interface{}{
		"create_timeout": timeouts.Create,
		"update_timeout": timeouts.Update,
		"delete_timeout": timeouts.Delete,
	}
}

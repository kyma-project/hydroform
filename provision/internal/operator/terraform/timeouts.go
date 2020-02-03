package terraform

import (
	"time"

	"github.com/kyma-incubator/hydroform/provision/types"
)

const (
	defaultCreateTimeout = 30 * time.Minute
	defaultUpdateTimeout = 30 * time.Minute
	defaultDeleteTimeout = 20 * time.Minute
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

func ExtendConfig(sourceCfg map[string]interface{}, extensions map[string]interface{}) {
	for k, v := range extensions {
		sourceCfg[k] = v
	}
}

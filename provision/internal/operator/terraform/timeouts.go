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

func applyTimeouts(cfg map[string]interface{}, timeouts types.Timeouts) {
	if timeouts.Create == 0 {
		timeouts.Create = defaultCreateTimeout
	}
	if timeouts.Update == 0 {
		timeouts.Update = defaultUpdateTimeout
	}
	if timeouts.Delete == 0 {
		timeouts.Delete = defaultDeleteTimeout
	}

	cfg["create_timeout"] = timeouts.Create
	cfg["update_timeout"] = timeouts.Update
	cfg["delete_timeout"] = timeouts.Delete
}

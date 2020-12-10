//Package config defines top-level configuration settings for library users.
package config

import (
	"log"
	"time"
)

//Configures various install/uninstall operation parameters.
//There are no different parameters for the "install" and "delete" operations.
//If you need different configurations, just use two different Installation instances.
type Config struct {
	//Number of parallel workers used for an install/uninstall operation
	WorkersCount int
	//After this time workers' context is canceled. Pending worker goroutines (if any) may continue if blocked by Helm client.
	CancelTimeout time.Duration
	//After this time install/delete operation is aborted and returns an error to the user.
	//Worker goroutines may still be working in the background.
	//Must be greater than CancelTimeout.
	QuitTimeout time.Duration
	//Timeout for the underlying Helm client
	HelmTimeoutSeconds int
	//Initial interval used for exponent backoff retry policy
	BackoffInitialIntervalSeconds int
	//Maximum time used for exponent backoff retry policy
	BackoffMaxElapsedTimeSeconds int
	//Logger to use
	Log func(format string, v ...interface{})
}

// TODO: Remove this variable. Search for occurrences of config.Log
// It is used in functions to avoid passing logger as a parameter
var Log = log.Printf

//It sets up the global logging function.
func SetupLogger(log func(string, ...interface{})) {
	Log = log
}

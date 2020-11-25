package config

import log "github.com/sirupsen/logrus"

//Configures various install/uninstall operation parameters.
//There are no different parameters for Install/Delete operations - if you need it to be different, just use two Installations with two different configs.
type Config struct {
	//Number of concurrent workers used for an install/delete operation.
	WorkersCount int
	//After this time workers' context is canceled. Pending worker goroutines (if any) may continue if blocked by Helm client.
	CancelTimeoutSeconds int
	//After this time install/delete operation is aborted and returns an error to the user.
	//Worker goroutines may still be working in the background.
	//Must be greater than CancelTimeoutSeconds.
	QuitTimeoutSeconds int
	//Timeout for Helm client
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
var Log = log.Infof

package types

import "time"

// Options contains all possible configuration options for Hydroform.
// Options need to be set each time a Hydroform function is called
type Options struct {
	DataDir    string
	Persistent bool
	Timeouts   *Timeouts
}

// Timeouts specifies timeouts on various operation
type Timeouts struct {
	Create time.Duration
	Update time.Duration
	Delete time.Duration
}

// Option is a function that allows to extensibly configure Hydroform.
type Option func(*Options)

// Make hydroform configuration files stay in the file system after running.
// By default files are always deleted after each call.
func Persistent() Option {
	return func(ops *Options) {
		ops.Persistent = true
	}
}

// Set a custom directory where all Hydroform files will be stored.
func WithDataDir(dir string) Option {
	return func(ops *Options) {
		ops.DataDir = dir
	}
}

func WithTimeouts(timeouts *Timeouts) Option {
	return func(ops *Options) {
		ops.Timeouts = timeouts
	}
}

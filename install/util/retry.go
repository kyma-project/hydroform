package util

import "time"

const (
	defaultMaxAttempts = 3
	defaultDelay       = 5 * time.Second
)

type operationFunc func() (interface{}, error)
type isRecoverableFunc func(error) bool

func WithDefaultRetry(operation operationFunc, isRecoverable isRecoverableFunc) (interface{}, error) {
	return withRetry(defaultMaxAttempts, defaultDelay, operation, isRecoverable)
}

func withRetry(maxAttempts int, delay time.Duration, operation operationFunc, isRecoverable isRecoverableFunc) (interface{}, error) {
	var err error
	var obj interface{}

	for i := 0; i < maxAttempts; i++ {
		obj, err = operation()

		if err == nil || !isRecoverable(err) {
			return obj, err
		}

		time.Sleep(delay)
	}

	return obj, err
}

package util

import (
	"fmt"
	"time"
)

const (
	defaultRetryNo       = 3
	defaultSleepDuration = 5 * time.Second
)

// WaitFor waits until isReady returns true, error or the timeout was reached
func WaitFor(interval, timeout time.Duration, isReady func() (bool, error)) error {
	done := time.After(timeout)

	for {
		if ready, err := isReady(); err != nil {
			return err
		} else if ready {
			return nil
		}

		select {
		case <-done:
			return fmt.Errorf("timeout waiting for condition")
		default:
			time.Sleep(interval)
		}
	}
}

func WithDefaultRetry(invocation func() (interface{}, error)) (interface{}, error) {
	return withRetry(defaultRetryNo, defaultSleepDuration, invocation)
}

func withRetry(count int, sleep time.Duration, invocation func() (interface{}, error)) (interface{}, error) {
	obj, err := invocation()
	for count := count - 1; count > 0 && err != nil; count-- {
		time.Sleep(sleep)
		obj, err = invocation()
	}

	return obj, err
}

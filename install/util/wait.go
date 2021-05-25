package util

import (
	"fmt"
	"time"
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


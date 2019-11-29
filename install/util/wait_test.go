package util

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
)

func TestWaitFor(t *testing.T) {

	t.Run("should wait until ready", func(t *testing.T) {
		// given
		start := time.Now().Unix()
		var timePassed int64

		// when
		err := WaitFor(time.Second, time.Second*10, func() (b bool, e error) {
			timePassed = time.Now().Unix() - start

			if time.Duration(timePassed)*time.Second > 3*time.Second {
				return true, nil
			}

			return false, nil
		})

		// then
		require.NoError(t, err)
	})

	t.Run("should return error on timeout", func(t *testing.T) {
		// given
		start := time.Now().Unix()
		var timePassed int64

		// when
		err := WaitFor(time.Second, time.Second*3, func() (b bool, e error) {
			timePassed = time.Now().Unix() - start

			if time.Duration(timePassed)*time.Second > 5*time.Second {
				return true, nil
			}

			return false, nil
		})

		// then
		require.Error(t, err)
		assert.Contains(t, err.Error(), "timeout")
	})

	t.Run("should return error when error occurred", func(t *testing.T) {
		// when
		err := WaitFor(time.Second, time.Second*3, func() (b bool, e error) {
			return false, fmt.Errorf("error")
		})

		// then
		require.Error(t, err)
		assert.Contains(t, err.Error(), "error")
	})

}

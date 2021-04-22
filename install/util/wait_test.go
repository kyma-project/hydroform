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

	t.Run("should not repeat invocation if no error", func(t *testing.T) {
		// when
		count := 0
		retries := 1
		obj, err := withRetry(retries, 0, func() (interface{}, error) {
			count++
			return nil, nil
		})

		// then
		require.NoError(t, err)
		require.Equal(t, 1, count)
		require.Nil(t, obj)
	})

	t.Run("should not repeat invocation if error resolved", func(t *testing.T) {
		// when
		count := 0
		retries := 3
		obj, err := withRetry(retries, 0, func() (interface{}, error) {
			count++
			if count < 2 {
				return nil, fmt.Errorf("error")
			} else {
				return nil, nil
			}
		})

		// then
		require.NoError(t, err)
		require.Equal(t, 2, count)
		require.Nil(t, obj)
	})

	t.Run("should repeat invocation on error", func(t *testing.T) {
		// when
		count := 0
		retries := 1
		obj, err := withRetry(retries, 0, func() (interface{}, error) {
			count++
			return nil, fmt.Errorf("error")
		})

		// then
		require.Error(t, err)
		require.Equal(t, retries, count)
		require.Nil(t, obj)
		assert.Contains(t, err.Error(), "error")
	})

	t.Run("should sleep on error for specific time", func(t *testing.T) {
		// given
		retries := 2
		start := time.Now()
		sleep := time.Second

		// when
		_, err := withRetry(retries, sleep, func() (interface{}, error) {
			return nil, nil
		})

		// then
		require.False(t, time.Now().After(start.Add(sleep)))

		// when
		_, err = withRetry(retries, sleep, func() (interface{}, error) {
			return nil, fmt.Errorf("")
		})

		// then
		require.True(t, time.Now().After(start.Add(sleep)))
		require.Error(t, err)
	})
}

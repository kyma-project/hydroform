package util

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
	"github.com/stretchr/testify/assert"
)

func TestRetry(t *testing.T) {

	t.Run("should not retry if no error", func(t *testing.T) {
		// given
		count := 0
		maxAttempts := 2

		// when
		obj, err := withRetry(maxAttempts, 0, func() (interface{}, error) {
			count++
			return nil, nil
		}, func (error) bool {return false})

		// then
		require.NoError(t, err)
		require.Equal(t, 1, count)
		require.Nil(t, obj)
	})

	t.Run("should retry on error", func(t *testing.T) {
		// when
		count := 0
		maxAttempts := 2
		obj, err := withRetry(maxAttempts, 0, func() (interface{}, error) {
			count++
			return nil, fmt.Errorf("error")
		}, func (error) bool {return true})

		// then
		require.Error(t, err)
		require.Equal(t, maxAttempts, count)
		require.Nil(t, obj)
		assert.Contains(t, err.Error(), "error")
	})

	t.Run("should skip retry on unrecoverable error", func(t *testing.T) {
		// when
		count := 0
		maxAttempts := 2
		obj, err := withRetry(maxAttempts, 0, func() (interface{}, error) {
			count++
			return nil, fmt.Errorf("error")
		}, func (error) bool {return false})

		// then
		require.Error(t, err)
		require.Equal(t, 1, count)
		require.Nil(t, obj)
		assert.Contains(t, err.Error(), "error")
	})

	t.Run("should sleep on error for specified time", func(t *testing.T) {
		// given
		maxAttempts := 2
		start := time.Now()
		sleep := time.Second

		// when
		_, err := withRetry(maxAttempts, sleep, func() (interface{}, error) {
			return nil, nil
		}, func (error) bool {return true})

		// then
		require.False(t, time.Now().After(start.Add(sleep)))

		// when
		_, err = withRetry(maxAttempts, sleep, func() (interface{}, error) {
			return nil, fmt.Errorf("")
		}, func (error) bool {return true})

		// then
		require.True(t, time.Now().After(start.Add(2 * sleep)))
		require.Error(t, err)
	})
}
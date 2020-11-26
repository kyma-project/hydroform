package helm

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNoBackoff(t *testing.T) {
	var count int = 0
	o := func() error {
		count += 1
		return nil
	}

	err := retryWithBackoff(context.TODO(), o, 1*time.Millisecond, 10*time.Millisecond)

	expectedCount := 1
	require.Equal(t, expectedCount, count, "Number of invocations not as expected")
	require.NoError(t, err)
}

func TestOneBackoff(t *testing.T) {
	var count int = 0
	o := func() error {
		count += 1
		if count < 2 {
			return errors.New("failure")
		}
		return nil
	}

	err := retryWithBackoff(context.TODO(), o, 1*time.Millisecond, 10*time.Millisecond)

	expectedCount := 2
	require.Equal(t, expectedCount, count, "Number of invocations not as expected")
	require.NoError(t, err)
}

func TestBackoffWithCancel(t *testing.T) {

	var count int = 0
	o1 := func() error {
		count += 1
		return errors.New("failure")
	}
	//Ensure more than 4 retries are done in 20[ms]
	err := retryWithBackoff(context.TODO(), o1, 1*time.Millisecond, 20*time.Millisecond)
	require.Error(t, err)
	require.Greater(t, count, 4)

	count = 0
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	o2 := func() error {
		count += 1
		t.Log("Operation run: #", count)

		//Cancel processing after 3rd retry
		if count == 3 {
			//Run async with a small delay. 4th retry should be scheduled
			go func() {
				time.Sleep(time.Millisecond * 2)
				cancel()
			}()
		}
		return errors.New("failure")
	}

	startTime := time.Now()
	err = retryWithBackoff(ctx, o2, 1*time.Millisecond, 2000*time.Millisecond)
	endTime := time.Now()
	timeDiff := endTime.Sub(startTime)
	t.Log("Total operations run count:", count)
	t.Logf("Total retrying time: %v[ms]", timeDiff.Milliseconds())

	expectedMaxTime := int64(10 * time.Millisecond) //10[ms]
	expectedMaxCount := 4
	require.Error(t, err)
	require.LessOrEqual(t, count, expectedMaxCount, "total retries count too big")
	require.Less(t, int64(timeDiff), expectedMaxTime, "total time of retries outside the expected range")
}

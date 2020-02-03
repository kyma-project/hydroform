package action

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFuncAction(t *testing.T) {
	f := FuncAction(func(args ...interface{}) (interface{}, error) {
		switch v := args[0].(type) {
		case int:
			return v * 2, nil
		case string:
			return fmt.Sprintf("I received the following string: %s", v), nil
		default:
			return nil, errors.New("Received argument is not supported")
		}
	})

	res, err := f.Run("arg1")
	require.NoError(t, err)
	require.Equal(t, "I received the following string: arg1", res.(string))

	res, err = f.Run(42)
	require.NoError(t, err)
	require.Equal(t, 84, res.(int))

	_, err = f.Run([]byte("arg1"))
	require.Error(t, err)
}

func TestPipe(t *testing.T) {
	p := Pipe{
		// return a number
		FuncAction(func(args ...interface{}) (interface{}, error) {
			return 27, nil
		}),
		// then do some operation with it and pass on the input and output
		FuncAction(func(args ...interface{}) (interface{}, error) {
			n := args[0].(int)
			return []interface{}{n, n * n}, nil
		}),
		// Put it into a message
		FuncAction(func(args ...interface{}) (interface{}, error) {
			n := args[0].(int)
			nn := args[1].(int)
			return fmt.Sprintf("%d times %d is %d", n, n, nn), nil
		}),
	}

	r, err := p.Run()
	require.NoError(t, err)
	require.Equal(t, "27 times 27 is 729", r)
}

func TestSequence(t *testing.T) {
	seq := Sequence{
		FuncAction(func(args ...interface{}) (interface{}, error) {
			return len(args), nil
		}),
		FuncAction(func(args ...interface{}) (interface{}, error) {
			return len(args), nil
		}),
		FuncAction(func(args ...interface{}) (interface{}, error) {
			return len(args), nil
		}),
	}

	res, err := seq.Run("arg1", "arg2", "arg3")
	require.NoError(t, err)
	// Check that all results are returned
	require.Len(t, res, 3)
	// check that all actions in the sequence get the exact number of arguments
	results := res.([]interface{})
	require.Equal(t, 3, results[0])
	require.Equal(t, 3, results[1])
	require.Equal(t, 3, results[2])
}

func TestParallel(t *testing.T) {
	p := Parallel{
		FuncAction(func(args ...interface{}) (interface{}, error) {
			return len(args), nil
		}),
		FuncAction(func(args ...interface{}) (interface{}, error) {
			return len(args), nil
		}),
		FuncAction(func(args ...interface{}) (interface{}, error) {
			return len(args), nil
		}),
	}

	res, err := p.Run("arg1", "arg2", "arg3")
	require.NoError(t, err)
	// Check that all results are returned
	require.Len(t, res, 3)
	// check that all actions in the parallel get the exact number of arguments
	results := res.([]interface{})
	require.Equal(t, 3, results[0])
	require.Equal(t, 3, results[1])
	require.Equal(t, 3, results[2])
}

func TestBefore(t *testing.T) {
	// No before action means nothing is returned
	require.NoError(t, Before())

	SetBefore(FuncAction(func(args ...interface{}) (interface{}, error) {
		return nil, errors.New("This action always fails")
	}))
	SetArgs("arg1", "arg2", "arg3")
	// Check that errors are forwarded
	require.Error(t, Before())
	// check that actions are cleared after running
	require.Nil(t, before)
}

func TestAfter(t *testing.T) {
	// No after action means nothing is returned
	require.NoError(t, After())

	SetAfter(FuncAction(func(args ...interface{}) (interface{}, error) {
		return nil, errors.New("This action always fails")
	}))
	SetArgs("arg1", "arg2", "arg3")
	// Check that errors are forwarded
	require.Error(t, After())
	// check that actions are cleared after running
	require.Nil(t, after)
}

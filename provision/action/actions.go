package action

import (
	"errors"
	"fmt"
	"strings"
)

// Sequence is an action formed by an ordered sequence of actions.
// It can be used the same way as any other Action.
type Sequence []Action

// Run executes all actions in order with the same input parameters, collects all results and errors and returns them.
// An error in an action does not stop the execution.
func (sq Sequence) Run(args ...interface{}) (interface{}, error) {
	results := make([]interface{}, 0)
	errStr := strings.Builder{}
	for _, a := range sq {
		res, err := a.Run(args...)
		if res != nil {
			results = append(results, res)
		}

		if err != nil {
			errStr.WriteString(fmt.Sprintf("\n%s", err.Error()))
		}
	}

	var err error = nil
	if errStr.Len() > 0 {
		err = errors.New(errStr.String())
	}

	return results, err
}

// Pipe is an action formed by an ordered sequence of actions that are piped to each other. The output of one action is passed as argument to the next one.
// It can be used the same way as any other Action.
type Pipe []Action

// Run executes all actions in order piping the output of one action into the input of the next one. If an action has an error the execution stops.
// Returns the output of the last action in the pipe
func (p Pipe) Run(args ...interface{}) (interface{}, error) {
	// turn args into interface{} so we can use it inside of the same loop as the returned interface{} on all other actions in the pipe.
	res := interface{}(args)
	var err error

	// run each action and use its output as input for the next one
	for _, a := range p {
		switch r := res.(type) {
		case []interface{}:
			res, err = a.Run(r...)
			if err != nil {
				return r, err
			}
		default:
			res, err = a.Run(r)
			if err != nil {
				return res, err
			}
		}
	}
	return res, nil
}

// Parallel is an action formed by a set of actions that will be run concurrently.
// It can be used the same way as any other Action.
type Parallel []Action

// Run executes all actions concurrently with the same input parameters, collects all results and errors and returns them.
// An error in an action does not stop the execution.
func (p Parallel) Run(args ...interface{}) (interface{}, error) {
	type resultSet struct {
		result interface{}
		err    error
	}
	ach := make(chan resultSet, len(p)) // chan is buffered to not block any sender

	// Run all actions concurrently and send their result and error through a channel
	for _, a := range p {
		go func(a Action, ch chan<- resultSet, args ...interface{}) {
			r := resultSet{}
			r.result, r.err = a.Run(args...)
			ch <- r
		}(a, ach, args...)

	}

	// Collect all results and errors from the channel
	results := make([]interface{}, 0)
	errStr := strings.Builder{}
	for i := 0; i < len(p); i++ {
		r := <-ach

		if r.result != nil {
			results = append(results, r.result)
		}
		if r.err != nil {
			errStr.WriteString(fmt.Sprintf("\n%s", r.err.Error()))
		}

	}

	var err error = nil
	if errStr.Len() > 0 {
		err = errors.New(errStr.String())
	}

	return results, err
}

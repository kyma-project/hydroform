package jobmanager

import (
	"time"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/config"
)

type component string

type executionTime int

const (
	Pre executionTime = iota
	Post
)

var duration time.Duration = 0.00

var preJobMap map[component][]job
var postJobMap map[component][]job

var jobs []jobs

// Define type for jobs
type job interface {
	execute(*config.Config, kubernetes.Interface) error
	when() (component, executionTime)
}

// Register job
func register(j job) {
	// TODO: Add job to corresponding map
}

// Function should be called before component is being deployed/upgraded
func ExecutePre(component string) {
	start := time.Now()
	// TODO: Executes the registered functions for given component; using maps
	//       If map for given key(aka component) is empty, nothing will be done
	//       Check installationType, to know which map should be used
	t := time.Now()
	duration += t.Sub(start)
}

// Function should be called after compoent is being deployed/upgraded
func ExecutePost(component string) {
	start := time.Now()
	// TODO: Executes the registered functions for given component; using maps
	//       If map for given key(aka component) is empty, nothing will be done
	//       Check installationType, to know which map should be used
	t := time.Now()
	duration += t.Sub(start)
}

// Returns duration of all jobs for benchmarking
func GetDuration() time.Duration {
	return duration
}

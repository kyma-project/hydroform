package jobmanager

import (
	"time"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/config"
	"k8s.io/client-go/kubernetes"
)

type component string

type executionTime int

const (
	Pre executionTime = iota
	Post
)

var duration time.Duration

var preJobMap map[component][]job
var postJobMap map[component][]job

// Define type for jobs
type job interface {
	execute(*config.Config, kubernetes.Interface) error
	when() (component, executionTime)
}

// Register job
func register(j job) {
	component, executionTime := j.when()
	if executionTime == Pre {
		preJobMap[component] = append(preJobMap[component], j)
	} else if executionTime == Post {
		postJobMap[component] = append(postJobMap[component], j)
	}
}

// Function should be called before component is being deployed/upgraded
func ExecutePre(c component) {
	start := time.Now()
	// TODO: Executes the registered functions for given component; using maps
	//       If map for given key(aka component) is empty, nothing will be done
	//       Check installationType, to know which map should be used
	jobs := preJobMap[c]
	if len(jobs) > 0 {
		for _, job := range jobs {
			job.execute() // TODO MIssing args, plus introduce worker to async
		}
	}

	t := time.Now()
	duration += t.Sub(start)
}

// Function should be called after compoent is being deployed/upgraded
func ExecutePost(c component) {
	start := time.Now()
	// TODO: Executes the registered functions for given component; using maps
	//       If map for given key(aka component) is empty, nothing will be done
	//       Check installationType, to know which map should be used
	jobs := postJobMap[c]
	if len(jobs) > 0 {
		for _, job := range jobs {
			job.execute()
		}
	}
	t := time.Now()
	duration += t.Sub(start)
}

// Returns duration of all jobs for benchmarking
func GetDuration() time.Duration {
	return duration
}

func init() {
	duration = 0 * time.Second
	preJobMap = make(map[component][]job)
	postJobMap = make(map[component][]job)
}

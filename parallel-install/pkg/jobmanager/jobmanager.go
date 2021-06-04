package jobmanager

import (
	"context"
	"sync"
	"time"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/config"
	"github.com/kyma-project/hydroform/parallel-install/pkg/logger"
	"k8s.io/client-go/kubernetes"
)

type component string
type executionTime int
type jobName string

type jobStatus struct {
	job    jobName
	status bool
}

// Define type for jobs
type job interface {
	execute(*config.Config, kubernetes.Interface) error
	when() (component, executionTime)
	identify() jobName
}

const (
	Pre executionTime = iota
	Post
)

var duration time.Duration

var preJobMap map[component][]job
var postJobMap map[component][]job

var Log logger.Interface

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
// If the Context is cancelled, the worker quits immediately, skipping the remaining components.
func ExecutePre(ctx context.Context, c component) {
	var wg sync.WaitGroup
	statusChan := make(chan jobStatus)

	go func() {
		wg.Wait()
		close(statusChan)
	}()

	start := time.Now()

	jobs := preJobMap[c]
	if len(jobs) > 0 {
		wg.Add(len(jobs))
		for _, job := range jobs {
			go worker(ctx, statusChan, &wg, job, &Log)
		}
	}

	for status := range statusChan {
		if status.status == false {
			Log.Fatal("Following job failed while execution: %v", status.job)
		}
	}

	t := time.Now()
	duration += t.Sub(start)
}

func worker(ctx context.Context, statusChan chan<- jobStatus, wg *sync.WaitGroup, j job, log *logger.Interface) {
	defer wg.Done()

	if err := j.execute(); err != nil { // TODO> Need to figure out how to pass config and K8s client
		statusChan <- jobStatus{j.identify(), false}
	} else {
		statusChan <- jobStatus{j.identify(), true}
	}

}

// Function should be called after compoent is being deployed/upgraded
func ExecutePost(c component) {

	var wg sync.WaitGroup
	statusChan := make(chan jobStatus)

	start := time.Now()
	// TODO: Executes the registered functions for given component; using maps
	//       If map for given key(aka component) is empty, nothing will be done
	//       Check installationType, to know which map should be used
	jobs := postJobMap[c]
	if len(jobs) > 0 {
		wg.Add(len(jobs))
		for _, job := range jobs {
			go worker(context.Background(), statusChan, &wg, job)
			//job.execute()
		}
	}
	for status := range statusChan {
		if status.status == false {
			// Maybe raise error
		}
	}
	wg.Wait()
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
	Log = logger.NewLogger(true)
}

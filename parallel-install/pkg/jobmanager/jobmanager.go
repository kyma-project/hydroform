package jobmanager

import (
	"context"
	"sync"
	"time"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/config"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/logger"
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

var preJobMap = make(map[component][]job)
var postJobMap = make(map[component][]job)

var kubeClient kubernetes.Interface
var cfg *config.Config

var Log logger.Interface

// Register job
func register(j job) int {
	component, executionTime := j.when()
	if executionTime == Pre {
		preJobMap[component] = append(preJobMap[component], j)
	} else if executionTime == Post {
		postJobMap[component] = append(postJobMap[component], j)
	}
	return 0
}

func SetConfig(config *config.Config) {
	cfg = config
}

func SetKubeClient(kc kubernetes.Interface) {
	kubeClient = kc
}

// Function should be called before component is being deployed/upgraded
// If the Context is cancelled, the worker quits immediately, skipping the remaining components.
func ExecutePre(ctx context.Context, c string) {
	execute(ctx, c, preJobMap)
}

// Function should be called after compoent is being deployed/upgraded
func ExecutePost(ctx context.Context, c string) {
	execute(ctx, c, postJobMap)
}

func worker(ctx context.Context, statusChan chan<- jobStatus, wg *sync.WaitGroup, j job) {
	defer wg.Done()
	if err := j.execute(cfg, kubeClient); err != nil { // TODO> Need to figure out how to pass config and K8s client
		statusChan <- jobStatus{j.identify(), false}
	} else {
		statusChan <- jobStatus{j.identify(), true}
	}

}

func execute(ctx context.Context, c string, executionMap map[component][]job) {
	var wg sync.WaitGroup
	statusChan := make(chan jobStatus)

	start := time.Now()

	jobs := executionMap[component(c)]
	if len(jobs) > 0 {
		wg.Add(len(jobs))
		for _, job := range jobs {
			go worker(ctx, statusChan, &wg, job)
		}
	}

	go func() {
		wg.Wait()
		close(statusChan)
	}()

	for status := range statusChan {
		if status.status == true {
			Log.Info("Following job executed: %v", status.job)
		} else if status.status == false {
			Log.Fatal("Following job failed while execution: %v", status.job)
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
	//Log = logger.NewLogger(true)
}

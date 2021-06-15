package jobmanager

import (
	"context"
	"sync"
	"time"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/config"
	"github.com/kyma-project/kyma/common/logging/logger"
	"k8s.io/client-go/kubernetes"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

type component string
type executionTime int
type jobName string

type jobStatus struct {
	job    jobName
	status bool
	err    error
}

type job interface {
	execute(*config.Config, kubernetes.Interface, context.Context) error
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

var zapLogger *zap.SugaredLogger

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

// Sets Installation Config at package level
func SetConfig(config *config.Config) {
	cfg = config
}

// Sets Kubernetes Cleint at package level
func SetKubeClient(kc kubernetes.Interface) {
	kubeClient = kc
}

// Function should be called before component is being deployed/upgraded
// If the Context is cancelled, the worker quits immediately, skipping the remaining components
func ExecutePre(ctx context.Context, c string) {
	execute(ctx, c, preJobMap)
}

// Function should be called after compoent is being deployed/upgraded
// If the Context is cancelled, the worker quits immediately, skipping the remaining components
func ExecutePost(ctx context.Context, c string) {
	execute(ctx, c, postJobMap)
}

// Used by ExecutePre() && ExecutePost()
// Used to start workers and grab jobs belonging to the respective component
func execute(ctx context.Context, c string, executionMap map[component][]job) {
	var wg sync.WaitGroup

	start := time.Now()

	jobs := executionMap[component(c)]
	statusChan := make(chan jobStatus, len(jobs))
	wg.Add(len(jobs))

	if len(jobs) > 0 {
		for _, job := range jobs {
			go worker(ctx, statusChan, &wg, job)
		}
	}
	go func() {
		wg.Wait()
		close(statusChan)
	}()

	emptyJob := jobStatus{}
	for status := range statusChan {
		zapLogger.Infof("Job Status: %v", status)
		if status != emptyJob {
			if status.status == true {
				zapLogger.Infof("Following job executed: %v", status.job)
			} else if status.status == false {
				zapLogger.Infof("Following job failed while execution: `%v` with error: %s", status.job, status.err)
			}
		}
	}

	t := time.Now()
	duration += t.Sub(start)
}

func worker(ctx context.Context, statusChan chan<- jobStatus, wg *sync.WaitGroup, j job) {
	defer wg.Done()
	if err := j.execute(cfg, kubeClient, ctx); err != nil {
		j := jobStatus{j.identify(), false, nil}
		statusChan <- j
	} else {
		j := jobStatus{j.identify(), true, err}
		statusChan <- j
	}
}

// Returns duration of all jobs for benchmarking
func GetDuration() time.Duration {
	zapLogger.Infof("Duration of runned jobs: %d", duration)
	return duration
}

func init() {
	duration = 0 * time.Second

	core, _ := observer.New(zap.DebugLevel)
	log, _ := logger.New(logger.TEXT, logger.INFO, core)
	zapLogger = log.WithContext()

	zapLogger.Desugar().WithOptions(zap.AddCaller())

}

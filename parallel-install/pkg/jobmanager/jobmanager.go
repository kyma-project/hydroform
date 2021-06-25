package jobmanager

import (
	"context"
	"sync"
	"time"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/logger"
	"k8s.io/client-go/kubernetes"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/config"
	istio "istio.io/client-go/pkg/clientset/versioned"
	"k8s.io/client-go/rest"
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
	execute(*config.Config, kubernetes.Interface, istio.Interface, context.Context) error
	when() (component, executionTime)
	identify() jobName
}

const (
	Pre executionTime = iota
	Post
)

var duration time.Duration
var durationMu sync.Mutex

var preJobMap = make(map[component][]job)
var postJobMap = make(map[component][]job)
var finishedJobs = []jobStatus{}

var kubeClient kubernetes.Interface
var cfg *config.Config
var restConfig *rest.Config
var istioClient *istio.Clientset

var log logger.Interface

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

// Sets Installation Config and KubeClient at package level
func RegisterJobManager(config *config.Config, kc kubernetes.Interface, rc *rest.Config) {
	cfg = config
	kubeClient = kc
	restConfig = rc
	istioClient, _ = istio.NewForConfig(restConfig)
}

// Sets Logger at package level
func SetLogger(logClient logger.Interface) {
	log = logClient
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
		log.Infof("Job Status: %v", status)
		if status != emptyJob {
			finishedJobs = append(finishedJobs, status)
			if status.status == true {
				log.Infof("Following job executed: %v", status.job)
			} else if status.status == false {
				log.Infof("Following job failed: `%v` with error: %s", status.job, status.err)
			}
		}
	}

	t := time.Now()
	addDuration(t.Sub(start))
}

func worker(ctx context.Context, statusChan chan<- jobStatus, wg *sync.WaitGroup, j job) {
	defer wg.Done()
	if err := j.execute(cfg, kubeClient, istioClient, ctx); err != nil {
		status := jobStatus{j.identify(), false, err}
		statusChan <- status
	} else {
		status := jobStatus{j.identify(), true, nil}
		statusChan <- status
	}
}

// Returns duration of all jobs for benchmarking
func GetDuration() time.Duration {
	durationMu.Lock()
	defer durationMu.Unlock()
	ret := duration
	resetDuration()
	return ret
}

func addDuration(t time.Duration) {
	durationMu.Lock()
	defer durationMu.Unlock()
	duration += t
}

func resetDuration() {
	duration = 0 * time.Microsecond
}

func resetMap(exec executionTime) {
	if exec == Pre {
		preJobMap = make(map[component][]job)
	} else if exec == Post {
		postJobMap = make(map[component][]job)
	}
}

func resetFinishedJobsMap() {
	finishedJobs = []jobStatus{}
}

func init() {
	duration = 0 * time.Microsecond
}

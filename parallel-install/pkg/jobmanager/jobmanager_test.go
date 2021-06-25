package jobmanager

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/config"
	installConfig "github.com/kyma-incubator/hydroform/parallel-install/pkg/config"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/logger"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

func TestJobManager(t *testing.T) {
	t.Run("concurrent pre-jobs sampleOne and sampleTwo should be triggered", func(t *testing.T) {
		// Init test setup
		resetFinishedJobsMap()
		SetLogger(logger.NewLogger(false))

		// Test the execution func
		jobMap := initJobMap()
		execute(context.TODO(), "componentOne", jobMap)

		// Check executed Jobs
		require.Contains(t, finishedJobs, jobStatus{job: "sampleOne", status: true, err: nil})
		require.Contains(t, finishedJobs, jobStatus{job: "sampleTwo", status: true, err: nil})
		require.NotContains(t, finishedJobs, jobStatus{job: "sampleThree", status: true, err: nil})
		require.NotContains(t, finishedJobs, jobStatus{job: "sampleFour", status: true, err: nil})
		require.NotContains(t, finishedJobs, jobStatus{job: "sampleFive", status: false, err: errors.New("JobFiveError")})
		require.NotEqual(t, 0*time.Second, GetDuration())
	})

	t.Run("single pre-job sampleThree should be triggered", func(t *testing.T) {
		// Init test setup
		resetFinishedJobsMap()
		SetLogger(logger.NewLogger(false))

		// Test the execution func
		jobMap := initJobMap()
		execute(context.TODO(), "componentTwo", jobMap)

		// Check executed Jobs
		require.NotContains(t, finishedJobs, jobStatus{job: "sampleOne", status: true, err: nil})
		require.NotContains(t, finishedJobs, jobStatus{job: "sampleTwo", status: true, err: nil})
		require.Contains(t, finishedJobs, jobStatus{job: "sampleThree", status: true, err: nil})
		require.NotContains(t, finishedJobs, jobStatus{job: "sampleFour", status: true, err: nil})
		require.NotContains(t, finishedJobs, jobStatus{job: "sampleFive", status: false, err: errors.New("JobFiveError")})
		require.NotEqual(t, 0*time.Second, GetDuration())
	})

	t.Run("no jobs should be triggered", func(t *testing.T) {
		// Init test setup
		resetFinishedJobsMap()
		SetLogger(logger.NewLogger(false))

		// Test the execution func
		jobMap := initJobMap()
		execute(context.TODO(), "nonExistingComponent", jobMap)

		// Check executed Jobs
		require.NotContains(t, finishedJobs, jobStatus{job: "sampleOne", status: true, err: nil})
		require.NotContains(t, finishedJobs, jobStatus{job: "sampleTwo", status: true, err: nil})
		require.NotContains(t, finishedJobs, jobStatus{job: "sampleThree", status: true, err: nil})
		require.NotContains(t, finishedJobs, jobStatus{job: "sampleFour", status: true, err: nil})
		require.NotContains(t, finishedJobs, jobStatus{job: "sampleFive", status: false, err: errors.New("JobFiveError")})
		require.Less(t, 0*time.Second, GetDuration())
	})

	t.Run("job error should be catched and user be informed", func(t *testing.T) {
		// Init test setup
		resetFinishedJobsMap()
		SetLogger(logger.NewLogger(false))

		// Test the execution func
		jobMap := make(map[component][]job)
		jobMap[component("componentFour")] = []job{sampleFive{}}
		execute(context.TODO(), "componentFour", jobMap)

		// Check executed Jobs
		require.NotContains(t, finishedJobs, jobStatus{job: "sampleOne", status: true, err: nil})
		require.NotContains(t, finishedJobs, jobStatus{job: "sampleTwo", status: true, err: nil})
		require.NotContains(t, finishedJobs, jobStatus{job: "sampleThree", status: true, err: nil})
		require.NotContains(t, finishedJobs, jobStatus{job: "sampleFour", status: true, err: nil})
		require.Contains(t, finishedJobs, jobStatus{job: "sampleFive", status: false, err: errors.New("JobFiveError")})

	})

	t.Run("duration should be reset", func(t *testing.T) {
		// Set duration to time unequal zero
		duration = 10*time.Second + 1*time.Hour
		// Reset duration
		resetDuration()
		// Check if duration is reset
		require.Equal(t, 0*time.Second, duration)
	})

	t.Run("preMap should be reset", func(t *testing.T) {
		// Fill map
		preJobMap[component("componentOne")] = []job{sampleOne{}, sampleTwo{}}
		preJobMap[component("componentTwo")] = []job{sampleThree{}}
		// Reset map
		resetMap(Pre)
		// Check if preJobMap is empty
		emptyMap := make(map[component][]job)
		require.Equal(t, emptyMap, preJobMap)
	})

	t.Run("postMap should be reset", func(t *testing.T) {
		// Fill map
		postJobMap[component("componentOne")] = []job{sampleOne{}, sampleTwo{}}
		postJobMap[component("componentTwo")] = []job{sampleThree{}}
		// Reset map
		resetMap(Post)
		// Check if postJobMap is empty
		emptyMap := make(map[component][]job)
		require.Equal(t, emptyMap, postJobMap)
	})
}

// ######## Helper Funcs #######

func initJobManager() {
	// Empty cluster, to check basic function og jobManager
	kubeClient := fake.NewSimpleClientset()
	installationCfg := &installConfig.Config{
		WorkersCount: 1,
	}
	// Set fake Kubernetes Client and empty installation config
	RegisterJobManager(installationCfg, kubeClient)
}

func initJobMap() map[component][]job {
	// Register jobs to corresponding component
	jobMap := make(map[component][]job)
	jobMap[component("componentOne")] = []job{sampleOne{}, sampleTwo{}}
	jobMap[component("componentTwo")] = []job{sampleThree{}}
	jobMap[component("componentThree")] = []job{sampleFour{}}
	jobMap[component("componentFour")] = []job{sampleFive{}}
	return jobMap
}

// ######### Test Jobs #########

type sampleOne struct {
	t *testing.T
}

func (j sampleOne) when() (component, executionTime) {
	return component("componentOne"), Pre
}
func (j sampleOne) identify() jobName {
	return jobName("sampleOne")
}
func (j sampleOne) execute(cfg *config.Config, kubeClient kubernetes.Interface, ctx context.Context) error {
	// testLogger.Debug("sampleOne triggered")
	return nil
}

type sampleTwo struct {
	t *testing.T
}

func (j sampleTwo) when() (component, executionTime) {
	return component("componentOne"), Pre
}
func (j sampleTwo) identify() jobName {
	return jobName("sampleTwo")
}
func (j sampleTwo) execute(cfg *config.Config, kubeClient kubernetes.Interface, ctx context.Context) error {
	// testLogger.Debug("sampleTwo triggered")
	return nil
}

type sampleThree struct {
	t *testing.T
}

func (j sampleThree) when() (component, executionTime) {
	return component("componentTwo"), Pre
}
func (j sampleThree) identify() jobName {
	return jobName("sampleThree")
}
func (j sampleThree) execute(cfg *config.Config, kubeClient kubernetes.Interface, ctx context.Context) error {
	// testLogger.Debug("sampleThree triggered")
	return nil
}

type sampleFour struct {
	t *testing.T
}

func (j sampleFour) when() (component, executionTime) {
	return component("componentThree"), Post
}
func (j sampleFour) identify() jobName {
	return jobName("sampleFour")
}
func (j sampleFour) execute(cfg *config.Config, kubeClient kubernetes.Interface, ctx context.Context) error {
	// testLogger.Debug("sampleFour triggered")
	return nil
}

type sampleFive struct {
	t *testing.T
}

func (j sampleFive) when() (component, executionTime) {
	return component("componentFour"), Post
}
func (j sampleFive) identify() jobName {
	return jobName("sampleFive")
}
func (j sampleFive) execute(cfg *config.Config, kubeClient kubernetes.Interface, ctx context.Context) error {
	// testLogger.Debug("sampleFive triggered")
	return errors.New("JobFiveError")
}

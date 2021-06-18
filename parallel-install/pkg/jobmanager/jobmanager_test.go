package jobmanager

import (
	"context"
	"testing"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/config"
	installConfig "github.com/kyma-incubator/hydroform/parallel-install/pkg/config"
	"github.com/kyma-project/kyma/common/logging/logger"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

var testLogger *zap.SugaredLogger

func TestJob(t *testing.T) {
	t.Run("concurrent pre-jobs sampleOne and sampleTwo should be triggered", func(t *testing.T) {

		observedLogs := initLogger(t)
		initJobManager()

		// Test the execution func
		jobMap := initJobMap()
		execute(context.TODO(), "componentOne", jobMap)

		// Copy logs into slice
		logs := []string{}
		for i := 0; i < observedLogs.Len(); i++ {
			logs = append(logs, observedLogs.All()[i].Message)
		}

		// Check if logs are correct
		require.Equal(t, 2, observedLogs.Len())
		require.Contains(t, logs, "sampleOne triggered")
		require.Contains(t, logs, "sampleTwo triggered")
		t.Log(observedLogs.All())
	})

	t.Run("single pre-job sampleThree should be triggered", func(t *testing.T) {

		observedLogs := initLogger(t)
		initJobManager()

		// Test the execution func
		jobMap := initJobMap()
		execute(context.TODO(), "componentTwo", jobMap)

		// Copy logs into slice
		logs := []string{}
		for i := 0; i < observedLogs.Len(); i++ {
			logs = append(logs, observedLogs.All()[i].Message)
		}

		// Check if logs are correct
		require.Equal(t, 1, observedLogs.Len())
		require.Contains(t, logs, "sampleThree triggered")
		t.Log(observedLogs.All())
	})

	t.Run("no jobs should be triggered", func(t *testing.T) {

		observedLogs := initLogger(t)
		initJobManager()

		// Test the execution func
		jobMap := initJobMap()
		execute(context.TODO(), "nonExistingComponent", jobMap)

		// Copy logs into slice
		logs := []string{}
		for i := 0; i < observedLogs.Len(); i++ {
			logs = append(logs, observedLogs.All()[i].Message)
		}

		// Check if logs are correct
		require.Equal(t, 0, observedLogs.Len())
		t.Log(observedLogs.All())
	})
}

// ######## Helper Funcs #######
func initLogger(t *testing.T) *observer.ObservedLogs {
	// Initialize new Logger with Observer
	core, observedLogs := observer.New(zap.DebugLevel)
	log, err := logger.New(logger.JSON, logger.DEBUG, core)
	require.NoError(t, err)
	testLogger = log.WithContext()
	testLogger.Desugar().WithOptions(zap.AddCaller())
	return observedLogs
}

func initJobManager() {
	// Empty cluster, to check basic function og jobManager
	kubeClient := fake.NewSimpleClientset()
	installationCfg := &installConfig.Config{
		WorkersCount: 1,
	}
	// Set fake Kubernetes Client and empty installation config
	SetKubeClient(kubeClient)
	SetConfig(installationCfg)
}

func initJobMap() map[component][]job {
	// Register jobs to corresponding component
	jobMap := make(map[component][]job)
	jobMap[component("componentOne")] = []job{sampleOne{}, sampleTwo{}}
	jobMap[component("componentTwo")] = []job{sampleThree{}}
	jobMap[component("componentThree")] = []job{sampleFour{}}

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
	testLogger.Debug("sampleOne triggered")
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
	testLogger.Debug("sampleTwo triggered")
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
	testLogger.Debug("sampleThree triggered")
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
	testLogger.Debug("sampleFour triggered")
	return nil
}

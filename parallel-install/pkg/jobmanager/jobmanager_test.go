package jobmanager

import (
	"context"
	"fmt"
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

func TestJob(t *testing.T) {
	// Initialize Logger with Observer
	core, observedLogs := observer.New(zap.DebugLevel)
	log, err := logger.New(logger.JSON, logger.DEBUG, core)
	require.NoError(t, err)
	zapLogger := log.WithContext()
	zapLogger.Desugar().WithOptions(zap.AddCaller())

	// Empty cluster, to check basic function og jobManager
	kubeClient := fake.NewSimpleClientset()
	installationCfg := &installConfig.Config{
		WorkersCount: 1,
	}

	// Set fake Kubernetes Client and empty installation config
	SetKubeClient(kubeClient)
	SetConfig(installationCfg)

	// Initialize jobMap
	jobMap := make(map[component][]job)
	jobMap[component("doNothing_test")] = []job{doNothing_test{}}

	// Test the execution func
	fmt.Println("BEfore")
	execute(context.TODO(), "doNothing_test", jobMap)
	//require.NoError(t, err)
	fmt.Println("After")
	// Check if logs are non empty
	require.NotEqual(t, 0, observedLogs.Len())
	t.Log(observedLogs.All())
}

// ######### Test Jobs #########

type doNothing_test struct {
	t *testing.T
}

func (j doNothing_test) when() (component, executionTime) {
	return component("doNothing_test"), Pre
}

func (j doNothing_test) identify() jobName {
	return jobName("doNothing_test")
}

func (j doNothing_test) execute(cfg *config.Config, kubeClient kubernetes.Interface, ctx context.Context) error {
	fmt.Println("Inside")
	zapLogger.Infof("Start of %s", j.identify())
	zapLogger.Debug("something")
	return nil
}

package deployment

import (
	"fmt"
	"sync"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/logger"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/config"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/engine"
	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"

	"testing"
	"time"
)

func TestDeployment_RetrieveProgressUpdates(t *testing.T) {
	procUpdChan := make(chan ProcessUpdate)

	// verify we received all expected events
	receivedEvents := make(map[string]int)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		for procUpd := range procUpdChan {
			receivedEvents[processUpdateString(procUpd)]++
		}
		expectedEvents := []string{
			"InstallPreRequisites-ProcessStart",
			"InstallPreRequisites-ProcessFinished",
			"InstallComponents-ProcessStart",
			"InstallComponents-ProcessRunning-test1-Installed",
			"InstallComponents-ProcessRunning-test2-Installed",
			"InstallComponents-ProcessRunning-test3-Installed",
			"InstallComponents-ProcessFinished",
		}

		assert.Equal(t, len(expectedEvents), len(receivedEvents),
			fmt.Sprintf("Amount of expected and received events differ (got %v)", receivedEvents))

		for _, expectedEvent := range expectedEvents {
			count, ok := receivedEvents[expectedEvent]
			assert.True(t, ok, fmt.Sprintf("Expected event '%s' missing", expectedEvent))
			assert.Equal(t, 1, count, fmt.Sprintf("Expected event '%s' missing, got %v", expectedEvent, receivedEvents))
		}
		wg.Done()
	}()

	kubeClient := fake.NewSimpleClientset()

	inst := newDeployment(t, procUpdChan, kubeClient)

	hc := &mockHelmClient{}
	provider := &mockProvider{
		hc: hc,
	}
	overridesProvider := &mockOverridesProvider{}
	eng := engine.NewEngine(overridesProvider, provider, engine.Config{
		WorkersCount: 2,
		Log:          logger.NewLogger(true),
	})
	err := inst.startKymaDeployment(provider, overridesProvider, eng)
	assert.NoError(t, err)

	close(procUpdChan)

	// wait until the test threat is ready
	wg.Wait()
}

func processUpdateString(procUpd ProcessUpdate) string {
	result := fmt.Sprintf("%s-%s", procUpd.Phase, procUpd.Event)
	if procUpd.Component.Status != "" {
		return fmt.Sprintf("%s-%s-%s", result, procUpd.Component.Name, procUpd.Component.Status)
	}
	return result
}

func TestDeployment_StartKymaDeployment(t *testing.T) {
	t.Parallel()

	kubeClient := fake.NewSimpleClientset()
	i := newDeployment(t, nil, kubeClient)

	t.Run("should deploy Kyma", func(t *testing.T) {
		hc := &mockHelmClient{}
		provider := &mockProvider{
			hc: hc,
		}
		overridesProvider := &mockOverridesProvider{}
		eng := engine.NewEngine(overridesProvider, provider, engine.Config{
			WorkersCount: 2,
			Log:          logger.NewLogger(true),
		})

		err := i.startKymaDeployment(provider, overridesProvider, eng)

		assert.NoError(t, err)
	})

	t.Run("should fail to deploy Kyma prerequisites", func(t *testing.T) {
		t.Run("due to cancel timeout", func(t *testing.T) {
			hc := &mockHelmClient{
				componentProcessingTime: 200,
			}
			provider := &mockProvider{
				hc: hc,
			}
			overridesProvider := &mockOverridesProvider{}
			eng := engine.NewEngine(overridesProvider, provider, engine.Config{
				WorkersCount: 2,
				Log:          logger.NewLogger(true),
			})

			start := time.Now()
			err := i.startKymaDeployment(provider, overridesProvider, eng)
			end := time.Now()

			elapsed := end.Sub(start)

			assert.Error(t, err)
			assert.EqualError(t, err, "Kyma prerequisites deployment failed due to the timeout")

			t.Logf("Elapsed time: %v", elapsed.Seconds())
			// Cancel timeout occurs at 150 ms
			// Quit timeout occurs at 250 ms
			// Blocking process (single component deployment) takes about 201[ms]
			// Quit condition should be detected before processing next component.
			// Check if program quits as expected after cancel timeout
			assert.GreaterOrEqual(t, elapsed.Milliseconds(), int64(150))
			assert.Less(t, elapsed.Milliseconds(), int64(220))
		})
		t.Run("due to quit timeout", func(t *testing.T) {
			hc := &mockHelmClient{
				componentProcessingTime: 300,
			}
			provider := &mockProvider{
				hc: hc,
			}
			overridesProvider := &mockOverridesProvider{}
			eng := engine.NewEngine(overridesProvider, provider, engine.Config{
				WorkersCount: 2,
				Log:          logger.NewLogger(true),
			})

			start := time.Now()
			err := i.startKymaDeployment(provider, overridesProvider, eng)
			end := time.Now()

			elapsed := end.Sub(start)

			assert.Error(t, err)
			assert.EqualError(t, err, "Force quit: Kyma prerequisites deployment failed due to the timeout")

			t.Logf("Elapsed time: %v", elapsed.Seconds())
			// One component deployment lasts 300 ms
			// Quit timeout occurs at 250 ms
			// Check if program ends just after quit timeout
			assert.GreaterOrEqual(t, elapsed.Milliseconds(), int64(250))
			assert.Less(t, elapsed.Milliseconds(), int64(260))
		})
	})

	t.Run("should deploy prerequisites and fail to deploy Kyma components", func(t *testing.T) {
		t.Run("due to cancel timeout", func(t *testing.T) {
			hc := &mockHelmClient{
				componentProcessingTime: 40,
			}
			provider := &mockProvider{
				hc: hc,
			}
			overridesProvider := &mockOverridesProvider{}
			eng := engine.NewEngine(overridesProvider, provider, engine.Config{
				WorkersCount: 2,
				Log:          logger.NewLogger(true),
			})

			start := time.Now()
			err := i.startKymaDeployment(provider, overridesProvider, eng)
			end := time.Now()

			elapsed := end.Sub(start)

			assert.Error(t, err)
			assert.EqualError(t, err, "Kyma deployment failed due to the timeout")

			t.Logf("Elapsed time: %v", elapsed.Seconds())
			// Cancel timeout occurs at 150 ms
			// Quit timeout occurs at 250 ms
			// Blocking process (component deployment) ends in the meantime (it's a multiple of 41[ms])
			// Check if program quits as expected after cancel timeout and before quit timeout
			assert.GreaterOrEqual(t, elapsed.Milliseconds(), int64(150))
			assert.Less(t, elapsed.Milliseconds(), int64(190))
		})
		t.Run("due to quit timeout", func(t *testing.T) {
			inst := newDeployment(t, nil, kubeClient)

			// Changing it to higher amounts to minimize difference between cancel and quit timeout
			// and give program enough time to process
			inst.cfg.CancelTimeout = 240 * time.Millisecond
			inst.cfg.QuitTimeout = 260 * time.Millisecond

			hc := &mockHelmClient{
				componentProcessingTime: 70,
			}
			provider := &mockProvider{
				hc: hc,
			}
			overridesProvider := &mockOverridesProvider{}
			eng := engine.NewEngine(overridesProvider, provider, engine.Config{
				WorkersCount: 2,
				Log:          logger.NewLogger(true),
			})

			start := time.Now()
			err := inst.startKymaDeployment(provider, overridesProvider, eng)
			end := time.Now()

			elapsed := end.Sub(start)

			assert.Error(t, err)
			assert.EqualError(t, err, "Force quit: Kyma deployment failed due to the timeout")

			// Prerequisites and two components deployment lasts over 280 ms (multiple of 71[ms], 2 workers deploying components in parallel)
			// Quit timeout occurs at 260 ms
			// Check if program ends just after quit timeout
			assert.GreaterOrEqual(t, elapsed.Milliseconds(), int64(260))
			assert.Less(t, elapsed.Milliseconds(), int64(270))
		})
	})
}

// Pass optionally an receiver-channel to get progress updates
func newDeployment(t *testing.T, procUpdates chan<- ProcessUpdate, kubeClient kubernetes.Interface) *Deployment {
	config := &config.Config{
		CancelTimeout:                 cancelTimeout,
		QuitTimeout:                   quitTimeout,
		BackoffInitialIntervalSeconds: 1,
		BackoffMaxElapsedTimeSeconds:  1,
		Log:                           logger.NewLogger(true),
		ComponentsListFile:            "../test/data/componentlist.yaml",
	}
	core, err := newCore(config, &OverridesBuilder{}, kubeClient, procUpdates)
	if err != nil {
		assert.NoError(t, err)
	}
	return &Deployment{core}
}

package deployment

import (
	"fmt"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/components"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/logger"
	log "github.com/sirupsen/logrus"
	"sync"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/config"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/engine"
	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"

	"testing"
	"time"
)

func callbackUpdate(update ProcessUpdate) {

	showCompStatus := func(comp components.KymaComponent) {
		if comp.Name != "" {
			log.Infof("Status of component '%s': %s", comp.Name, comp.Status)
		}
	}

	switch update.Event {
	case ProcessStart:
		log.Infof("Starting installation phase '%s'", update.Phase)
	case ProcessRunning:
		showCompStatus(update.Component)
	case ProcessFinished:
		log.Infof("Finished installation phase '%s' successfully", update.Phase)
	default:
		//any failure case
		log.Infof("Process failed in phase '%s' with error state '%s':", update.Phase, update.Event)
		showCompStatus(update.Component)
	}
}

func TestDeployment_RetrieveProgressUpdates(t *testing.T) {

	//verify we received all expected events
	receivedEvents := make(map[string]int)
	var wg sync.WaitGroup
	wg.Add(1)

	mutex := &sync.Mutex{}

	procUpd := func(procUpd ProcessUpdate){
		mutex.Lock()
		receivedEvents[processUpdateString(procUpd)]++
		mutex.Unlock()
	}

	kubeClient := fake.NewSimpleClientset()

	inst := newDeployment(t, procUpd, kubeClient)

	hc := &mockHelmClient{}
	provider := &mockProvider{
		hc: hc,
	}
	overridesProvider := &mockOverridesProvider{}
	prerequisitesEng := engine.NewEngine(overridesProvider, provider, engine.Config{
		WorkersCount: 1,
		Log:          logger.NewLogger(true),
	})
	componentsEng := engine.NewEngine(overridesProvider, provider, engine.Config{
		WorkersCount: 2,
		Log:          logger.NewLogger(true),
	})

	// blocking function call here. Exits when done.
	err := inst.startKymaDeployment(overridesProvider, prerequisitesEng, componentsEng)
	assert.NoError(t, err)

	expectedEvents := []string{
		"InstallPreRequisites-ProcessStart",
		"InstallPreRequisites-ProcessRunning-test1-Installed",
		"InstallPreRequisites-ProcessRunning-test2-Installed",
		"InstallPreRequisites-ProcessRunning-test3-Installed",
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
		prerequisitesEng := engine.NewEngine(overridesProvider, provider, engine.Config{
			WorkersCount: 1,
			Log:          logger.NewLogger(true),
		})
		componentsEng := engine.NewEngine(overridesProvider, provider, engine.Config{
			WorkersCount: 2,
			Log:          logger.NewLogger(true),
		})

		err := i.startKymaDeployment(overridesProvider, prerequisitesEng, componentsEng)

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
			prerequisitesEng := engine.NewEngine(overridesProvider, provider, engine.Config{
				WorkersCount: 1,
				Log:          logger.NewLogger(true),
			})
			componentsEng := engine.NewEngine(overridesProvider, provider, engine.Config{
				WorkersCount: 2,
				Log:          logger.NewLogger(true),
			})

			start := time.Now()
			err := i.startKymaDeployment(overridesProvider, prerequisitesEng, componentsEng)
			end := time.Now()

			elapsed := end.Sub(start)

			assert.Error(t, err)
			assert.EqualError(t, err, "Kyma deployment failed due to the timeout")

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
			prerequisitesEng := engine.NewEngine(overridesProvider, provider, engine.Config{
				WorkersCount: 1,
				Log:          logger.NewLogger(true),
			})
			componentsEng := engine.NewEngine(overridesProvider, provider, engine.Config{
				WorkersCount: 2,
				Log:          logger.NewLogger(true),
			})

			start := time.Now()
			err := i.startKymaDeployment(overridesProvider, prerequisitesEng, componentsEng)
			end := time.Now()

			elapsed := end.Sub(start)

			assert.Error(t, err)
			assert.EqualError(t, err, "Force quit: Kyma deployment failed due to the timeout")

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
			prerequisitesEng := engine.NewEngine(overridesProvider, provider, engine.Config{
				WorkersCount: 1,
				Log:          logger.NewLogger(true),
			})
			componentsEng := engine.NewEngine(overridesProvider, provider, engine.Config{
				WorkersCount: 2,
				Log:          logger.NewLogger(true),
			})

			start := time.Now()
			err := i.startKymaDeployment(overridesProvider, prerequisitesEng, componentsEng)
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
			prerequisitesEng := engine.NewEngine(overridesProvider, provider, engine.Config{
				WorkersCount: 1,
				Log:          logger.NewLogger(true),
			})
			componentsEng := engine.NewEngine(overridesProvider, provider, engine.Config{
				WorkersCount: 2,
				Log:          logger.NewLogger(true),
			})

			start := time.Now()
			err := inst.startKymaDeployment(overridesProvider, prerequisitesEng, componentsEng)
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
func newDeployment(t *testing.T, procUpdates func(ProcessUpdate), kubeClient kubernetes.Interface) *Deployment {
	compList, err := config.NewComponentList("../test/data/componentlist.yaml")
	assert.NoError(t, err)
	config := &config.Config{
		CancelTimeout:                 cancelTimeout,
		QuitTimeout:                   quitTimeout,
		BackoffInitialIntervalSeconds: 1,
		BackoffMaxElapsedTimeSeconds:  1,
		Log:                           logger.NewLogger(true),
		ComponentList:                 compList,
	}
	core, err := newCore(config, &OverridesBuilder{}, kubeClient, procUpdates)
	assert.NoError(t, err)
	return &Deployment{core}
}

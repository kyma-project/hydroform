package deployment

import (
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/config"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/engine"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/logger"
	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"

	"testing"
	"time"
)

func TestDeployment_StartKymaUninstallation(t *testing.T) {

	kubeClient := fake.NewSimpleClientset()
	i := newDeletion(t, nil, kubeClient)

	t.Run("should uninstall Kyma", func(t *testing.T) {
		hc := &mockHelmClient{}
		provider := &mockProvider{
			hc: hc,
		}
		overridesProvider := &mockOverridesProvider{}
		eng := engine.NewEngine(overridesProvider, provider, engine.Config{
			WorkersCount: 2,
			Log:          logger.NewLogger(true),
		})

		err := i.startKymaUninstallation(provider, eng)

		assert.NoError(t, err)
	})

	t.Run("should fail to uninstall Kyma components", func(t *testing.T) {
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
			err := i.startKymaUninstallation(provider, eng)
			end := time.Now()

			elapsed := end.Sub(start)

			assert.Error(t, err)
			assert.EqualError(t, err, "Kyma uninstallation failed due to the timeout")

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
			err := i.startKymaUninstallation(provider, eng)
			end := time.Now()

			elapsed := end.Sub(start)

			assert.Error(t, err)
			assert.EqualError(t, err, "Force quit: Kyma uninstallation failed due to the timeout")

			t.Logf("Elapsed time: %v", elapsed.Seconds())
			// One component deployment lasts 300 ms
			// Quit timeout occurs at 250 ms
			// Check if program ends just after quit timeout
			assert.GreaterOrEqual(t, elapsed.Milliseconds(), int64(250))
			assert.Less(t, elapsed.Milliseconds(), int64(260))
		})
	})

	t.Run("should uninstall components and fail to deploy Kyma prerequisites", func(t *testing.T) {
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
			err := i.startKymaUninstallation(provider, eng)
			end := time.Now()

			elapsed := end.Sub(start)

			assert.Error(t, err)
			assert.EqualError(t, err, "Kyma prerequisites uninstallation failed due to the timeout")

			t.Logf("Elapsed time: %v", elapsed.Seconds())
			// Cancel timeout occurs at 150 ms
			// Quit timeout occurs at 250 ms
			// Blocking process (component deployment) ends in the meantime (it's a multiple of 41[ms])
			// Check if program quits as expected after cancel timeout and before quit timeout
			assert.GreaterOrEqual(t, elapsed.Milliseconds(), int64(150))
			assert.Less(t, elapsed.Milliseconds(), int64(190))
		})
		t.Run("due to quit timeout", func(t *testing.T) {
			kubeClient := fake.NewSimpleClientset()

			inst := newDeletion(t, nil, kubeClient)

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
			err := inst.startKymaUninstallation(provider, eng)
			end := time.Now()

			elapsed := end.Sub(start)

			assert.Error(t, err)
			assert.EqualError(t, err, "Force quit: Kyma prerequisites uninstallation failed due to the timeout")

			t.Logf("Elapsed time: %v", elapsed.Seconds())
			// Prerequisites and two components deployment lasts over 280 ms (multiple of 71[ms], 2 workers uninstalling components in parallel)
			// Quit timeout occurs at 260 ms
			// Check if program ends just after quit timeout
			assert.GreaterOrEqual(t, elapsed.Milliseconds(), int64(260))
			assert.Less(t, elapsed.Milliseconds(), int64(270))
		})
	})
}

// Pass optionally an receiver-channel to get progress updates
func newDeletion(t *testing.T, procUpdates chan<- ProcessUpdate, kubeClient kubernetes.Interface) *Deletion {
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
	return &Deletion{core}
}

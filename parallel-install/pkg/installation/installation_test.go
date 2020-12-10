package installation

import (
	"context"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/components"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/config"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/engine"
	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/kubernetes/fake"

	//"k8s.io/client-go/kubernetes/fake"
	"log"
	"testing"
	"time"
)

const cancelTimeout = 150 * time.Millisecond
const quitTimeout = 250 * time.Millisecond

func TestInstallation_StartKymaInstallation(t *testing.T) {
	t.Parallel()

	i := newInstallation()

	t.Run("should install Kyma", func(t *testing.T) {
		kubeClient := fake.NewSimpleClientset()

		hc := &mockHelmClient{}
		provider := &mockProvider{
			hc: hc,
		}
		overridesProvider := &mockOverridesProvider{}
		eng := engine.NewEngine(overridesProvider, provider, engine.Config{
			WorkersCount: 2,
			Log:          log.Printf,
		})

		err := i.startKymaInstallation(kubeClient, provider, overridesProvider, eng)

		assert.NoError(t, err)
	})

	t.Run("should fail to install Kyma prerequisites", func(t *testing.T) {
		t.Run("due to cancel timeout", func(t *testing.T) {
			kubeClient := fake.NewSimpleClientset()

			hc := &mockHelmClient{
				componentProcessingTime: 200,
			}
			provider := &mockProvider{
				hc: hc,
			}
			overridesProvider := &mockOverridesProvider{}
			eng := engine.NewEngine(overridesProvider, provider, engine.Config{
				WorkersCount: 2,
				Log:          log.Printf,
			})

			start := time.Now()
			err := i.startKymaInstallation(kubeClient, provider, overridesProvider, eng)
			end := time.Now()

			elapsed := end.Sub(start)

			assert.Error(t, err)
			assert.EqualError(t, err, "Kyma prerequisites installation failed due to the timeout")

			t.Logf("Elapsed time: %v", elapsed.Seconds())
			// Cancel timeout occurs at 150 ms
			// Quit timeout occurs at 250 ms
			// Blocking process (single component installation) takes about 201[ms]
			// Quit condition should be detected before processing next component.
			// Check if program quits as expected after cancel timeout
			assert.GreaterOrEqual(t, elapsed.Milliseconds(), int64(150))
			assert.Less(t, elapsed.Milliseconds(), int64(220))
		})
		t.Run("due to quit timeout", func(t *testing.T) {
			kubeClient := fake.NewSimpleClientset()

			hc := &mockHelmClient{
				componentProcessingTime: 300,
			}
			provider := &mockProvider{
				hc: hc,
			}
			overridesProvider := &mockOverridesProvider{}
			eng := engine.NewEngine(overridesProvider, provider, engine.Config{
				WorkersCount: 2,
				Log:          log.Printf,
			})

			start := time.Now()
			err := i.startKymaInstallation(kubeClient, provider, overridesProvider, eng)
			end := time.Now()

			elapsed := end.Sub(start)

			assert.Error(t, err)
			assert.EqualError(t, err, "Force quit: Kyma prerequisites installation failed due to the timeout")

			t.Logf("Elapsed time: %v", elapsed.Seconds())
			// One component installation lasts 300 ms
			// Quit timeout occurs at 250 ms
			// Check if program ends just after quit timeout
			assert.GreaterOrEqual(t, elapsed.Milliseconds(), int64(250))
			assert.Less(t, elapsed.Milliseconds(), int64(260))
		})
	})

	t.Run("should install prerequisites and fail to install Kyma components", func(t *testing.T) {
		t.Run("due to cancel timeout", func(t *testing.T) {
			kubeClient := fake.NewSimpleClientset()

			hc := &mockHelmClient{
				componentProcessingTime: 40,
			}
			provider := &mockProvider{
				hc: hc,
			}
			overridesProvider := &mockOverridesProvider{}
			eng := engine.NewEngine(overridesProvider, provider, engine.Config{
				WorkersCount: 2,
				Log:          log.Printf,
			})

			start := time.Now()
			err := i.startKymaInstallation(kubeClient, provider, overridesProvider, eng)
			end := time.Now()

			elapsed := end.Sub(start)

			assert.Error(t, err)
			assert.EqualError(t, err, "Kyma installation failed due to the timeout")

			t.Logf("Elapsed time: %v", elapsed.Seconds())
			// Cancel timeout occurs at 150 ms
			// Quit timeout occurs at 250 ms
			// Blocking process (component installation) ends in the meantime (it's a multiple of 41[ms])
			// Check if program quits as expected after cancel timeout and before quit timeout
			assert.GreaterOrEqual(t, elapsed.Milliseconds(), int64(150))
			assert.Less(t, elapsed.Milliseconds(), int64(190))
		})
		t.Run("due to quit timeout", func(t *testing.T) {
			kubeClient := fake.NewSimpleClientset()

			inst := newInstallation()

			// Changing it to higher amounts to minimize difference between cancel and quit timeout
			// and give program enough time to process
			inst.Cfg.CancelTimeout = 240 * time.Millisecond
			inst.Cfg.QuitTimeout = 260 * time.Millisecond

			hc := &mockHelmClient{
				componentProcessingTime: 70,
			}
			provider := &mockProvider{
				hc: hc,
			}
			overridesProvider := &mockOverridesProvider{}
			eng := engine.NewEngine(overridesProvider, provider, engine.Config{
				WorkersCount: 2,
				Log:          log.Printf,
			})

			start := time.Now()
			err := inst.startKymaInstallation(kubeClient, provider, overridesProvider, eng)
			end := time.Now()

			elapsed := end.Sub(start)

			assert.Error(t, err)
			assert.EqualError(t, err, "Force quit: Kyma installation failed due to the timeout")

			// Prerequisites and two components installation lasts over 280 ms (multiple of 71[ms], 2 workers installing components in parallel)
			// Quit timeout occurs at 260 ms
			// Check if program ends just after quit timeout
			assert.GreaterOrEqual(t, elapsed.Milliseconds(), int64(260))
			assert.Less(t, elapsed.Milliseconds(), int64(270))
		})
	})
}

func TestInstallation_StartKymaUninstallation(t *testing.T) {

	i := newInstallation()

	t.Run("should uninstall Kyma", func(t *testing.T) {
		kubeClient := fake.NewSimpleClientset()

		hc := &mockHelmClient{}
		provider := &mockProvider{
			hc: hc,
		}
		overridesProvider := &mockOverridesProvider{}
		eng := engine.NewEngine(overridesProvider, provider, engine.Config{
			WorkersCount: 2,
			Log:          log.Printf,
		})

		err := i.startKymaUninstallation(kubeClient, provider, eng)

		assert.NoError(t, err)
	})

	t.Run("should fail to uninstall Kyma components", func(t *testing.T) {
		t.Run("due to cancel timeout", func(t *testing.T) {
			kubeClient := fake.NewSimpleClientset()

			hc := &mockHelmClient{
				componentProcessingTime: 200,
			}
			provider := &mockProvider{
				hc: hc,
			}
			overridesProvider := &mockOverridesProvider{}
			eng := engine.NewEngine(overridesProvider, provider, engine.Config{
				WorkersCount: 2,
				Log:          log.Printf,
			})

			start := time.Now()
			err := i.startKymaUninstallation(kubeClient, provider, eng)
			end := time.Now()

			elapsed := end.Sub(start)

			assert.Error(t, err)
			assert.EqualError(t, err, "Kyma uninstallation failed due to the timeout")

			t.Logf("Elapsed time: %v", elapsed.Seconds())
			// Cancel timeout occurs at 150 ms
			// Quit timeout occurs at 250 ms
			// Blocking process (single component installation) takes about 201[ms]
			// Quit condition should be detected before processing next component.
			// Check if program quits as expected after cancel timeout
			assert.GreaterOrEqual(t, elapsed.Milliseconds(), int64(150))
			assert.Less(t, elapsed.Milliseconds(), int64(220))
		})
		t.Run("due to quit timeout", func(t *testing.T) {
			kubeClient := fake.NewSimpleClientset()

			hc := &mockHelmClient{
				componentProcessingTime: 300,
			}
			provider := &mockProvider{
				hc: hc,
			}
			overridesProvider := &mockOverridesProvider{}
			eng := engine.NewEngine(overridesProvider, provider, engine.Config{
				WorkersCount: 2,
				Log:          log.Printf,
			})

			start := time.Now()
			err := i.startKymaUninstallation(kubeClient, provider, eng)
			end := time.Now()

			elapsed := end.Sub(start)

			assert.Error(t, err)
			assert.EqualError(t, err, "Force quit: Kyma uninstallation failed due to the timeout")

			t.Logf("Elapsed time: %v", elapsed.Seconds())
			// One component installation lasts 300 ms
			// Quit timeout occurs at 250 ms
			// Check if program ends just after quit timeout
			assert.GreaterOrEqual(t, elapsed.Milliseconds(), int64(250))
			assert.Less(t, elapsed.Milliseconds(), int64(260))
		})
	})

	t.Run("should uninstall components and fail to install Kyma prerequisites", func(t *testing.T) {
		t.Run("due to cancel timeout", func(t *testing.T) {
			kubeClient := fake.NewSimpleClientset()

			hc := &mockHelmClient{
				componentProcessingTime: 40,
			}
			provider := &mockProvider{
				hc: hc,
			}
			overridesProvider := &mockOverridesProvider{}
			eng := engine.NewEngine(overridesProvider, provider, engine.Config{
				WorkersCount: 2,
				Log:          log.Printf,
			})

			start := time.Now()
			err := i.startKymaUninstallation(kubeClient, provider, eng)
			end := time.Now()

			elapsed := end.Sub(start)

			assert.Error(t, err)
			assert.EqualError(t, err, "Kyma prerequisites uninstallation failed due to the timeout")

			t.Logf("Elapsed time: %v", elapsed.Seconds())
			// Cancel timeout occurs at 150 ms
			// Quit timeout occurs at 250 ms
			// Blocking process (component installation) ends in the meantime (it's a multiple of 41[ms])
			// Check if program quits as expected after cancel timeout and before quit timeout
			assert.GreaterOrEqual(t, elapsed.Milliseconds(), int64(150))
			assert.Less(t, elapsed.Milliseconds(), int64(190))
		})
		t.Run("due to quit timeout", func(t *testing.T) {
			kubeClient := fake.NewSimpleClientset()

			inst := newInstallation()

			// Changing it to higher amounts to minimize difference between cancel and quit timeout
			// and give program enough time to process
			inst.Cfg.CancelTimeout = 240 * time.Millisecond
			inst.Cfg.QuitTimeout = 260 * time.Millisecond

			hc := &mockHelmClient{
				componentProcessingTime: 70,
			}
			provider := &mockProvider{
				hc: hc,
			}
			overridesProvider := &mockOverridesProvider{}
			eng := engine.NewEngine(overridesProvider, provider, engine.Config{
				WorkersCount: 2,
				Log:          log.Printf,
			})

			start := time.Now()
			err := inst.startKymaUninstallation(kubeClient, provider, eng)
			end := time.Now()

			elapsed := end.Sub(start)

			assert.Error(t, err)
			assert.EqualError(t, err, "Force quit: Kyma prerequisites uninstallation failed due to the timeout")

			t.Logf("Elapsed time: %v", elapsed.Seconds())
			// Prerequisites and two components installation lasts over 280 ms (multiple of 71[ms], 2 workers uninstalling components in parallel)
			// Quit timeout occurs at 260 ms
			// Check if program ends just after quit timeout
			assert.GreaterOrEqual(t, elapsed.Milliseconds(), int64(260))
			assert.Less(t, elapsed.Milliseconds(), int64(270))
		})
	})
}

type mockHelmClient struct {
	componentProcessingTime int
}

func (c *mockHelmClient) InstallRelease(ctx context.Context, chartDir, namespace, name string, overrides map[string]interface{}) error {
	time.Sleep(1 * time.Millisecond)
	time.Sleep(time.Duration(c.componentProcessingTime) * time.Millisecond)
	return nil
}
func (c *mockHelmClient) UninstallRelease(ctx context.Context, namespace, name string) error {
	time.Sleep(1 * time.Millisecond)
	time.Sleep(time.Duration(c.componentProcessingTime) * time.Millisecond)
	return nil
}

func newInstallation() Installation {
	return Installation{

		Cfg: config.Config{
			CancelTimeout: cancelTimeout,
			QuitTimeout:   quitTimeout,
			Log:           log.Printf,
		},
	}
}

type mockProvider struct {
	hc *mockHelmClient
}

func (p *mockProvider) GetComponents() ([]components.Component, error) {
	return []components.Component{
		{
			Name:            "test1",
			Namespace:       "test1",
			OverridesGetter: func() map[string]interface{} { return nil },
			HelmClient:      p.hc,
			Log:             log.Printf,
		},
		{
			Name:            "test2",
			Namespace:       "test2",
			OverridesGetter: func() map[string]interface{} { return nil },
			HelmClient:      p.hc,
			Log:             log.Printf,
		},
		{
			Name:            "test3",
			Namespace:       "test3",
			OverridesGetter: func() map[string]interface{} { return nil },
			HelmClient:      p.hc,
			Log:             log.Printf,
		},
	}, nil
}

type mockOverridesProvider struct{}

func (o *mockOverridesProvider) OverridesGetterFunctionFor(name string) func() map[string]interface{} {
	return nil
}
func (o *mockOverridesProvider) ReadOverridesFromCluster() error {
	return nil
}

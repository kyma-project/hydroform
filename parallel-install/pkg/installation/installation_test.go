package installation

import (
	"context"
	"fmt"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/components"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/config"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/engine"
	"github.com/stretchr/testify/assert"
	//"k8s.io/client-go/kubernetes/fake"
	"log"
	"testing"
	"time"
)

const cancelTimeout = 1
const quitTimeout = 2
const cancelTimeoutMillisecond = time.Duration(50) * time.Millisecond
const quitTimeoutMillisecond = time.Duration(150) * time.Millisecond

func TestInstallation_StartKymaInstallation(t *testing.T) {
	t.Parallel()
	i := newInstallation()

	t.Run("should install Kyma", func(t *testing.T) {
		hc := &mockHelmClient{}
		provider := &mockProvider{
			hc: hc,
		}
		overridesProvider := &mockOverridesProvider{}
		eng := engine.NewEngine(overridesProvider, provider, "", engine.Config{
			WorkersCount: 2,
			Log:          log.Printf,
		})

		err := i.StartKymaInstallation(provider, overridesProvider, eng)

		assert.NoError(t, err)
	})

	t.Run("should fail to install Kyma prerequisites due to cancel timeout", func(t *testing.T) {
		hc := &mockHelmClient{
			cancelWithTimeout: true,
		}
		provider := &mockProvider{
			hc: hc,
		}
		overridesProvider := &mockOverridesProvider{}
		eng := engine.NewEngine(overridesProvider, provider, "", engine.Config{
			WorkersCount: 1,
			Log:          log.Printf,
		})

		err := i.StartKymaInstallation(provider, overridesProvider, eng)

		assert.Error(t, err)
		assert.EqualError(t, err, "Kyma prerequisites installation failed due to the timeout")
	})

	t.Run("should fail to install Kyma prerequisites due to quit timeout", func(t *testing.T) {
		hc := &mockHelmClient{
			quitWithTimeout: true,
		}
		provider := &mockProvider{
			hc: hc,
		}
		overridesProvider := &mockOverridesProvider{}
		eng := engine.NewEngine(overridesProvider, provider, "", engine.Config{
			WorkersCount: 1,
			Log:          log.Printf,
		})

		err := i.StartKymaInstallation(provider, overridesProvider, eng)

		assert.Error(t, err)
		assert.EqualError(t, err, "Force quit: Kyma prerequisites installation failed due to the timeout")
	})

	t.Run("should install prerequisites and fail to install Kyma components due to timeout", func(t *testing.T) {
		hc := &mockHelmClient{
			cancelAfterFirstComponent: true,
		}
		provider := &mockProvider{
			hc: hc,
		}
		overridesProvider := &mockOverridesProvider{}
		eng := engine.NewEngine(overridesProvider, provider, "", engine.Config{
			WorkersCount: 1,
			Log:          log.Printf,
		})

		err := i.StartKymaInstallation(provider, overridesProvider, eng)

		fmt.Println(err)

		assert.Error(t, err)
		assert.EqualError(t, err, "Kyma installation failed due to the timeout")
	})
}

//func Test_installPrerequisites(t *testing.T) {
//	i := newInstallation()
//
//	t.Run("should install prerequisites with no error", func(t *testing.T) {
//		ctx, cancelFunc := context.WithCancel(context.TODO())
//		defer cancelFunc()
//
//		hc := &mockHelmClient{}
//		comps := newComponents(hc)
//
//		//when
//		err := i.installPrerequisites(ctx, cancelFunc, comps, cancelTimeoutMillisecond, quitTimeoutMillisecond)
//
//		//then
//		assert.NoError(t, err)
//	})
//
//	t.Run("should cancel installation after given timeout and exit with error", func(t *testing.T) {
//		ctx, cancelFunc := context.WithCancel(context.TODO())
//		defer cancelFunc()
//
//		hc := &mockHelmClient{
//			cancelWithTimeout: true,
//		}
//		comps := newComponents(hc)
//
//		//when
//		err := i.installPrerequisites(ctx, cancelFunc, comps, cancelTimeoutMillisecond, quitTimeoutMillisecond)
//
//		//then
//		assert.Error(t, err)
//		assert.EqualError(t, err, "Kyma installation failed due to the timeout")
//	})
//
//	t.Run("should quit installation after given timeout and exit with error", func(t *testing.T) {
//		ctx, cancelFunc := context.WithCancel(context.TODO())
//		defer cancelFunc()
//
//		hc := &mockHelmClient{
//			quitWithTimeout: true,
//		}
//		comps := newComponents(hc)
//
//		//when
//		err := i.installPrerequisites(ctx, cancelFunc, comps, cancelTimeoutMillisecond, quitTimeoutMillisecond)
//
//		//then
//		assert.Error(t, err)
//		assert.EqualError(t, err, "Force quit: Kyma installation failed due to the timeout")
//	})
//
//	t.Run("should not install next components after timeout", func(t *testing.T) {
//		ctx, cancelFunc := context.WithCancel(context.TODO())
//		defer cancelFunc()
//
//		hc := &mockHelmClient{
//			cancelAfterFirstComponent: true,
//		}
//		comps := newComponents(hc)
//
//		//when
//		startTime := time.Now()
//		err := i.installPrerequisites(ctx, cancelFunc, comps, cancelTimeoutMillisecond, quitTimeoutMillisecond)
//		endTime := time.Now()
//
//		elapsed := endTime.Sub(startTime)
//
//		//then
//		assert.Error(t, err)
//		// One component installation takes 80 ms
//		// So to install one component this assertion must be true
//		// 80 <= componentInstallationTime < 160
//		assert.GreaterOrEqual(t, elapsed.Milliseconds(), int64(80))
//		assert.Less(t, elapsed.Milliseconds(), int64(160))
//	})
//}

type mockHelmClient struct {
	cancelWithTimeout         bool
	quitWithTimeout           bool
	cancelAfterFirstComponent bool
}

func (c *mockHelmClient) InstallRelease(ctx context.Context, chartDir, namespace, name string, overrides map[string]interface{}) error {
	if c.cancelWithTimeout {
		time.Sleep(1100 * time.Millisecond)
	}
	if c.quitWithTimeout {
		time.Sleep(2100 * time.Millisecond)
	}
	if c.cancelAfterFirstComponent {
		time.Sleep(300 * time.Millisecond)
	}
	return nil
}
func (c *mockHelmClient) UninstallRelease(ctx context.Context, namespace, name string) error {
	return nil
}

func newInstallation() Installation {
	return Installation{

		Cfg: config.Config{
			CancelTimeoutSeconds: cancelTimeout,
			QuitTimeoutSeconds:   quitTimeout,
			Log:                  log.Printf,
		},
	}
}

func newComponents(hc *mockHelmClient) []components.Component {
	return []components.Component{
		{
			Name:            "test1",
			Namespace:       "test1",
			OverridesGetter: func() map[string]interface{} { return nil },
			HelmClient:      hc,
			Log:             log.Printf,
		},
		{
			Name:            "test2",
			Namespace:       "test2",
			OverridesGetter: func() map[string]interface{} { return nil },
			HelmClient:      hc,
			Log:             log.Printf,
		},
		{
			Name:            "test3",
			Namespace:       "test3",
			OverridesGetter: func() map[string]interface{} { return nil },
			HelmClient:      hc,
			Log:             log.Printf,
		},
	}

}

type mockProvider struct {
	hc	*mockHelmClient
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

type mockOverridesProvider struct {

}

func (o *mockOverridesProvider) OverridesGetterFunctionFor(name string) func() map[string]interface{} {
	return nil
}
func (o *mockOverridesProvider) ReadOverridesFromCluster() error {
	return nil
}
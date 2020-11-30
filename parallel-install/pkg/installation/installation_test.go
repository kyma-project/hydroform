package installation

import (
	"context"
	"fmt"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/components"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/config"
	"github.com/stretchr/testify/assert"
	//"k8s.io/client-go/kubernetes/fake"
	"log"
	"testing"
	"time"
)

const cancelTimeout = 100
const quitTimeout = 150
const cancelTimeoutMillisecond = time.Duration(50) * time.Millisecond
const quitTimeoutMillisecond = time.Duration(150) * time.Millisecond

func Test_installPrerequisites(t *testing.T) {
	i := newInstallation()

	t.Run("should install prerequisites with no error", func(t *testing.T) {
		ctx, cancelFunc := context.WithCancel(context.TODO())
		defer cancelFunc()

		hc := &mockHelmClient{}
		comps := newComponents(hc)

		//when
		err := i.installPrerequisites(ctx, cancelFunc, comps, cancelTimeoutMillisecond, quitTimeoutMillisecond)

		//then
		assert.NoError(t, err)
	})

	t.Run("should cancel installation after given timeout and exit with error", func(t *testing.T) {
		ctx, cancelFunc := context.WithCancel(context.TODO())
		defer cancelFunc()

		hc := &mockHelmClient{
			cancelWithTimeout: true,
		}
		comps := newComponents(hc)

		//when
		err := i.installPrerequisites(ctx, cancelFunc, comps, cancelTimeoutMillisecond, quitTimeoutMillisecond)

		//then
		assert.Error(t, err)
		assert.EqualError(t, err, "Kyma installation failed due to the timeout")
	})

	t.Run("should quit installation after given timeout and exit with error", func(t *testing.T) {
		ctx, cancelFunc := context.WithCancel(context.TODO())
		defer cancelFunc()

		hc := &mockHelmClient{
			quitWithTimeout: true,
		}
		comps := newComponents(hc)

		//when
		err := i.installPrerequisites(ctx, cancelFunc, comps, cancelTimeoutMillisecond, quitTimeoutMillisecond)

		//then
		assert.Error(t, err)
		assert.EqualError(t, err, "Force quit: Kyma installation failed due to the timeout")
	})

	t.Run("should not install next components after timeout", func(t *testing.T) {
		ctx, cancelFunc := context.WithCancel(context.TODO())
		defer cancelFunc()

		hc := &mockHelmClient{
			cancelAfterFirstComponent: true,
		}
		comps := newComponents(hc)

		//when
		startTime := time.Now()
		err := i.installPrerequisites(ctx, cancelFunc, comps, cancelTimeoutMillisecond, quitTimeoutMillisecond)
		endTime := time.Now()

		elapsed := endTime.Sub(startTime)

		fmt.Printf("%d", elapsed.Milliseconds())

		//then
		assert.Error(t, err)
		// One component installation takes 80 ms
		// So to install one component this assertion must be true
		// 80 <= componentInstallationTime < 160
		assert.GreaterOrEqual(t, elapsed.Milliseconds(), int64(80))
		assert.Less(t, elapsed.Milliseconds(), int64(160))
	})
}

type mockHelmClient struct {
	cancelWithTimeout         bool
	quitWithTimeout           bool
	cancelAfterFirstComponent bool
}

func (c *mockHelmClient) InstallRelease(ctx context.Context, chartDir, namespace, name string, overrides map[string]interface{}) error {
	if c.cancelWithTimeout {
		time.Sleep(120 * time.Millisecond)
	}
	if c.quitWithTimeout {
		time.Sleep(500 * time.Millisecond)
	}
	if c.cancelAfterFirstComponent {
		time.Sleep(80 * time.Millisecond)
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

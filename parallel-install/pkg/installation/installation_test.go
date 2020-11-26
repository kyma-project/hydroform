package installation

import (
	"context"
	"fmt"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/components"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/config"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
	"time"
)

const cancelTimeout = 50
const quitTimeout = 150
const cancelTimeoutMillisecond = time.Duration(50) * time.Millisecond
const quitTimeoutMillisecond = time.Duration(150) * time.Millisecond


func Test_CancelTimeout(t *testing.T) {
	//given
	i := newInstallation()

	ctx, cancelFunc := context.WithCancel(context.TODO())

	hc := &mockHelmClient{
		cancelWithTimeout: true,
	}
	component := newComponent(hc)

	//when
	err := i.installPrerequisites(ctx, cancelFunc, []components.Component{component}, cancelTimeoutMillisecond, quitTimeoutMillisecond)

	//then
	assert.Error(t, err, "Kyma installation failed due to the timeout")
}

func Test_QuitTimeout(t *testing.T) {
	//given
	i := newInstallation()

	ctx, cancelFunc := context.WithCancel(context.TODO())

	hc := &mockHelmClient{
		quitWithTimeout: true,
	}
	component := newComponent(hc)

	//when
	err := i.installPrerequisites(ctx, cancelFunc, []components.Component{component}, cancelTimeoutMillisecond, quitTimeoutMillisecond)

	fmt.Println(err)
	//then
	assert.Error(t, err, "Kyma installation failed due to the timeout")
}

type mockHelmClient struct {
	cancelWithTimeout bool
	quitWithTimeout bool
}

func (c *mockHelmClient) InstallRelease(chartDir, namespace, name string, overrides map[string]interface{}) error {
	if c.cancelWithTimeout {
		time.Sleep(100 * time.Millisecond)
	}
	if c.quitWithTimeout{
		time.Sleep(200 * time.Millisecond)
	}
	return nil
}
func (c *mockHelmClient) UninstallRelease(namespace, name string) error {
	return nil
}

func newInstallation() Installation {
	return Installation{
		Cfg:            config.Config{
			CancelTimeoutSeconds:          cancelTimeout,
			QuitTimeoutSeconds:            quitTimeout,
			Log:                           log.Printf,
		},
	}
}

func newComponent(hc *mockHelmClient) components.Component {
	return components.Component{
		Name:            "test",
		Namespace:       "test",
		OverridesGetter: func() map[string]interface{} { return nil },
		HelmClient:      hc,
		Log:             log.Printf,
	}
}
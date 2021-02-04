package deployment

import (
	"context"
	"log"
	"time"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/components"
)

const cancelTimeout = 150 * time.Millisecond
const quitTimeout = 250 * time.Millisecond

type mockHelmClient struct {
	componentProcessingTime int
}

func (c *mockHelmClient) DeployRelease(ctx context.Context, chartDir, namespace, name string, overrides map[string]interface{}, profile string) error {
	time.Sleep(1 * time.Millisecond)
	time.Sleep(time.Duration(c.componentProcessingTime) * time.Millisecond)
	return nil
}
func (c *mockHelmClient) UninstallRelease(ctx context.Context, namespace, name string) error {
	time.Sleep(1 * time.Millisecond)
	time.Sleep(time.Duration(c.componentProcessingTime) * time.Millisecond)
	return nil
}

type mockProvider struct {
	hc *mockHelmClient
}

func (p *mockProvider) GetComponents() ([]components.KymaComponent, error) {
	return []components.KymaComponent{
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

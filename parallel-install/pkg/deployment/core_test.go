package deployment

import (
	"context"
	"time"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/logger"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/components"
)

//these constants are used in Deletion and Deployment tests
const cancelTimeout = 150 * time.Millisecond
const quitTimeout = 250 * time.Millisecond

//mockHelmClient is used in test-cases of core extending objects, like Deletion an Deployment tests
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

func (c *mockHelmClient) Template(chartDir, namespace, name string, overrides map[string]interface{}, profile string) (string, error) {
	return "Templating is not supported by this mock", nil
}

//mockProvider is used in test-cases of core extending objects, like Deletion an Deployment tests
type mockProvider struct {
	hc *mockHelmClient
}

func (p *mockProvider) GetComponents(reversed bool) []components.KymaComponent {
	return []components.KymaComponent{
		{
			Name:            "test1",
			Namespace:       "test1",
			OverridesGetter: func() map[string]interface{} { return nil },
			HelmClient:      p.hc,
			Log:             logger.NewLogger(true),
		},
		{
			Name:            "test2",
			Namespace:       "test2",
			OverridesGetter: func() map[string]interface{} { return nil },
			HelmClient:      p.hc,
			Log:             logger.NewLogger(true),
		},
		{
			Name:            "test3",
			Namespace:       "test3",
			OverridesGetter: func() map[string]interface{} { return nil },
			HelmClient:      p.hc,
			Log:             logger.NewLogger(true),
		},
	}
}

//mockOverridesProvider is used in test-cases of core extending objects, like Deletion an Deployment tests
type mockOverridesProvider struct{}

func (o *mockOverridesProvider) OverridesGetterFunctionFor(name string) func() map[string]interface{} {
	return nil
}

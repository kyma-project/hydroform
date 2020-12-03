package engine

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/components"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/config"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/helm"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/semaphore"
)

var testComponentsNames = []string{"test0", "test1", "test2", "test3", "test4"}

const (
	componentProcessingTimeInMilliseconds = 50
	defualtWorkersCount                   = 4
)

func TestOneWorkerIsSpawned(t *testing.T) {
	//Test that only one worker is spawned if configured so.
	expected := []bool{true, true, true, true, true}
	tokensAcquiredChan := make(chan bool)

	overridesProvider := &mockOverridesProvider{}

	installationCfg := config.Config{
		WorkersCount: 1,
		Log:          log.Printf,
	}
	hc := &mockHelmClient{
		semaphore:          semaphore.NewWeighted(int64(1)),
		tokensAcquiredChan: tokensAcquiredChan,
	}
	componentsProvider := &mockComponentsProvider{
		hc: hc,
	}
	engineCfg := Config{WorkersCount: installationCfg.WorkersCount, Log: log.Printf}
	e := NewEngine(overridesProvider, componentsProvider, "", engineCfg)

	_, err := e.Install(context.TODO())

	// This variable holds results of semaphore.TryAcquire from mockHelmClient.InstallRelease
	// If token during installation of the nth component was acquired it holds true on the nth position
	var tokensAcquired []bool
Loop:
	for {
		select {
		case tokenAcquired, ok := <-tokensAcquiredChan:
			if ok {
				tokensAcquired = append(tokensAcquired, tokenAcquired)
				if len(tokensAcquired) == len(testComponentsNames) {
					close(tokensAcquiredChan)
				}
			} else {
				break Loop
			}
		}
	}

	require.NoError(t, err)
	// Semaphore has one token in pool
	// If there was more than one worker
	// Some workers wouldn't be able to acquire a token
	require.Equal(t, expected, tokensAcquired)
}
func TestFourWorkersAreSpawned(t *testing.T) {
	//Test that four workers are spawned if configured so.
	expected := []bool{true, true, true, false, true}
	tokensAcquiredChan := make(chan bool)

	overridesProvider := &mockOverridesProvider{}
	installationCfg := config.Config{
		WorkersCount: 4,
		Log:          log.Printf,
	}
	hc := &mockHelmClient{
		semaphore:          semaphore.NewWeighted(int64(3)),
		tokensAcquiredChan: tokensAcquiredChan,
	}
	componentsProvider := &mockComponentsProvider{
		hc: hc,
	}
	engineCfg := Config{WorkersCount: installationCfg.WorkersCount, Log: log.Printf}
	e := NewEngine(overridesProvider, componentsProvider, "", engineCfg)

	_, err := e.Install(context.TODO())

	// This variable holds results of semaphore.TryAcquire from mockHelmClient.InstallRelease
	// If token during installation of the nth component was acquired it holds true on the nth position
	var tokensAcquired []bool
Loop:
	for {
		select {
		case tokenAcquired, ok := <-tokensAcquiredChan:
			if ok {
				tokensAcquired = append(tokensAcquired, tokenAcquired)
				if len(tokensAcquired) == len(testComponentsNames) {
					close(tokensAcquiredChan)
				}
			} else {
				break Loop
			}
		}
	}

	require.NoError(t, err)
	// Semaphore has 3 tokens in pool and there should be 4 workers
	// After completing work workers release their tokens
	// If worker installing 4th component is not able to acquire a token (a "loser"),
	// it means the token pool got exhausted and no tokens were released yet - there are more than 3 workers
	// If worker installing 5th component is able to acquire a token it means a token was released during
	// previous operations and worker didn't start immediately - there are less than 5 workers
	// Caveat:
	// If there are five workers we'd see a pattern: [true, true, true, false, false]
	// The same pattern occurs for four workers, when "loser" worker is processing very fast and is rejected token twice.
	// We ensure the "loser" worker is not rejected the token twice
	// by a making it working for a longer time than the workers that acquired token successfully.
	require.Equal(t, expected, tokensAcquired)
}

func TestSuccessScenario(t *testing.T) {
	//Test success scenario:
	//Expected: All configured components are processed and reported via statusChan
	overridesProvider := &mockOverridesProvider{}

	helmClient := &mockSimpleHelmClient{}

	componentsProvider := &mockComponentsProvider{helmClient}

	componentsToBeProcessed, err := componentsProvider.GetComponents()
	require.NoError(t, err)

	goPath := os.Getenv("GOPATH")
	require.NotEmpty(t, goPath)

	engineCfg := Config{WorkersCount: defualtWorkersCount}

	e := NewEngine(overridesProvider, componentsProvider, "", engineCfg)
	statusChan, err := e.Install(context.TODO())
	require.NoError(t, err)

	waitFor := time.Duration(((len(componentsToBeProcessed)/defualtWorkersCount)+1)*componentProcessingTimeInMilliseconds) * 2 * time.Millisecond // time required to process all components doubled

	// wait until channel is filled with all components' statuses
	require.Eventually(t, func() bool {
		return len(statusChan) == len(componentsToBeProcessed)
	}, waitFor, 10*time.Millisecond)

	// check if each component has status "Installed"
	for componentsCount := 0; componentsCount < len(componentsToBeProcessed); componentsCount++ {
		componentStatus := <-statusChan
		require.Equal(t, components.StatusInstalled, componentStatus.Status)
	}

	// make sure that the status channel does not contain any unexpected statuses
	require.Zero(t, len(statusChan))
}
func TestErrorScenario(t *testing.T) {
	//Test error scenario: Configure some components to report error on install.
	//Expected: All configured components are processed, success and error statuses are reported via statusChan
}
func TestContextCancelScenario(t *testing.T) {
	//Test cancel scenario: Configure two workers and six components (A, B, C, D, E, F), then after B is reported via statusChan, cancel the context.
	//Expected: Components A, B, C, D are reported via statusChan. This is because context is canceled after B, but workers should already start processing C and D.
}

type mockHelmClient struct {
	semaphore          *semaphore.Weighted
	tokensAcquiredChan chan bool
}

func (c *mockHelmClientWithSemaphore) InstallRelease(ctx context.Context, chartDir, namespace, name string, overrides map[string]interface{}) error {
	token := c.semaphore.TryAcquire(1)
	go func() {
		c.tokensAcquiredChan <- token
	}()

	if token {
		time.Sleep(20 * time.Millisecond)
		c.semaphore.Release(1)
	} else {
		time.Sleep(30 * time.Millisecond)
	}
	return nil
}
func (c *mockHelmClientWithSemaphore) UninstallRelease(ctx context.Context, namespace, name string) error {
	time.Sleep(1 * time.Millisecond)
	return nil
}

type mockComponentsProvider struct {
	hc helm.ClientInterface
}

func (p *mockComponentsProvider) GetComponents() ([]components.Component, error) {
	var comps []components.Component
	for _, name := range testComponentsNames {
		component := components.Component{
			Name:            name,
			Namespace:       "test",
			OverridesGetter: func() map[string]interface{} { return nil },
			HelmClient:      p.hc,
			Log:             log.Printf,
		}
		comps = append(comps, component)
	}

	return comps, nil
}

type mockSimpleHelmClient struct{}

func (c *mockSimpleHelmClient) InstallRelease(ctx context.Context, chartDir, namespace, name string, overrides map[string]interface{}) error {
	time.Sleep(1 * time.Millisecond)
	time.Sleep(time.Duration(componentProcessingTimeInMilliseconds) * time.Millisecond)
	return nil
}
func (c *mockSimpleHelmClient) UninstallRelease(ctx context.Context, namespace, name string) error {
	time.Sleep(1 * time.Millisecond)
	time.Sleep(time.Duration(componentProcessingTimeInMilliseconds) * time.Millisecond)
	return nil
}

type mockOverridesProvider struct{}

func (o *mockOverridesProvider) ReadOverridesFromCluster() error {
	return nil
}

func (o *mockOverridesProvider) OverridesGetterFunctionFor(name string) func() map[string]interface{} {
	return nil
}

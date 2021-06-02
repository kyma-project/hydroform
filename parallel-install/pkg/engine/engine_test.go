package engine

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/components"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/config"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/helm"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/logger"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/semaphore"
)

var testComponentsNames = []string{"test0", "test1", "test2", "test3", "test4", "test5"}

const (
	componentProcessingTimeInMilliseconds = 50
	defualtWorkersCount                   = 4
)

func TestOneWorkerIsSpawned(t *testing.T) {
	//Test that only one worker is spawned if configured so.
	expected := []bool{true, true, true, true, true, true}
	tokensAcquiredChan := make(chan bool)

	overridesProvider := &mockOverridesProvider{}

	installationCfg := config.Config{
		WorkersCount: 1,
		Log:          logger.NewLogger(true),
	}
	hc := &mockHelmClientWithSemaphore{
		semaphore:          semaphore.NewWeighted(int64(1)),
		tokensAcquiredChan: tokensAcquiredChan,
	}
	componentsProvider := &mockComponentsProvider{t, hc}
	engineCfg := Config{
		WorkersCount: installationCfg.WorkersCount,
		Log:          installationCfg.Log,
	}
	e := NewEngine(overridesProvider, componentsProvider, engineCfg)

	_, err := e.Deploy(context.TODO())

	// This variable holds results of semaphore.TryAcquire from mockHelmClientWithSemaphore.InstallRelease
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
	expected := []bool{true, true, true, false, true, true}
	tokensAcquiredChan := make(chan bool)

	overridesProvider := &mockOverridesProvider{}
	installationCfg := config.Config{
		WorkersCount: 4,
		Log:          logger.NewLogger(true),
	}
	hc := &mockHelmClientWithSemaphore{
		semaphore:          semaphore.NewWeighted(int64(3)),
		tokensAcquiredChan: tokensAcquiredChan,
	}
	componentsProvider := &mockComponentsProvider{t, hc}
	engineCfg := Config{
		WorkersCount: installationCfg.WorkersCount,
		Log:          installationCfg.Log,
	}
	e := NewEngine(overridesProvider, componentsProvider, engineCfg)

	_, err := e.Deploy(context.TODO())

	// This variable holds results of semaphore.TryAcquire from mockHelmClientWithSemaphore.InstallRelease
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
	//
	// Caveat1:
	// If there are five workers we'd see a pattern: [true, true, true, false, false]
	// The same pattern occurs for four workers, when "loser" worker is processing very fast and is rejected token twice.
	// We ensure the "loser" worker is not rejected the token twice
	// by a making it working significantly longer than the workers that acquired token successfully.
	require.Equal(t, expected, tokensAcquired)
}

func TestSuccessScenario(t *testing.T) {
	//Test success scenario:
	//Expected: All configured components are processed and reported via statusChan
	overridesProvider := &mockOverridesProvider{}

	helmClient := &mockSimpleHelmClient{}

	componentsProvider := &mockComponentsProvider{t, helmClient}

	componentsToBeProcessed := componentsProvider.GetComponents(false)

	engineCfg := Config{
		WorkersCount: defualtWorkersCount,
		Log:          logger.NewLogger(true),
	}
	e := NewEngine(overridesProvider, componentsProvider, engineCfg)
	statusChan, err := e.Deploy(context.TODO())
	require.NoError(t, err)

	maxIterationsForWorker := ((len(componentsToBeProcessed) / defualtWorkersCount) + 1) //at least one worker runs that many iterations.
	processingTimeDoubled := time.Duration(maxIterationsForWorker*componentProcessingTimeInMilliseconds) * 2 * time.Millisecond

	// wait until channel is filled with all components' statuses
	require.Eventually(t, func() bool {
		return len(statusChan) == len(componentsToBeProcessed)
	}, processingTimeDoubled, 10*time.Millisecond, "Invalid statusChan length")

	// check if each component has status "Installed"
	for componentsCount := 0; componentsCount < len(componentsToBeProcessed); componentsCount++ {
		componentStatus := <-statusChan
		require.Equal(t, components.StatusInstalled, componentStatus.Status)
	}

	// make sure that the status channel does not contain any unexpected additional statuses
	require.Zero(t, len(statusChan))
}
func TestErrorScenario(t *testing.T) {
	//Test error scenario: Configure some components to report error on install.
	//Expected: All configured components are processed, success and error statuses are reported via statusChan
	overridesProvider := &mockOverridesProvider{}
	expectedFailedComponents := []string{"test0", "test4"}

	helmClient := &mockSimpleHelmClient{
		componentsToFail: expectedFailedComponents,
	}

	componentsProvider := &mockComponentsProvider{t, helmClient}

	componentsToBeProcessed := componentsProvider.GetComponents(false)

	engineCfg := Config{
		WorkersCount: defualtWorkersCount,
		Log:          logger.NewLogger(true),
	}
	e := NewEngine(overridesProvider, componentsProvider, engineCfg)
	statusChan, err := e.Deploy(context.TODO())
	require.NoError(t, err)

	maxIterationsForWorker := ((len(componentsToBeProcessed) / defualtWorkersCount) + 1) //at least one worker runs that many iterations.
	processingTimeDoubled := time.Duration(maxIterationsForWorker*componentProcessingTimeInMilliseconds) * 2 * time.Millisecond

	// wait until channel is filled with all components' statuses
	require.Eventually(t, func() bool {
		return len(statusChan) == len(componentsToBeProcessed)
	}, processingTimeDoubled, 10*time.Millisecond, "Invalid statusChan length")

	// check if components that should fail have status "Error", and the rest have status "Installed"
	for componentsCount := 0; componentsCount < len(componentsToBeProcessed); componentsCount++ {
		componentStatus := <-statusChan
		if componentStatus.Name == expectedFailedComponents[0] || componentStatus.Name == expectedFailedComponents[1] {
			require.Equal(t, components.StatusError, componentStatus.Status)
		} else {
			require.Equal(t, components.StatusInstalled, componentStatus.Status)
		}
	}

	// make sure that the status channel does not contain any unexpected additional statuses
	require.Zero(t, len(statusChan))
}

func TestContextCancelScenario(t *testing.T) {
	//Test cancel scenario: Configure two workers and six components (A, B, C, D, E, F), then after B is reported via statusChan, cancel the context.
	//Expected: Components A, B, C, D are reported via statusChan. This is because context is canceled after B, but workers should already start processing C and D.
	expectedInstalledComponents := []string{"test0", "test1", "test2", "test3"}
	expectedNotInstalledComponents := []string{"test4", "test5"}

	overridesProvider := &mockOverridesProvider{}
	installationCfg := config.Config{
		WorkersCount: 2,
		Log:          logger.NewLogger(true),
	}
	hc := &mockSimpleHelmClient{}
	componentsProvider := &mockComponentsProvider{t, hc}
	engineCfg := Config{
		WorkersCount: installationCfg.WorkersCount,
		Log:          installationCfg.Log,
	}
	e := NewEngine(overridesProvider, componentsProvider, engineCfg)

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()
	statusChan, err := e.Deploy(ctx)

	require.NoError(t, err)

	var installedComponents []string
Loop:
	for {
		select {
		case component, ok := <-statusChan:
			if ok {
				installedComponents = append(installedComponents, component.Name)
				if component.Name == testComponentsNames[1] {
					//Cancel asynchronously after 10[ms]
					go func() {
						time.Sleep(10 * time.Millisecond)
						cancel()
					}()
				}
			} else {
				break Loop
			}
		}
	}

	require.ElementsMatch(t, installedComponents, expectedInstalledComponents)
	require.NotSubset(t, installedComponents, expectedNotInstalledComponents)
}

type mockHelmClientWithSemaphore struct {
	semaphore          *semaphore.Weighted
	tokensAcquiredChan chan bool
}

func (c *mockHelmClientWithSemaphore) DeployRelease(ctx context.Context, chartDir, namespace, name string, overrides map[string]interface{}, profile string) error {
	token := c.semaphore.TryAcquire(1)

	if token {
		time.Sleep(20 * time.Millisecond)
		c.semaphore.Release(1)
	} else {
		time.Sleep(30 * time.Millisecond)
	}

	go func() {
		c.tokensAcquiredChan <- token
	}()

	return nil
}
func (c *mockHelmClientWithSemaphore) UninstallRelease(ctx context.Context, namespace, name string) error {
	time.Sleep(1 * time.Millisecond)
	return nil
}

type mockComponentsProvider struct {
	t  *testing.T
	hc helm.ClientInterface
}

func (p *mockComponentsProvider) GetComponents(reversed bool) []components.KymaComponent {
	var comps []components.KymaComponent
	for _, name := range testComponentsNames {
		component := components.KymaComponent{
			Name:            name,
			Namespace:       "test",
			OverridesGetter: func() map[string]interface{} { return nil },
			HelmClient:      p.hc,
			Log:             logger.NewLogger(true),
		}
		comps = append(comps, component)
	}

	return comps
}

type mockSimpleHelmClient struct {
	componentsToFail []string
}

func (c *mockSimpleHelmClient) DeployRelease(ctx context.Context, chartDir, namespace, name string, overrides map[string]interface{}, profile string) error {
	time.Sleep(time.Duration(componentProcessingTimeInMilliseconds) * time.Millisecond)
	for i := 0; i < len(c.componentsToFail); i++ {
		if name == c.componentsToFail[i] {
			return fmt.Errorf("failed to install %s", name)
		}
	}
	return nil
}
func (c *mockSimpleHelmClient) UninstallRelease(ctx context.Context, namespace, name string) error {
	time.Sleep(time.Duration(componentProcessingTimeInMilliseconds) * time.Millisecond)
	for i := 0; i < len(c.componentsToFail); i++ {
		if name == c.componentsToFail[i] {
			return fmt.Errorf("failed to uninstall %s", name)
		}
	}
	return nil
}

type mockOverridesProvider struct{}

func (o *mockOverridesProvider) ReadOverridesFromCluster() error {
	return nil
}

func (o *mockOverridesProvider) OverridesGetterFunctionFor(name string) func() map[string]interface{} {
	return nil
}

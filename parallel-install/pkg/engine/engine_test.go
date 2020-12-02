package engine

import (
	"context"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/components"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/config"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/overrides"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"k8s.io/client-go/kubernetes/fake"
	"log"
	"os"
	"path/filepath"
	"testing"
)

func TestOneWorkerIsSpawned(t *testing.T) {
	//Test that only one worker is spawned if configured so.
}
func TestFourWorkersAreSpawned(t *testing.T) {
	//Test that four workers are spawned if configured so.
}
func TestSuccessScenario(t *testing.T) {
	//Test success scenario:
	//Expected: All configured components are processed and reported via statusChan
	k8sMock := fake.NewSimpleClientset()
	overridesProvider, err := overrides.New(k8sMock, []string{""}, log.Printf)
	require.NoError(t, err)

	content, err := ioutil.ReadFile("../test/data/installationCR.yaml")
	require.NoError(t, err)

	installationCfg := config.Config{
		WorkersCount: 4,
		Log: log.Printf,
	}

	componentsProvider := components.NewComponentsProvider(overridesProvider, "", string(content), installationCfg)

	componentsToBeProcessed, err := componentsProvider.GetComponents()
	require.NoError(t, err)

	goPath := os.Getenv("GOPATH")
	require.NotEmpty(t, goPath)
	resourcesPath := filepath.Join(goPath, "src", "github.com", "kyma-project", "kyma", "resources")

	engineCfg := Config{WorkersCount: installationCfg.WorkersCount}

	e := NewEngine(overridesProvider, componentsProvider, resourcesPath, engineCfg)
	statusChan, err := e.Install(context.TODO())
	require.NoError(t, err)

	for componentsCount := 0; componentsCount < len(componentsToBeProcessed); componentsCount++ {
		componentStatus := <-statusChan
		require.Equal(t, components.StatusInstalled, componentStatus.Status)
	}

	//require.Equal(t, len(componentsToBeProcessed), componentsCount)
}
func TestErrorScenario(t *testing.T) {
	//Test error scenario: Configure some components to report error on install.
	//Expected: All configured components are processed, success and error statuses are reported via statusChan
}
func TestContextCancelScenario(t *testing.T) {
	//Test cancel scenario: Configure two workers and six components (A, B, C, D, E, F), then after B is reported via statusChan, cancel the context.
	//Expected: Components A, B, C, D are reported via statusChan. This is because context is canceled after B, but workers should already start processing C and D.
}

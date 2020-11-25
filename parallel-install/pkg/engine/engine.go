package engine

import (
	"context"
	"sync"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/config"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/components"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/overrides"
)

var statusMap map[string]string

const logPrefix = "[engine/engine.go]"

type Config struct {
	WorkersCount int
	Log          func(format string, v ...interface{})
}

type Engine struct {
	overridesProvider     overrides.OverridesProvider
	prerequisitesProvider components.Provider
	componentsProvider    components.Provider
	resourcesPath         string
	cfg                   Config
}

func NewEngine(overridesProvider overrides.OverridesProvider, componentsProvider components.Provider, resourcesPath string, cfg Config) *Engine {
	statusMap = make(map[string]string)
	return &Engine{
		overridesProvider:  overridesProvider,
		componentsProvider: componentsProvider,
		resourcesPath:      resourcesPath,
		cfg:                cfg,
	}
}

type Installation interface {
	//Installs components. ctx is used for cancellation of the operation. returned channel receives every processed component and is closed after all components are processed.
	Install(ctx context.Context) (<-chan components.Component, error)
	//Uninstalls components. ctx is used for cancellation of the operation. returned channel receives every processed component and is closed after all components are processed.
	Uninstall(ctx context.Context) (<-chan components.Component, error)
}

func (e *Engine) Install(ctx context.Context) (<-chan components.Component, error) {

	cmps, err := e.componentsProvider.GetComponents()
	if err != nil {
		return nil, err
	}

	//TODO: Size dependent on number of components?
	statusChan := make(chan components.Component, 30)

	//TODO: Can we avoid this goroutine? Maybe refactor run() so it's non-blocking ?
	go func() {
		defer close(statusChan)

		err = e.overridesProvider.ReadOverridesFromCluster()
		if err != nil {
			e.cfg.Log("%s error while reading overrides: %v", logPrefix, err)
			return
		}

		run(ctx, statusChan, cmps, "install", e.cfg.WorkersCount)

	}()

	return statusChan, nil
}

func (e *Engine) Uninstall(ctx context.Context) (<-chan components.Component, error) {
	cmps, err := e.componentsProvider.GetComponents()
	if err != nil {
		return nil, err
	}

	//TODO: Size dependent on number of components?
	statusChan := make(chan components.Component, 30)

	go func() {
		defer close(statusChan)

		run(ctx, statusChan, cmps, "uninstall", e.cfg.WorkersCount)
	}()

	return statusChan, nil
}

func run(ctx context.Context, statusChan chan<- components.Component, cmps []components.Component, installationType string, concurrency int) {
	//TODO: Size dependent on number of components?
	jobChan := make(chan components.Component, 30)

	//Fill the queue with jobs
	for _, comp := range cmps {
		if !enqueueJob(comp, jobChan) {
			config.Log("%s Max capacity reached, component dismissed: %s", logPrefix, comp.Name)
		}
	}

	//Spawn workers
	var wg sync.WaitGroup

	//TODO: Configurable number of workers
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go worker(ctx, &wg, jobChan, statusChan, installationType)
	}

	// to stop the workers, first close the job channel
	close(jobChan)

	// block until workers quit
	wg.Wait()
}

func worker(ctx context.Context, wg *sync.WaitGroup, jobChan <-chan components.Component, statusChan chan<- components.Component, installationType string) {
	defer wg.Done()

	for {
		select {
		//TODO: Perhaps this should be removed/refactored. Golang choses cases randomly if both are possible, so it might chose processing component instead, and that is invalid.
		case <-ctx.Done():
			config.Log("%s Finishing work: %v", logPrefix, ctx.Err())
			return

		case component, ok := <-jobChan:
			//TODO: Is there a better way to find out if Context is canceled?
			if err := ctx.Err(); err != nil {
				config.Log("%s Finishing work: %v.", logPrefix, err)
				return
			}
			if ok {
				if installationType == "install" {
					if err := component.InstallComponent(); err != nil {
						component.Status = components.StatusError
					} else {
						component.Status = components.StatusInstalled
					}
					statusChan <- component
				} else if installationType == "uninstall" {
					if err := component.UninstallComponent(); err != nil {
						component.Status = components.StatusError
					} else {
						component.Status = components.StatusUninstalled
					}
					statusChan <- component
				}
			} else {
				config.Log("%s Finishing work: no more jobs in queue.", logPrefix)
				return
			}
		}
	}
}

func enqueueJob(job components.Component, jobChan chan<- components.Component) bool {
	select {
	case jobChan <- job:
		return true
	default:
		return false
	}
}

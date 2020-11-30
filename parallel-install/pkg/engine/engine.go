//Package engine implements concurrent processing of components.
//The engine is configured with number of workers that perform operations concurrently.
//If only single worker is configured, the processing becomes sequential.
//If you need different configuration for installation and uninstallation, just create two different Engine instances with different configurations.
//
//The code in the package uses user-provided function for logging.
package engine

import (
	"context"
	"sync"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/config"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/components"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/overrides"
)

const logPrefix = "[engine/engine.go]"

//Config defines configuration values for the engine
type Config struct {
	WorkersCount int                                   //number of concurrent processes for install/uninstall operations
	Log          func(format string, v ...interface{}) //Logging function
}

//Engine implements Installation interface
type Engine struct {
	overridesProvider  overrides.OverridesProvider
	componentsProvider components.Provider
	cfg                Config
}

//NewEngine returns new Engine instance
func NewEngine(overridesProvider overrides.OverridesProvider, componentsProvider components.Provider, cfg Config) *Engine {
	return &Engine{
		overridesProvider:  overridesProvider,
		componentsProvider: componentsProvider,
		cfg:                cfg,
	}
}

//Installation interface defines contract for the Engine
type Installation interface {
	//Install performs concurrent components installation.
	//Errors are not stopping the processing, because it's assumed components are independent of each other,
	//and error condition in one component should not influence others.
	//
	//The returned channel receives every processed component and is closed after all components are processed or the process is cancelled.
	//
	//ctx is used for cancellation of the operation.
	//It is not guaranteed that cancellation is handled immediately, because the underlying Helm operation are blocking and do not support Context-based cancellation.
	//However, once the underlying concurrent operations end, the cancel condition is detected and the return channel is closed.
	//All remaining components are not processed then.
	Install(ctx context.Context) (<-chan components.Component, error)
	//Uninstall performs concurrent components uninstallation.
	//Errors are not stopping the processing, because it's assumed components are independent of each other,
	//and error condition in one component should not influence others.
	//
	//The returned channel receives every processed component and is closed after all components are processed or the process is cancelled.
	//
	//ctx is used for cancellation of the operation.
	//It is not guaranteed that cancellation is handled immediately, because the underlying Helm operation are blocking and do not support Context-based cancellation.
	//However, once the underlying concurrent operations end, the cancel condition is detected and the return channel is closed.
	//All remaining components are not processed then.
	//
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

//Blocking function used to spawn configured number of workers and then await their completion
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

//Non-blocking concurrent worker.
//Detects Context cancellation.
//Context cancellation is not detected immediately, but between component processing operations - because such operations are blocking.
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
					if err := component.InstallComponent(ctx); err != nil {
						component.Status = components.StatusError
					} else {
						component.Status = components.StatusInstalled
					}
					statusChan <- component
				} else if installationType == "uninstall" {
					if err := component.UninstallComponent(ctx); err != nil {
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

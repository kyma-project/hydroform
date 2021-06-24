//Package engine defines the contract and implements parallel processing of components.
//The Engine is configured with a number of workers that run in parallel.
//If only a single worker is configured, the processing becomes sequential.
//If you need different configuration for installation and uninstallation,
//just create two different Engine instances with different configurations.
//
//The code in the package uses the user-provided function for logging.
package engine

import (
	"context"
	"sync"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/logger"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/components"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/jobmanager"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/overrides"
)

const (
	logPrefix                  = "[engine/engine.go]"
	deploy    installationType = "deploy"
	uninstall installationType = "uninstall"
)

type installationType string

//Config defines configuration values for the Engine.
type Config struct {
	WorkersCount int              //Number of parallel processes for install/uninstall operations
	Log          logger.Interface //Logger to be used
}

//Engine implements Installation interface
type Engine struct {
	overridesProvider  overrides.Provider
	componentsProvider components.Provider
	cfg                Config
}

//NewEngine returns new Engine instance
func NewEngine(overridesProvider overrides.Provider, componentsProvider components.Provider, cfg Config) *Engine {
	return &Engine{
		overridesProvider:  overridesProvider,
		componentsProvider: componentsProvider,
		cfg:                cfg,
	}
}

//Installation interface defines contract for the Engine
type Installation interface {
	//Deploy performs parallel components installation.
	//Errors are not stopping the processing because it's assumed components are independent of one another.
	//An error condition in one component should not influence others.
	//
	//The returned channel receives every processed component and is closed after all components are processed or the process is cancelled.
	//
	//ctx is used for the operation cancellation.
	//It is not guaranteed that the cancellation is handled immediately because the underlying Helm operations are blocking and do not support the Context-based cancellation.
	//However, once the underlying parallel operations end, the cancel condition is detected and the return channel is closed.
	//All remaining components are not processed then.
	Deploy(ctx context.Context) (<-chan components.KymaComponent, error)
	//Uninstall performs parallel components uninstallation.
	//Errors are not stopping the processing because it's assumed components are independent of one another.
	//An error condition in one component should not influence others.
	//
	//The returned channel receives every processed component and is closed after all components are processed or the process is cancelled.
	//
	//ctx is used for the operation cancellation.
	//It is not guaranteed that the cancellation is handled immediately because the underlying Helm operations are blocking and do not support the Context-based cancellation.
	//However, once the underlying parallel operations end, the cancel condition is detected and the return channel is closed.
	//All remaining components are not processed then.
	//
	Uninstall(ctx context.Context) (<-chan components.KymaComponent, error)

	//Returns the rendered Kubernetes manifests of Kyma components.
	Manifests() ([]*components.Manifest, error)
}

type manifestResult struct {
	manifest *components.Manifest
	err      error
}

func (e *Engine) Manifests(isPrerequisite bool) ([]*components.Manifest, error) {
	cmps := e.componentsProvider.GetComponents(false)
	manifests := make(chan manifestResult, len(cmps))

	//run templating in parallel
	var wg sync.WaitGroup

	sem := make(chan int, e.cfg.WorkersCount)
	for _, cmp := range cmps {
		wg.Add(1)
		sem <- 1 // will block if there is the limit defined in e.cfg.WorkersCount reached
		go func(cmp components.KymaComponent, manifests chan manifestResult, wg *sync.WaitGroup) {
			manifest, err := cmp.Manifest(isPrerequisite)
			manifests <- manifestResult{
				manifest: manifest,
				err:      err,
			}
			<-sem
			wg.Done()
		}(cmp, manifests, &wg)
	}
	wg.Wait()
	close(manifests)

	//extract manifests and check for errors
	result := []*components.Manifest{}
	for manifestResult := range manifests {
		if manifestResult.err != nil { //rendering of one of the components failed
			return nil, manifestResult.err
		}
		result = append(result, manifestResult.manifest)
	}
	return result, nil
}

func (e *Engine) Deploy(ctx context.Context) (<-chan components.KymaComponent, error) {
	cmps := e.componentsProvider.GetComponents(false)

	//TODO: Size dependent on number of components?
	statusChan := make(chan components.KymaComponent, 30)

	//TODO: Can we avoid this goroutine? Maybe refactor run() so it's non-blocking ?
	go func() {
		defer close(statusChan)

		err := e.overridesProvider.ReadOverridesFromCluster()
		if err != nil {
			e.cfg.Log.Errorf("%s error while reading overrides: %v", logPrefix, err)
			return
		}

		e.run(ctx, statusChan, cmps, deploy)
	}()

	return statusChan, nil
}

func (e *Engine) Uninstall(ctx context.Context) (<-chan components.KymaComponent, error) {
	cmps := e.componentsProvider.GetComponents(true)

	//TODO: Size dependent on number of components?
	statusChan := make(chan components.KymaComponent, 30)

	go func() {
		defer close(statusChan)

		e.run(ctx, statusChan, cmps, uninstall)
	}()

	return statusChan, nil
}

//Blocking function used to spawn a configured number of workers and then await their completion.
func (e *Engine) run(ctx context.Context, statusChan chan<- components.KymaComponent, cmps []components.KymaComponent, installType installationType) {
	//TODO: Size dependent on number of components?
	jobChan := make(chan components.KymaComponent, 30)

	//Fill the queue with jobs
	for _, comp := range cmps {
		if !e.enqueueJob(comp, jobChan) {
			e.cfg.Log.Errorf("%s Max capacity reached, component dismissed: %s", logPrefix, comp.Name)
		}
	}

	//Spawn workers
	var wg sync.WaitGroup

	//TODO: Configurable number of workers
	for i := 0; i < e.cfg.WorkersCount; i++ {
		wg.Add(1)
		go e.worker(ctx, &wg, jobChan, statusChan, installType)
	}

	// to stop the workers, first close the job channel
	close(jobChan)

	// block until workers quit
	wg.Wait()
}

//Non-blocking worker.
//Designed to run in parallel (several workers are processing the same jobChan).
//Detects Context cancellation.
//Context cancellation is not detected immediately. It's detected between component processing operations because such operations are blocking.
//If the Context is cancelled, the worker quits immediately, skipping the remaining components.
func (e *Engine) worker(ctx context.Context, wg *sync.WaitGroup, jobChan <-chan components.KymaComponent, statusChan chan<- components.KymaComponent, installType installationType) {
	defer wg.Done()

	for {
		select {
		//TODO: Perhaps this should be removed/refactored. Golang choses cases randomly if both are possible, so it might chose processing component instead, and that is invalid.
		case <-ctx.Done():
			e.cfg.Log.Infof("%s Finishing work: %v", logPrefix, ctx.Err())
			return

		case component, ok := <-jobChan:
			//TODO: Is there a better way to find out if Context is canceled?
			if err := ctx.Err(); err != nil {
				e.cfg.Log.Infof("%s Finishing work: %v.", logPrefix, err)
				return
			}
			if ok {
				if installType == deploy {
					jobmanager.ExecutePre(ctx, component.Name)
					if err := component.Deploy(ctx); err != nil {
						component.Status = components.StatusError
						component.Error = err
					} else {
						component.Status = components.StatusInstalled
						jobmanager.ExecutePost(ctx, component.Name)
					}
					statusChan <- component
				} else if installType == uninstall {
					if err := component.Uninstall(ctx); err != nil {
						component.Status = components.StatusError
						component.Error = err
					} else {
						component.Status = components.StatusUninstalled
					}
					statusChan <- component
				}
			} else {
				e.cfg.Log.Infof("%s Finishing work: no more jobs in queue.", logPrefix)
				return
			}
		}
	}
}

func (e *Engine) enqueueJob(job components.KymaComponent, jobChan chan<- components.KymaComponent) bool {
	select {
	case jobChan <- job:
		return true
	default:
		return false
	}
}

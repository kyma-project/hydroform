package engine

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/kyma-incubator/hydroform/installation-poc/pkg/components"
	"github.com/kyma-incubator/hydroform/installation-poc/pkg/overrides"
)

var statusMap map[string]string

type Config struct {
	WorkersCount int
}

type Engine struct {
	overridesProvider     overrides.OverridesProvider
	prerequisitesProvider components.Provider
	componentsProvider    components.Provider
	resourcesPath         string
	cfg                   Config
}

func NewEngine(overridesProvider overrides.OverridesProvider, prerequisitesProvider components.Provider, componentsProvider components.Provider, resourcesPath string, cfg Config) *Engine {
	statusMap = make(map[string]string)
	return &Engine{
		overridesProvider:     overridesProvider,
		prerequisitesProvider: prerequisitesProvider,
		componentsProvider:    componentsProvider,
		resourcesPath:         resourcesPath,
		cfg:                   cfg,
	}
}

type Installation interface {
	//Installs components. ctx is used for cancellation of the operation. returned channel receives every processed component and is closed after all components are processed.
	Install(ctx context.Context) (<-chan components.Component, error)
	//Uninstalls components. ctx is used for cancellation of the operation. returned channel receives every processed component and is closed after all components are processed.
	Uninstall(ctx context.Context) (<-chan components.Component, error)
}

func (e *Engine) installPrerequisites(ctx context.Context, statusChan chan<- components.Component, prerequisites []components.Component) {

	for _, prerequisite := range prerequisites {
		//TODO: Is there a better way to find out if Context is canceled?
		if ctx.Err() != nil {
			//Context is canceled or timed-out. Skip processing
			return
		}
		err := prerequisite.InstallComponent()
		if err != nil {
			prerequisite.Status = components.StatusError
		} else {
			prerequisite.Status = components.StatusInstalled
		}
		statusChan <- prerequisite
	}
}

func (e *Engine) uninstallPrerequisites(ctx context.Context, statusChan chan<- components.Component, prerequisites []components.Component) {

	for i := len(prerequisites) - 1; i >= 0; i-- {
		//TODO: Is there a better way to find out if Context is canceled?
		if ctx.Err() != nil {
			//Context is canceled or timed-out. Skip processing
			return
		}
		prereq := prerequisites[i]
		err := prereq.UninstallComponent()
		if err != nil {
			prereq.Status = components.StatusError
		} else {
			prereq.Status = components.StatusUninstalled
		}
		statusChan <- prereq
	}
}

func (e *Engine) Install(ctx context.Context) (<-chan components.Component, error) {

	prerequisites, err := e.prerequisitesProvider.GetComponents()
	if err != nil {
		return nil, err
	}

	cmps, err := e.componentsProvider.GetComponents()
	if err != nil {
		return nil, err
	}

	err = e.overridesProvider.ReadOverridesFromCluster()
	if err != nil {
		return nil, fmt.Errorf("error while reading overrides: %v", err)
	}

	//TODO: Size dependent on number of components?
	statusChan := make(chan components.Component, 30)

	//TODO: I'd prefer to avoid this goroutine. Because goroutines are cheap, for now it's OK.
	//A better approach would be a dedicated data type containing a list of operations.
	//Every operation would be a non-empty set of components, along with information about how many workers should be used (default = 1)
	//Then we could use generic `run` subroutine to process such list.
	go func() {
		defer close(statusChan)

		e.installPrerequisites(ctx, statusChan, prerequisites)
		if ctx.Err() != nil {
			return
		}

		err = e.overridesProvider.ReadOverridesFromCluster()
		if err != nil {
			log.Printf("error while reading overrides: %v", err)
			return
		}

		//Install the rest of the components
		run(ctx, statusChan, cmps, "install", e.cfg.WorkersCount)

	}()

	return statusChan, nil
}

func (e *Engine) Uninstall(ctx context.Context) (<-chan components.Component, error) {
	cmps, err := e.componentsProvider.GetComponents()
	if err != nil {
		return nil, err
	}

	prerequisites, err := e.prerequisitesProvider.GetComponents()
	if err != nil {
		return nil, err
	}

	//TODO: Size dependent on number of components?
	statusChan := make(chan components.Component, 30)

	//TODO: Perhaps we should refactor to get rid of this goroutine.
	go func() {
		defer close(statusChan)

		//Uninstall the "standard" components
		run(ctx, statusChan, cmps, "uninstall", e.cfg.WorkersCount)

		if ctx.Err() == nil {
			//Uninstall the prequisite components
			e.uninstallPrerequisites(ctx, statusChan, prerequisites)
		}
	}()

	return statusChan, nil
}

func run(ctx context.Context, statusChan chan<- components.Component, cmps []components.Component, installationType string, concurrency int) {
	//TODO: Size dependent on number of components?
	jobChan := make(chan components.Component, 30)

	//Fill the queue with jobs
	for _, comp := range cmps {
		if !enqueueJob(comp, jobChan) {
			log.Printf("Max capacity reached, component dismissed: %s", comp.Name)
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
			log.Printf("Finishing work: %v", ctx.Err())
			return

		case component, ok := <-jobChan:
			//TODO: Is there a better way to find out if Context is canceled?
			if err := ctx.Err(); err != nil {
				log.Printf("Finishing work: %v.", err)
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
				log.Printf("Finishing work: no more jobs in queue.")
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

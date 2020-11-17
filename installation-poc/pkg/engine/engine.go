package engine

import (
	"context"
	"log"
	"sync"

	"github.com/kyma-incubator/hydroform/installation-poc/pkg/overrides"

	"github.com/kyma-incubator/hydroform/installation-poc/pkg/components"
)

var statusMap map[string]string

type Engine struct {
	overridesProvider     overrides.OverridesProvider
	prerequisitesProvider components.Provider
	componentsProvider    components.Provider
	resourcesPath         string
}

func NewEngine(overridesProvider overrides.OverridesProvider, prerequisitesProvider components.Provider, componentsProvider components.Provider, resourcesPath string) *Engine {
	statusMap = make(map[string]string)
	return &Engine{
		overridesProvider:     overridesProvider,
		prerequisitesProvider: prerequisitesProvider,
		componentsProvider:    componentsProvider,
		resourcesPath:         resourcesPath,
	}
}

type Installation interface {
	//Installs components. ctx is used for cancellation of the operation. returned channel receives every processed component and is closed after all components are processed.
	Install(ctx context.Context) (<-chan components.Component, error)
	//Uninstalls components. ctx is used for cancellation of the operation. returned channel receives every processed component and is closed after all components are processed.
	Uninstall(ctx context.Context) (<-chan components.Component, error)
}

func (e *Engine) installPrerequisites(statusChan chan<- components.Component, prerequisites []components.Component) {

	for _, prerequisite := range prerequisites {
		err := prerequisite.InstallComponent()
		if err != nil {
			prerequisite.Status = components.StatusError
		} else {
			prerequisite.Status = components.StatusInstalled
		}
		statusChan <- prerequisite
	}
}

func (e *Engine) uninstallPrerequisites(statusChan chan<- components.Component, prerequisites []components.Component) {

	for i := len(prerequisites) - 1; i >= 0; i-- {
		prq := prerequisites[i]
		err := prq.UninstallComponent()
		if err != nil {
			prq.Status = components.StatusError
		} else {
			prq.Status = components.StatusUninstalled
		}
		statusChan <- prq
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

	//TODO: Size dependent on number of components?
	statusChan := make(chan components.Component, 30)

	go func() {
		defer close(statusChan)

		e.installPrerequisites(statusChan, prerequisites)

		//Install the rest of the components
		run(ctx, statusChan, cmps, "install")
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

	go func() {
		defer close(statusChan)

		//Uninstall the "standard" components
		run(ctx, statusChan, cmps, "uninstall")

		if ctx.Err() == nil {
			//Uninstall the prequisite components
			e.uninstallPrerequisites(statusChan, prerequisites)
		}
	}()

	return statusChan, nil
}

func run(ctx context.Context, statusChan chan<- components.Component, cmps []components.Component, installationType string) {
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
	for i := 0; i < 4; i++ {
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
		case <-ctx.Done():
			log.Printf("Finishing work: context cancelled.")
			return

		case component, ok := <-jobChan:
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
				if err := ctx.Err(); err != nil {
					log.Printf("Finishing work: context error: %s.", err.Error())
				} else {
					log.Printf("Finishing work: no more jobs in queue.")
				}
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

package engine

import (
	"context"
	"fmt"
	"log"
	"path"
	"sync"
	"time"

	"github.com/kyma-incubator/hydroform/installation-poc/pkg/overrides"

	"github.com/kyma-incubator/hydroform/installation-poc/pkg/components"
	"github.com/kyma-incubator/hydroform/installation-poc/pkg/helm"
)

var statusMap map[string]string

type Engine struct {
	overridesProvider  overrides.OverridesProvider
	componentsProvider components.ComponentsProvider
	resourcesPath      string
}

func NewEngine(overridesProvider overrides.OverridesProvider, componentsProvider components.ComponentsProvider, resourcesPath string) *Engine {
	statusMap = make(map[string]string)
	return &Engine{
		overridesProvider:  overridesProvider,
		componentsProvider: componentsProvider,
		resourcesPath:      resourcesPath,
	}
}

type Installation interface {
	Install() error
	Uninstall() error
}

func (e *Engine) installPrerequisites() error {
	helmClient := &helm.Client{}

	clusterEssentials := &components.Component{
		Name:       "cluster-essentials",
		Namespace:  "kyma-system",
		Overrides:  e.overridesProvider.OverridesFor("cluster-essentials"),
		ChartDir:   path.Join(e.resourcesPath, "cluster-essentials"),
		HelmClient: helmClient,
	}
	err := clusterEssentials.InstallComponent()
	if err != nil {
		return err
	}

	istio := &components.Component{
		Name:       "istio",
		Namespace:  "istio-system",
		Overrides:  e.overridesProvider.OverridesFor("istio"),
		ChartDir:   path.Join(e.resourcesPath, "istio"),
		HelmClient: helmClient,
	}
	err = istio.InstallComponent()
	if err != nil {
		return err
	}

	xipPatch := &components.Component{
		Name:       "xip-patch",
		Namespace:  "kyma-installer",
		Overrides:  e.overridesProvider.OverridesFor("xip-patch"),
		ChartDir:   path.Join(e.resourcesPath, "xip-patch"),
		HelmClient: helmClient,
	}
	err = xipPatch.InstallComponent()
	if err != nil {
		return err
	}

	return nil
}

func (e *Engine) uninstallPrerequisites() error {
	helmClient := &helm.Client{}

	xipPatch := &components.Component{
		Name:       "xip-patch",
		Namespace:  "kyma-installer",
		Overrides:  e.overridesProvider.OverridesFor("xip-patch"),
		ChartDir:   path.Join(e.resourcesPath, "xip-patch"),
		HelmClient: helmClient,
	}
	err := xipPatch.UninstallComponent()
	if err != nil {
		return err
	}

	istio := &components.Component{
		Name:       "istio",
		Namespace:  "istio-system",
		Overrides:  e.overridesProvider.OverridesFor("istio"),
		ChartDir:   path.Join(e.resourcesPath, "istio"),
		HelmClient: helmClient,
	}
	err = istio.UninstallComponent()
	if err != nil {
		return err
	}

	clusterEssentials := &components.Component{
		Name:       "cluster-essentials",
		Namespace:  "kyma-system",
		Overrides:  e.overridesProvider.OverridesFor("cluster-essentials"),
		ChartDir:   path.Join(e.resourcesPath, "cluster-essentials"),
		HelmClient: helmClient,
	}
	err = clusterEssentials.UninstallComponent()
	if err != nil {
		return err
	}

	return nil
}

func (e *Engine) Install() error {
	err := e.installPrerequisites()
	if err != nil {
		return err
	}

	cmps, err := e.componentsProvider.GetComponents()
	if err != nil {
		return err
	}

	//Install the rest of the components
	return run(cmps, "install")
}

func (e *Engine) Uninstall() error {
	cmps, err := e.componentsProvider.GetComponents()
	if err != nil {
		return err
	}

	//Uninstall the components
	err = run(cmps, "uninstall")
	if err != nil {
		return err
	}

	//Uninstall the prequisite components
	err = e.uninstallPrerequisites()
	if err != nil {
		return err
	}

	return nil
}

func run(cmps []components.Component, installationType string) error {
	jobChan := make(chan components.Component, 30)
	for _, comp := range cmps {
		if !enqueueJob(comp, jobChan) {
			log.Printf("Max capacity reached, component dismissed: %s", comp.Name)
		}
	}

	statusChan := make(chan components.Component, 30)
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for i := 0; i < 4; i++ {
		wg.Add(1)
		go worker(ctx, &wg, jobChan, statusChan, installationType)
	}

	// to stop the workers, first close the job channel
	close(jobChan)
	return wait(&wg, 10*time.Minute, statusChan, cmps)
}

func worker(ctx context.Context, wg *sync.WaitGroup, jobChan <-chan components.Component, statusChan chan<- components.Component, installationType string) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return

		case component, ok := <-jobChan:
			if ctx.Err() != nil || !ok {
				return
			}
			if ok {
				if installationType == "install" {
					if err := component.InstallComponent(); err != nil {
						component.Status = "Error"
					} else {
						component.Status = "Installed"
					}
					statusChan <- component
				} else if installationType == "uninstall" {
					if err := component.UninstallComponent(); err != nil {
						component.Status = "Error"
					} else {
						component.Status = "Uninstalled"
					}
					statusChan <- component
				}
			}
		}
	}
}

func wait(wg *sync.WaitGroup, timeout time.Duration, statusChan <-chan components.Component, cmps []components.Component) error {
	ch := make(chan struct{})
	go func() {
		wg.Wait()
		close(ch)
	}()
	for {
		select {
		case component, ok := <-statusChan:
			if ok {
				log.Printf("Operation in progress.. Component: %v, Status: %v", component.Name, component.Status)
				statusMap[component.Name] = component.Status
			}
		case <-ch:
			operationErrored := false
			for _, cmp := range cmps {
				componentStatus, ok := statusMap[cmp.Name]
				if !ok {
					log.Printf("Component: %s, Status: %s", cmp.Name, "Error")
					operationErrored = true
					continue
				}
				log.Printf("Component: %s, Status: %s", cmp.Name, componentStatus)
				if componentStatus == "Error" {
					operationErrored = true
				}
			}
			if operationErrored {
				return fmt.Errorf("Operation was unsuccessful! Check the previous logs to see the problem.")
			}
			return nil
		case <-time.After(timeout):
			return fmt.Errorf("Timeout occurred after %v minutes", timeout.Minutes())
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

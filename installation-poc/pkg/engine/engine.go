package engine

import (
	"log"
	"path"
	"sync"
	"time"
	"context"

	"github.com/kyma-incubator/hydroform/installation-poc/pkg/components"
	"github.com/kyma-incubator/hydroform/installation-poc/pkg/helm"
)

type Engine struct {
	componentsProvider components.ComponentsProvider
	resourcesPath      string
}

func NewEngine(componentsProvider components.ComponentsProvider, resourcesPath string) *Engine {
	return &Engine{
		componentsProvider: componentsProvider,
		resourcesPath:      resourcesPath,
	}
}

type Installation interface {
	Install() error
}

func (e *Engine) installPrerequisites() error {
	//TODO need to have overrides for this 3 components as well
	helmClient := &helm.Client{}

	clusterEssentials := &components.Component{
		Name:       "cluster-essentials",
		Namespace:  "kyma-system",
		Overrides:  nil,
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
		Overrides:  nil,
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
		Overrides:  nil,
		ChartDir:   path.Join(e.resourcesPath, "xip-patch"),
		HelmClient: helmClient,
	}
	err = xipPatch.InstallComponent()
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
	jobChan := make(chan components.Component, 30)
	for _, comp := range cmps {
		if !enqueueJob(comp, jobChan) {
			log.Printf("Max capacity reached, component dismissed: %s", comp.Name)
		}
	}

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())

	for i := 0; i < 4; i++ {
		wg.Add(1)
		go worker(ctx, &wg, jobChan)
	}

	// to stop the workers, first close the job channel
	close(jobChan)
	wait(&wg, 10*time.Minute)
	cancel()

	return nil
}

func worker(ctx context.Context, wg *sync.WaitGroup, jobChan <-chan components.Component) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return

		case job, ok := <-jobChan:
			if ctx.Err() != nil || !ok {
				return
			}
			if ok {
				job.InstallComponent()
			}
		}
	}
}

func wait(wg *sync.WaitGroup, timeout time.Duration) bool {
	ch := make(chan struct{})
	go func() {
		wg.Wait()
		close(ch)
	}()
	select {
	case <-ch:
		return true
	case <-time.After(timeout):
		log.Println("Timeout occurred!")
		return false
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

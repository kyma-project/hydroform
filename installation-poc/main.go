package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"

	"github.com/kyma-incubator/hydroform/installation-poc/pkg/components"
	"github.com/kyma-incubator/hydroform/installation-poc/pkg/helm"
	"github.com/kyma-incubator/hydroform/installation-poc/pkg/overrides"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var resourcesPath string

type Engine struct {
	componentsProvider components.ComponentsProvider
}

func NewEngine(componentsProvider components.ComponentsProvider) *Engine {
	return &Engine{
		componentsProvider: componentsProvider,
	}
}

type Installation interface {
	Install(components []components.Component) error
}

func (e *Engine) installPrerequisites() error {
	//TODO need to have overrides for this 3 components as well
	helmClient := &helm.Client{}

	clusterEssentials := &components.Component{
		Name:       "cluster-essentials",
		Namespace:  "kyma-system",
		Overrides:  nil,
		ChartDir:   path.Join(resourcesPath, "cluster-essentials"),
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
		ChartDir:   path.Join(resourcesPath, "istio"),
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
		ChartDir:   path.Join(resourcesPath, "xip-patch"),
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

func main() {
	goPath := os.Getenv("GOPATH")
	if goPath == "" {
		log.Fatalf("Please set GOPATH")
	}

	resourcesPath = filepath.Join(goPath, "src", "github.com", "kyma-project", "kyma", "resources")

	kubeconfigPath := "/Users/i304607/Downloads/mst.yml"

	config, err := getClientConfig(kubeconfigPath)
	if err != nil {
		log.Fatalf("Unable to build kubernetes configuration. Error: %v", err)
	}

	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Unable to create internal client. Error: %v", err)
	}

	overridesProvider := overrides.New(kubeClient)

	componentsProvider := components.NewComponents(overridesProvider, resourcesPath)

	engine := NewEngine(componentsProvider)

	err = engine.Install()
	if err != nil {
		log.Fatalf("Kyma installation fialed. Error: %v", err)
	}

	fmt.Println("Kyma installed")
}

func getClientConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	return rest.InClusterConfig()
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

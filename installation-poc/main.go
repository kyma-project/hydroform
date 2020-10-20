package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/kyma-incubator/hydroform/installation-poc/pkg/helm"
	"github.com/kyma-incubator/hydroform/installation-poc/pkg/overrides"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var resourcesPath string
var kubeconfigPath string
var commonListOpts = metav1.ListOptions{LabelSelector: "installer=overrides, !component"}
var componentListOpts = metav1.ListOptions{LabelSelector: "installer=overrides, component"}
var componentOverrides map[string]map[string]interface{}
var globalOverrides map[string]interface{}

type Engine struct {
	componentsProvider ComponentsProvider
}

func NewEngine(componentsProvider ComponentsProvider) *Engine {
	return &Engine{
		componentsProvider: componentsProvider,
	}
}

type Provider struct {
	overridesProvider overrides.OverridesProvider
}

func NewComponents(overridesProvider overrides.OverridesProvider) *Provider {
	return &Provider{
		overridesProvider: overridesProvider,
	}
}

type ComponentsProvider interface {
	GetComponents() ([]Component,error)
}

type Installation interface {
	Install(components []Component) error
}

type Component struct {
	Name       string
	Namespace  string
	Overrides  map[string]interface{}
	HelmClient helm.ClientInterface
}

type ComponentInstallation interface {
	InstallComponent() error
}

func (c *Component) InstallComponent() error {
	chartDir := path.Join(resourcesPath, c.Name)
	log.Printf("MST Installing %s in %s from %s", c.Name, c.Namespace, chartDir)

	err := c.HelmClient.InstallRelease(chartDir, c.Namespace, c.Name, c.Overrides)
	if err != nil {
		log.Printf("MST Error installing %s: %v", c.Name, err)
		return err
	}

	log.Printf("MST Installed %s in %s", c.Name, c.Namespace)

	return nil
}

func (p *Provider) GetComponents() ([]Component, error) {
	helmClient := &helm.Client{}

	err := p.overridesProvider.ReadOverridesFromCluster()
	if err != nil {
		return nil, err
	}

	return []Component{
		Component{
			Name:       "istio-kyma-patch",
			Namespace:  "istio-system",
			Overrides:  p.overridesProvider.OverridesFor("istio-kyma-patch"),
			HelmClient: helmClient,
		},
		Component{
			Name:       "knative-eventing",
			Namespace:  "knative-eventing",
			Overrides:  p.overridesProvider.OverridesFor("knative-eventing"),
			HelmClient: helmClient,
		},
		Component{
			Name:       "dex",
			Namespace:  "kyma-system",
			Overrides:  p.overridesProvider.OverridesFor("dex"),
			HelmClient: helmClient,
		},
		Component{
			Name:       "ory",
			Namespace:  "kyma-system",
			Overrides:  p.overridesProvider.OverridesFor("ory"),
			HelmClient: helmClient,
		},
		Component{
			Name:       "api-gateway",
			Namespace:  "kyma-system",
			Overrides:  p.overridesProvider.OverridesFor("api-gateway"),
			HelmClient: helmClient,
		},
		Component{
			Name:       "rafter",
			Namespace:  "kyma-system",
			Overrides:  p.overridesProvider.OverridesFor("rafter"),
			HelmClient: helmClient,
		},
		Component{
			Name:       "service-catalog",
			Namespace:  "kyma-system",
			Overrides:  p.overridesProvider.OverridesFor("service-catalog"),
			HelmClient: helmClient,
		},
		Component{
			Name:       "service-catalog-addons",
			Namespace:  "kyma-system",
			Overrides:  p.overridesProvider.OverridesFor("service-catalog-addons"),
			HelmClient: helmClient,
		},
		Component{
			Name:       "nats-streaming",
			Namespace:  "natss",
			Overrides:  p.overridesProvider.OverridesFor("nats-streaming"),
			HelmClient: helmClient,
		},
		Component{
			Name:       "core",
			Namespace:  "kyma-system",
			Overrides:  p.overridesProvider.OverridesFor("core"),
			HelmClient: helmClient,
		},
		Component{
			Name:       "cluster-users",
			Namespace:  "kyma-system",
			Overrides:  p.overridesProvider.OverridesFor("cluster-users"),
			HelmClient: helmClient,
		},
		Component{
			Name:       "permission-controller",
			Namespace:  "kyma-system",
			Overrides:  p.overridesProvider.OverridesFor("permission-controller"),
			HelmClient: helmClient,
		},
		Component{
			Name:       "apiserver-proxy",
			Namespace:  "kyma-system",
			Overrides:  p.overridesProvider.OverridesFor("apiserver-proxy"),
			HelmClient: helmClient,
		},
		Component{
			Name:       "iam-kubeconfig-service",
			Namespace:  "kyma-system",
			Overrides:  p.overridesProvider.OverridesFor("iam-kubeconfig-service"),
			HelmClient: helmClient,
		},
		Component{
			Name:       "serverless",
			Namespace:  "kyma-system",
			Overrides:  p.overridesProvider.OverridesFor("serverless"),
			HelmClient: helmClient,
		},
		Component{
			Name:       "knative-provisioner-natss",
			Namespace:  "knative-eventing",
			Overrides:  p.overridesProvider.OverridesFor("knative-provisioner-natss"),
			HelmClient: helmClient,
		},
		Component{
			Name:       "event-sources",
			Namespace:  "kyma-system",
			Overrides:  p.overridesProvider.OverridesFor("event-sources"),
			HelmClient: helmClient,
		},
		Component{
			Name:       "application-connector",
			Namespace:  "kyma-integration",
			Overrides:  p.overridesProvider.OverridesFor("application-connector"),
			HelmClient: helmClient,
		},
		Component{
			Name:       "console",
			Namespace:  "kyma-system",
			Overrides:  p.overridesProvider.OverridesFor("console"),
			HelmClient: helmClient,
		},
	}, nil
}

func (e *Engine) installPrerequisites() error {
	//TODO need to have overrides for this 3 components as well
	helmClient := &helm.Client{}

	clusterEssentials := &Component{
		Name:       "cluster-essentials",
		Namespace:  "kyma-system",
		Overrides:  nil,
		HelmClient: helmClient,
	}
	err := clusterEssentials.InstallComponent()
	if err != nil {
		return err
	}

	istio := &Component{
		Name:       "istio",
		Namespace:  "istio-system",
		Overrides:  nil,
		HelmClient: helmClient,
	}
	err = istio.InstallComponent()
	if err != nil {
		return err
	}

	xipPatch := &Component{
		Name:       "xip-patch",
		Namespace:  "kyma-installer",
		Overrides:  nil,
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

	components, err := e.componentsProvider.GetComponents()
	if err != nil {
		return err
	}

	//Install the rest of the components
	jobChan := make(chan Component, 30)
	for _, comp := range components {
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

	kubeconfigPath = "/Users/i304607/Downloads/mst.yml"

	config, err := getClientConfig(kubeconfigPath)
	if err != nil {
		log.Fatalf("Unable to build kubernetes configuration. Error: %v", err)
	}

	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Unable to create internal client. Error: %v", err)
	}

	overridesProvider := overrides.New(kubeClient)

	componentsProvider := NewComponents(overridesProvider)

	engine := NewEngine(componentsProvider)

	err = engine.Install()
	if err != nil {
		log.Fatalf("Kyma installation fialed. Error: %v", err)
	}

	fmt.Println("Kyma installed")
}

func unflattenToMap(sourceMap map[string]interface{}) map[string]interface{} {
	mergedMap := map[string]interface{}{}
	if len(sourceMap) == 0 {
		return mergedMap
	}

	for key, value := range sourceMap {
		keys := strings.Split(key, ".")
		mergeIntoMap(keys, value.(string), mergedMap)
	}

	return mergedMap
}

func mergeIntoMap(keys []string, value string, dstMap map[string]interface{}) {
	currentKey := keys[0]
	//Last key points directly to string value
	if len(keys) == 1 {

		//Conversion to boolean to satisfy Helm requirements.yaml: "enable:true/false syntax"
		var vv interface{} = value
		if value == "true" || value == "false" {
			vv, _ = strconv.ParseBool(value)
		}

		dstMap[currentKey] = vv
		return
	}

	//All keys but the last one should point to a nested map
	nestedMap, isMap := dstMap[currentKey].(map[string]interface{})

	if !isMap {
		nestedMap = map[string]interface{}{}
		dstMap[currentKey] = nestedMap
	}

	mergeIntoMap(keys[1:], value, nestedMap)
}

func getClientConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	return rest.InClusterConfig()
}

func worker(ctx context.Context, wg *sync.WaitGroup, jobChan <-chan Component) {
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

func enqueueJob(job Component, jobChan chan<- Component) bool {
	select {
	case jobChan <- job:
		return true
	default:
		return false
	}
}

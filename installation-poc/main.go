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
	"helm.sh/helm/v3/pkg/strvals"
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
	//values.yaml
}

func NewEngine(componentsProvider ComponentsProvider) *Engine {
	return &Engine{
		componentsProvider: componentsProvider,
	}
}

type Components struct{}

type ComponentsProvider interface {
	GetComponents() []Component
}

type Installation interface {
	Install(components []Component) error
}

type Component struct {
	Name              string
	Namespace         string
	OverridesProvider OverridesProvider
	HelmClient        helm.ClientInterface
}

type ComponentInstallation interface {
	InstallComponent() error
}

type Overrides struct{}

type OverridesProvider interface {
	OverridesFor(component *Component) map[string]interface{}
}

func (overrides *Overrides) OverridesFor(component *Component) map[string]interface{} {
	if val, ok := componentOverrides[component.Name]; ok {
		log.Printf("Overrides for %s: %v", component.Name, val)
		return val
	}
	log.Printf("Overrides for %s: %v", component.Name, globalOverrides)
	return globalOverrides
}

func readOverridesFromCluster() error {
	config, err := getClientConfig(kubeconfigPath)
	if err != nil {
		log.Fatalf("Unable to build kubernetes configuration. Error: %v", err)
	}

	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Unable to create internal client. Error: %v", err)
	}

	//Read global overrides
	globalOverrideCMs, err := kubeClient.CoreV1().ConfigMaps("kyma-installer").List(context.TODO(), commonListOpts)

	var globalValues []string
	for _, cm := range globalOverrideCMs.Items {
		log.Printf("%s data %v", cm.Name, cm.Data)
		for k, v := range cm.Data {
			globalValues = append(globalValues, k+"="+v)
		}
	}

	globalOverrides = make(map[string]interface{})
	for _, value := range globalValues {
		if err := strvals.ParseInto(value, globalOverrides); err != nil {
			log.Printf("Error parsing global overrides: %v", err)
			return err
		}
	}

	//Read component overrides
	componentOverrides = make(map[string]map[string]interface{})

	componentOverrideCMs, err := kubeClient.CoreV1().ConfigMaps("kyma-installer").List(context.TODO(), componentListOpts)

	for _, cm := range componentOverrideCMs.Items {
		log.Printf("%s data %v", cm.Name, cm.Data)
		var componentValues []string
		name := cm.Labels["component"]

		for k, v := range cm.Data {
			componentValues = append(componentValues, k+"="+v)
		}

		//Merge global overrides to component overrides for each component
		componentValues = append(globalValues, componentValues...)

		componentOverrides[name] = make(map[string]interface{})
		for _, value := range componentValues {
			if err := strvals.ParseInto(value, componentOverrides[name]); err != nil {
				log.Printf("Error parsing overrides for %s: %v", name, err)
				return err
			}
		}
	}

	log.Println("Reading the overrides from the cluster completed successfully!")
	return nil
}



func (c *Component) InstallComponent() error {
	chartDir := path.Join(resourcesPath, c.Name)
	log.Printf("MST Installing %s in %s from %s", c.Name, c.Namespace, chartDir)

	overrides := c.OverridesProvider.OverridesFor(c)

	err := c.HelmClient.InstallRelease(chartDir, c.Namespace, c.Name, overrides)
	if err != nil {
		log.Printf("MST Error installing %s: %v", c.Name, err)
		return err
	}

	log.Printf("MST Installed %s in %s", c.Name, c.Namespace)

	return nil
}

func (c *Components) GetComponents() []Component {
	overridesProvider := &Overrides{}
	helmClient := &helm.Client{}
	return []Component{
		Component{
			Name:              "istio-kyma-patch",
			Namespace:         "istio-system",
			OverridesProvider: overridesProvider,
			HelmClient:        helmClient,
		},
		Component{
			Name:              "knative-eventing",
			Namespace:         "knative-eventing",
			OverridesProvider: overridesProvider,
			HelmClient:        helmClient,
		},
		Component{
			Name:              "dex",
			Namespace:         "kyma-system",
			OverridesProvider: overridesProvider,
			HelmClient:        helmClient,
		},
		Component{
			Name:              "ory",
			Namespace:         "kyma-system",
			OverridesProvider: overridesProvider,
			HelmClient:        helmClient,
		},
		Component{
			Name:              "api-gateway",
			Namespace:         "kyma-system",
			OverridesProvider: overridesProvider,
			HelmClient:        helmClient,
		},
		Component{
			Name:              "rafter",
			Namespace:         "kyma-system",
			OverridesProvider: overridesProvider,
			HelmClient:        helmClient,
		},
		Component{
			Name:              "service-catalog",
			Namespace:         "kyma-system",
			OverridesProvider: overridesProvider,
			HelmClient:        helmClient,
		},
		Component{
			Name:              "service-catalog-addons",
			Namespace:         "kyma-system",
			OverridesProvider: overridesProvider,
			HelmClient:        helmClient,
		},
		Component{
			Name:              "nats-streaming",
			Namespace:         "natss",
			OverridesProvider: overridesProvider,
			HelmClient:        helmClient,
		},
		Component{
			Name:              "core",
			Namespace:         "kyma-system",
			OverridesProvider: overridesProvider,
			HelmClient:        helmClient,
		},
		Component{
			Name:              "cluster-users",
			Namespace:         "kyma-system",
			OverridesProvider: overridesProvider,
			HelmClient:        helmClient,
		},
		Component{
			Name:              "permission-controller",
			Namespace:         "kyma-system",
			OverridesProvider: overridesProvider,
			HelmClient:        helmClient,
		},
		Component{
			Name:              "apiserver-proxy",
			Namespace:         "kyma-system",
			OverridesProvider: overridesProvider,
			HelmClient:        helmClient,
		},
		Component{
			Name:              "iam-kubeconfig-service",
			Namespace:         "kyma-system",
			OverridesProvider: overridesProvider,
			HelmClient:        helmClient,
		},
		Component{
			Name:              "serverless",
			Namespace:         "kyma-system",
			OverridesProvider: overridesProvider,
			HelmClient:        helmClient,
		},
		Component{
			Name:              "knative-provisioner-natss",
			Namespace:         "knative-eventing",
			OverridesProvider: overridesProvider,
			HelmClient:        helmClient,
		},
		Component{
			Name:              "event-sources",
			Namespace:         "kyma-system",
			OverridesProvider: overridesProvider,
			HelmClient:        helmClient,
		},
		Component{
			Name:              "application-connector",
			Namespace:         "kyma-integration",
			OverridesProvider: overridesProvider,
			HelmClient:        helmClient,
		},
		Component{
			Name:              "console",
			Namespace:         "kyma-system",
			OverridesProvider: overridesProvider,
			HelmClient:        helmClient,
		},
	}
}

func installPrerequisites() error {
	overridesProvider := &Overrides{}
	helmClient := &helm.Client{}

	clusterEssentials := &Component{
		Name:              "cluster-essentials",
		Namespace:         "kyma-system",
		OverridesProvider: overridesProvider,
		HelmClient:        helmClient,
	}
	err := clusterEssentials.InstallComponent()
	if err != nil {
		return err
	}

	istio := &Component{
		Name:              "istio",
		Namespace:         "istio-system",
		OverridesProvider: overridesProvider,
		HelmClient:        helmClient,
	}
	err = istio.InstallComponent()
	if err != nil {
		return err
	}

	xipPatch := &Component{
		Name:              "xip-patch",
		Namespace:         "kyma-installer",
		OverridesProvider: overridesProvider,
		HelmClient:        helmClient,
	}
	err = xipPatch.InstallComponent()
	if err != nil {
		return err
	}

	err = readOverridesFromCluster()
	if err != nil {
		return err
	}

	return nil
}

func (e *Engine) Install() error {
	err := installPrerequisites()
	if err != nil {
		return err
	}

	components := e.componentsProvider.GetComponents()

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

	componentsProvider := &Components{}

	engine := NewEngine(componentsProvider)

	err := engine.Install()
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



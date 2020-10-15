package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"sync"
	"time"
	"context"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"sigs.k8s.io/yaml"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var resources = "/Users/i304607/Yaas/go/src/github.com/kyma-project/kyma/resources"
var overridesFile = "/Users/i304607/overrides.yaml"
//var kubeconfig = "/Users/i304607/.kube/config"
var kubeconfig = "/Users/i304607/Downloads/mst.yml"
var commonListOpts = metav1.ListOptions{LabelSelector: "installer=overrides"}

type Component struct {
	Name      string
	Namespace string
}

var components = []Component{
	Component{
		Name:      "istio-kyma-patch",
		Namespace: "istio-system",
	},
	Component{
		Name:      "knative-serving",
		Namespace: "knative-serving",
	},
	Component{
		Name:      "knative-eventing",
		Namespace: "knative-eventing",
	},
	Component{
		Name:      "dex",
		Namespace: "kyma-system",
	},
	Component{
		Name:      "ory",
		Namespace: "kyma-system",
	},
	Component{
		Name:      "api-gateway",
		Namespace: "kyma-system",
	},
	Component{
		Name:      "rafter",
		Namespace: "kyma-system",
	},
	Component{
		Name:      "service-catalog",
		Namespace: "kyma-system",
	},
	Component{
		Name:      "service-catalog-addons",
		Namespace: "kyma-system",
	},
	Component{
		Name:      "nats-streaming",
		Namespace: "natss",
	},
	Component{
		Name:      "core",
		Namespace: "kyma-system",
	},
	Component{
		Name:      "cluster-users",
		Namespace: "kyma-system",
	},
	Component{
		Name:      "permission-controller",
		Namespace: "kyma-system",
	},
	Component{
		Name:      "apiserver-proxy",
		Namespace: "kyma-system",
	},
	Component{
		Name:      "iam-kubeconfig-service",
		Namespace: "kyma-system",
	},
	Component{
		Name:      "serverless",
		Namespace: "kyma-system",
	},
	Component{
		Name:      "knative-provisioner-natss",
		Namespace: "knative-eventing",
	},
	Component{
		Name:      "event-sources",
		Namespace: "kyma-system",
	},
	Component{
		Name:      "application-connector",
		Namespace: "kyma-integration",
	},
	Component{
		Name:      "console",
		Namespace: "kyma-system",
	},
}

func main() {
	////pre-req for kyma
	//err := installKymaComponent("cluster-essentials", "kyma-system")
	//if err != nil {
	//	log.Fatalf("Error: %v", err)
	//}
	//err = installKymaComponent("istio", "istio-system")
	//if err != nil {
	//	log.Fatalf("Error: %v", err)
	//}
	//err = installKymaComponent("xip-patch", "kyma-installer")
	//if err != nil {
	//	log.Fatalf("Error: %v", err)
	//}

	//read overrides produced by xip-patch
	config, err := getClientConfig(kubeconfig)
	if err != nil {
		log.Fatalf("Unable to build kubernetes configuration. Error: %v", err)
	}

	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Unable to create internal client. Error: %v", err)
	}

	configmaps, err := kubeClient.CoreV1().ConfigMaps("kyma-installer").List(context.TODO(), commonListOpts)
	overrides := make(map[string]interface{})

	for _, cm := range configmaps.Items {
		log.Printf("%s data %v", cm.Name, cm.Data)

		yamlData, err := yaml.Marshal(cm.Data)

		//save to .yaml
		tmpFile, err := ioutil.TempFile(os.TempDir(), cm.Name+"-")
		if err != nil {
			log.Fatal("Cannot create temporary file", err)
		}

		fmt.Println("Created File: " + tmpFile.Name())
		defer os.Remove(tmpFile.Name())

		if _, err = tmpFile.Write(yamlData); err != nil {
			log.Fatal("Failed to write to temporary file", err)
		}

		// Close the file
		if err := tmpFile.Close(); err != nil {
			log.Fatal(err)
		}

		//read from file
		var data map[string]interface{}
		bs, err := ioutil.ReadFile(tmpFile.Name())
		if err != nil {
			panic(err)
		}
		if err := yaml.Unmarshal(bs, &data); err != nil {
			panic(err)
		}

		for k, v := range data {
			overrides[k] = v
		}

	}

	//save to .yaml
	overridesData, err := yaml.Marshal(overrides)
	if err := ioutil.WriteFile(overridesFile, overridesData, 0644); err != nil {
		panic(err)
	}

	// Install the rest of the components
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


	time.Sleep(5000 * time.Millisecond)
	fmt.Println("Kyma installed")
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
				installKymaComponent(job.Name, job.Namespace)
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

func installRelease(chartDir, namespace, name string) error {
	cfg, err := newActionConfig(namespace)
	if err != nil {
		return err
	}

	chart, err := loader.Load(chartDir)
	if err != nil {
		return err
	}

	install := action.NewInstall(cfg)
	install.ReleaseName = name
	install.Namespace = namespace
	install.Atomic = true
	install.Wait = true
	install.CreateNamespace = true

	values, err := chartutil.ReadValuesFile(overridesFile)
	if err != nil {
		return err
	}

	_, err = install.Run(chart, values)

	if err != nil {
		return err
	}

	return nil
}

func installKymaComponent(name, namespace string) error {
	chartDir := path.Join(resources, name)
	log.Printf("MST Installing %s in %s from %s", name, namespace, chartDir)

	err :=installRelease(chartDir, namespace, name)
	if err != nil {
		log.Printf("MST Error installing %s: %v", name, err)
		return err
	}

	log.Printf("MST Installed %s in %s", name, namespace)
	return nil
}

func newActionConfig(namespace string) (*action.Configuration, error) {
	clientGetter := genericclioptions.NewConfigFlags(false)
	clientGetter.Namespace = &namespace

	cfg := new(action.Configuration)
	if err := cfg.Init(clientGetter, namespace, "secrets", log.Printf); err != nil {
		return nil, err
	}

	return cfg, nil
}

/*
* CODE GENERATED AUTOMATICALLY WITH devops/internal/config
 */

package main

import (
	"context"
	"fmt"
	"github.com/docopt/docopt-go"
	"github.com/kyma-incubator/hydroform/function/pkg/client"
	"github.com/kyma-incubator/hydroform/function/pkg/operator"
	unstructfn "github.com/kyma-incubator/hydroform/function/pkg/resources/unstructured"
	"github.com/kyma-incubator/hydroform/function/pkg/workspace"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"os"
	"path"
	"path/filepath"
)

const (
	usage = `apply description

Usage:
	apply [ --dir=<DIR> --dry-run --kube-config=<FILE> ] [options]

Options:
	--debug                 Enable verbose output.
	-h --help               Show this screen.
	--version               Show version.`

	version = "0.0.1"
)

type config struct {
	Name       string `docopt:"--name" json:"name"`
	Debug      bool   `docopt:"--debug" json:"debug"`
	Dir        string `docopt:"--dir"`
	DryRun     bool   `docopt:"--dry-run"`
	KubeConfig string `docopt:"--kube-config"`
}

func newConfig() (*config, error) {
	arguments, err := docopt.ParseArgs(usage, nil, version)
	if err != nil {
		return nil, err
	}
	var cfg config
	if err = arguments.Bind(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func getClient(cfg *config) dynamic.Interface {
	home := homedir.HomeDir()

	if cfg.KubeConfig == "" && home == "" {
		log.Fatal("unable to find kubeconfig file")
	}

	if cfg.KubeConfig == "" {
		cfg.KubeConfig = filepath.Join(home, ".kube", "config")
	}

	entry := log.WithField("kubeConfig", cfg.KubeConfig)
	entry.Debug("building dynamic client")
	config, err := clientcmd.BuildConfigFromFlags("", cfg.KubeConfig)
	if err != nil {
		entry.Fatal(err)
	}

	result, err := dynamic.NewForConfig(config)
	if err != nil {
		entry.Fatal(err)
	}
	return result
}

func main() {
	// parse command arguments
	cfg, err := newConfig()
	if err != nil {
		log.Fatal(err)
	}

	if cfg.Debug {
		log.SetLevel(log.DebugLevel)
	}

	var stages []string
	if cfg.DryRun {
		stages = append(stages, metav1.DryRunAll)
	}

	entry := log.WithField("dir", cfg.Dir)
	entry.Debug("opening workspace")

	file, err := os.Open(path.Join(cfg.Dir, workspace.CfgFilename))
	if err != nil {
		entry.Fatal(err)
	}

	// Load project configuration
	var configuration workspace.Cfg
	if err := yaml.NewDecoder(file).Decode(&configuration); err != nil {
		entry.Fatal(err)
	}
	configuration.SourcePath = cfg.Dir
	entry = log.NewEntry(log.StandardLogger())
	entryFromCfg(entry, configuration).Debug("creating function from configuration")
	function, err := unstructfn.NewFunction(configuration)
	if err != nil {
		entry.Fatal(err)
	}
	entryFromUnstructured(entry, &function).Debug("function created")

	entryFromCfg(entry, configuration).Debug("creating triggers from configuration")
	triggers, err := unstructfn.NewTriggers(configuration)
	if err != nil {
		entry.Fatal(err)
	}
	for _, trigger := range triggers {
		entryFromUnstructured(entry, &trigger).Debug("trigger created")
	}

	dynamicInterface := getClient(cfg)
	parent, fnClient := newFunctionParentOperatorWithClient(function, dynamicInterface)

	entryFromUnstructured(entry, &function).Debug("applying function")
	status, err := parent.Apply(fnClient, operator.ApplyOptions{
		DryRun: stages,
	})
	if err != nil {
		entry.Fatal(err)
	}
	entryFromStatus(entry, status[0]).Debug(fmt.Sprintf("function %s", status[0].StatusType))

	// WE NEED THAT IN HYDROFORM
	resourceInterface := dynamicInterface.Resource(schema.GroupVersionResource{
		Group:    "eventing.knative.dev",
		Version:  "v1alpha1",
		Resource: "triggers",
	}).Namespace(configuration.Namespace)

	ctx := context.Background()
	list, err := resourceInterface.List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("ownerUID=%s", status[0].GetUID()),
	})

	if err != nil {
		entry.Fatal(err)
	}

	for _, item := range list.Items {
		if !contains(triggers, item.GetName()) {
			if err := resourceInterface.Delete(ctx, item.GetName(), metav1.DeleteOptions{
				DryRun: stages,
			}); err != nil {
				entry.Fatal(err)
			}
			entryFromUnstructured(entry, &item).Debug("deleted")
		}
	}

	entry.WithField("len", len(triggers)).Debug("applying triggers")
	for _, trigger := range triggers {
		resourceOperator, trClient := newTriggerResourceOperatorWithClient(trigger, dynamicInterface)
		status, err := resourceOperator.Apply(trClient, operator.ApplyOptions{
			DryRun:          stages,
			OwnerReferences: status.GetOwnerReferences(),
			Labels: map[string]string{
				"ownerUID": string(status[0].GetUID()),
			},
		})

		if err != nil {
			entry = entryFromStatus(entry, status[0])
			entry.Error(err)
			entry.Error("wiping workspace")
			status, err := parent.Delete(fnClient, operator.DeleteOptions{
				DeletionPropagation: metav1.DeletePropagationForeground,
			})
			if err != nil {
				entry.Fatal(err)
			}
			entryFromStatus(entry, status[0]).Info("function wiped")
			break
		}
		entryFromUnstructured(entry, &trigger).
			Debug(fmt.Sprintf("trigger %s", status[0].StatusType))
	}
}

func newFunctionParentOperatorWithClient(u unstructured.Unstructured, c dynamic.Interface) (operator.Parent, client.Client) {
	parentOperator := operator.NewParentFunction(u)
	resourceInterface := c.Resource(parentOperator.GetGroupVersionResource()).Namespace(u.GetNamespace())
	return parentOperator, resourceInterface
}

func newTriggerResourceOperatorWithClient(u unstructured.Unstructured, c dynamic.Interface) (operator.Resource, client.Client) {
	resourceTrigger := operator.NewResourceTrigger(u)
	resourceInterface := c.Resource(resourceTrigger.GetGroupVersionResource()).Namespace(u.GetNamespace())
	return resourceTrigger, resourceInterface
}

func entryFromCfg(e *log.Entry, cfg workspace.Cfg) *log.Entry {
	return e.WithFields(map[string]interface{}{
		"workspaceName":       cfg.Name,
		"workspaceNamespace":  cfg.Namespace,
		"workspaceSourcePath": cfg.SourcePath,
		"workspaceIsGit":      cfg.Git,
	})
}

func entryFromUnstructured(e *log.Entry, u *unstructured.Unstructured) *log.Entry {
	return e.WithFields(map[string]interface{}{
		"name":        u.GetName(),
		"namespace":   u.GetNamespace(),
		"kind":        u.GetKind(),
		"api-version": u.GetAPIVersion(),
	})
}

func entryFromStatus(e *log.Entry, s client.StatusEntry) *log.Entry {
	return e.WithFields(map[string]interface{}{
		"name":       s.GetName(),
		"uid":        s.GetUID(),
		"apiVersion": s.GetAPIVersion(),
	})
}

func contains(s []unstructured.Unstructured, name string) bool {
	for _, u := range s {
		if u.GetName() == name {
			return true
		}
	}
	return false
}

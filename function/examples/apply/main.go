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
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	//_ "k8s.io/client-go/plugin/pkg/client/auth"
	"os"
	"path"
	"path/filepath"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

const (
	usage = `apply description

Usage:
	apply [ --dir=<DIR> ] [ --dry-run ] [ --kube-config=<FILE> ] [options]

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

func statusLoggingCallback(e *log.Entry) func(interface{}, error) error {
	return func(v interface{}, err error) error {
		s, ok := v.(client.PostStatusEntry)
		if !ok {
			return fmt.Errorf("invalid callback argument type")
		}
		entryFromStatus(e, s).Debug(fmt.Sprintf("object %s", s.StatusType))
		return err
	}
}

func callbackIgnoreNotFound(_ interface{}, err error) error {
	if !errors.IsNotFound(err) {
		return err
	}
	return nil
}

func callbackStatusGetter(in *client.PostStatusEntry) func(interface{}, error) error {
	return func(v interface{}, err error) error {
		entry, ok := v.(client.PostStatusEntry)
		if !ok {
			return fmt.Errorf("invalid callback argument type")
		}
		*in = entry
		return err
	}
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

	entry = log.NewEntry(log.StandardLogger())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var statusEntry client.PostStatusEntry
	var fnOwnerReferences []metav1.OwnerReference

	dynamicInterface := getClient(cfg)

	if configuration.Source.Type == workspace.SourceTypeGit {
		gitRepository, err := unstructfn.NewPublicGitRepository(configuration)
		if err != nil {
			entry.Fatal(err)
		}

		gitRepositoryOperator := newOperator(
			operator.NewGenericOperator,
			operator.GVRGitRepository,
			configuration.Namespace,
			dynamicInterface, []unstructured.Unstructured{
				gitRepository,
			})

		if err = gitRepositoryOperator.Apply(ctx, operator.ApplyOptions{
			Options: operator.Options{
				DryRun: stages,
				Callbacks: operator.Callbacks{
					Post: []operator.Callback{
						statusLoggingCallback(entry),
						callbackStatusGetter(&statusEntry),
					},
				},
			},
		}); err != nil {
			entry.Fatal(err)
		}

		fnOwnerReferences = []metav1.OwnerReference{
			{
				APIVersion: statusEntry.GetAPIVersion(),
				Kind:       statusEntry.GetKind(),
				Name:       statusEntry.GetName(),
				UID:        statusEntry.GetUID(),
			},
		}
	}

	entryFromCfg(entry, configuration).Debug("generating function from configuration")
	function, err := unstructfn.NewFunction(configuration)
	if err != nil {
		entry.Fatal(err)
	}
	entryFromUnstructured(entry, &function).Debug("function generated")

	entryFromCfg(entry, configuration).Debug("generating triggers from configuration")
	triggers, err := unstructfn.NewTriggers(configuration)
	if err != nil {
		entry.Fatal(err)
	}
	for _, trigger := range triggers {
		entryFromUnstructured(entry, &trigger).Debug("trigger generated")
	}

	// Build function operator
	fnOperator := newOperator(
		operator.NewGenericOperator,
		operator.GVKFunction,
		configuration.Namespace,
		dynamicInterface,
		[]unstructured.Unstructured{function},
	)

	// Try to apply function
	if err := fnOperator.Apply(
		ctx,
		operator.ApplyOptions{
			OwnerReferences: fnOwnerReferences,
			Options: operator.Options{
				DryRun: stages,
				Callbacks: operator.Callbacks{
					Post: []operator.Callback{
						statusLoggingCallback(entry),
						callbackStatusGetter(&statusEntry),
					},
				},
			},
		},
	); err != nil {
		entry.Fatal(err)
	}

	// Build triggers operator
	trOperator := newOperator(
		operator.NewTriggersOperator,
		operator.GVKTriggers,
		configuration.Namespace,
		dynamicInterface,
		triggers,
	)

	// Try to apply triggers
	err = trOperator.Apply(
		ctx,
		operator.ApplyOptions{
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: statusEntry.GetAPIVersion(),
					Kind:       statusEntry.GetKind(),
					Name:       statusEntry.GetName(),
					UID:        statusEntry.GetUID(),
				},
			},
			Options: operator.Options{
				DryRun: stages,
				Callbacks: operator.Callbacks{
					Pre: nil,
					Post: []operator.Callback{
						statusLoggingCallback(entry),
					},
				},
			},
		},
	)
}

type Provider = func(client.Client, ...unstructured.Unstructured) operator.Operator

func newOperator(p Provider, gvr schema.GroupVersionResource, namespace string, dI dynamic.Interface, u []unstructured.Unstructured) operator.Operator {
	resourceInterface := dI.Resource(gvr).Namespace(namespace)
	resourceOperator := p(resourceInterface, u...)
	return resourceOperator
}

func entryFromCfg(e *log.Entry, cfg workspace.Cfg) *log.Entry {
	return e.WithFields(map[string]interface{}{
		"workspaceName":      cfg.Name,
		"workspaceNamespace": cfg.Namespace,
		"workspaceSourceType": func() string {
			if cfg.Source.Type == workspace.SourceTypeGit {
				return "git"
			}
			return "inline"
		}(),
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

func entryFromStatus(e *log.Entry, s client.PostStatusEntry) *log.Entry {
	return e.WithFields(map[string]interface{}{
		"name":       s.GetName(),
		"uid":        s.GetUID(),
		"apiVersion": s.GetAPIVersion(),
	})
}

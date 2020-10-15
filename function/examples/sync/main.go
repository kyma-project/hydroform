/*
* CODE GENERATED AUTOMATICALLY WITH devops/internal/config
 */

package main

import (
	"context"
	"github.com/docopt/docopt-go"
	"github.com/kyma-incubator/hydroform/function/pkg/resources/types"
	"github.com/kyma-incubator/hydroform/function/pkg/workspace"
	"github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	usage = `sync description

Usage:
	sync <name> [ --kubeconfig=<FILE> ] [ --output=<DIR> ] [options]

Options:
	-n --namespace=<NAMESPACE>  Choose namespace for your function [default: default].
	--debug                 Enable verbose output.
	-h --help               Show this screen.
	--version               Show version.`

	version = "0.0.1"

	functions = "functions"
	gitrepositories = "gitrepositories"
	git = "git"
)

type config struct {
	Name      string `json:"name"`
	Namespace string `docopt:"--namespace"`
	KubeConfig string `docopt:"--kubeconfig"`
	OutputPath string `docopt:"--output"`
	Debug      bool   `docopt:"--debug" json:"debug"`
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

func prepareCrdConfig(config config) (*rest.Config, error){
	crdConfig, err := clientcmd.BuildConfigFromFlags("", config.KubeConfig)
	if err != nil {
		return nil, err
	}

	crdConfig.ContentConfig.GroupVersion = &schema.GroupVersion{Group: v1alpha1.GroupVersion.Group, Version: v1alpha1.GroupVersion.Version}
	crdConfig.APIPath = "/apis"
	crdConfig.NegotiatedSerializer = serializer.NewCodecFactory(scheme.Scheme)
	return crdConfig, nil
}

func prepareWorkspace(config config, function v1alpha1.Function, restClient *rest.RESTClient) error{
	var source workspace.Source
	if function.Spec.Type == git {
		gitRepo := &v1alpha1.GitRepository{}

		err := restClient.Get().Resource(gitrepositories).Namespace(config.Namespace).Name(config.Name).Do(context.Background()).Into(gitRepo)
		if err != nil {
			return err
		}

		source = workspace.Source{
			Type: workspace.SourceTypeGit,
			SourceGit: workspace.SourceGit{
				URL:        gitRepo.Spec.URL,
				Reference:  function.Spec.Reference,
				BaseDir:    function.Spec.BaseDir,
			},
		}

		configuration := workspace.Cfg{
			Name:      function.Name,
			Namespace: config.Namespace,
			Runtime:   types.Runtime(function.Spec.Runtime),
			Source:    source,
		}

		if err := workspace.Initialize(configuration, config.OutputPath); err != nil {
			return err
		}
	} else {
		configuration := workspace.Cfg{
			Name:      function.Name,
			Namespace: function.Namespace,
			Labels:    function.Labels,
			Runtime:   types.Runtime(function.Spec.Runtime),
			Source: workspace.Source{
				Type: workspace.SourceTypeInline,
				SourceInline: workspace.SourceInline{
					SourcePath:        config.OutputPath,
				},
			},
		}

		if err := workspace.InitializeFromFunction(function,configuration, config.OutputPath); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	config, err := newConfig()
	if err != nil {
		panic(err.Error())
	}

	crdConfig, err := prepareCrdConfig(*config)
	if err != nil {
		panic(err.Error())
	}

	restClient, err := rest.UnversionedRESTClientFor(crdConfig)
	if err != nil {
		panic(err.Error())
	}

	function := &v1alpha1.Function{}
	//kf-test-func kf-test-inline
	err = restClient.Get().Resource(functions).Namespace(config.Namespace).Name(config.Name).Do(context.Background()).Into(function)
	if err != nil {
		panic(err.Error())
	}

	err = prepareWorkspace(*config,*function,restClient)
	if err != nil {
		panic(err.Error())
	}

	log.Println("Syncing completed.")
}

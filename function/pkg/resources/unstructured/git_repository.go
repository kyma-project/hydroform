package unstructured

import (
	"github.com/kyma-incubator/hydroform/function/pkg/workspace"
	"github.com/mitchellh/mapstructure"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const gitRepositoryApiVersion = "serverless.kyma-project.io/v1alpha1"

func NewPublicGitRepository(cfg workspace.Cfg) (out unstructured.Unstructured, err error) {
	var source workspace.SourceGit
	if err = mapstructure.Decode(cfg.Source, &source); err != nil {
		return
	}

	decorators := Decorators{
		withLabels(cfg.Labels),
		withMetadata(source.Repository, cfg.Namespace),
		withURL(source.URL),
		withGitRepository,
	}

	err = decorate(&out, decorators)
	return
}

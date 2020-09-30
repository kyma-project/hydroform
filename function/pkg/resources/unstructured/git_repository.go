package unstructured

import (
	"github.com/kyma-incubator/hydroform/function/pkg/workspace"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const gitRepositoryApiVersion = "serverless.kyma-project.io/v1alpha1"

func NewPublicGitRepository(cfg workspace.Cfg) (out unstructured.Unstructured, err error) {
	decorators := Decorators{
		withLabels(cfg.Labels),
		withMetadata(cfg.Source.Repository, cfg.Namespace),
		withURL(cfg.Source.URL),
		withGitRepository,
	}

	err = decorate(&out, decorators)
	return
}

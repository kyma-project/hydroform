package unstructured

import (
	"github.com/kyma-incubator/hydroform/function/pkg/workspace"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const gitRepositoryApiVersion = "serverless.kyma-project.io/v1alpha1"

func NewPublicGitRepository(cfg workspace.Cfg) (out unstructured.Unstructured, err error) {
	decorators := Decorators{
		decorateWithLabels(cfg.Labels),
		decorateWithMetadata(cfg.Source.Repository, cfg.Namespace),
		decorateWithField(cfg.Source.URL, "spec", "url"),
		decorateWithGitRepository,
	}

	err = decorate(&out, decorators)
	return
}

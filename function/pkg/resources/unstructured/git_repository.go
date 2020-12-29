package unstructured

import (
	"github.com/kyma-incubator/hydroform/function/pkg/resources/types"
	"github.com/kyma-incubator/hydroform/function/pkg/workspace"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

const gitRepositoryApiVersion = "serverless.kyma-project.io/v1alpha1"

func NewPublicGitRepository(cfg workspace.Cfg) (out unstructured.Unstructured, err error) {
	name := cfg.Name
	if cfg.Source.Repository != "" {
		name = cfg.Source.Repository
	}

	gitRepo := types.GitRepository{
		ApiVersion: functionApiVersion,
		Kind:       "GitRepository",
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: cfg.Namespace,
		},
		Spec: types.GitRepositorySpec{
			URL:  cfg.Source.URL,
			Auth: nil,
		},
	}

	unstructuredRepo, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&gitRepo)
	if err != nil {
		return unstructured.Unstructured{}, err
	}
	out = unstructured.Unstructured{Object: unstructuredRepo}

	return
}

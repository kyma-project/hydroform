package unstructured

import (
	"github.com/kyma-project/hydroform/function/pkg/resources/types"
	"github.com/kyma-project/hydroform/function/pkg/workspace"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

const gitRepositoryAPIVersion = "serverless.kyma-project.io/v1alpha1"

func NewPublicGitRepository(cfg workspace.Cfg) (out unstructured.Unstructured, err error) {
	name := cfg.Name
	if cfg.Source.Repository != "" {
		name = cfg.Source.Repository
	}

	credentialsType := "basic"
	if cfg.Source.CredentialsType != "" {
		credentialsType = cfg.Source.CredentialsType
	}

	var auth *types.RepositoryAuth
	if cfg.Source.CredentialsSecretName != "" {
		auth = &types.RepositoryAuth{
			Type:       credentialsType,
			SecretName: cfg.Source.CredentialsSecretName,
		}
	}

	gitRepo := types.GitRepository{
		APIVersion: functionAPIVersion,
		Kind:       "GitRepository",
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: cfg.Namespace,
		},
		Spec: types.GitRepositorySpec{
			URL:  cfg.Source.URL,
			Auth: auth,
		},
	}

	unstructuredRepo, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&gitRepo)
	if err != nil {
		return unstructured.Unstructured{}, err
	}
	out = unstructured.Unstructured{Object: unstructuredRepo}

	return
}

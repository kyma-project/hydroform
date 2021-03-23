package unstructured

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/kyma-incubator/hydroform/function/pkg/resources/types"
	"github.com/kyma-incubator/hydroform/function/pkg/workspace"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	apiRuleApiVersion = "gateway.kyma-project.io/v1alpha1"
	apiRuleKind       = "APIRule"
	apiRuleGateway    = "kyma-gateway.kyma-system.svc.cluster.local"
)

func NewApiRule(cfg workspace.Cfg) (out unstructured.Unstructured, err error) {

	apiRule := types.ApiRule{
		ApiVersion: apiRuleApiVersion,
		Kind:       apiRuleKind,
		ObjectMeta: metav1.ObjectMeta{
			Name:      cfg.ApiRule.Name,
			Namespace: cfg.Namespace,
			Labels:    cfg.Labels,
		},
		Spec: types.ApiRuleSpec{
			Service: types.Service{
				Host: cfg.ApiRule.Host,
				Name: cfg.Name,
				Port: cfg.ApiRule.Port,
			},
			Rules: []types.Rules{{
				//		AccessStrategies: []types.AccessStrategies{{
				//			Config:  types.Config{
				//				JwksUrls:       cfg.ApiRule.JwksUrls,
				//				TrustedIssuers: cfg.ApiRule.TrustedIssuers,
				//			},
				//			Handler: cfg.ApiRule.Handler,
				//		}},
				//		Methods: cfg.ApiRule.Methods,
				Path: cfg.ApiRule.Path,
			}},
			Gateway: apiRuleGateway,
		},
	}

	unstructuredRepo, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&apiRule)
	if err != nil {
		return unstructured.Unstructured{}, err
	}
	out = unstructured.Unstructured{Object: unstructuredRepo}

	return
}

package unstructured

import (
	"fmt"
	"reflect"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/kyma-incubator/hydroform/function/pkg/resources/types"
	"github.com/kyma-incubator/hydroform/function/pkg/workspace"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	apiRuleApiVersion = "gateway.kyma-project.io/v1alpha1"
	apiRuleKind       = "APIRule"
)

func NewApiRule(cfg workspace.Cfg, clusterAddress string) ([]unstructured.Unstructured, error) {
	var out []unstructured.Unstructured
	for _, cfgApiRule := range cfg.ApiRules {
		apiRule := prepareApiRule(cfg.Name, cfg.Namespace, clusterAddress, cfg.Labels, cfgApiRule)

		unstructuredRepo, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&apiRule)
		if err != nil {
			return []unstructured.Unstructured{}, err
		}

		out = append(out, unstructured.Unstructured{Object: unstructuredRepo})
	}

	return out, nil
}

func prepareApiRule(name, namespace, host string, labels map[string]string, apiRule workspace.ApiRule) types.ApiRule {
	return types.ApiRule{
		ApiVersion: apiRuleApiVersion,
		Kind:       apiRuleKind,
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultString(apiRule.Name, name),
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: types.ApiRuleSpec{
			Service: types.Service{
				Host: defaultString(apiRule.Service.Host, fmt.Sprintf("%s.%s", name, host)),
				Port: defaultInt64(apiRule.Service.Port, workspace.ApiRulePort),
				Name: name,
			},
			Rules:   prepareRules(apiRule.Rules),
			Gateway: defaultString(apiRule.Gateway, workspace.ApiRuleGateway),
		},
	}
}

func prepareRules(rules []workspace.Rule) []types.Rule {
	var typesRules []types.Rule
	for _, rule := range rules {
		typesRules = append(typesRules, types.Rule{
			AccessStrategies: prepareAccessStrategies(rule.AccessStrategies),
			Methods:          rule.Methods,
			Path:             defaultString(rule.Path, workspace.ApiRulePath),
		})
	}
	return typesRules
}

func prepareAccessStrategies(accessStrategies []workspace.AccessStrategie) []types.AccessStrategie {
	if len(accessStrategies) == 0 {
		return []types.AccessStrategie{
			{
				Handler: workspace.ApiRuleHandler,
			},
		}
	}

	strategies := []types.AccessStrategie{}
	for _, strategie := range accessStrategies {
		as := types.AccessStrategie{
			Handler: defaultString(strategie.Handler, workspace.ApiRuleHandler),
		}
		if !reflect.DeepEqual(strategie.Config, workspace.AccessStrategieConfig{}) {
			as.Config = &types.Config{
				JwksUrls:       strategie.Config.JwksUrls,
				TrustedIssuers: strategie.Config.TrustedIssuers,
				RequiredScope:  strategie.Config.RequiredScope,
			}
		}
		strategies = append(strategies, as)
	}
	return strategies
}

func defaultString(val, or string) string {
	if val == "" {
		return or
	}
	return val
}

func defaultInt64(val, or int64) int64 {
	if val == 0 {
		return or
	}
	return val
}

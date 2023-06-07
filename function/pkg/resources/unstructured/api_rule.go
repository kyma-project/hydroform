package unstructured

import (
	"reflect"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/kyma-project/hydroform/function/pkg/resources/types"
	"github.com/kyma-project/hydroform/function/pkg/workspace"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	apiRuleAPIVersion = "gateway.kyma-project.io/v1beta1"
	apiRuleKind       = "APIRule"
)

func NewAPIRule(cfg workspace.Cfg) ([]unstructured.Unstructured, error) {
	var out []unstructured.Unstructured
	for _, cfgAPIRule := range cfg.APIRules {
		apiRule := prepareAPIRule(cfg.Name, cfg.Namespace, cfg.Labels, cfgAPIRule)

		unstructuredRepo, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&apiRule)
		if err != nil {
			return []unstructured.Unstructured{}, err
		}

		out = append(out, unstructured.Unstructured{Object: unstructuredRepo})
	}

	return out, nil
}

func prepareAPIRule(name, namespace string, labels map[string]string, apiRule workspace.APIRule) types.APIRule {
	return types.APIRule{
		APIVersion: apiRuleAPIVersion,
		Kind:       apiRuleKind,
		ObjectMeta: metav1.ObjectMeta{
			Name:   defaultString(apiRule.Name, name),
			Labels: labels,
		},
		Spec: types.APIRuleSpec{
			Gateway: defaultString(apiRule.Gateway, workspace.APIRuleGateway),
			Host:    defaultString(apiRule.Service.Host, name),
			Service: types.Service{
				Name:      name,
				Namespace: namespace,
				Port:      defaultInt64(apiRule.Service.Port, workspace.APIRulePort),
			},
			Rules: prepareRules(apiRule.Rules),
		},
	}
}

func prepareRules(rules []workspace.Rule) []types.Rule {
	var typesRules []types.Rule
	for _, rule := range rules {
		typesRules = append(typesRules, types.Rule{
			Path:             defaultString(rule.Path, workspace.APIRulePath),
			Methods:          rule.Methods,
			AccessStrategies: prepareAccessStrategies(rule.AccessStrategies),
		})
	}
	return typesRules
}

func prepareAccessStrategies(accessStrategies []workspace.AccessStrategie) []types.AccessStrategie {
	if len(accessStrategies) == 0 {
		return []types.AccessStrategie{
			{
				Handler: workspace.APIRuleHandler,
			},
		}
	}

	strategies := []types.AccessStrategie{}
	for _, strategie := range accessStrategies {
		as := types.AccessStrategie{
			Handler: defaultString(strategie.Handler, workspace.APIRuleHandler),
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

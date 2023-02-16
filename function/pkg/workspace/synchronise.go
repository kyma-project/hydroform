package workspace

import (
	"context"

	"github.com/kyma-project/hydroform/function/pkg/client"
	"github.com/kyma-project/hydroform/function/pkg/operator"
	"github.com/kyma-project/hydroform/function/pkg/resources/types"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	coretypes "k8s.io/apimachinery/pkg/types"
)

const (
	APIRuleGateway = "kyma-gateway.kyma-system.svc.cluster.local"
	APIRulePath    = "/.*"
	APIRuleHandler = "allow"
	APIRulePort    = int64(80)
)

func Synchronise(ctx context.Context, config Cfg, outputPath string, build client.Build) error {
	return synchronise(ctx, config, outputPath, build, DefaultWriterProvider)
}

func synchronise(ctx context.Context, config Cfg, outputPath string, build client.Build, writerProvider WriterProvider) error {

	u, err := build(config.Namespace, operator.GVRFunction).Get(ctx, config.Name, v1.GetOptions{})
	if err != nil {
		return err
	}

	var function types.Function
	if err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &function); err != nil {
		return err
	}

	if config.Resources.Limits != nil {
		config.Resources.Limits = function.Spec.ResourceLimits()
	}
	if config.Resources.Requests != nil {
		config.Resources.Requests = function.Spec.ResourceRequests()
	}

	config.Runtime = function.Spec.Runtime
	config.RuntimeImageOverride = function.Spec.RuntimeImageOverride
	config.Labels = function.ObjectMeta.Labels
	config.Env = toWorkspaceEnvVar(function.Spec.Env)

	ul, err := build(config.Namespace, operator.GVRSubscription).List(ctx, v1.ListOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return err
	}

	if ul != nil {
		for _, item := range ul.Items {
			var subscription types.SubscriptionV1alpha1
			if err = runtime.DefaultUnstructuredConverter.FromUnstructured(item.Object, &subscription); err != nil {
				return err
			}

			isRef := subscription.IsReference(function.Name, function.Namespace)
			isOwnerRef := (len(subscription.OwnerReferences) == 0 || isOwnerReference(subscription.OwnerReferences, u.GetUID()))
			if !isRef || !isOwnerRef {
				continue
			}

			filterLen := subscription.Spec.Filter.Filters
			if len(filterLen) == 0 {
				continue
			}

			var filters []EventFilter
			for _, fromFilter := range subscription.Spec.Filter.Filters {
				toFilter := toWorkspaceEnvFilter(fromFilter)
				filters = append(filters, toFilter)
			}

			config.Subscriptions = append(config.Subscriptions, Subscription{
				Name: subscription.Name,
				V0: SubscriptionV0{
					Protocol: subscription.Spec.Protocol,
					Filter: Filter{
						Dialect: subscription.Spec.Filter.Dialect,
						Filters: filters,
					},
				},
			})
		}
	}

	ul, err = build(config.Namespace, operator.GVRApiRule).List(ctx, v1.ListOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return err
	}

	if ul != nil {
		for _, item := range ul.Items {
			var apiRule types.APIRule
			if err = runtime.DefaultUnstructuredConverter.FromUnstructured(item.Object, &apiRule); err != nil {
				return err
			}

			isRef := apiRule.IsReference(function.Name)
			isOwnerRef := (len(apiRule.OwnerReferences) == 0 || isOwnerReference(apiRule.OwnerReferences, u.GetUID()))
			if !isRef || !isOwnerRef {
				continue
			}

			newAPIRule := APIRule{
				Name:    setIfNotEqual(apiRule.Name, function.Name),
				Gateway: setIfNotEqual(apiRule.Spec.Gateway, APIRuleGateway),
				Service: Service{
					Host: apiRule.Spec.Host,
				},
				Rules: toWorkspaceRules(apiRule.Spec.Rules),
			}

			if apiRule.Spec.Service.Port != APIRulePort {
				newAPIRule.Service.Port = apiRule.Spec.Service.Port
			}

			config.APIRules = append(config.APIRules, newAPIRule)
		}
	}

	var ws workspace
	if function.Spec.Source.Inline != nil {
		ws, err = createInlineWorkspace(&config, outputPath, function)
	}
	if function.Spec.Source.GitRepository != nil {
		createGitConfig(&config, function)
	}
	if err != nil {
		return err
	}
	return ws.build(config, outputPath, writerProvider)
}

func createGitConfig(config *Cfg, function types.Function) {
	config.Source = Source{
		Type: SourceTypeGit,
		SourceGit: SourceGit{
			URL:       function.Spec.Source.GitRepository.URL,
			Reference: function.Spec.Source.GitRepository.Reference,
			BaseDir:   function.Spec.Source.GitRepository.BaseDir,
		},
	}
	if function.Spec.Source.GitRepository.Auth != nil {
		config.Source.SourceGit.CredentialsSecretName = function.Spec.Source.GitRepository.Auth.SecretName
		config.Source.SourceGit.CredentialsType = string(function.Spec.Source.GitRepository.Auth.Type)
	}
}

func createInlineWorkspace(config *Cfg, outputPath string, function types.Function) (workspace, error) {
	config.Source = Source{
		Type: SourceTypeInline,
		SourceInline: SourceInline{
			SourcePath: outputPath,
		},
	}
	return fromSources(function.Spec.Runtime, function.Spec.Source.Inline.Source, function.Spec.Source.Inline.Dependencies)
}

func toWorkspaceEnvVar(envs []corev1.EnvVar) []EnvVar {
	outEnvs := make([]EnvVar, 0)
	for _, env := range envs {

		newEnv := EnvVar{
			Name:  env.Name,
			Value: env.Value,
		}

		if env.ValueFrom != nil {
			newEnv.ValueFrom = &EnvVarSource{}

			if env.ValueFrom.SecretKeyRef != nil {
				newEnv.ValueFrom.SecretKeyRef = &SecretKeySelector{
					Name: env.ValueFrom.SecretKeyRef.Name,
					Key:  env.ValueFrom.SecretKeyRef.Key,
				}
			}

			if env.ValueFrom.ConfigMapKeyRef != nil {
				newEnv.ValueFrom.ConfigMapKeyRef = &ConfigMapKeySelector{
					Name: env.ValueFrom.ConfigMapKeyRef.Name,
					Key:  env.ValueFrom.ConfigMapKeyRef.Key,
				}
			}
		}
		outEnvs = append(outEnvs, newEnv)
	}
	return outEnvs
}

func toWorkspaceEnvFilter(filter types.EventFilter) EventFilter {
	return EventFilter{
		EventSource: EventSource{
			Property: filter.EventSource.Property,
			Type:     filter.EventSource.Type,
			Value:    filter.EventSource.Value,
		},
		EventType: EventType{
			Property: filter.EventType.Property,
			Type:     filter.EventType.Type,
			Value:    filter.EventType.Value,
		},
	}
}

func toWorkspaceRules(rules []types.Rule) []Rule {
	var out []Rule
	for _, rule := range rules {
		out = append(out, Rule{
			Path:             setIfNotEqual(rule.Path, APIRulePath),
			Methods:          rule.Methods,
			AccessStrategies: toWorkspaceAccessStrategies(rule.AccessStrategies),
		})
	}

	return out
}

func toWorkspaceAccessStrategies(accessStrategies []types.AccessStrategie) []AccessStrategie {
	var out []AccessStrategie
	for _, as := range accessStrategies {
		strategie := AccessStrategie{
			Handler: as.Handler,
		}
		if as.Config != nil {
			strategie.Config.JwksUrls = as.Config.JwksUrls
			strategie.Config.TrustedIssuers = as.Config.TrustedIssuers
			strategie.Config.RequiredScope = as.Config.RequiredScope
		}
		out = append(out, strategie)
	}

	return out
}

func setIfNotEqual(val, defVal string) string {
	if val != defVal {
		return val
	}
	return ""
}

func isOwnerReference(references []v1.OwnerReference, ownerUID coretypes.UID) bool {
	for _, ref := range references {
		if ref.UID == ownerUID {
			return true
		}
	}

	return false
}

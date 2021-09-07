package workspace

import (
	"context"

	"github.com/kyma-incubator/hydroform/function/pkg/client"
	"github.com/kyma-incubator/hydroform/function/pkg/operator"
	"github.com/kyma-incubator/hydroform/function/pkg/resources/types"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	APIRuleGateway     = "kyma-gateway.kyma-system.svc.cluster.local"
	APIRulePath        = "/.*"
	APIRuleHandler     = "allow"
	APIRulePort        = int64(80)
	functionApiVersion = "serverless.kyma-project.io/v1alpha1"
	functionKind       = "Function"
)

func Synchronise(ctx context.Context, config Cfg, outputPath string, build client.Build) error {
	return synchronise(ctx, config, outputPath, build, defaultWriterProvider)
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
	config.Labels = function.Spec.Labels
	config.Env = toWorkspaceEnvVar(function.Spec.Env)

	ul, err := build(config.Namespace, operator.GVRSubscription).List(ctx, v1.ListOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return err
	}

	if ul != nil {
		for _, item := range ul.Items {
			var subscription types.Subscription
			if err = runtime.DefaultUnstructuredConverter.FromUnstructured(item.Object, &subscription); err != nil {
				return err
			}

			if !subscription.IsReference(function.Name, function.Namespace) ||
				(!isOwnerReference(subscription.OwnerReferences, config.Name) && len(subscription.OwnerReferences) != 0) {
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
				Name:     subscription.Name,
				Protocol: subscription.Spec.Protocol,
				Filter: Filter{
					Dialect: subscription.Spec.Filter.Dialect,
					Filters: filters,
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

			if !apiRule.IsReference(function.Name) ||
				(!isOwnerReference(apiRule.OwnerReferences, config.Name) && len(apiRule.OwnerReferences) != 0) {
				continue
			}

			newAPIRule := APIRule{
				Name:    setIfNotEqual(apiRule.Name, function.Name),
				Gateway: setIfNotEqual(apiRule.Spec.Gateway, APIRuleGateway),
				Service: Service{
					Host: apiRule.Spec.Service.Host,
				},
				Rules: toWorkspaceRules(apiRule.Spec.Rules),
			}

			if apiRule.Spec.Service.Port != APIRulePort {
				newAPIRule.Service.Port = apiRule.Spec.Service.Port
			}

			config.APIRules = append(config.APIRules, newAPIRule)
		}
	}

	if function.Spec.Type == "git" {
		gitRepository := types.GitRepository{}

		u, err := build(config.Namespace, operator.GVRGitRepository).Get(ctx, function.Spec.Source, v1.GetOptions{})
		if err != nil {
			return err
		}

		if err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &gitRepository); err != nil {
			return err
		}

		config.Source = Source{
			Type: SourceTypeGit,
			SourceGit: SourceGit{
				URL:        gitRepository.Spec.URL,
				Repository: function.Spec.Source,
				Reference:  function.Spec.Reference,
				BaseDir:    function.Spec.BaseDir,
			},
		}
		return initialize(config, outputPath, writerProvider)
	}

	config.Source = Source{
		Type: SourceTypeInline,
		SourceInline: SourceInline{
			SourcePath: outputPath,
		},
	}
	ws, err := fromSources(function.Spec.Runtime, function.Spec.Source, function.Spec.Deps)
	if err != nil {
		return err
	}

	return ws.build(config, outputPath, writerProvider)
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
		EventSource: EventFilterProperty{
			Property: filter.EventSource.Property,
			Type:     filter.EventSource.Type,
			Value:    filter.EventSource.Value,
		},
		EventType: EventFilterProperty{
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

func isOwnerReference(references []v1.OwnerReference, owner string) bool {
	for _, ref := range references {
		if ref.APIVersion == functionApiVersion &&
			ref.Kind == functionKind &&
			ref.Name == owner {
			return true
		}
	}

	return false
}

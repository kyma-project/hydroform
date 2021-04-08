package workspace

import (
	"context"
	"io"
	"os"

	corev1 "k8s.io/api/core/v1"

	"github.com/kyma-incubator/hydroform/function/pkg/client"
	"github.com/kyma-incubator/hydroform/function/pkg/operator"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/kyma-incubator/hydroform/function/pkg/resources/types"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

type FileName string

type workspace []file

func (ws workspace) build(cfg Cfg, dirPath string, writerProvider WriterProvider) error {
	workspaceFiles := append(ws, cfg)
	for _, fileTemplate := range workspaceFiles {
		if err := writerProvider.write(dirPath, fileTemplate, cfg); err != nil {
			return err
		}
	}
	return nil
}

var defaultWriterProvider = func(outFilePath string) (io.Writer, func() error, error) {
	file, err := os.Create(outFilePath)
	if err != nil {
		return nil, nil, err
	}
	return file, file.Close, nil
}

var errUnsupportedRuntime = errors.New("unsupported runtime")

func Initialize(cfg Cfg, dirPath string) error {
	return initialize(cfg, dirPath, defaultWriterProvider)
}

func initialize(cfg Cfg, dirPath string, writerProvider WriterProvider) (err error) {
	ws := workspace{}
	if cfg.Source.Type != SourceTypeGit {
		ws, err = fromRuntime(cfg.Runtime)
	}
	if err != nil {
		return err
	}
	return ws.build(cfg, dirPath, writerProvider)
}

func fromSources(runtime string, source, deps string) (workspace, error) {
	switch runtime {
	case types.Nodejs10, types.Nodejs12:
		return workspace{
			newTemplatedFile(source, FileNameHandlerJs),
			newTemplatedFile(deps, FileNamePackageJSON),
		}, nil
	case types.Python38:
		return workspace{
			newTemplatedFile(source, FileNameHandlerPy),
			newTemplatedFile(deps, FileNameRequirementsTxt),
		}, nil
	default:
		return workspace{}, errUnsupportedRuntime
	}
}

func fromRuntime(runtime types.Runtime) (workspace, error) {
	switch runtime {
	case types.Nodejs12, types.Nodejs10:
		return workspaceNodeJs, nil
	case types.Python38:
		return workspacePython, nil
	default:
		return nil, errUnsupportedRuntime
	}
}

const (
	ApiRuleGateway = "kyma-gateway.kyma-system.svc.cluster.local"
	ApiRulePath    = "/.*"
	ApiRuleHandler = "allow"
	ApiRulePort    = int64(80)
)

func Synchronise(ctx context.Context, config Cfg, outputPath string, build client.Build, kymaAddress string) error {
	return synchronise(ctx, config, outputPath, build, defaultWriterProvider, kymaAddress)
}

func synchronise(ctx context.Context, config Cfg, outputPath string, build client.Build, writerProvider WriterProvider, kymaAddress string) error {

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

	ul, err := build("", operator.GVRSubscription).List(ctx, v1.ListOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return err
	}

	if ul != nil {
		for _, item := range ul.Items {
			var subscription types.Subscription
			if err = runtime.DefaultUnstructuredConverter.FromUnstructured(item.Object, &subscription); err != nil {
				return err
			}

			if !subscription.IsReference(function.Name, function.Namespace) {
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
			var apiRule types.ApiRule
			if err = runtime.DefaultUnstructuredConverter.FromUnstructured(item.Object, &apiRule); err != nil {
				return err
			}

			if !apiRule.IsReference(function.Name) {
				continue
			}

			newApiRule := ApiRule{
				Name:    setIfNotEqual(apiRule.Name, function.Name),
				Gateway: setIfNotEqual(apiRule.Spec.Gateway, ApiRuleGateway),
				Service: Service{
					Host: apiRule.Spec.Service.Host,
				},
				Rules: toWorkspaceRules(apiRule.Spec.Rules),
			}

			if apiRule.Spec.Service.Port != ApiRulePort {
				newApiRule.Service.Port = apiRule.Spec.Service.Port
			}

			config.ApiRules = append(config.ApiRules, newApiRule)
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

type SourceFileName = string

type DepsFileName = string

func InlineFileNames(r types.Runtime) (SourceFileName, DepsFileName, bool) {
	switch r {
	case types.Nodejs10, types.Nodejs12:
		return string(FileNameHandlerJs), string(FileNamePackageJSON), true
	case types.Python38:
		return string(FileNameHandlerPy), string(FileNameRequirementsTxt), true
	default:
		return "", "", false
	}
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
			Path:             setIfNotEqual(rule.Path, ApiRulePath),
			Methods:          rule.Methods,
			AccessStrategies: toWorkspaceAccessStrategies(rule.AccessStrategies),
		})
	}

	return out
}

func toWorkspaceAccessStrategies(accessStrategies []types.AccessStrategie) []AccessStrategie {
	var out []AccessStrategie
	for _, as := range accessStrategies {
		out = append(out, AccessStrategie{
			Handler: as.Handler,
			Config: AccessStrategieConfig{
				JwksUrls:       as.Config.JwksUrls,
				TrustedIssuers: as.Config.TrustedIssuers,
				RequiredScope:  as.Config.RequiredScope,
			},
		})
	}

	return out
}

func setIfNotEqual(val, defVal string) string {
	if val != defVal {
		return val
	}
	return ""
}

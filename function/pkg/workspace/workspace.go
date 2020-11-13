package workspace

import (
	"context"
	"io"
	"os"

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

func Synchronise(ctx context.Context, config Cfg, outputPath string, build client.Build) error {
	return synchronise(ctx, config, outputPath, build, defaultWriterProvider)
}

func synchronise(ctx context.Context, config Cfg, outputPath string, build client.Build, writerProvider WriterProvider) error {

	u, err := build(config.Namespace, operator.GVKFunction).Get(ctx, config.Name, v1.GetOptions{})
	if err != nil {
		return err
	}

	var function types.Function
	if err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &function); err != nil {
		return err
	}

	config.Runtime = function.Spec.Runtime
	config.Resources.Limits = function.Spec.ResourceLimits()
	config.Resources.Requests = function.Spec.ResourceRequests()

	ul, err := build("", operator.GVKTriggers).List(ctx, v1.ListOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return err
	}

	if ul != nil {
		for _, item := range ul.Items {
			var trigger types.Trigger
			if err = runtime.DefaultUnstructuredConverter.FromUnstructured(item.Object, &trigger); err != nil {
				return err
			}

			if !trigger.IsReference(function.Name, function.Namespace) {
				continue
			}

			config.Triggers = append(config.Triggers, Trigger{
				Version: trigger.Spec.Filter.Attributes.Eventtypeversion,
				Source:  trigger.Spec.Filter.Attributes.Source,
				Type:    trigger.Spec.Filter.Attributes.Type,
				Name:    trigger.Metadata.Name,
			})
		}
	}

	if function.Spec.Type == "git" {
		gitRepository := types.GitRepository{}
		u, err := build(config.Namespace, operator.GVRGitRepository).Get(ctx, config.Name, v1.GetOptions{})
		if err != nil {
			return err
		}

		if err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &gitRepository); err != nil {
			return err
		}

		config.Source = Source{
			Type: SourceTypeGit,
			SourceGit: SourceGit{
				URL:       gitRepository.Spec.URL,
				Reference: function.Spec.Reference,
				BaseDir:   function.Spec.BaseDir,
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

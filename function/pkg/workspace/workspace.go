package workspace

import (
	"github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"io"
	"os"

	"github.com/kyma-incubator/hydroform/function/pkg/resources/types"
	"github.com/pkg/errors"
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

func InitializeFromFunction(function v1alpha1.Function,cfg Cfg, dirPath string) error {
	return initializeFromFunction(function,cfg, dirPath, defaultWriterProvider)
}

func initializeFromFunction(function v1alpha1.Function,cfg Cfg, dirPath string, writerProvider WriterProvider) (err error) {

	var sourceFileName FileName
	var depsFileName FileName

	switch function.Spec.Runtime {
	case v1alpha1.Nodejs12, v1alpha1.Nodejs10:
		sourceFileName = FileNameHandlerJs
		depsFileName = FileNamePackageJSON
	case v1alpha1.Python38:
		sourceFileName = FileNameHandlerPy
		depsFileName = FileNameRequirementsTxt
	default:
		return errUnsupportedRuntime
	}

	ws := workspace{
		newTemplatedFile(function.Spec.Source, sourceFileName),
		newTemplatedFile(function.Spec.Deps, depsFileName),
	}
	return ws.build(cfg, dirPath, writerProvider)
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

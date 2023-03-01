package workspace

import (
	"io"
	"os"

	"github.com/kyma-project/hydroform/function/pkg/resources/types"
	"github.com/pkg/errors"
)

type FileName string

type workspace []File

func (ws workspace) build(cfg Cfg, dirPath string, writerProvider WriterProvider) error {
	workspaceFiles := append(ws, cfg)
	for _, fileTemplate := range workspaceFiles {
		if err := writerProvider.Write(dirPath, fileTemplate, cfg); err != nil {
			return err
		}
	}
	return nil
}

var DefaultWriterProvider = func(outFilePath string) (io.Writer, func() error, error) {
	file, err := os.Create(outFilePath)
	if err != nil {
		return nil, nil, err
	}
	return file, file.Close, nil
}

var errUnsupportedRuntime = errors.New("unsupported runtime")

func Initialize(cfg Cfg, dirPath string) error {
	return initialize(cfg, dirPath, DefaultWriterProvider)
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
	case types.Nodejs14, types.Nodejs16, types.Nodejs18:
		return workspace{
			NewTemplatedFile(source, FileNameHandlerJs),
			NewTemplatedFile(deps, FileNamePackageJSON),
		}, nil
	case types.Python39:
		return workspace{
			NewTemplatedFile(source, FileNameHandlerPy),
			NewTemplatedFile(deps, FileNameRequirementsTxt),
		}, nil
	default:
		return workspace{}, errUnsupportedRuntime
	}
}

func fromRuntime(runtime types.Runtime) (workspace, error) {
	switch runtime {
	case types.Nodejs14, types.Nodejs16, types.Nodejs18:
		return workspaceNodeJs, nil
	case types.Python39:
		return workspacePython, nil
	default:
		return nil, errUnsupportedRuntime
	}
}

type SourceFileName = string

type DepsFileName = string

func InlineFileNames(r types.Runtime) (SourceFileName, DepsFileName, bool) {
	switch r {
	case types.Nodejs14, types.Nodejs16, types.Nodejs18:
		return string(FileNameHandlerJs), string(FileNamePackageJSON), true
	case types.Python39:
		return string(FileNameHandlerPy), string(FileNameRequirementsTxt), true
	default:
		return "", "", false
	}
}

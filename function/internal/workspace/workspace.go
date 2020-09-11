package workspace

import (
	"github.com/kyma-incubator/hydroform/function/internal/resources/types"
	"github.com/pkg/errors"
	"io"
	"os"
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

func initialize(cfg Cfg, dirPath string, writerProvider WriterProvider) error {
	ws, err := fromRuntime(cfg.Runtime)
	if err != nil {
		return err
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

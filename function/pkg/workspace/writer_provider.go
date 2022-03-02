package workspace

import (
	"io"
	"path"
)

type Cancel = func() error

type WriterProvider func(path string) (io.Writer, Cancel, error)

func (p WriterProvider) Write(destinationDirPath string, fileTemplate File, cfg interface{}) error {
	outFilePath := path.Join(destinationDirPath, fileTemplate.FileName())
	writer, closeFn, err := p(outFilePath)
	if err != nil {
		return err
	}
	defer func() {
		if closeFn == nil {
			return
		}
		_ = closeFn()
	}()

	err = fileTemplate.Write(writer, cfg)
	if err != nil {
		return err
	}

	return nil
}

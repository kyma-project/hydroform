package workspace

import (
	"io"
	"path"
)

type Cancel = func() error

type WriterProvider func(path string) (io.Writer, Cancel, error)

func (p WriterProvider) write(destinationDirPath string, fileTemplate file, cfg Cfg) error {
	outFilePath := path.Join(destinationDirPath, fileTemplate.fileName())
	writer, closeFn, err := p(outFilePath)
	if err != nil {
		return err
	}
	defer func() {
		if closeFn == nil {
			return
		}
		closeFn()
	}()

	err = fileTemplate.write(writer, cfg)
	if err != nil {
		return err
	}

	return nil
}

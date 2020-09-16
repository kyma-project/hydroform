package workspace

import (
	"io"
	"path"
)

type WriterProvider func(path string) (io.Writer, func() error, error)

func (p WriterProvider) write(destinationDirPath string, fileTemplate file, cfg Cfg) error {
	outFilePath := path.Join(destinationDirPath, fileTemplate.fileName())
	writer, closeFn, err := p(outFilePath)
	defer func() {
		err := closeFn()
		if err != nil {
			return
		}
	}()

	err = fileTemplate.write(writer, cfg)
	if err != nil {
		return err
	}

	return nil
}

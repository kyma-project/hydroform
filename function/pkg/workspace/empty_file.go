package workspace

import "io"

type emptyFile struct {
	name FileName
}

func (e emptyFile) Write(_ io.Writer, _ interface{}) error {
	return nil
}

func (e emptyFile) FileName() string {
	return string(e.name)
}

func newEmptyFile(name FileName) File {
	return emptyFile{
		name: name,
	}
}

package workspace

import (
	"io"
	"text/template"
)

var _ File = TemplatedFile{}

type TemplatedFile struct {
	name FileName
	tpl  string
}

func (t TemplatedFile) FileName() string {
	return string(t.name)
}

func (t TemplatedFile) Write(writer io.Writer, cfg interface{}) error {
	tpl, err := template.New("templatedFile").Parse(t.tpl)
	if err != nil {
		return err
	}

	return tpl.Execute(writer, cfg)
}

func NewTemplatedFile(tpl string, name FileName) File {
	return &TemplatedFile{
		tpl:  tpl,
		name: name,
	}
}

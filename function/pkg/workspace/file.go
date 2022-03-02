package workspace

import "io"

type File interface {
	Write(io.Writer, interface{}) error
	FileName() string
}

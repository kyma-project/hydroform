package workspace

import "io"

//go:generate mockgen -source=file.go -destination=automock/file.go

type file interface {
	write(io.Writer, interface{}) error
	fileName() string
}

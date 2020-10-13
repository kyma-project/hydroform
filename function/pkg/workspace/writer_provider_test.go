package workspace

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/kyma-incubator/hydroform/function/pkg/resources/types"
)

type errWriter struct{}

func (w *errWriter) Write(p []byte) (n int, err error) {
	return -1, fmt.Errorf("write error")
}

func TestWriterProvider_write(t *testing.T) {
	type args struct {
		destinationDirPath string
		fileTemplate       file
		cfg                Cfg
	}
	tests := []struct {
		name    string
		p       WriterProvider
		args    args
		wantErr bool
	}{
		{
			name: "writer provider error",
			p: func(path string) (io.Writer, Cancel, error) {
				return nil, nil, fmt.Errorf("writer provider error")
			},
			wantErr: true,
			args: args{
				destinationDirPath: "/testme",
				fileTemplate:       newTemplatedFile("test", "test"),
				cfg: Cfg{
					Name:      "test-name",
					Labels:    map[string]string{},
					Namespace: "test-namespace",
					Resources: Resources{},
					Runtime:   types.Nodejs10,
					Triggers:  []Trigger{},
				},
			},
		},
		{
			name: "write error",
			p: func(path string) (io.Writer, Cancel, error) {
				return func() io.Writer {
					return &errWriter{}
				}(), func() error { return nil }, nil
			},
			wantErr: true,
			args: args{
				destinationDirPath: "/testme",
				fileTemplate:       newTemplatedFile("test", "test"),
				cfg: Cfg{
					Name:      "test-name",
					Labels:    map[string]string{},
					Namespace: "test-namespace",
					Resources: Resources{},
					Runtime:   types.Nodejs10,
					Triggers:  []Trigger{},
				},
			},
		},
		{
			name: "happy path",
			p: func(path string) (io.Writer, Cancel, error) {
				return &bytes.Buffer{}, func() error { return nil }, nil
			},
			wantErr: false,
			args: args{
				destinationDirPath: "/testme",
				fileTemplate:       newTemplatedFile("test", "test"),
				cfg: Cfg{
					Name:      "test-name",
					Labels:    map[string]string{},
					Namespace: "test-namespace",
					Resources: Resources{},
					Runtime:   types.Nodejs10,
					Triggers:  []Trigger{},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.p.write(tt.args.destinationDirPath, tt.args.fileTemplate, tt.args.cfg); (err != nil) != tt.wantErr {
				t.Errorf("WriterProvider.write() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

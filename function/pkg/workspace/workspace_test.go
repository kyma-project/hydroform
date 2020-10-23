package workspace

import (
	"bytes"
	"io"
	"reflect"
	"testing"

	"github.com/kyma-incubator/hydroform/function/pkg/resources/types"
)

func Test_workspace_build(t *testing.T) {
	type args struct {
		cfg            Cfg
		dirPath        string
		writerProvider WriterProvider
	}
	tests := []struct {
		name    string
		ws      workspace
		args    args
		wantErr bool
	}{
		{
			name:    "write error",
			wantErr: true,
			args: args{
				writerProvider: func() WriterProvider {
					return func(path string) (io.Writer, Cancel, error) {
						return &errWriter{}, nil, nil
					}
				}(),
			},
		},
		{
			name:    "happy path",
			wantErr: false,
			args: args{
				writerProvider: func(path string) (io.Writer, Cancel, error) {
					return &bytes.Buffer{}, nil, nil
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.ws.build(tt.args.cfg, tt.args.dirPath, tt.args.writerProvider); (err != nil) != tt.wantErr {
				t.Errorf("workspace.build() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_initialize(t *testing.T) {
	type args struct {
		cfg            Cfg
		dirPath        string
		writerProvider WriterProvider
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "unsupported runtime",
			wantErr: true,
			args: args{
				cfg: Cfg{
					Runtime: types.Runtime("unsupported runtime"),
				},
			},
		},
		{
			name:    "happy path",
			wantErr: false,
			args: args{
				cfg: Cfg{
					Runtime: types.Python38,
					Triggers: []Trigger{
						{
							Version: "test-version",
							Source:           "test-source",
							Type:             "test-type",
						},
					},
				},
				writerProvider: func(path string) (io.Writer, Cancel, error) {
					return &bytes.Buffer{}, nil, nil
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := initialize(tt.args.cfg, tt.args.dirPath, tt.args.writerProvider); (err != nil) != tt.wantErr {
				t.Errorf("initialize() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_fromRuntime(t *testing.T) {
	type args struct {
		runtime types.Runtime
	}
	tests := []struct {
		name    string
		args    args
		want    workspace
		wantErr bool
	}{
		{
			name: "unsupported runtime error",
			args: args{
				runtime: types.Runtime("unsupported"),
			},
			wantErr: true,
		},
		{
			name: "nodejs10",
			args: args{
				runtime: types.Nodejs10,
			},
			want:    workspaceNodeJs,
			wantErr: false,
		},
		{
			name: "nodejs12",
			args: args{
				runtime: types.Nodejs12,
			},
			want:    workspaceNodeJs,
			wantErr: false,
		},
		{
			name: "python38",
			args: args{
				runtime: types.Python38,
			},
			want:    workspacePython,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fromRuntime(tt.args.runtime)
			if (err != nil) != tt.wantErr {
				t.Errorf("fromRuntime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("fromRuntime() = %v, want %v", got, tt.want)
			}
		})
	}
}

package workspace

import (
	"bytes"
	"io"
	"reflect"
	"testing"

	"github.com/kyma-project/hydroform/function/pkg/resources/types"
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
					Runtime: types.Python39,
					Subscriptions: []Subscription{
						{
							Name: "fixme",
							V0: &SubscriptionV0{
								Protocol: "fixme",
								Filter: Filter{
									Dialect: "fixme",
									Filters: []EventFilter{
										{
											EventSource: EventSource{
												Property: "source",
												Type:     "exact",
												Value:    "test-source",
											},
											EventType: EventType{
												Property: "type",
												Type:     "exact",
												Value:    "test-type.test-version",
											},
										},
									},
								},
							},
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
			name: "nodejs18",
			args: args{
				runtime: types.Nodejs18,
			},
			want:    workspaceNodeJs,
			wantErr: false,
		},
		{
			name: "nodejs16",
			args: args{
				runtime: types.Nodejs16,
			},
			want:    workspaceNodeJs,
			wantErr: false,
		},
		{
			name: "python39",
			args: args{
				runtime: types.Python39,
			},
			want:    workspacePython,
			wantErr: false,
		},
		{
			name: "python312",
			args: args{
				runtime: types.Python312,
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

func newStrWriterProvider() WriterProvider {
	return func(_ string) (io.Writer, Cancel, error) {
		var buffer bytes.Buffer
		return &buffer, func() error {
			return nil
		}, nil
	}
}

func Test_fromSources(t *testing.T) {
	type args struct {
		runtime types.Runtime
		source  string
		deps    string
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
				runtime: "unsupported",
				source:  "",
				deps:    "",
			},
			want:    workspace{},
			wantErr: true,
		},
		{
			name: "python39",
			args: args{
				runtime: types.Python39,
				source:  handlerPython,
				deps:    "deps",
			},
			want: workspace{
				NewTemplatedFile(handlerPython, FileNameHandlerPy),
				NewTemplatedFile("deps", FileNameRequirementsTxt),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fromSources(tt.args.runtime, tt.args.source, tt.args.deps)
			if (err != nil) != tt.wantErr {
				t.Errorf("fromSources() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("fromSources() = %v, want %v", got, tt.want)
			}
		})
	}
}

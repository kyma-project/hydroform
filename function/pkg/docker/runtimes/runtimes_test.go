package runtimes

import (
	"fmt"
	"github.com/kyma-incubator/hydroform/function/pkg/resources/types"
	"reflect"
	"testing"
)

func TestContainerEnvs(t *testing.T) {
	type args struct {
		runtime types.Runtime
		debug   bool
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "should return envs for empty runtime",
			args: args{
				runtime: "",
				debug:   false,
			},
			want: []string{
				"FUNC_RUNTIME=",
				"FUNC_HANDLER=main",
				"MOD_NAME=handler",
				"FUNC_PORT=8080",
				Nodejs12Path,
			},
		},
		{
			name: "should return envs for empty runtime with debug",
			args: args{
				runtime: "",
				debug:   true,
			},
			want: []string{
				"FUNC_RUNTIME=",
				"FUNC_HANDLER=main",
				"MOD_NAME=handler",
				"FUNC_PORT=8080",
				Nodejs12Path,
				fmt.Sprintf("NODE_OPTIONS=%s", Nodejs12DebugOption),
			},
		},
		{
			name: "should return envs for nodejs12",
			args: args{
				runtime: types.Nodejs12,
				debug:   false,
			},
			want: []string{
				"FUNC_RUNTIME=nodejs12",
				"FUNC_HANDLER=main",
				"MOD_NAME=handler",
				"FUNC_PORT=8080",
				Nodejs12Path,
			},
		},
		{
			name: "should return envs for nodejs12 with debug",
			args: args{
				runtime: types.Nodejs12,
				debug:   true,
			},
			want: []string{
				"FUNC_RUNTIME=nodejs12",
				"FUNC_HANDLER=main",
				"MOD_NAME=handler",
				"FUNC_PORT=8080",
				Nodejs12Path,
				fmt.Sprintf("NODE_OPTIONS=%s", Nodejs12DebugOption),
			},
		},
		{
			name: "should return envs for nodejs10",
			args: args{
				runtime: types.Nodejs10,
				debug:   false,
			},
			want: []string{
				"FUNC_RUNTIME=nodejs10",
				"FUNC_HANDLER=main",
				"MOD_NAME=handler",
				"FUNC_PORT=8080",
				Nodejs10Path,
			},
		},
		{
			name: "should return envs for nodejs10 with debug",
			args: args{
				runtime: types.Nodejs10,
				debug:   true,
			},
			want: []string{
				"FUNC_RUNTIME=nodejs10",
				"FUNC_HANDLER=main",
				"MOD_NAME=handler",
				"FUNC_PORT=8080",
				Nodejs10Path,
				fmt.Sprintf("NODE_OPTIONS=%s", Nodejs10DebugOption),
			},
		},
		{
			name: "should return envs for python38",
			args: args{
				runtime: types.Python38,
				debug:   false,
			},
			want: []string{
				"FUNC_RUNTIME=python38",
				"FUNC_HANDLER=main",
				"MOD_NAME=handler",
				"FUNC_PORT=8080",
				Python38Path,
			},
		},
		{
			name: "should return envs for python38 with debug",
			args: args{
				runtime: types.Python38,
				debug:   false,
			},
			want: []string{
				"FUNC_RUNTIME=python38",
				"FUNC_HANDLER=main",
				"MOD_NAME=handler",
				"FUNC_PORT=8080",
				Python38Path,
				// TODO
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ContainerEnvs(tt.args.runtime, tt.args.debug); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ContainerEnvs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDockerfile(t *testing.T) {
	tests := []struct {
		name    string
		runtime types.Runtime
		want    string
	}{
		{
			name:    "should return default dockerfile",
			runtime: "",
			want:    Nodejs12Dockerfile,
		},

		{
			name:    "should return Nodejs12 dockerfile",
			runtime: types.Nodejs12,
			want:    Nodejs12Dockerfile,
		},
		{
			name:    "should return Nodejs10 dockerfile",
			runtime: types.Nodejs10,
			want:    Nodejs10Dockerfile,
		},
		{
			name:    "should return Python38 dockerfile",
			runtime: types.Python38,
			want:    Python38Dockerfile,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Dockerfile(tt.runtime); got != tt.want {
				t.Errorf("Dockerfile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRuntimeDebugPort(t *testing.T) {
	tests := []struct {
		name    string
		runtime types.Runtime
		want    string
	}{
		{
			name:    "should return default port",
			runtime: "",
			want:    "9229",
		},
		{
			name:    "should return nodejs12 debug port",
			runtime: types.Nodejs12,
			want:    Nodejs12DebugEndpoint,
		},
		{
			name:    "should return nodejs10 debug port",
			runtime: types.Nodejs10,
			want:    Nodejs10DebugEndpoint,
		},
		{
			name:    "should return python38 debug port",
			runtime: types.Python38,
			want:    Python38DebugEndpoint,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RuntimeDebugPort(tt.runtime); got != tt.want {
				t.Errorf("RuntimeDebugPort() = %v, want %v", got, tt.want)
			}
		})
	}
}

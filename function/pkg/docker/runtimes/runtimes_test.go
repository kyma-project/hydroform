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
				"KUBELESS_INSTALL_VOLUME=/kubeless",
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
				"KUBELESS_INSTALL_VOLUME=/kubeless",
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
				"KUBELESS_INSTALL_VOLUME=/kubeless",
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
				"KUBELESS_INSTALL_VOLUME=/kubeless",
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
				"KUBELESS_INSTALL_VOLUME=/kubeless",
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
				"KUBELESS_INSTALL_VOLUME=/kubeless",
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
				"KUBELESS_INSTALL_VOLUME=/kubeless",
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
				"KUBELESS_INSTALL_VOLUME=/kubeless",
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

func TestContainerCommands(t *testing.T) {
	type args struct {
		runtime types.Runtime
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "should return commands for empty runtime",
			args: args{
				runtime: "",
			},
			want: []string{
				"/kubeless-npm-install.sh", "node kubeless.js",
			},
		},
		{
			name: "should return commands for Nodejs12",
			args: args{
				runtime: types.Nodejs12,
			},
			want: []string{
				"/kubeless-npm-install.sh", "node kubeless.js",
			},
		},
		{
			name: "should return commands for Nodejs10",
			args: args{
				runtime: types.Nodejs10,
			},
			want: []string{
				"/kubeless-npm-install.sh", "node kubeless.js",
			},
		},
		{
			name: "should return commands for Python38",
			args: args{
				runtime: types.Python38,
			},
			want: []string{
				"pip install -r $KUBELESS_INSTALL_VOLUME/requirements.txt", "python kubeless.py",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ContainerCommands(tt.args.runtime); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ContainerCommands() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContainerImage(t *testing.T) {
	type args struct {
		runtime types.Runtime
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "should return image for empty runtime",
			args: args{
				runtime: "",
			},
			want: "eu.gcr.io/kyma-project/function-runtime-nodejs12:cc7dd53f",
		},
		{
			name: "should return image for Nodejs12",
			args: args{
				runtime: types.Nodejs12,
			},
			want: "eu.gcr.io/kyma-project/function-runtime-nodejs12:cc7dd53f",
		},
		{
			name: "should return image for Nodejs10",
			args: args{
				runtime: types.Nodejs10,
			},
			want: "eu.gcr.io/kyma-project/function-runtime-nodejs10:cc7dd53f",
		},
		{
			name: "should return image for Python38",
			args: args{
				runtime: types.Python38,
			},
			want: "eu.gcr.io/kyma-project/function-runtime-python38:cc7dd53f",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ContainerImage(tt.args.runtime); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ContainerImage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContainerUser(t *testing.T) {
	type args struct {
		runtime types.Runtime
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "should return user for empty runtime",
			args: args{
				runtime: "",
			},
			want: "1000",
		},
		{
			name: "should return user for Nodejs12",
			args: args{
				runtime: types.Nodejs12,
			},
			want: "1000",
		},
		{
			name: "should return user for Nodejs10",
			args: args{
				runtime: types.Nodejs10,
			},
			want: "1000",
		},
		{
			name: "should return user for Python38",
			args: args{
				runtime: types.Python38,
			},
			want: "root",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ContainerUser(tt.args.runtime); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ContainerUser() = %v, want %v", got, tt.want)
			}
		})
	}
}
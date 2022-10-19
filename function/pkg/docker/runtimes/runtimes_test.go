package runtimes

import (
	"reflect"
	"testing"

	"github.com/docker/docker/api/types/mount"
	"github.com/kyma-project/hydroform/function/pkg/workspace"

	"github.com/kyma-project/hydroform/function/pkg/resources/types"
)

func TestContainerEnvs(t *testing.T) {
	type args struct {
		runtime   types.Runtime
		hotDeploy bool
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "should return envs for empty runtime",
			args: args{
				runtime:   "",
				hotDeploy: false,
			},
			want: []string{
				"KUBELESS_INSTALL_VOLUME=/kubeless",
				"FUNC_RUNTIME=",
				"FUNC_HANDLER=main",
				"MOD_NAME=handler",
				"FUNC_PORT=8080",
				"SERVICE_NAMESPACE=default",
				NodejsPath,
			},
		},
		{
			name: "should return envs for empty runtime with hotDeploy",
			args: args{
				runtime:   "",
				hotDeploy: true,
			},
			want: []string{
				"KUBELESS_INSTALL_VOLUME=/kubeless",
				"FUNC_RUNTIME=",
				"FUNC_HANDLER=main",
				"MOD_NAME=handler",
				"FUNC_PORT=8080",
				"SERVICE_NAMESPACE=default",
				NodejsPath,
			},
		},
		{
			name: "should return envs for nodejs16",
			args: args{
				runtime:   types.Nodejs16,
				hotDeploy: false,
			},
			want: []string{
				"FUNC_RUNTIME=nodejs16",
				"FUNC_HANDLER=main",
				"MOD_NAME=handler",
				"FUNC_PORT=8080",
				"SERVICE_NAMESPACE=default",
				NodejsPath,
				"HOME=/home/node",
			},
		},
		{
			name: "should return envs for nodejs14",
			args: args{
				runtime:   types.Nodejs14,
				hotDeploy: false,
			},
			want: []string{
				"KUBELESS_INSTALL_VOLUME=/kubeless",
				"FUNC_RUNTIME=nodejs14",
				"FUNC_HANDLER=main",
				"MOD_NAME=handler",
				"FUNC_PORT=8080",
				"SERVICE_NAMESPACE=default",
				NodejsPath,
				"HOME=/home/node",
			},
		},
		{
			name: "should return envs for nodejs14 with hotDeploy",
			args: args{
				runtime:   types.Nodejs14,
				hotDeploy: true,
			},
			want: []string{
				"KUBELESS_INSTALL_VOLUME=/kubeless",
				"FUNC_RUNTIME=nodejs14",
				"FUNC_HANDLER=main",
				"MOD_NAME=handler",
				"FUNC_PORT=8080",
				"SERVICE_NAMESPACE=default",
				NodejsPath,
				"HOME=/home/node",
			},
		},
		{
			name: "should return envs for python39",
			args: args{
				runtime:   types.Python39,
				hotDeploy: false,
			},
			want: []string{
				"KUBELESS_INSTALL_VOLUME=/kubeless",
				"FUNC_RUNTIME=python39",
				"FUNC_HANDLER=main",
				"MOD_NAME=handler",
				"FUNC_PORT=8080",
				"SERVICE_NAMESPACE=default",
				Python39Path,
				"PYTHONUNBUFFERED=TRUE",
			},
		},
		{
			name: "should return envs for python39 with debug",
			args: args{
				runtime:   types.Python39,
				hotDeploy: false,
			},
			want: []string{
				"KUBELESS_INSTALL_VOLUME=/kubeless",
				"FUNC_RUNTIME=python39",
				"FUNC_HANDLER=main",
				"MOD_NAME=handler",
				"FUNC_PORT=8080",
				"SERVICE_NAMESPACE=default",
				Python39Path,
				"PYTHONUNBUFFERED=TRUE",
			},
		},
		{
			name: "should return envs for python39 with hotDeploy",
			args: args{
				runtime:   types.Python39,
				hotDeploy: true,
			},
			want: []string{
				"KUBELESS_INSTALL_VOLUME=/kubeless",
				"FUNC_RUNTIME=python39",
				"FUNC_HANDLER=main",
				"MOD_NAME=handler",
				"FUNC_PORT=8080",
				"SERVICE_NAMESPACE=default",
				Python39Path,
				"PYTHONUNBUFFERED=TRUE",
				"CHERRYPY_RELOADED=true",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ContainerEnvs(tt.args.runtime, tt.args.hotDeploy); !reflect.DeepEqual(got, tt.want) {
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
			name:    "should return nodejs14 debug port",
			runtime: types.Nodejs14,
			want:    NodejsDebugEndpoint,
		},
		{
			name:    "should return nodejs16 debug port",
			runtime: types.Nodejs16,
			want:    NodejsDebugEndpoint,
		},
		{
			name:    "should return python39 debug port",
			runtime: types.Python39,
			want:    Python39DebugEndpoint,
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
		runtime   types.Runtime
		debug     bool
		hotDeploy bool
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
				"npm install --production --prefix=$KUBELESS_INSTALL_VOLUME", "node kubeless.js",
			},
		},
		{
			name: "should return commands for empty runtime with hotDeploy",
			args: args{
				runtime:   "",
				hotDeploy: true,
			},
			want: []string{
				"npm install --production --prefix=$KUBELESS_INSTALL_VOLUME", "npx nodemon --watch /kubeless/*.js /kubeless_rt/kubeless.js",
			},
		},
		{
			name: "should return commands for empty runtime with hotDeploy",
			args: args{
				runtime:   "",
				hotDeploy: true,
			},
			want: []string{
				"npm install --production --prefix=$KUBELESS_INSTALL_VOLUME", "npx nodemon --watch /kubeless/*.js /kubeless_rt/kubeless.js",
			},
		},
		{
			name: "should return commands for Nodejs14",
			args: args{
				runtime: types.Nodejs14,
			},
			want: []string{
				"npm install --production --prefix=$KUBELESS_INSTALL_VOLUME", "node kubeless.js",
			},
		},
		{
			name: "should return commands for Nodejs14 with hotDeploy",
			args: args{
				runtime:   types.Nodejs14,
				hotDeploy: true,
			},
			want: []string{
				"npm install --production --prefix=$KUBELESS_INSTALL_VOLUME", "npx nodemon --watch /kubeless/*.js /kubeless_rt/kubeless.js",
			},
		},
		{
			name: "should return commands for Nodejs16",
			args: args{
				runtime:   types.Nodejs16,
				hotDeploy: false,
			},
			want: []string{
				"npm install --production", "node server.js",
			},
		},
		{
			name: "should return commands for Nodejs16 with hotDeploy",
			args: args{
				runtime:   types.Nodejs16,
				hotDeploy: true,
			},
			want: []string{
				"npm install --production", "npx nodemon --watch /usr/src/app/function/*.js /usr/src/app/server.js",
			},
		},
		{
			name: "should return commands for Python39",
			args: args{
				runtime: types.Python39,
			},
			want: []string{
				"pip install -r $KUBELESS_INSTALL_VOLUME/requirements.txt", "python kubeless.py",
			},
		},
		{
			name: "should return commands for Python39 with hotDeploy",
			args: args{
				runtime:   types.Python39,
				hotDeploy: true,
			},
			want: []string{
				"pip install -r $KUBELESS_INSTALL_VOLUME/requirements.txt", "python kubeless.py",
			},
		},
		{
			name: "should return commands for Python39 with hotDeploy",
			args: args{
				runtime:   types.Python39,
				hotDeploy: true,
			},
			want: []string{
				"pip install -r $KUBELESS_INSTALL_VOLUME/requirements.txt", "python kubeless.py",
			},
		},
		{
			name: "should return commands for Python39 with debug",
			args: args{
				runtime: types.Python39,
				debug:   true,
			},
			want: []string{
				"pip install -r $KUBELESS_INSTALL_VOLUME/requirements.txt", "pip install debugpy", "python -m debugpy --listen 0.0.0.0:5678 kubeless.py",
			},
		},
		{
			name: "should return commands for Python39 with hotDeploy and debug",
			args: args{
				runtime:   types.Python39,
				hotDeploy: true,
				debug:     true,
			},
			want: []string{
				"pip install -r $KUBELESS_INSTALL_VOLUME/requirements.txt", "pip install debugpy", "python -m debugpy --listen 0.0.0.0:5678 kubeless.py",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ContainerCommands(tt.args.runtime, tt.args.debug, tt.args.hotDeploy); !reflect.DeepEqual(got, tt.want) {
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
			want: "eu.gcr.io/kyma-project/function-runtime-nodejs14:e1491c46",
		},
		{
			name: "should return image for Nodejs14",
			args: args{
				runtime: types.Nodejs14,
			},
			want: "eu.gcr.io/kyma-project/function-runtime-nodejs14:e1491c46",
		},
		{
			name: "should return image for Nodejs16",
			args: args{
				runtime: types.Nodejs16,
			},
			want: "eu.gcr.io/kyma-project/function-runtime-nodejs16:e1491c46",
		},
		{
			name: "should return image for Python39",
			args: args{
				runtime: types.Python39,
			},
			want: "eu.gcr.io/kyma-project/function-runtime-python39:e1491c46",
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

func TestGetMounts(t *testing.T) {
	type args struct {
		runtime    types.Runtime
		sourceType workspace.SourceType
		workDir    string
	}
	tests := []struct {
		name string
		args args
		want []mount.Mount
	}{
		{
			name: "should return mount for source type inline",
			args: args{
				runtime:    types.Nodejs14,
				sourceType: workspace.SourceTypeInline,
				workDir:    "/your/work/dir",
			},
			want: []mount.Mount{
				{
					Type:   mount.TypeBind,
					Source: "/your/work/dir",
					Target: KubelessTmpPath,
				},
				{
					Type:   mount.TypeVolume,
					Target: KubelessPath,
				},
			},
		},
		{
			name: "should return mount for source type git",
			args: args{
				runtime:    types.Nodejs14,
				sourceType: workspace.SourceTypeGit,
			},
			want: []mount.Mount{
				{
					Type:   mount.TypeBind,
					Source: "",
					Target: KubelessPath,
				},
			},
		},
		{
			name: "should return mount for nodejs16",
			args: args{
				runtime:    types.Nodejs16,
				sourceType: workspace.SourceTypeInline,
				workDir:    "/your/work/dir",
			},
			want: []mount.Mount{
				{
					Type:   mount.TypeBind,
					Source: "/your/work/dir",
					Target: KubelessTmpPath,
				},
				{
					Type:   mount.TypeVolume,
					Target: FunctionMountPath,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetMounts(tt.args.runtime, tt.args.sourceType, tt.args.workDir); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetMounts() = %v, want %v", got, tt.want)
			}
		})
	}
}

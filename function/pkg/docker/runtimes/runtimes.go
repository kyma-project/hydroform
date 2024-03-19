package runtimes

import (
	"fmt"
	"path/filepath"

	"github.com/docker/docker/api/types/mount"
	"github.com/kyma-project/hydroform/function/pkg/workspace"

	"github.com/kyma-project/hydroform/function/pkg/resources/types"
)

const (
	ServerPort        = "8080"
	FunctionMountPath = "/usr/src/app/function"
	KubelessPath      = "/kubeless"
	KubelessTmpPath   = "/tmp/kubeless"
	ContainerUser     = "root"

	NodejsPath          = "NODE_PATH=$(KUBELESS_INSTALL_VOLUME)/node_modules"
	NodejsDebugEndpoint = `9229`

	Python39Path        = "PYTHONPATH=$(KUBELESS_INSTALL_VOLUME)/lib.python3.9/site-packages:$(KUBELESS_INSTALL_VOLUME)"
	Python312Path       = "PYTHONPATH=$(KUBELESS_INSTALL_VOLUME)/lib.python3.12/site-packages:$(KUBELESS_INSTALL_VOLUME)"
	PythonHotDeploy     = "CHERRYPY_RELOADED=true"
	PythonUnbuffered    = "PYTHONUNBUFFERED=TRUE"
	PythonDebugEndpoint = `5678`
)

func ContainerEnvs(runtime types.Runtime, hotDeploy bool) []string {
	envs := []string{}
	if runtime != types.Nodejs18 && runtime != types.Nodejs20 {
		envs = append(envs, fmt.Sprintf("KUBELESS_INSTALL_VOLUME=%s", KubelessPath))
	}
	envs = append(envs, []string{
		fmt.Sprintf("FUNC_RUNTIME=%s", runtime),
		"FUNC_HANDLER=main",
		"MOD_NAME=handler",
		fmt.Sprintf("FUNC_PORT=%s", ServerPort),
		"SERVICE_NAMESPACE=default",
	}...)
	return append(envs, runtimeEnvs(runtime, hotDeploy)...)
}

func runtimeEnvs(runtime types.Runtime, hotDeploy bool) []string {
	switch runtime {
	case types.Nodejs18, types.Nodejs20:
		return []string{NodejsPath, "HOME=/home/node"}
	case types.Python39:
		envs := []string{Python39Path, PythonUnbuffered}
		if hotDeploy {
			envs = append(envs, PythonHotDeploy)
		}
		return envs
	case types.Python312:
		envs := []string{Python312Path, PythonUnbuffered}
		if hotDeploy {
			envs = append(envs, PythonHotDeploy)
		}
		return envs
	default:
		return []string{NodejsPath}
	}
}

func RuntimeDebugPort(runtime types.Runtime) string {
	switch runtime {
	case types.Nodejs18, types.Nodejs20:
		return NodejsDebugEndpoint
	case types.Python39, types.Python312:
		return PythonDebugEndpoint
	default:
		return NodejsDebugEndpoint
	}
}

func ContainerCommands(runtime types.Runtime, debug bool, hotDeploy bool) []string {
	switch runtime {
	case types.Nodejs18, types.Nodejs20:
		runCommand := ""
		if hotDeploy && debug {
			runCommand = "npx nodemon --watch /usr/src/app/function/*.js --inspect=0.0.0.0 --exitcrash server.js"
		} else if hotDeploy {
			runCommand = "npx nodemon --watch /usr/src/app/function/*.js /usr/src/app/server.js"
		} else if debug {
			runCommand = "node --inspect=0.0.0.0 server.js"
		} else {
			//npm start ?
			runCommand = "node server.js"
		}
		return []string{"npm install --production", runCommand}
	case types.Python39, types.Python312:
		if debug {
			return []string{"pip install -r $KUBELESS_INSTALL_VOLUME/requirements.txt", "pip install debugpy", "python -m debugpy --listen 0.0.0.0:5678 kubeless.py"}
		}
		return []string{"pip install -r $KUBELESS_INSTALL_VOLUME/requirements.txt", "python kubeless.py"}

	default:
		if hotDeploy {
			return []string{"npm install --production --prefix=$KUBELESS_INSTALL_VOLUME", "npx nodemon --watch /kubeless/*.js /kubeless_rt/kubeless.js"}
		}
		return []string{"npm install --production --prefix=$KUBELESS_INSTALL_VOLUME", "node kubeless.js"}
	}
}

func GetMounts(runtime types.Runtime, sourceType workspace.SourceType, workDir string) []mount.Mount {
	sourceMountPoint := KubelessPath
	if !isKubelessRuntime(runtime) {
		sourceMountPoint = FunctionMountPath
	}
	if sourceType == workspace.SourceTypeInline {
		return []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: workDir,
				Target: KubelessTmpPath,
			},
			{
				Type:   mount.TypeVolume,
				Target: sourceMountPoint,
			},
		}
	}
	return []mount.Mount{
		{
			Type:   mount.TypeBind,
			Source: workDir,
			Target: sourceMountPoint,
		},
	}
}

func MoveInlineCommand(runtime types.Runtime, sourcePath, depsPath string) []string {
	sourceMountPoint := KubelessPath
	if !isKubelessRuntime(runtime) {
		sourceMountPoint = FunctionMountPath
	}
	sourcePathFull := filepath.Join(KubelessTmpPath, sourcePath)
	sourceDestFull := filepath.Join(sourceMountPoint, filepath.Base(sourcePath))

	depsPathFull := filepath.Join(KubelessTmpPath, depsPath)
	depsDestFull := filepath.Join(sourceMountPoint, filepath.Base(depsPath))

	linkedPaths := []string{
		fmt.Sprintf("ln -s -f %s %s", sourcePathFull, sourceDestFull),
		fmt.Sprintf("ln -s -f %s %s", depsPathFull, depsDestFull),
	}
	return linkedPaths
}

func ContainerImage(runtime types.Runtime) string {
	switch runtime {
	case types.Nodejs18:
		return "eu.gcr.io/kyma-project/function-runtime-nodejs18:v20230228-b2981e80"
	case types.Nodejs20:
		return "europe-docker.pkg.dev/kyma-project/prod/function-runtime-nodejs20:v20240314-8476bf34"
	case types.Python39:
		return "eu.gcr.io/kyma-project/function-runtime-python39:v20230223-ec41ec1e"
	case types.Python312:
		return "europe-docker.pkg.dev/kyma-project/prod/function-runtime-python312:v20240307-8e7d9941"
	default:
		return "eu.gcr.io/kyma-project/function-runtime-nodejs18:v20230228-b2981e80"
	}
}

func isKubelessRuntime(runtime types.Runtime) bool {
	if runtime == types.Nodejs18 || runtime == types.Nodejs20 {
		return false
	}
	return true
}

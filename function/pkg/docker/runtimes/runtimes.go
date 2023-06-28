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

	Python39Path          = "PYTHONPATH=$(KUBELESS_INSTALL_VOLUME)/lib.python3.9/site-packages:$(KUBELESS_INSTALL_VOLUME)"
	Python39HotDeploy     = "CHERRYPY_RELOADED=true"
	Python39Unbuffered    = "PYTHONUNBUFFERED=TRUE"
	Python39DebugEndpoint = `5678`
)

func ContainerEnvs(runtime types.Runtime, hotDeploy bool) []string {
	envs := []string{}
	if runtime != types.Nodejs16 && runtime != types.Nodejs18 {
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
	case types.Nodejs16, types.Nodejs18:
		return []string{NodejsPath, "HOME=/home/node"}
	case types.Python39:
		envs := []string{Python39Path, Python39Unbuffered}
		if hotDeploy {
			envs = append(envs, Python39HotDeploy)
		}
		return envs
	default:
		return []string{NodejsPath}
	}
}

func RuntimeDebugPort(runtime types.Runtime) string {
	switch runtime {
	case types.Nodejs16, types.Nodejs18:
		return NodejsDebugEndpoint
	case types.Python39:
		return Python39DebugEndpoint
	default:
		return NodejsDebugEndpoint
	}
}

func ContainerCommands(runtime types.Runtime, debug bool, hotDeploy bool) []string {
	switch runtime {
	case types.Nodejs16, types.Nodejs18:
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
	case types.Python39:
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
	case types.Nodejs16:
		return "eu.gcr.io/kyma-project/function-runtime-nodejs16:v20230228-b2981e80"
	case types.Nodejs18:
		return "eu.gcr.io/kyma-project/function-runtime-nodejs18:v20230228-b2981e80"
	case types.Python39:
		return "eu.gcr.io/kyma-project/function-runtime-python39:v20230223-ec41ec1e"
	default:
		return "eu.gcr.io/kyma-project/function-runtime-nodejs18:v20230228-b2981e80"
	}
}

func isKubelessRuntime(runtime types.Runtime) bool {
	if runtime == types.Nodejs16 || runtime == types.Nodejs18 {
		return false
	}
	return true
}

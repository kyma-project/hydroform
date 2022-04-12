package runtimes

import (
	"fmt"
	"github.com/docker/docker/api/types/mount"
	"github.com/kyma-incubator/hydroform/function/pkg/workspace"
	"path/filepath"

	"github.com/kyma-incubator/hydroform/function/pkg/resources/types"
)

const (
	ServerPort      = "8080"
	KubelessPath    = "/kubeless"
	KubelessTmpPath = "/tmp/kubeless"
	ContainerUser   = "root"

	NodejsPath          = "NODE_PATH=$(KUBELESS_INSTALL_VOLUME)/node_modules"
	NodejsDebugEndpoint = `9229`

	Python39Path          = "PYTHONPATH=$(KUBELESS_INSTALL_VOLUME)/lib.python3.9/site-packages:$(KUBELESS_INSTALL_VOLUME)"
	Python39HotDeploy     = "CHERRYPY_RELOADED=true"
	Python39Unbuffered    = "PYTHONUNBUFFERED=TRUE"
	Python39DebugEndpoint = `5678`
)

func ContainerEnvs(runtime types.Runtime, hotDeploy bool) []string {
	return append([]string{
		fmt.Sprintf("KUBELESS_INSTALL_VOLUME=%s", KubelessPath),
		fmt.Sprintf("FUNC_RUNTIME=%s", runtime),
		"FUNC_HANDLER=main",
		"MOD_NAME=handler",
		fmt.Sprintf("FUNC_PORT=%s", ServerPort),
	}, runtimeEnvs(runtime, hotDeploy)...)
}

func runtimeEnvs(runtime types.Runtime, hotDeploy bool) []string {
	switch runtime {
	case types.Nodejs12, types.Nodejs14:
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
	case types.Nodejs12, types.Nodejs14:
		return NodejsDebugEndpoint
	case types.Python39:
		return Python39DebugEndpoint
	default:
		return NodejsDebugEndpoint
	}
}

func ContainerCommands(runtime types.Runtime, debug bool, hotDeploy bool) []string {
	switch runtime {
	case types.Nodejs12, types.Nodejs14:
		runCommand := ""
		if hotDeploy && debug {
			runCommand = "npx nodemon --watch /kubeless/*.js --inspect=0.0.0.0 --exitcrash kubeless.js "
		} else if hotDeploy {
			runCommand = "npx nodemon --watch /kubeless/*.js /kubeless_rt/kubeless.js"
		} else if debug {
			runCommand = "node --inspect=0.0.0.0 kubeless.js "
		} else {
			runCommand = "node kubeless.js"
		}
		return []string{"npm install --production --prefix=$KUBELESS_INSTALL_VOLUME", runCommand}
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

func GetMounts(sourceType workspace.SourceType, workDir string) []mount.Mount {
	if sourceType == workspace.SourceTypeInline {
		return []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: workDir,
				Target: KubelessTmpPath,
			},
			{
				Type:   mount.TypeVolume,
				Target: KubelessPath,
			},
		}
	} else {
		return []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: workDir,
				Target: KubelessPath,
			},
		}
	}
}

func MoveInlineCommand(sourcePath, depsPath string) []string {
	return []string{
		fmt.Sprintf("ln -s -f %s %s", filepath.Join(KubelessTmpPath, sourcePath), filepath.Join(KubelessPath, filepath.Base(sourcePath))),
		fmt.Sprintf("ln -s -f %s %s", filepath.Join(KubelessTmpPath, depsPath), filepath.Join(KubelessPath, filepath.Base(depsPath))),
	}
}

func ContainerImage(runtime types.Runtime) string {
	switch runtime {
	case types.Nodejs12:
		return "eu.gcr.io/kyma-project/function-runtime-nodejs12:245170b1"
	case types.Nodejs14:
		return "eu.gcr.io/kyma-project/function-runtime-nodejs14:245170b1"
	case types.Python39:
		return "eu.gcr.io/kyma-project/function-runtime-python39:245170b1"
	default:
		return "eu.gcr.io/kyma-project/function-runtime-nodejs14:245170b1"
	}
}

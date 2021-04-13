package runtimes

import (
	"fmt"

	"github.com/kyma-incubator/hydroform/function/pkg/resources/types"
)

const (
	ServerPort   = "8080"
	KubelessPath = "/kubeless"

	NodejsPath          = "NODE_PATH=$(KUBELESS_INSTALL_VOLUME)/node_modules"
	NodejsDebugEndpoint = `9229`

	Python38Path          = "PYTHONPATH=$(KUBELESS_INSTALL_VOLUME)/lib.python3.8/site-packages:$(KUBELESS_INSTALL_VOLUME)"
	Python38HotDeploy     = "CHERRYPY_RELOADED=true"
	Python38DebugEndpoint = `5678`
)

func ContainerEnvs(runtime types.Runtime, hotDeploy bool) []string {
	return append([]string{
		fmt.Sprintf("KUBELESS_INSTALL_VOLUME=%s", KubelessPath),
		fmt.Sprintf("FUNC_RUNTIME=%s", runtime),
		"FUNC_HANDLER=main",
		"MOD_NAME=handler",
		"FUNC_PORT=8080",
	}, runtimeEnvs(runtime, hotDeploy)...)
}

func runtimeEnvs(runtime types.Runtime, hotDeploy bool) []string {
	switch runtime {
	case types.Nodejs12, types.Nodejs14:
		return []string{NodejsPath}
	case types.Python38:
		envs := []string{Python38Path}
		if hotDeploy {
			envs = append(envs, Python38HotDeploy)
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
	case types.Python38:
		return Python38DebugEndpoint
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
		return []string{"/kubeless-npm-install.sh", runCommand}
	case types.Python38:
		if debug {
			return []string{"pip install -r $KUBELESS_INSTALL_VOLUME/requirements.txt", "pip install debugpy", "python -m debugpy --listen 0.0.0.0:5678 kubeless.py"}
		}
		return []string{"pip install -r $KUBELESS_INSTALL_VOLUME/requirements.txt", "python kubeless.py"}

	default:
		if hotDeploy {
			return []string{"/kubeless-npm-install.sh", "npx nodemon --watch /kubeless/*.js /kubeless_rt/kubeless.js"}
		}
		return []string{"/kubeless-npm-install.sh", "node kubeless.js"}
	}
}

func ContainerImage(runtime types.Runtime) string {
	switch runtime {
	case types.Nodejs12:
		return "eu.gcr.io/kyma-project/function-runtime-nodejs12:4bed80da"
	case types.Nodejs14:
		return "eu.gcr.io/kyma-project/function-runtime-Nodejs14:4bed80da"
	case types.Python38:
		return "eu.gcr.io/kyma-project/function-runtime-python38:4bed80da"
	default:
		return "eu.gcr.io/kyma-project/function-runtime-nodejs12:4bed80da"
	}
}

func ContainerUser(runtime types.Runtime) string {
	switch runtime {
	case types.Nodejs12:
		return "1000"
	case types.Nodejs14:
		return "1000"
	case types.Python38:
		return "root"
	default:
		return "1000"
	}
}

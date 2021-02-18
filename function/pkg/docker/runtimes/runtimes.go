package runtimes

import (
	"fmt"

	"github.com/kyma-incubator/hydroform/function/pkg/resources/types"
)

const (
	ServerPort = "8080"

	Nodejs10Path          = "NODE_PATH=$(KUBELESS_INSTALL_VOLUME)/node_modules"
	Nodejs10DebugOption   = "--inspect=0.0.0.0"
	Nodejs10DebugEndpoint = `9229`

	Nodejs12Path          = "NODE_PATH=$(KUBELESS_INSTALL_VOLUME)/node_modules"
	Nodejs12DebugOption   = "--inspect=0.0.0.0"
	Nodejs12DebugEndpoint = `9229`

	Python38Path          = "PYTHONPATH=$(KUBELESS_INSTALL_VOLUME)/lib.python3.8/site-packages:$(KUBELESS_INSTALL_VOLUME)"
	Python38HotDeploy     = "CHERRYPY_RELOADED=true"
	Python38DebugEndpoint = `5678`
)

func ContainerEnvs(runtime types.Runtime, debug bool, hotDeploy bool) []string {
	return append([]string{
		fmt.Sprintf("FUNC_RUNTIME=%s", runtime),
		"FUNC_HANDLER=main",
		"MOD_NAME=handler",
		"FUNC_PORT=8080",
		"KUBELESS_INSTALL_VOLUME=/kubeless",
	}, runtimeEnvs(runtime, debug, hotDeploy)...)
}

func runtimeEnvs(runtime types.Runtime, debug bool, hotDeploy bool) []string {
	switch runtime {
	case types.Nodejs12:
		envs := []string{Nodejs12Path}
		if debug {
			envs = append(envs, fmt.Sprintf("NODE_OPTIONS=%s", Nodejs12DebugOption))
		}
		return envs
	case types.Nodejs10:
		envs := []string{Nodejs10Path}
		if debug {
			envs = append(envs, fmt.Sprintf("NODE_OPTIONS=%s", Nodejs10DebugOption))
		}
		return envs
	case types.Python38:
		envs := []string{Python38Path}
		if hotDeploy {
			envs = append(envs, Python38HotDeploy)
		}
		// TODO
		//if debug { }
		return envs
	default:
		envs := []string{Nodejs12Path}
		if debug {
			envs = append(envs, fmt.Sprintf("NODE_OPTIONS=%s", Nodejs12DebugOption))
		}
		return envs
	}
}

func RuntimeDebugPort(runtime types.Runtime) string {
	switch runtime {
	case types.Nodejs12:
		return Nodejs12DebugEndpoint
	case types.Nodejs10:
		return Nodejs10DebugEndpoint
	case types.Python38:
		return Python38DebugEndpoint
	default:
		return Nodejs12DebugEndpoint
	}
}

func ContainerCommands(runtime types.Runtime, hotDeploy bool) []string {
	switch runtime {
	case types.Nodejs12, types.Nodejs10:
		if hotDeploy {
			return []string{"/kubeless-npm-install.sh", "npx nodemon --watch /kubeless/*.js --inspect=0.0.0.0 /kubeless_rt/kubeless.js "}
		} else {
			return []string{"/kubeless-npm-install.sh", "node kubeless.js"}
		}
	case types.Python38:
		return []string{"pip install -r $KUBELESS_INSTALL_VOLUME/requirements.txt", "python kubeless.py"}
	default:
		if hotDeploy {
			return []string{"/kubeless-npm-install.sh", "npx nodemon --watch /kubeless/*.js /kubeless_rt/kubeless.js"}
		} else {
			return []string{"/kubeless-npm-install.sh", "node kubeless.js"}
		}
	}
}

func ContainerImage(runtime types.Runtime) string {
	switch runtime {
	case types.Nodejs12:
		return "eu.gcr.io/kyma-project/function-runtime-nodejs12:e7698eb5"
	case types.Nodejs10:
		return "eu.gcr.io/kyma-project/function-runtime-nodejs10:e7698eb5"
	case types.Python38:
		return "eu.gcr.io/kyma-project/function-runtime-python38:e7698eb5"
	default:
		return "eu.gcr.io/kyma-project/function-runtime-nodejs12:e7698eb5"
	}
}

func ContainerUser(runtime types.Runtime) string {
	switch runtime {
	case types.Nodejs12:
		return "1000"
	case types.Nodejs10:
		return "1000"
	case types.Python38:
		return "root"
	default:
		return "1000"
	}
}

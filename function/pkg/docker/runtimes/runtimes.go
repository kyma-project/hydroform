package runtimes

import (
	"fmt"

	"github.com/kyma-incubator/hydroform/function/pkg/resources/types"
)

func Dockerfile(runtime types.Runtime) string {
	switch runtime {
	case types.Nodejs12:
		return Node12Dockerfile
	case types.Nodejs10:
		return Node10Dockerfile
	case types.Python38:
		return Python38Dockerfile
	default:
		return Node12Dockerfile
	}
}

func ContainerEnvs(runtime types.Runtime, debug bool) []string {
	envs := []string{
		fmt.Sprintf("FUNC_RUNTIME=%s", runtime),
		"FUNC_HANDLER=main",
		"MOD_NAME=handler",
		"FUNC_PORT=8080",
	}

	return append(envs, runtimeEnvs(runtime, debug)...)
}

func runtimeEnvs(runtime types.Runtime, debug bool) []string {
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
		if debug {
			// TODO
		}
		return envs
	default:
		envs := []string{Nodejs12Path}
		if debug {
			envs = append(envs, fmt.Sprintf("NODE_OPTIONS=%s", Nodejs12DebugOption))
		}
		return envs
	}
}

func ContainerPorts(runtime types.Runtime, exposedPort string, debug bool) map[string]string {
	ports := map[string]string{
		"8080": exposedPort,
	}
	if debug {
		port := runtimeDebugPort(runtime)
		ports[port] = port
	}
	return ports
}

func runtimeDebugPort(runtime types.Runtime) string {
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

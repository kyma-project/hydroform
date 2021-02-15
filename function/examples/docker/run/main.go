package main

import (
	"context"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/kyma-incubator/hydroform/function/pkg/docker"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

func main() {
	runOpts := docker.RunOpts{
		Ports:         map[string]string{"8080": "8080"},
		Envs:          []string{"FUNC_HANDLER=main", "MOD_NAME=handler", "FUNC_PORT=8080", "FUNC_RUNTIME=nodejs12", "NODE_PATH='$(KUBELESS_INSTALL_VOLUME)/node_modules'", "KUBELESS_INSTALL_VOLUME=/kubeless"},
		ContainerName: "test123",
		Image:         "eu.gcr.io/kyma-project/function-runtime-nodejs12:cc7dd53f",
		WorkDir:       "/tmp/tmpfunc/",
		Commands:      []string{"/kubeless-npm-install.sh", "node kubeless.js"},
	}
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		fmt.Println(err)
	}
	ctx := context.Background()
	docker.RunContainer(ctx, cli, runOpts)
}

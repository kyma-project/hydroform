package main

import (
	"context"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/kyma-incubator/hydroform/function/pkg/docker"
	"github.com/kyma-incubator/hydroform/function/pkg/docker/runtimes"
	"github.com/kyma-incubator/hydroform/function/pkg/resources/types"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

func main() {
	runOpts := docker.RunOpts{
		Ports:         map[string]string{"8080": "8080"},
		Envs:          runtimes.ContainerEnvs(types.Nodejs12, false),
		ContainerName: "test123",
		Image:         runtimes.ContainerImage(types.Nodejs12),
		WorkDir:       "/tmp/tmpfunc/",
		Commands:      runtimes.ContainerCommands(types.Nodejs12),
		User:          runtimes.ContainerUser(types.Nodejs12),
	}
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		fmt.Println(err)
	}
	ctx := context.Background()
	docker.RunContainer(ctx, cli, runOpts)
}

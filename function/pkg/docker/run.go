package docker

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"io"
)

type RunOpts struct {
	Ports         map[string]string
	Envs          []string
	ContainerName string
	ImageName     string
}

func RunContainer(c *client.Client, ctx context.Context, opts RunOpts) (string, error) {
	body, err := c.ContainerCreate(ctx, &container.Config{
		Env:          opts.Envs,
		ExposedPorts: portSet(opts.Ports),
		Image:        opts.ImageName,
	}, &container.HostConfig{
		PortBindings: portMap(opts.Ports),
		AutoRemove:   true,
	}, nil, nil,
		opts.ContainerName)
	if err != nil {
		return "", err
	}

	err = c.ContainerStart(ctx, body.ID, types.ContainerStartOptions{})
	if err != nil {
		return "", err
	}

	return body.ID, nil
}

func FollowRun(c *client.Client, ctx context.Context, ID string, log func(...interface{})) error {
	buf, err := c.ContainerAttach(ctx, ID, types.ContainerAttachOptions{
		Stdout: true,
		Stderr: true,
		Stream: true,
	})
	if err != nil {
		return err
	}
	defer buf.Close()

	for {
		line, e := buf.Reader.ReadBytes('\n')
		if e == io.EOF {
			break
		}
		if e != nil {
			err = e
		}

		log(string(line))
	}

	return err
}

func Stop(c *client.Client, ctx context.Context, ID string, log func(...interface{})) func() {
	return func() {
		log("\r- Ctrl+C pressed in Terminal\n", fmt.Sprintf("Removing container %s...\n", ID))
		c.ContainerStop(ctx, ID, nil)
	}
}

func portSet(ports map[string]string) nat.PortSet {
	portSet := nat.PortSet{}
	for from, _ := range ports {
		portSet[nat.Port(from)] = struct{}{}
	}
	return portSet
}

func portMap(ports map[string]string) nat.PortMap {
	portMap := nat.PortMap{}
	for from, to := range ports {
		portMap[nat.Port(from)] = []nat.PortBinding{
			{
				HostPort: to,
			},
		}

	}
	return portMap
}

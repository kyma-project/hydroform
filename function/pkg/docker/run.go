package docker

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/docker/docker/api/types/mount"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	apiclient "github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
)

//go:generate mockgen -source=run.go -destination=automock/run.go

type DockerClient interface {
	ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig,
		networkingConfig *network.NetworkingConfig, platform *specs.Platform, containerName string) (container.ContainerCreateCreatedBody, error)
	ContainerStart(ctx context.Context, containerID string, options types.ContainerStartOptions) error
	ContainerAttach(ctx context.Context, container string, options types.ContainerAttachOptions) (types.HijackedResponse, error)
	ContainerStop(ctx context.Context, containerID string, timeout *time.Duration) error
	ImagePull(ctx context.Context, refStr string, options types.ImagePullOptions) (io.ReadCloser, error)
}

type RunOpts struct {
	Ports         map[string]string
	Envs          []string
	ContainerName string
	Image         string
	WorkDir       string
	Commands      []string
	User          string
}

func RunContainer(ctx context.Context, c DockerClient, opts RunOpts) (string, error) {
	body, err := pullAndRun(ctx, c, &container.Config{
		Env:          opts.Envs,
		ExposedPorts: portSet(opts.Ports),
		Image:        opts.Image,
		Cmd:          []string{"/bin/sh", "-c", strings.Join(opts.Commands[:], ";")},
		User:         opts.User,
	}, &container.HostConfig{
		PortBindings: portMap(opts.Ports),
		AutoRemove:   true,
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: opts.WorkDir,
				Target: "/kubeless",
			},
		},
	}, opts.ContainerName)
	if err != nil {
		return "", err
	}

	err = c.ContainerStart(ctx, body.ID, types.ContainerStartOptions{})
	if err != nil {
		return "", err
	}

	return body.ID, nil
}

func pullAndRun(ctx context.Context, c DockerClient, config *container.Config, hostConfig *container.HostConfig, containerName string) (container.ContainerCreateCreatedBody, error) {
	body, err := c.ContainerCreate(ctx, config, hostConfig, nil, nil, containerName)
	if apiclient.IsErrNotFound(err) {
		_, err = c.ImagePull(ctx, config.Image, types.ImagePullOptions{})
		if err != nil {
			return body, err
		}
		body, err = c.ContainerCreate(ctx, config, hostConfig, nil, nil, containerName)
	}
	return body, err
}

func FollowRun(ctx context.Context, c DockerClient, ID string, log func(...interface{})) error {
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
			break
		}

		log(string(line))
	}

	return err
}

func Stop(ctx context.Context, c DockerClient, ID string, log func(...interface{})) func() {
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

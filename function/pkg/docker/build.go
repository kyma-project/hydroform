package docker

import (
	"bufio"
	"context"
	"github.com/pkg/errors"
	"io"

	"gopkg.in/square/go-jose.v2/json"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
)

type BuildOpts struct {
	Context string
	Tags    []string
}

func BuildImage(c *client.Client, ctx context.Context, opts BuildOpts) (*types.ImageBuildResponse, error) {
	tar := &archive.TarOptions{}
	reader, err := archive.TarWithOptions(opts.Context, tar)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	response, err := c.ImageBuild(ctx, reader, types.ImageBuildOptions{
		Tags:           opts.Tags,
	})
	if err != nil {
		return nil, err
	}
	return &response, nil
}

type ErrorDetail struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type ResultEntry struct {
	Stream      string      `json:"stream,omitempty"`
	ErrorDetail ErrorDetail `json:"errorDetail,omitempty"`
	Error       string      `json:"error,omitempty"`
}

func FollowBuild(readCloser io.ReadCloser, log func(...interface{})) error {
	buf := bufio.NewReader(readCloser)
	for {
		line, err := buf.ReadBytes('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		var entryResult ResultEntry
		if err = json.Unmarshal(line, &entryResult); err != nil {
			return err
		}
		if entryResult.Error != "" {
			err := errors.Errorf("image build failed", entryResult.Error)
			return err
		}
		log(entryResult.Stream)
	}
	return nil
}


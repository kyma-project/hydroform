package docker

import (
	"bufio"
	"context"
	"github.com/pkg/errors"
	"io"

	"gopkg.in/square/go-jose.v2/json"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/archive"
)

//go:generate mockgen -source=build.go -destination=automock/build.go

type ImageClient interface {
	ImageBuild(ctx context.Context, buildContext io.Reader, options types.ImageBuildOptions) (types.ImageBuildResponse, error)
}

type BuildOpts struct {
	Context string
	Tags    []string
}

func BuildImage(ctx context.Context, c ImageClient, opts BuildOpts) (*types.ImageBuildResponse, error) {
	reader, err := archive.TarWithOptions(opts.Context, &archive.TarOptions{})
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	response, err := c.ImageBuild(ctx, reader, types.ImageBuildOptions{
		Tags: opts.Tags,
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

func FollowBuild(readCloser io.Reader, log func(...interface{})) error {
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
			err := errors.Errorf("image build failed: %s", entryResult.ErrorDetail.Message)
			return err
		}
		log(entryResult.Stream)
	}
	return nil
}

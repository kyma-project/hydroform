package docker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/golang/mock/gomock"
	mock_docker "github.com/kyma-incubator/hydroform/function/pkg/docker/automock"
	"github.com/stretchr/testify/require"
)

func TestBuildImage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	path, err := ioutil.TempDir(os.TempDir(), "test-")
	require.Equal(t, nil, err)
	defer os.RemoveAll(path)

	type args struct {
		ctx  context.Context
		c    ImageClient
		opts BuildOpts
	}
	tests := []struct {
		name    string
		args    args
		want    *types.ImageBuildResponse
		wantErr bool
	}{
		{
			name: "should build image",
			args: args{
				ctx: ctx,
				c: func() ImageClient {
					mock := mock_docker.NewMockImageClient(ctrl)

					mock.EXPECT().ImageBuild(ctx, gomock.Not(gomock.Nil()), types.ImageBuildOptions{
						Tags: []string{"test:tag", "test"},
					}).Return(types.ImageBuildResponse{}, nil).Times(1)

					return mock
				}(),
				opts: BuildOpts{
					Context: path,
					Tags:    []string{"test:tag", "test"},
				},
			},
			want:    &types.ImageBuildResponse{},
			wantErr: false,
		},
		{
			name: "should return error during the build",
			args: args{
				ctx: ctx,
				c: func() ImageClient {
					mock := mock_docker.NewMockImageClient(ctrl)

					mock.EXPECT().ImageBuild(ctx, gomock.Not(gomock.Nil()), types.ImageBuildOptions{}).
						Return(types.ImageBuildResponse{}, errors.New("test error")).Times(1)

					return mock
				}(),
				opts: BuildOpts{
					Context: path,
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BuildImage(tt.args.ctx, tt.args.c, tt.args.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildImage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BuildImage() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFollowBuild(t *testing.T) {
	tests := []struct {
		name        string
		readCloser  io.Reader
		wantCounter int
		wantErr     bool
	}{
		{
			name: "should read from reader and return nil",
			readCloser: strings.NewReader(
				fmt.Sprintf("%s\n%s\n%s\n",
					fixBuildResult("test1", ""),
					fixBuildResult("test2", ""),
					fixBuildResult("test3", ""),
				)),
			wantCounter: 3,
			wantErr:     false,
		},
		{
			name: "should return error got error from buffer",
			readCloser: strings.NewReader(
				fmt.Sprintf("%s\n",
					fixBuildResult("", "sad test error :("))),
			wantCounter: 0,
			wantErr:     true,
		},
		{
			name:        "should return error when can't unmarshal result",
			readCloser:  strings.NewReader("test bad result\n"),
			wantCounter: 0,
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			counter := 0
			counterFunc := func(i ...interface{}) {
				counter++
			}
			if err := FollowBuild(tt.readCloser, counterFunc); (err != nil) != tt.wantErr {
				t.Errorf("FollowBuild() error = %v, wantErr %v", err, tt.wantErr)
			}
			require.Equal(t, tt.wantCounter, counter)
		})
	}
}

func fixBuildResult(stream, err string) []byte {
	res, _ := json.Marshal(&ResultEntry{
		Stream:      stream,
		ErrorDetail: ErrorDetail{Message: err},
		Error:       err,
	})
	return res
}

package docker

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/docker/docker/pkg/ioutils"
)

type ContextOpts struct {
	DirPrefix  string
	Dockerfile string
	SrcDir     string
	SrcFiles   []string
}

const (
	tmpDirFormat       = "%s-"
	codeDir            = "src"
	dockerfileFilename = "Dockerfile"
)

func InlineContext(args ContextOpts, logf func(string, ...interface{})) (string, error) {
	tmpDir, err := ioutils.TempDir(os.TempDir(), fmt.Sprintf(tmpDirFormat, args.DirPrefix))
	if err != nil {
		return tmpDir, err
	}
	logf("context created: %s", tmpDir)

	sourceDir := filepath.Join(tmpDir, codeDir)
	err = os.Mkdir(sourceDir, os.ModePerm)
	if err != nil {
		return tmpDir, err
	}

	for _, file := range args.SrcFiles {
		from := filepath.Join(args.SrcDir, file)
		to := filepath.Join(sourceDir, file)
		logf("Copy file from: %s, to: %s", from, to)
		err = copyFile(from, to)
		if err != nil {
			return tmpDir, err
		}
	}

	dockerfile := filepath.Join(tmpDir, dockerfileFilename)
	logf("Create Dockerfile: %s", dockerfile)
	file, err := os.Create(dockerfile)
	if err != nil {
		return tmpDir, err
	}

	_, err = file.Write([]byte(args.Dockerfile))

	return tmpDir, nil
}

func copyFile(from, to string) error {
	out, err := os.Create(to)
	if err != nil {
		return err
	}

	defer out.Close()

	in, err := os.Open(from)
	defer in.Close()
	if err != nil {
		return err
	}

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	return nil
}

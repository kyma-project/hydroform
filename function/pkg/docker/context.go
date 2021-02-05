package docker

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
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

func Inline(args ContextOpts, logf func(string, ...interface{})) (string, error) {
	tmpDir, err := ioutil.TempDir(os.TempDir(), fmt.Sprintf(tmpDirFormat, args.DirPrefix))
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

	return tmpDir, err
}

func copyFile(from, to string) error {
	out, err := os.Create(to)
	if err != nil {
		return err
	}
	defer out.Close()

	in, err := os.Open(from)
	if err != nil {
		return err
	}
	defer in.Close()

	_, err = io.Copy(out, in)

	return err
}

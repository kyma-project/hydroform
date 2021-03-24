package archive

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/download"
	"github.com/stretchr/testify/assert"
)

const (
	testDir string = "tmp"
	zipDst  string = "tmp/zipDst"
	tarDst  string = "tmp/tarDst"
)

var (
	// The files that we should get after decompressing the test zip file
	zipFiles []string = []string{
		"Chart.yaml",
		"values.yaml",
	}
	// The directories that we should get after decompressing the test zip file
	zipDirs []string = []string{
		"templates",
	}
	// The files that we should get after decompressing the test tar file
	tarFiles []string = []string{
		"Chart.yaml",
		"values.yaml",
	}
	// The directories that we should get after decompressing the test tar file
	tarDirs []string = []string{
		"templates",
		"charts",
	}
)

func TestMain(m *testing.M) {
	// create zip dst folder to which zip file will be decompressed
	if err := os.MkdirAll(zipDst, os.ModePerm); err != nil {
		panic(err)
	}
	// create tar dst folder to which tar file will be decompressed
	if err := os.MkdirAll(tarDst, os.ModePerm); err != nil {
		panic(err)
	}
	exitVal := m.Run()
	// remove tmp folder
	if err := os.RemoveAll(testDir); err != nil {
		panic(err)
	}
	os.Exit(exitVal)
}

func Test_Unzip(t *testing.T) {
	file, _ := download.GetFile("https://storage.googleapis.com/kyma-mps-dev-artifacts/prometheus-config-updater.zip", zipDst)
	err := Unzip(file, zipDst)
	assert.NoError(t, err, "Unzipping file should not error")
	for _, f := range zipFiles {
		assert.FileExists(t, filepath.Join(zipDst, f), "All files in the zip file should be found in dst dir after unzipping")
	}
	for _, d := range zipDirs {
		assert.DirExists(t, filepath.Join(zipDst, d), "All folders in the zip file should be found in dst dir after unzipping")
	}
}

func Test_Untar(t *testing.T) {
	file, _ := download.GetFile("https://storage.googleapis.com/kyma-mps-dev-artifacts/avs-bridge-noparent-1.3.5.tgz", tarDst)
	err := Untar(file, tarDst)
	assert.NoError(t, err, "Untarring file should not error")
	for _, f := range tarFiles {
		assert.FileExists(t, filepath.Join(tarDst, f), "All files in the tar file should be found in dst dir after untarring")
	}
	for _, d := range tarDirs {
		assert.DirExists(t, filepath.Join(tarDst, d), "All folders in the tar file should be found in dst dir after untarring")
	}
}

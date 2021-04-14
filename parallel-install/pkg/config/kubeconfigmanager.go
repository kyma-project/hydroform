package config

import (
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	temporaryFilePattern = "kubeconfig-*.yaml"
)

// kubeConfigManager manages resolving kubeconfig from two possible sources: a file path or a kubeconfig content.
type kubeConfigManager struct {
	kubeconfigSource KubeconfigSource
	temporaryPath    string // Mutable!
}

// CleanupFunc defines contract for the temporary kubeconfig file cleanup function.
type CleanupFunc func() error

// NewKubeConfigManager creates a new instance of KubeConfigManager.
// The logic follows contract defined by KubeconfigSource
func NewKubeConfigManager(kubeconfigSource KubeconfigSource) (*kubeConfigManager, error) {

	pathExists := exists(kubeconfigSource.Path)
	contentExists := exists(kubeconfigSource.Content)

	if !pathExists && !contentExists {
		return nil, errors.New("Either kubeconfig path or kubeconfig content property must be set")
	}

	return &kubeConfigManager{
		kubeconfigSource: kubeconfigSource,
	}, nil
}

// Path returns a path to the kubeconfig file.
// It may render kubeconfig to a temporary file, returned CleanupFunc should be used to remove it.
func (k *kubeConfigManager) Path() (string, CleanupFunc, error) {
	pathExists := exists(k.kubeconfigSource.Path)
	contentExists := exists(k.kubeconfigSource.Content)

	var resPath string
	var cleanupFunc CleanupFunc

	if pathExists {
		// return exiting file path if exists
		resPath = k.kubeconfigSource.Path
		cleanupFunc = func() error { return nil }
	} else if contentExists {

		// return exiting file path if exists
		if k.temporaryPath == "" {
			tempPath, err := createTemporaryFile(k.kubeconfigSource.Content)
			if err != nil {
				return "", nil, err
			}
			k.temporaryPath = tempPath
		}

		cleanupFunc = func() error {
			return os.Remove(k.temporaryPath)
		}

		resPath = k.temporaryPath
	}

	return resPath, cleanupFunc, nil
}

// Config returns a kubeconfig REST Config used by k8s clients.
func (k *kubeConfigManager) Config() (*rest.Config, error) {
	if exists(k.kubeconfigSource.Path) {
		return clientcmd.BuildConfigFromFlags("", k.kubeconfigSource.Path)
	} else {
		return clientcmd.RESTConfigFromKubeConfig([]byte(k.kubeconfigSource.Content))
	}
}

func exists(property string) bool {
	if property == "" {
		return false
	}

	return true
}

func createTemporaryFile(kubeconfigContent string) (string, error) {
	tmpFile, err := ioutil.TempFile(os.TempDir(), temporaryFilePattern)
	if err != nil {
		return "", errors.Wrap(err, "Failed to generate a temporary file for kubeconfig")
	}

	resPath := tmpFile.Name()
	if _, err = tmpFile.Write([]byte(kubeconfigContent)); err != nil {
		return "", errors.Wrapf(err, "Failed to write to the temporary file: %s", resPath)
	}

	if err := tmpFile.Close(); err != nil {
		return "", errors.Wrapf(err, "Failed to close the temporary file: %s", resPath)
	}

	return resPath, nil
}

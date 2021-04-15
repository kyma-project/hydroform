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

// CleanupFunc defines the contract for removing a temporary kubeconfig file.
type CleanupFunc func() error

// Path returns a filesystem path to the kubeconfig file.
// It may render the kubeconfig to a temporary file.
// In order to ensure proper cleanup you should always call the returned CleanupFunc using `defer` statement.
func Path(kubeconfigSource KubeconfigSource) (resPath string, cf CleanupFunc, err error) {

	pathSet := notEmpty(kubeconfigSource.Path)
	contentSet := notEmpty(kubeconfigSource.Content)

	if !pathSet && !contentSet {
		return "", nil, errors.New("Either kubeconfig path or kubeconfig content property must be set")
	}

	if pathSet {
		// return exiting file path
		resPath = kubeconfigSource.Path
		cf = func() error { return nil }
	} else {
		resPath, err = createTemporaryFile(kubeconfigSource.Content)
		if err != nil {
			return "", nil, err
		}

		cf = func() error {
			if _, err := os.Stat(resPath); err == nil {
				return os.Remove(resPath)
			}
			return nil
		}
	}

	return
}

// RestConfig returns a kubeconfig REST Config used by k8s clients.
func RestConfig(kubeconfigSource KubeconfigSource) (*rest.Config, error) {

	pathSet := notEmpty(kubeconfigSource.Path)
	contentSet := notEmpty(kubeconfigSource.Content)

	if !pathSet && !contentSet {
		return nil, errors.New("Either kubeconfig path or kubeconfig content property must be set")
	}

	if notEmpty(kubeconfigSource.Path) {
		return clientcmd.BuildConfigFromFlags("", kubeconfigSource.Path)
	} else {
		return clientcmd.RESTConfigFromKubeConfig([]byte(kubeconfigSource.Content))
	}
}

func notEmpty(property string) bool {
	return property != ""
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

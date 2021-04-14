package config

import (
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// kubeConfigManager handles kubeconfig path and kubeconfig content.
type kubeConfigManager struct {
	path    string
	content string
}

// NewKubeConfigManager creates a new instance of KubeConfigManager.
// TODO: opisac priorytet brania configow
func NewKubeConfigManager(kubeconfigSource KubeconfigSource) (*kubeConfigManager, error) {
	pathExists := exists(kubeconfigSource.Path)
	contentExists := exists(kubeconfigSource.Content)
	var resolvedPath string
	var resolvedContent string

	if pathExists && contentExists {
		resolvedPath = kubeconfigSource.Path
	} else if pathExists {
		resolvedPath = kubeconfigSource.Path
	} else if contentExists {
		resolvedContent = kubeconfigSource.Content
		tmpFile, err := ioutil.TempFile(os.TempDir(), "kubeconfig-*.yaml")
		if err != nil {
			return nil, errors.Wrap(err, "Failed to generate a temporary file for kubeconfig")
		}

		if _, err = tmpFile.Write([]byte(resolvedContent)); err != nil {
			return nil, errors.Wrap(err, "Failed to write to the temporary file")
		}

		resolvedPath = tmpFile.Name()
		if err := tmpFile.Close(); err != nil {
			return nil, errors.Wrap(err, "Failed to close the temporary file")
		}
	} else {
		return nil, errors.New("either kubeconfig or kubeconfigcontent property has to be set")
	}

	return &kubeConfigManager{
		path:    resolvedPath,
		content: resolvedContent,
	}, nil
}

// Path returns a path to the kubeconfig file.
func (k *kubeConfigManager) Path() string {
	return k.path
}

// Path returns a path to the kubeconfig file.
func (k *kubeConfigManager) Cleanup() error {
	return os.Remove(k.path)
}

// Config returns a kubeconfig REST Config used by k8s clients.
func (k *kubeConfigManager) Config() (*rest.Config, error) {
	if exists(k.path) {
		return clientcmd.BuildConfigFromFlags("", k.path)
	} else {
		return clientcmd.RESTConfigFromKubeConfig([]byte(k.content))
	}
}

func exists(property string) bool {
	if property == "" {
		return false
	}

	return true
}

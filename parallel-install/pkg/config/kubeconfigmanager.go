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
func NewKubeConfigManager(path, content *string) (*kubeConfigManager, error) {
	pathExists := exists(path)
	contentExists := exists(content)
	var resolvedPath string
	var resolvedContent string

	if pathExists && contentExists {
		resolvedPath = *path
	} else if pathExists {
		resolvedPath = *path
	} else if contentExists {
		resolvedContent = *content
	} else {
		return nil, errors.New("either kubeconfig or kubeconfigcontent property has to be set")
	}

	return &kubeConfigManager{
		path:    resolvedPath,
		content: resolvedContent,
	}, nil
}

// Path returns a path to the kubeconfig file.
func (k *kubeConfigManager) Path() (string, error) {
	return k.resolvePath()
}

// Config returns a kubeconfig REST Config used by k8s clients.
func (k *kubeConfigManager) Config() (*rest.Config, error) {
	if exists(&k.path) {
		resolvedPath, err := k.resolvePath()
		if err != nil {
			return nil, err
		}

		return clientcmd.BuildConfigFromFlags("", resolvedPath)
	} else {
		return clientcmd.RESTConfigFromKubeConfig([]byte(k.content))
	}
}

func (k *kubeConfigManager) resolvePath() (path string, err error) {
	if exists(&k.path) {
		path = k.path
	} else {
		tmpFile, err := ioutil.TempFile(os.TempDir(), "kubeconfig-*.yaml")
		if err != nil {
			return "", errors.Wrap(err, "Failed to generate a temporary file for kubeconfig")
		}

		if _, err = tmpFile.Write([]byte(k.content)); err != nil {
			return "", errors.Wrap(err, "Failed to write to the temporary file")
		}

		path = tmpFile.Name()

		if err := tmpFile.Close(); err != nil {
			return "", errors.Wrap(err, "Failed to close the temporary file")
		}
	}

	return path, nil
}

func exists(property *string) bool {
	if property == nil || *property == "" {
		return false
	}

	return true
}

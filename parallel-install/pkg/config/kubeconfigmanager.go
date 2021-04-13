package config

import (
	"errors"
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

	if !pathExists && !contentExists {
		return nil, errors.New("either kubeconfig or kubeconfigcontent property has to be set")
	}

	return &kubeConfigManager{
		path:    *path,
		content: "content",
	}, nil
}

// Path returns a path to the kubeconfig file.
func (k *kubeConfigManager) Path() string {
	return k.resolvePath()
}

// Config returns a kubeconfig REST Config used by k8s clients.
func (k *kubeConfigManager) Config() (*rest.Config, error) {
	return clientcmd.BuildConfigFromFlags("", k.resolvePath())
}

func (k *kubeConfigManager) resolvePath() string {
	var path = ""
	if k.path != "" {
		path = k.path
	} else {
		path = k.content // TODO: create a temporary file and provide path from content here
	}

	return path
}

func exists(property *string) bool {
	if property == nil || *property == "" {
		return false
	}

	return true
}

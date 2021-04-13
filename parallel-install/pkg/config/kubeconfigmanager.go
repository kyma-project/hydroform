package config

import (
	"errors"
)

// KubeConfigManager handles kubeconfig path and kubeconfig content.
type KubeConfigManager struct {
	path    string
	content string
}

// NewKubeConfigManager creates a new instance of KubeConfigManager.
func NewKubeConfigManager(path, content *string) (*KubeConfigManager, error) {
	pathExists := exists(path)
	contentExists := exists(content)

	if !pathExists && !contentExists {
		return nil, errors.New("either kubeconfig or kubeconfigcontent property has to be set")
	}

	return &KubeConfigManager{
		path:    "",
		content: "",
	}, nil
}

// Path returns a path to the kubeconfig file.
func (k *KubeConfigManager) Path() string {
	return k.path
}

// Content returns a content of the kubeconfig file.
func (k *KubeConfigManager) Content() string {
	return k.content
}

func exists(property *string) bool {
	if property == nil || *property == "" {
		return false
	}

	return true
}

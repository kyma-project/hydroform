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
	resolvedPath := ""
	resolvedContent := ""

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
func (k *kubeConfigManager) Path() string {
	return k.resolvePath()
}

// Config returns a kubeconfig REST Config used by k8s clients.
func (k *kubeConfigManager) Config() (*rest.Config, error) {
	if k.path != "" {
		return clientcmd.BuildConfigFromFlags("", k.resolvePath())
	} else {
		return clientcmd.RESTConfigFromKubeConfig([]byte(k.content))
	}
}

func (k *kubeConfigManager) resolvePath() string {
	var path = ""
	if k.path != "" {
		path = k.path
	} else {
		// TODO: check if it's correct, error handling, proper path (generated?)
		APIConfig, _ := clientcmd.Load([]byte(k.content))
		clientcmd.WriteToFile(*APIConfig, "test")
		path = "test"
	}

	return path
}

func exists(property *string) bool {
	if property == nil || *property == "" {
		return false
	}

	return true
}

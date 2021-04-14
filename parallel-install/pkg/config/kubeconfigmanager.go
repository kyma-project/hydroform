package config

import (
	"errors"
	"io/ioutil"
	"log"
	"os"

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
	var path string
	if k.path != "" {
		path = k.path
	} else {
		// TODO: use clientcmd.Load and clientcmd.WriteToFile

		tmpFile, err := ioutil.TempFile(os.TempDir(), "kubeconfig-*.yaml")
		if err != nil {
			log.Fatal(err)
		}
		path = tmpFile.Name()
		log.Print(path)

		if _, err = tmpFile.Write([]byte(k.content)); err != nil {
			log.Fatal("Failed to write to temporary file", err)
		}

		if err := tmpFile.Close(); err != nil {
			log.Fatal(err)
		}
	}

	return path
}

func exists(property *string) bool {
	if property == nil || *property == "" {
		return false
	}

	return true
}

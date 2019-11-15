package main

import (
	"context"
	"fmt"
	"github.com/kyma-incubator/hydroform/install/config"
	"github.com/kyma-incubator/hydroform/install/installation"
	"io/ioutil"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

type logger struct{}

func (l logger) Infof(format string, a ...interface{}) {
	log.Println(fmt.Sprintf(format, a...))
}

const (
	tillerYamlUrl = "https://raw.githubusercontent.com/kyma-project/kyma/release-1.7/installation/resources/tiller.yaml" //TODO: Check if there is some better url to fetch tiller yaml
	installerYamlUrl = "https://github.com/kyma-project/kyma/releases/download/1.7.0/kyma-installer-local.yaml"
	configYamlUrl = "https://github.com/kyma-project/kyma/releases/download/1.7.0/kyma-config-local.yaml"
)

func main() {
	log.Printf("Fetching kubeconfig...")
	kubeconfigPath := filepath.Join(homedir.HomeDir(), ".kube", "config")
	k8sConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	logAndExitOnError(err)

	log.Printf("Fetching Tiller config file...")
	tillerYamlContent, err := get(tillerYamlUrl)
	logAndExitOnError(err)

	log.Printf("Fetching Kyma Installer config files...")
	installerYamlContent, err := get(installerYamlUrl)
	logAndExitOnError(err)

	log.Printf("Fetching Kyma Config file...")
	kymaConfigYamlContent, err := get(configYamlUrl)
	decoder, err := installation.DefaultDecoder()
	logAndExitOnError(err)
	configuration, err := config.YAMLToConfiguration(kymaConfigYamlContent, decoder)
	logAndExitOnError(err)

	log.Printf("Creating new Kyma Installer...")
	installer, err := installation.NewKymaInstaller(k8sConfig, installation.WithLogger(logger{}))
	logAndExitOnError(err)

	artifacts := installation.Installation{
		TillerYaml:    tillerYamlContent,
		InstallerYaml: installerYamlContent,
		Configuration: configuration,
	}

	log.Printf("Preparing installation...")
	err = installer.PrepareInstallation(artifacts)
	logAndExitOnError(err)

	log.Printf("Starting installation...")
	stateChannel, errorChannel, err := installer.StartInstallation(context.Background())
	logAndExitOnError(err)

	log.Printf("Waiting for installation to finish...")
	waitForInstallation(stateChannel, errorChannel)

	log.Printf("Installation finished!")
}

func logAndExitOnError(err error) {
	if err != nil {
		log.Printf("Exitting. An error occurred: %v", err)
		os.Exit(1)
	}
}

func get(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("received non OK status code while getting a file from url: %s", url)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func waitForInstallation(stateChannel <-chan installation.InstallationState, errorChannel <-chan error) {
	for {
		select {
		case state, ok := <-stateChannel:
			if !ok {
				log.Println("State channel closed")
				return
			}
			log.Printf("Description: %s, State: %s", state.Description, state.State)
		case err := <-errorChannel:
			log.Printf("An error occurred: %v", err)
		default:
		}
	}
}

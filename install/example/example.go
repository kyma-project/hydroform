package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/kyma-incubator/hydroform/install/scheme"

	"github.com/kyma-incubator/hydroform/install/config"
	"github.com/kyma-incubator/hydroform/install/installation"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type logger struct{}

func (l logger) Infof(format string, a ...interface{}) {
	log.Println(fmt.Sprintf(format, a...))
}

const (
	tillerYamlUrl      = "https://raw.githubusercontent.com/kyma-project/kyma/release-1.12/installation/resources/tiller.yaml"
	installerYamlUrl   = "https://raw.githubusercontent.com/kyma-project/kyma/1.12.0/installation/resources/installer-local.yaml"
	configYamlUrl      = "https://raw.githubusercontent.com/kyma-project/kyma/1.12.0/installation/resources/installer-config-local.yaml.tpl"
	installerCRYamlUrl = "https://raw.githubusercontent.com/kyma-project/kyma/1.12.0/installation/resources/installer-cr.yaml.tpl"

	upgradeTillerYamlUrl      = "https://raw.githubusercontent.com/kyma-project/kyma/release-1.13/installation/resources/tiller.yaml"
	upgradeInstallerYamlUrl   = "https://raw.githubusercontent.com/kyma-project/kyma/1.13.0/installation/resources/installer-local.yaml"
	upgradeConfigYamlUrl      = "https://raw.githubusercontent.com/kyma-project/kyma/1.13.0/installation/resources/installer-config-local.yaml.tpl"
	upgradeInstallerCRYamlUrl = "https://raw.githubusercontent.com/kyma-project/kyma/1.13.0/installation/resources/installer-cr.yaml.tpl"
)

func main() {
	minikubeIp := flag.String("minikubeIP", "192.168.64.2	", "IP of Minikube instance")
	flag.Parse()

	if minikubeIp == nil || *minikubeIp == "" {
		log.Print("minikubeIP is required")
		os.Exit(1)
	}

	log.Printf("Fetching kubeconfig...")
	kubeconfigPath := filepath.Join(homedir.HomeDir(), ".kube", "config")
	k8sConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	logAndExitOnError(err)

	log.Printf("Fetching Tiller config file...")
	tillerYamlContent, err := fetchFile(tillerYamlUrl)
	logAndExitOnError(err)

	log.Printf("Fetching Kyma Installer config files...")
	installerYamlContent, err := fetchFile(installerYamlUrl)
	logAndExitOnError(err)

	log.Printf("Fetching Kyma Installer config files...")
	installerCRYamlContent, err := fetchFile(installerCRYamlUrl)
	logAndExitOnError(err)

	log.Printf("Fetching Kyma Config file...")
	kymaConfigYamlContent, err := fetchFile(configYamlUrl)
	logAndExitOnError(err)
	decoder, err := scheme.DefaultDecoder()
	logAndExitOnError(err)
	configuration, err := config.YAMLToConfiguration(decoder, kymaConfigYamlContent)
	logAndExitOnError(err)

	configuration.Configuration.Set("global.minikubeIP", *minikubeIp, false)

	log.Printf("Creating new Kyma Installer...")
	installer, err := installation.NewKymaInstaller(k8sConfig, installation.WithLogger(logger{}))
	logAndExitOnError(err)

	artifacts := installation.Installation{
		TillerYaml:      tillerYamlContent,
		InstallerYaml:   installerYamlContent,
		InstallerCRYaml: installerCRYamlContent,
		Configuration:   configuration,
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

	log.Printf("Fetching upgraded Tiller config file...")
	tillerYamlContentUpg, err := fetchFile(upgradeTillerYamlUrl)
	logAndExitOnError(err)

	log.Printf("Fetching upgraded Kyma Installer config files...")
	installerYamlContentUpg, err := fetchFile(upgradeInstallerYamlUrl)
	logAndExitOnError(err)

	log.Printf("Fetching upgraded Kyma Installer config files...")
	installerCRYamlContentUpg, err := fetchFile(upgradeInstallerCRYamlUrl)
	logAndExitOnError(err)

	log.Printf("Fetching upgraded Kyma Config file...")
	kymaConfigYamlContentUpg, err := fetchFile(upgradeConfigYamlUrl)
	logAndExitOnError(err)

	configuration, err = config.YAMLToConfiguration(decoder, kymaConfigYamlContentUpg)
	logAndExitOnError(err)

	configuration.Configuration.Set("global.minikubeIP", *minikubeIp, false)

	artifacts = installation.Installation{
		TillerYaml:      tillerYamlContentUpg,
		InstallerYaml:   installerYamlContentUpg,
		InstallerCRYaml: installerCRYamlContentUpg,
		Configuration:   configuration,
	}

	log.Printf("Preparing upgrade...")

	err = installer.PrepareUpgrade(artifacts)
	logAndExitOnError(err)

	log.Printf("Starting upgrade...")
	stateChannel, errorChannel, err = installer.StartInstallation(context.Background())
	logAndExitOnError(err)

	log.Printf("Waiting for upgrade to finish...")
	waitForInstallation(stateChannel, errorChannel)

	log.Printf("Upgrade finished!")
}

func logAndExitOnError(err error) {
	if err != nil {
		log.Printf("Exitting. An error occurred: %v", err)
		os.Exit(1)
	}
}

func fetchFile(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("received non OK status code while getting a file from url: %s", url)
	}

	var data bytes.Buffer
	if _, err := io.Copy(&data, resp.Body); err != nil {
		return "", err
	}

	return data.String(), nil
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
		case err, ok := <-errorChannel:
			if !ok {
				log.Println("Error channel closed")
				continue
			}
			log.Printf("An error occurred: %v", err)

			installationError := installation.InstallationError{}
			if ok := errors.As(err, &installationError); ok {
				log.Printf("Installation errors:")
				for _, e := range installationError.ErrorEntries {
					log.Printf("Component: %s", e.Component)
					log.Printf(e.Log)
				}
			}
		default:
			log.Printf("Waiting for next event...")
			time.Sleep(5 * time.Second)
		}
	}
}

package helm

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

type Client struct {
}

type ClientInterface interface {
	InstallRelease(chartDir, namespace, name string, overrides map[string]interface{}) error
	UninstallRelease(namespace, name string) error
}

func (c *Client) UninstallRelease(namespace, name string) error {

	cfg, err := newActionConfig(namespace)
	if err != nil {
		return err
	}

	uninstall := action.NewUninstall(cfg)
	//TODO: Make configurable
	uninstall.Timeout = 5 * time.Minute

	operation := func() error {
		rel, err := uninstall.Run(name)
		if err != nil {
			//TODO: Find a better way. Maybe explicit check before uninstalling?
			if strings.HasSuffix(err.Error(), "release: not found") {
				return nil
			}
			return err
		}

		if rel == nil || rel.Release == nil || rel.Release.Info == nil {
			return fmt.Errorf("Failed to uninstall %s. Status: %v", name, "Unknown")
		}

		if rel.Release.Info.Status != release.StatusUninstalled {
			return fmt.Errorf("Failed to uninstall %s. Status: %v", name, rel.Release.Info.Status)
		}

		return nil
	}

	//TODO: Find a way to stop backoff once we have Context cancel() function invoked by the global installation timetout.
	exponentialBackoff := backoff.NewExponentialBackOff()
	exponentialBackoff.InitialInterval = 3 * time.Second
	exponentialBackoff.MaxElapsedTime = 10 * time.Minute

	err = backoff.Retry(operation, exponentialBackoff)
	if err != nil {
		return fmt.Errorf("Failed to uninstall %s within the given timeout. Error: %v", name, err)
	}

	return nil
}

func (c *Client) InstallRelease(chartDir, namespace, name string, overrides map[string]interface{}) error {
	cfg, err := newActionConfig(namespace)
	if err != nil {
		return err
	}

	chart, err := loader.Load(chartDir)
	if err != nil {
		return err
	}

	install := action.NewInstall(cfg)
	install.ReleaseName = name
	install.Namespace = namespace
	install.Atomic = true
	install.Wait = true
	install.CreateNamespace = true
	install.Timeout = 3 * time.Minute

	operation := func() error {
		rel, err := install.Run(chart, overrides)
		if err != nil {
			return err
		}

		if rel == nil || rel.Info == nil {
			return fmt.Errorf("Failed to install %s. Status: %v", name, "Unknown")
		}

		if rel.Info.Status != release.StatusDeployed {
			return fmt.Errorf("Failed to install %s. Status: %v", name, rel.Info.Status)
		}

		return nil
	}

	//TODO: Find a way to stop backoff once we have Context cancel() function invoked by the global installation timetout.
	exponentialBackoff := backoff.NewExponentialBackOff()
	exponentialBackoff.InitialInterval = 3 * time.Second
	exponentialBackoff.MaxElapsedTime = 10 * time.Minute

	err = backoff.Retry(operation, exponentialBackoff)
	if err != nil {
		return fmt.Errorf("Failed to install %s within the given timeout. Error: %v", name, err)
	}

	return err
}

func newActionConfig(namespace string) (*action.Configuration, error) {
	clientGetter := genericclioptions.NewConfigFlags(false)
	clientGetter.Namespace = &namespace

	cfg := new(action.Configuration)
	if err := cfg.Init(clientGetter, namespace, "secrets", log.Printf); err != nil {
		return nil, err
	}

	return cfg, nil
}

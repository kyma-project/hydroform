package helm

import (
	"fmt"
	"log"
	"time"

	"github.com/avast/retry-go"
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

	maxAttempts := 3
	fixedDelay := 3

	err = retry.Do(
		func() error {
			rel, err := uninstall.Run(name)
			if err != nil {
				return err
			}

			if rel == nil || rel.Release == nil || rel.Release.Info == nil {
				return fmt.Errorf("Failed to uninstall %s. Status: %v", name, "Unknown")
			}

			if rel.Release.Info.Status != release.StatusUninstalled {
				return fmt.Errorf("Failed to uninstall %s. Status: %v", name, rel.Release.Info.Status)
			}

			return nil
		},
		retry.Attempts(uint(maxAttempts)),
		retry.DelayType(func(attempt uint, config *retry.Config) time.Duration {
			log.Printf("Retry number %d on uninstalling %s.\n", attempt+1, name)
			return time.Duration(fixedDelay) * time.Second
		}),
	)

	return err
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

	maxAttempts := 3
	fixedDelay := 3

	err = retry.Do(
		func() error {
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
		},
		retry.Attempts(uint(maxAttempts)),
		retry.DelayType(func(attempt uint, config *retry.Config) time.Duration {
			log.Printf("Retry number %d on installing %s.\n", attempt+1, name)
			return time.Duration(fixedDelay) * time.Second
		}),
	)

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

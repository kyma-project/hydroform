package helm

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/config"

	"github.com/cenkalti/backoff/v4"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

const logPrefix = "[helm/client.go]"

type Config struct {
	HelmTimeoutSeconds            int
	BackoffInitialIntervalSeconds int
	BackoffMaxElapsedTimeSeconds  int
	Log                           func(format string, v ...interface{})
}

type Client struct {
	cfg Config
}

type ClientInterface interface {
	InstallRelease(ctx context.Context, chartDir, namespace, name string, overrides map[string]interface{}) error
	UninstallRelease(ctx context.Context, namespace, name string) error
}

func NewClient(cfg Config) *Client {
	return &Client{
		cfg: cfg,
	}
}

func (c *Client) UninstallRelease(ctx context.Context, namespace, name string) error {

	cfg, err := newActionConfig(namespace)
	if err != nil {
		return err
	}

	uninstall := action.NewUninstall(cfg)
	uninstall.Timeout = time.Duration(c.cfg.HelmTimeoutSeconds) * time.Second

	operation := func() error {
		c.cfg.Log("%s Starting uninstall for release %s in namespace %s", logPrefix, name, namespace)
		rel, err := uninstall.Run(name)
		if err != nil {
			//TODO: Find a better way. Maybe explicit check before uninstalling?
			if strings.HasSuffix(err.Error(), "release: not found") {
				return nil
			}
			c.cfg.Log("%s Error: %v", logPrefix, err)
			return err
		}

		if rel == nil || rel.Release == nil || rel.Release.Info == nil {
			err = fmt.Errorf("Failed to uninstall %s. Status: %v", name, "Unknown")
			c.cfg.Log("%s Error: %v", logPrefix, err)
			return err
		}

		if rel.Release.Info.Status != release.StatusUninstalled {
			err = fmt.Errorf("Failed to uninstall %s. Status: %v", name, rel.Release.Info.Status)
			c.cfg.Log("%s Error: %v", logPrefix, err)
			return err
		}

		return nil
	}

	initialInterval := time.Duration(c.cfg.BackoffInitialIntervalSeconds) * time.Second
	maxElapsedTime := time.Duration(c.cfg.BackoffMaxElapsedTimeSeconds) * time.Second
	err = retryWithBackoff(ctx, operation, initialInterval, maxElapsedTime)
	if err != nil {
		return fmt.Errorf("Error: Failed to uninstall %s within the configured time. Error: %v", name, err)
	}

	return nil
}

func (c *Client) InstallRelease(ctx context.Context, chartDir, namespace, name string, overrides map[string]interface{}) error {
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
	install.Timeout = time.Duration(c.cfg.HelmTimeoutSeconds) * time.Second

	operation := func() error {
		c.cfg.Log("%s Starting install for release %s in namespace %s", logPrefix, name, namespace)
		rel, err := install.Run(chart, overrides)
		if err != nil {
			c.cfg.Log("%s Error: %v", logPrefix, err)
			return err
		}

		if rel == nil || rel.Info == nil {
			err = fmt.Errorf("Failed to install %s. Status: %v", name, "Unknown")
			c.cfg.Log("%s Error: %v", logPrefix, err)
			return err
		}

		if rel.Info.Status != release.StatusDeployed {
			err = fmt.Errorf("Failed to install %s. Status: %v", name, rel.Info.Status)
			c.cfg.Log("%s Error: %v", logPrefix, err)
			return err
		}

		return nil
	}

	initialInterval := time.Duration(c.cfg.BackoffInitialIntervalSeconds) * time.Second
	maxElapsedTime := time.Duration(c.cfg.BackoffMaxElapsedTimeSeconds) * time.Second
	err = retryWithBackoff(ctx, operation, initialInterval, maxElapsedTime)
	if err != nil {
		return fmt.Errorf("Error: Failed to install %s within the configured time. Error: %v", name, err)
	}

	return nil
}

func retryWithBackoff(ctx context.Context, operation func() error, initialInterval, maxTime time.Duration) error {

	exponentialBackoff := backoff.NewExponentialBackOff()
	exponentialBackoff.InitialInterval = initialInterval
	exponentialBackoff.MaxElapsedTime = maxTime

	err := backoff.Retry(operation, backoff.WithContext(exponentialBackoff, ctx))
	if err != nil {
		return err
	}
	return nil
}

func newActionConfig(namespace string) (*action.Configuration, error) {
	clientGetter := genericclioptions.NewConfigFlags(false)
	clientGetter.Namespace = &namespace

	cfg := new(action.Configuration)
	if err := cfg.Init(clientGetter, namespace, "secrets", config.Log); err != nil {
		return nil, err
	}

	return cfg, nil
}

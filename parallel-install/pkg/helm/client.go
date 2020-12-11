//Package helm implements a wrapper over a native Helm client.
//The wrapper exposes a simple installation API and the configuration.
//
//The code in the package uses the user-provided function for logging.
package helm

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/config"
	"helm.sh/helm/v3/pkg/storage/driver"

	"github.com/cenkalti/backoff/v4"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

const logPrefix = "[helm/client.go]"

//Config provides configuration for the Client.
type Config struct {
	HelmTimeoutSeconds            int                                   //Underlying native Helm client processing timeout
	BackoffInitialIntervalSeconds int                                   //Initial interval for the exponential backoff retry algorithm
	BackoffMaxElapsedTimeSeconds  int                                   //Maximum time for the exponential backoff retry algorithm
	MaxHistory                    int                                   //Maximum number of revisions saved per release
	Log                           func(format string, v ...interface{}) //Used for logging
}

//Client implements the ClientInterface.
type Client struct {
	cfg Config
}

//ClientInterface defines the contract for the Helm-related installation processes.
type ClientInterface interface {
	//DeployRelease deploys a named chart from a local filesystem directory with specific overrides.
	//The function retries on errors according to Config provided to the Client.
	//
	//ctx is used for the operation cancellation.
	//Cancellation of the successful operation is not possible
	//because the underlying Helm operations are blocking and do not support Context-based cancellation.
	//Cancellation is possible when errors occur and the operation is re-tried.
	//When the operation is re-tried, it is not guaranteed that the cancellation is handled immediately due to the blocking nature of Helm client calls.
	//However, once the underlying Helm operation ends, the "cancel" condition is detected and the operation's result is returned without further retries.
	DeployRelease(ctx context.Context, chartDir, namespace, name string, overrides map[string]interface{}) error
	//UninstallRelease uninstalls a named chart from the cluster.
	//The function retries on errors according to Config provided to the Client.
	//
	//ctx is used for the operation cancellation.
	//Cancellation of the successful operation is not possible
	//because the underlying Helm operations are blocking and do not support the Context-based cancellation.
	//Cancellation is possible when errors occur and the operation is re-tried.
	//When the operation is re-tried, it is not guaranteed that the cancellation is handled immediately due to the blocking nature of Helm client calls.
	//However, once the underlying Helm operation ends, the cancel condition is detected and the operation's result is returned without further retries.
	UninstallRelease(ctx context.Context, namespace, name string) error
}

//NewClient returns a new Client instance.
//If you need different configurations for installation and uninstallation,
//just create two different Client instances with different configurations.
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

func (c *Client) upgradeRelease(ctx context.Context, chartDir, namespace, name string, overrides map[string]interface{}, cfg *action.Configuration, chart *chart.Chart) error {
	upgrade := action.NewUpgrade(cfg)
	upgrade.Atomic = false
	upgrade.CleanupOnFail = true
	upgrade.Wait = true
	upgrade.ReuseValues = false
	upgrade.Recreate = false
	upgrade.MaxHistory = c.cfg.MaxHistory
	upgrade.Timeout = time.Duration(c.cfg.HelmTimeoutSeconds) * time.Second

	operation := func() error {
		c.cfg.Log("%s Starting deploy for release %s in namespace %s", logPrefix, name, namespace)
		rel, err := upgrade.Run(name, chart, overrides)
		if err != nil {
			c.cfg.Log("%s Error: %v", logPrefix, err)
			return err
		}

		if rel == nil || rel.Info == nil {
			err = fmt.Errorf("Failed to deploy %s. Status: %v", name, "Unknown")
			c.cfg.Log("%s Error: %v", logPrefix, err)
			return err
		}

		if rel.Info.Status != release.StatusDeployed {
			err = fmt.Errorf("Failed to deploy %s. Status: %v", name, rel.Info.Status)
			c.cfg.Log("%s Error: %v", logPrefix, err)
			return err
		}

		return nil
	}

	initialInterval := time.Duration(c.cfg.BackoffInitialIntervalSeconds) * time.Second
	maxElapsedTime := time.Duration(c.cfg.BackoffMaxElapsedTimeSeconds) * time.Second
	err := retryWithBackoff(ctx, operation, initialInterval, maxElapsedTime)
	if err != nil {
		return fmt.Errorf("Error: Failed to deploy %s within the configured time. Error: %v", name, err)
	}

	return nil
}

func (c *Client) installRelease(ctx context.Context, chartDir, namespace, name string, overrides map[string]interface{}, cfg *action.Configuration, chart *chart.Chart) error {
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
	err := retryWithBackoff(ctx, operation, initialInterval, maxElapsedTime)
	if err != nil {
		return fmt.Errorf("Error: Failed to install %s within the configured time. Error: %v", name, err)
	}

	return nil
}

func (c *Client) DeployRelease(ctx context.Context, chartDir, namespace, name string, overrides map[string]interface{}) error {
	cfg, err := newActionConfig(namespace)
	if err != nil {
		return err
	}

	chart, err := loader.Load(chartDir)
	if err != nil {
		return err
	}

	upgrade, err := isUpgrade(name, cfg)
	if err != nil {
		return err
	}

	if upgrade {
		err = c.upgradeRelease(ctx, chartDir, namespace, name, overrides, cfg, chart)
		if err != nil {
			return err
		}
	} else {
		err = c.installRelease(ctx, chartDir, namespace, name, overrides, cfg, chart)
		if err != nil {
			return err
		}
	}
	return nil
}

func isUpgrade(name string, cfg *action.Configuration) (bool, error) {
	history := action.NewHistory(cfg)
	history.Max = 1

	_, err := history.Run(name)
	if err != nil {
		if err == driver.ErrReleaseNotFound {
			return false, nil
		} else {
			return false, err
		}
	}

	return true, nil
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

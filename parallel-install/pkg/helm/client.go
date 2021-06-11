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

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/logger"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/overrides"
	"helm.sh/helm/v3/pkg/chartutil"

	"helm.sh/helm/v3/pkg/storage/driver"

	"github.com/cenkalti/backoff/v4"
	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

const logPrefix = "[helm/client.go]"

//Config provides configuration for the Client.
type Config struct {
	HelmTimeoutSeconds            int              //Underlying native Helm client processing timeout
	BackoffInitialIntervalSeconds int              //Initial interval for the exponential backoff retry algorithm
	BackoffMaxElapsedTimeSeconds  int              //Maximum time for the exponential backoff retry algorithm
	MaxHistory                    int              //Maximum number of revisions saved per release
	Log                           logger.Interface //Used for logging
	Atomic                        bool
	KymaComponentMetadataTemplate *KymaComponentMetadataTemplate
	KubeconfigSource              config.KubeconfigSource
	ReuseValues                   bool //Reuse values for component upgrade
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
	DeployRelease(ctx context.Context, chartDir, namespace, name string, overrides map[string]interface{}, profile string) error
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
	path, cleanupFunc, err := config.Path(c.cfg.KubeconfigSource)
	if err != nil {
		return err
	}

	defer func() {
		cleanupErr := cleanupFunc()
		if cleanupErr != nil {
			c.cfg.Log.Error(cleanupErr)
		}
	}()

	cfg, err := c.newActionConfig(namespace, path)
	if err != nil {
		return err
	}

	uninstall := action.NewUninstall(cfg)
	uninstall.Timeout = time.Duration(c.cfg.HelmTimeoutSeconds) * time.Second

	operation := func() error {
		c.cfg.Log.Infof("%s Starting uninstall for release %s in namespace %s", logPrefix, name, namespace)
		rel, err := uninstall.Run(name)
		if err != nil {
			//TODO: Find a better way. Maybe explicit check before uninstalling?
			if strings.HasSuffix(err.Error(), "release: not found") {
				return nil
			}
			c.cfg.Log.Errorf("%s Error: %v", logPrefix, err)
			return err
		}

		if rel == nil || rel.Release == nil || rel.Release.Info == nil {
			err = fmt.Errorf("Failed to uninstall %s. Status: %v", name, "Unknown")
			c.cfg.Log.Errorf("%s Error: %v", logPrefix, err)
			return err
		}

		if rel.Release.Info.Status != release.StatusUninstalled {
			err = fmt.Errorf("Failed to uninstall %s. Status: %v", name, rel.Release.Info.Status)
			c.cfg.Log.Errorf("%s Error: %v", logPrefix, err)
			return err
		}

		return nil
	}

	initialInterval := time.Duration(c.cfg.BackoffInitialIntervalSeconds) * time.Second
	maxElapsedTime := time.Duration(c.cfg.BackoffMaxElapsedTimeSeconds) * time.Second
	err = c.retryWithBackoff(ctx, operation, initialInterval, maxElapsedTime)
	if err != nil {
		return fmt.Errorf("Error: Failed to uninstall %s within the configured time. Error: %v", name, err)
	}

	return nil
}

func (c *Client) newUpgrade(cfg *action.Configuration) *action.Upgrade {
	upgrade := action.NewUpgrade(cfg)
	upgrade.Atomic = c.cfg.Atomic
	upgrade.CleanupOnFail = true
	upgrade.Wait = true
	upgrade.ReuseValues = c.cfg.ReuseValues
	upgrade.Recreate = false
	upgrade.MaxHistory = c.cfg.MaxHistory
	upgrade.Timeout = time.Duration(c.cfg.HelmTimeoutSeconds) * time.Second
	return upgrade
}

func (c *Client) upgradeRelease(namespace, name string, overrides map[string]interface{}, cfg *action.Configuration, chart *chart.Chart) error {
	upgrade := c.newUpgrade(cfg)
	c.cfg.Log.Infof("%s Starting upgrade for release %s in namespace %s", logPrefix, name, namespace)
	rel, err := upgrade.Run(name, chart, overrides)
	if rel != nil {
		if errUpdateMeta := c.updateKymaMetadata(cfg, rel); errUpdateMeta != nil {
			if err != nil {
				return errors.Wrap(err, errUpdateMeta.Error())
			}
			return errUpdateMeta
		}
	}

	if err != nil {
		c.cfg.Log.Errorf("%s Error: %v", logPrefix, err)
		return err
	}

	if rel == nil || rel.Info == nil {
		err = fmt.Errorf("Failed to upgrade %s. Status: %v", name, "Unknown")
		c.cfg.Log.Errorf("%s Error: %v", logPrefix, err)
		return err
	}

	if rel.Info.Status != release.StatusDeployed {
		err = fmt.Errorf("Failed to upgrade %s. Status: %v", name, rel.Info.Status)
		c.cfg.Log.Errorf("%s Error: %v", logPrefix, err)
		return err
	}

	return nil
}

func (c *Client) installRelease(namespace, name string, overrides map[string]interface{}, cfg *action.Configuration, chart *chart.Chart) error {
	install := action.NewInstall(cfg)
	install.ReleaseName = name
	install.Namespace = namespace
	install.Atomic = c.cfg.Atomic
	install.Wait = true
	install.CreateNamespace = true
	install.Timeout = time.Duration(c.cfg.HelmTimeoutSeconds) * time.Second

	c.cfg.Log.Infof("%s Starting install for release %s in namespace %s", logPrefix, name, namespace)
	rel, err := install.Run(chart, overrides)
	if rel != nil {
		if errUpdateMeta := c.updateKymaMetadata(cfg, rel); errUpdateMeta != nil {
			if err != nil {
				return errors.Wrap(err, errUpdateMeta.Error())
			}
			return errUpdateMeta
		}
	}

	if err != nil {
		c.cfg.Log.Errorf("%s Error: %v", logPrefix, err)
		return err
	}

	if rel == nil || rel.Info == nil {
		err = fmt.Errorf("Failed to install %s. Status: %v", name, "Unknown")
		c.cfg.Log.Errorf("%s Error: %v", logPrefix, err)
		return err
	}

	if rel.Info.Status != release.StatusDeployed {
		err = fmt.Errorf("Failed to install %s. Status: %v", name, rel.Info.Status)
		c.cfg.Log.Errorf("%s Error: %v", logPrefix, err)
		return err
	}

	return nil
}

func (c *Client) rollbackRelease(name string, cfg *action.Configuration) error {
	rollback := action.NewRollback(cfg)
	rollback.CleanupOnFail = true
	rollback.Wait = true
	rollback.Timeout = time.Duration(c.cfg.HelmTimeoutSeconds) * time.Second

	c.cfg.Log.Infof("%s Starting rollback of release %s", logPrefix, name)
	err := rollback.Run(name)
	if err != nil {
		c.cfg.Log.Errorf("%s Error: %v", logPrefix, err)
	}
	return err
}

func (c *Client) DeployRelease(ctx context.Context, chartDir, namespace, name string, overridesValues map[string]interface{}, profile string) error {
	path, cleanupFunc, err := config.Path(c.cfg.KubeconfigSource)
	if err != nil {
		return err
	}

	defer func() {
		cleanupErr := cleanupFunc()
		if cleanupErr != nil {
			c.cfg.Log.Error(cleanupErr)
		}
	}()

	operation := func() error {
		cfg, err := c.newActionConfig(namespace, path)
		if err != nil {
			return err
		}

		chart, err := loader.Load(chartDir)
		if err != nil {
			return err
		}

		profileValues, err := getProfileValues(*chart, profile)
		if err != nil {
			return err
		}

		comboValues := overrides.MergeMaps(profileValues, overridesValues)

		isInstalled, err := c.isReleaseInstalled(ctx, namespace, name, cfg)
		if err != nil {
			return err
		}

		if isInstalled {
			err = c.upgradeRelease(namespace, name, comboValues, cfg, chart)
		} else {
			err = c.installRelease(namespace, name, comboValues, cfg, chart)
		}
		return err
	}

	initialInterval := time.Duration(c.cfg.BackoffInitialIntervalSeconds) * time.Second
	maxElapsedTime := time.Duration(c.cfg.BackoffMaxElapsedTimeSeconds) * time.Second
	err = c.retryWithBackoff(ctx, operation, initialInterval, maxElapsedTime)
	if err != nil {
		return fmt.Errorf("Error: Failed to deploy %s within the configured time. Error: %v", name, err)
	}

	return nil
}

func (c *Client) isReleaseInstalled(ctx context.Context, namespace, name string, cfg *action.Configuration) (bool, error) {
	history := action.NewHistory(cfg)
	history.Max = 2

	rels, err := history.Run(name)

	if err != nil {
		if err == driver.ErrReleaseNotFound {
			//release was never installed
			c.cfg.Log.Infof("%s Release '%s' wasn't installed yet", logPrefix, name)
			return false, nil
		}
		return false, err
	}

	//ensure last release is in consistent status
	relsCount := len(rels)
	lastRelease := rels[relsCount-1]
	if c.isPendingReleaseStatus(lastRelease.Info.Status) {
		c.cfg.Log.Infof("%s Release '%s' is in pending state '%s': starting cleanup", logPrefix, name, lastRelease.Info.Status)
		if relsCount > 1 {
			//rollback to previous release
			c.cfg.Log.Infof("%s Release '%s' was already installed before: trigger rollback of pending release", logPrefix, name)
			if err := c.rollbackRelease(name, cfg); err != nil {
				return true, err
			}
		} else {
			//first release installation wasn't finished: delete incomplete release
			c.cfg.Log.Infof("%s Release '%s' was not installed before: trigger uninstall of pending release", logPrefix, name)
			if err := c.UninstallRelease(ctx, namespace, name); err != nil {
				return false, err
			}
			return false, nil
		}
	}

	c.cfg.Log.Infof("%s Release '%s' is installed and has non-pending status", logPrefix, name)
	return true, nil
}

func (c *Client) isPendingReleaseStatus(relStatus release.Status) bool {
	return relStatus == release.StatusPendingInstall || relStatus == release.StatusPendingUpgrade || relStatus == release.StatusPendingRollback
}

func getProfileValues(ch chart.Chart, profileName string) (map[string]interface{}, error) {
	var profile *chart.File
	for _, f := range ch.Files {
		if (f.Name == fmt.Sprintf("profile-%s.yaml", profileName)) || (f.Name == fmt.Sprintf("%s.yaml", profileName)) {
			profile = f
			break
		}
	}
	if profile == nil {
		return ch.Values, nil
	}
	profileValues, err := chartutil.ReadValues(profile.Data)
	if err != nil {
		return nil, err
	}
	return profileValues, nil
}

func (c *Client) retryWithBackoff(ctx context.Context, operation func() error, initialInterval, maxTime time.Duration) error {

	exponentialBackoff := backoff.NewExponentialBackOff()
	exponentialBackoff.InitialInterval = initialInterval
	exponentialBackoff.MaxElapsedTime = maxTime

	err := backoff.Retry(operation, backoff.WithContext(exponentialBackoff, ctx))
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) newActionConfig(namespace string, kubeconfigPath string) (*action.Configuration, error) {
	clientGetter := genericclioptions.NewConfigFlags(false)
	clientGetter.Namespace = &namespace
	clientGetter.KubeConfig = &kubeconfigPath

	cfg := new(action.Configuration)

	debugLogFunc := func(format string, args ...interface{}) { //leverage debugLog function to use logger instance
		c.cfg.Log.Info(fmt.Sprintf(format, args...))
	}
	if err := cfg.Init(clientGetter, namespace, "secrets", debugLogFunc); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Client) updateKymaMetadata(cfg *action.Configuration, rel *release.Release) error {
	//add Kyma metadata to Helm release secret
	kubeClient, err := cfg.KubernetesClientSet()
	if err == nil {
		err = (&KymaMetadataProvider{kubeClient: kubeClient}).Set(rel, c.cfg.KymaComponentMetadataTemplate)
	}
	if err != nil {
		c.cfg.Log.Errorf("%s Error: %v", logPrefix, err)
	}
	return err
}

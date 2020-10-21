package helm

import (
	"log"
	"time"

	"github.com/avast/retry-go"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

type Client struct {
}

type ClientInterface interface {
	InstallRelease(chartDir, namespace, name string, overrides map[string]interface{}) error
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

	maxAttempts := 3
	fixedDelay := 3

	install := action.NewInstall(cfg)
	install.ReleaseName = name
	install.Namespace = namespace
	install.Atomic = true
	install.Wait = true
	install.CreateNamespace = true

	err = retry.Do(
		func() error {
			_, err = install.Run(chart, overrides)

			if err != nil {
				return err
			}
			return nil
		},
		retry.Attempts(uint(maxAttempts)),
		retry.DelayType(func(attempt uint, config *retry.Config) time.Duration {
			log.Printf("Retry number %d on getting release status.\n", attempt+1)
			return time.Duration(fixedDelay) * time.Second
		}),
	)

	return nil
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

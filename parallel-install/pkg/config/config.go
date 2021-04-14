package config

import (
	"fmt"
	"os"
	"time"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/logger"
)

//Configures various install/uninstall operation parameters.
//There are no different parameters for the "install" and "delete" operations.
//If you need different configurations, just use two different Installation instances.
type Config struct {
	//Number of parallel workers used for an install/uninstall operation
	WorkersCount int
	//After this time workers' context is canceled. Pending worker goroutines (if any) may continue if blocked by Helm client.
	CancelTimeout time.Duration
	//After this time install/delete operation is aborted and returns an error to the user.
	//Worker goroutines may still be working in the background.
	//Must be greater than CancelTimeout.
	QuitTimeout time.Duration
	//Timeout for the underlying Helm client
	HelmTimeoutSeconds int
	//Initial interval used for exponent backoff retry policy
	BackoffInitialIntervalSeconds int
	//Maximum time used for exponent backoff retry policy
	BackoffMaxElapsedTimeSeconds int
	//Logger to use
	Log logger.Interface
	//Maximum number of Helm revision saved per release
	HelmMaxRevisionHistory int
	//Installation / Upgrade profile: evaluation|production
	Profile string
	// Kyma components list
	ComponentList *ComponentList
	// Path to Kyma resources
	ResourcePath string
	// Path to Kyma installation resources
	InstallationResourcePath string
	// Path to Kubeconfig
	KubeconfigPath string
	// Content of Kubeconfig
	KubeconfigContent string
	//Kyma version
	Version string
	//Atomic deployment
	Atomic bool
}

// validate verifies that mandatory options are provided
func (c *Config) validate() error {
	if c.WorkersCount <= 0 {
		return fmt.Errorf("Workers count cannot be <= 0")
	}
	if c.ComponentList == nil {
		return fmt.Errorf("Component list undefined")
	}
	return nil
}

// ValidateDeletion verifies that deletion specific options are provided
func (c *Config) ValidateDeletion() error {
	if err := c.validate(); err != nil { //deployment requires all core options
		return err
	}
	return nil
}

// ValidateDeployment verifies that deployment specific options are provided
func (c *Config) ValidateDeployment() error {
	if err := c.validate(); err != nil { //deployment requires all core options
		return err
	}
	if err := c.pathExists(c.ResourcePath, "Resource path"); err != nil {
		return err
	}
	if err := c.pathExists(c.InstallationResourcePath, "Installation resource path"); err != nil {
		return err
	}
	if c.Version == "" {
		return fmt.Errorf("Version is empty")
	}
	return nil
}

func (c *Config) pathExists(path string, description string) error {
	if path == "" {
		return fmt.Errorf("%s is empty", description)
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("%s '%s' not found", description, path)
	}
	return nil
}

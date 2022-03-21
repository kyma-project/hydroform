package terraform

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/hashicorp/terraform-svchost/disco"
	"github.com/hashicorp/terraform/command"
	"github.com/hashicorp/terraform/command/cliconfig"
	"github.com/hashicorp/terraform/command/format"
	"github.com/hashicorp/terraform/plugin/discovery"
	"github.com/kyma-incubator/hydroform/provision/types"
	"github.com/mitchellh/colorstring"

	hashiCli "github.com/mitchellh/cli"
)

const defaultDataDir = "./.hydroform/"

// Options contains all configuration for the terraform operator
type Options struct {
	command.Meta
	// Persistent allows to configure if terraform files should stay in the file system or be cleaned up after each operation.
	Persistent bool
	// TODO add module source property here corresponding to flag -from-module

	// Timeouts specifies the timeouts of the operations
	Timeouts types.Timeouts

	// Print terraform log for debugging
	Verbose bool
}

// Option is a function that allows to extensibly configure the terraform operator.
type Option func(ops *Options)

// Set a custom UI for the Meta configuration.
func WithUI(ui hashiCli.Ui) Option {
	return func(ops *Options) {
		ops.Meta.Ui = ui
	}
}

// Set a custom directory where all Hydroform files will be stored.
func WithDataDir(dir string) Option {
	return func(ops *Options) {
		ops.Meta.OverrideDataDir = dir
	}
}

// Make files persistent after using terraform
func Persistent() Option {
	return func(ops *Options) {
		ops.Persistent = true
	}
}

// Sets operation timeouts
func WithTimeouts(timeouts types.Timeouts) Option {
	return func(ops *Options) {
		ops.Timeouts = timeouts
	}
}

func Verbose(verbose bool) Option {
	return func(ops *Options) {
		ops.Verbose = verbose
	}
}

// ToTerraformOptions turns Hydroform options into terraform operator specific options
func ToTerraformOptions(ops *types.Options) (tfOps []Option) {

	if ops.DataDir != "" {
		tfOps = append(tfOps, WithDataDir(ops.DataDir))
	}

	if ops.Persistent {
		tfOps = append(tfOps, Persistent())
	}

	if ops.Timeouts != nil {
		tfOps = append(tfOps, WithTimeouts(*ops.Timeouts))
	}

	if ops.Verbose {
		tfOps = append(tfOps, Verbose(ops.Verbose))
	}

	return tfOps
}

// options creates a configuration for the terraform operator
// Use Option functions to configure its fields.
func options(ops ...Option) Options {
	// create default meta
	ui := &HydroUI{}

	pluginsDirs, err := globalPluginDirs()
	if err != nil {
		ui.Error(fmt.Sprintf("Error setting terraform plugins directory: %s", err))
		return Options{}
	}

	config, diags := cliconfig.LoadConfig()
	if len(diags) > 0 {
		ui.Error("There are some problems with the CLI configuration:")
		for _, diag := range diags {
			earlyColor := &colorstring.Colorize{
				Colors:  colorstring.DefaultColors,
				Disable: true,
				Reset:   true,
			}
			ui.Error(format.Diagnostic(diag, nil, earlyColor, 78))
		}
		if diags.HasErrors() {
			ui.Error("As a result of the above problems, Terraform may not behave as intended.\n\n")
			// We continue to run anyway, since Terraform has reasonable defaults.
		}
	}

	helperPlugins := discovery.FindPlugins("credentials", pluginsDirs)
	credsSrc, err := config.CredentialsSource(helperPlugins)
	if err != nil {
		ui.Error(fmt.Sprintf("Error loading terraform provider plugins: %s", err))
		return Options{}
	}
	services := disco.NewWithCredentialsSource(credsSrc)

	configDir, err := cliconfig.ConfigDir()
	if err != nil {
		configDir = ""
	}

	tfOps := Options{
		Meta: command.Meta{
			GlobalPluginDirs:    pluginsDirs,
			Ui:                  ui,
			Services:            services,
			RunningInAutomation: true,
			CLIConfigDir:        configDir,
			PluginCacheDir:      pluginsDirs[0],
			OverrideDataDir:     defaultDataDir,
			ShutdownCh:          makeShutdownCh(),
		},
	}

	// apply custom configs
	for _, o := range ops {
		o(&tfOps)
	}

	return tfOps
}

func makeShutdownCh() <-chan struct{} {
	resultCh := make(chan struct{})

	signalCh := make(chan os.Signal, 4)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		for {
			<-signalCh
			resultCh <- struct{}{}
		}
	}()
	return resultCh
}

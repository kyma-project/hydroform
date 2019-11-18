package terraform

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"

	"github.com/hashicorp/terraform-svchost/disco"
	"github.com/hashicorp/terraform/command"
	"github.com/hashicorp/terraform/command/cliconfig"
	"github.com/hashicorp/terraform/command/format"
	"github.com/hashicorp/terraform/plugin/discovery"
	"github.com/mitchellh/colorstring"

	hashiCli "github.com/mitchellh/cli"
)

const defaultDataDir = "./.hydroform/"

// Options contains all configuration for the terraform operator
type Options struct {
	command.Meta
	// TODO add module source property here corresponding to flag -from-module
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

// options creates a configuration for the terraform operator
// Use Option functions to configure its fields.
func options(ops ...Option) Options {
	// create default meta

	// Problems loading the default terraform config will just be output to stderr
	// If the config is loaded successfully we will start using the custom UI if any
	Ui := &hashiCli.BasicUi{
		Reader:      os.Stdin,
		Writer:      os.Stdout,
		ErrorWriter: os.Stderr,
	}

	pluginsDirs, err := globalPluginDirs()
	if err != nil {
		Ui.Error(fmt.Sprintf("Error setting terraform plugins directory: %s", err))
		return Options{}
	}

	config, diags := cliconfig.LoadConfig()
	if len(diags) > 0 {
		Ui.Error("There are some problems with the CLI configuration:")
		for _, diag := range diags {
			earlyColor := &colorstring.Colorize{
				Colors:  colorstring.DefaultColors,
				Disable: true,
				Reset:   true,
			}
			Ui.Error(format.Diagnostic(diag, nil, earlyColor, 78))
		}
		if diags.HasErrors() {
			Ui.Error("As a result of the above problems, Terraform may not behave as intended.\n\n")
			// We continue to run anyway, since Terraform has reasonable defaults.
		}
	}

	helperPlugins := discovery.FindPlugins("credentials", pluginsDirs)
	credsSrc, err := config.CredentialsSource(helperPlugins)
	if err != nil {
		Ui.Error(fmt.Sprintf("Error loading terraform provider plugins: %s", err))
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
			Ui:                  Ui,
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

func globalPluginDirs() ([]string, error) {
	var ret []string
	// Look in ~/.terraform.d/plugins/ , or its equivalent on non-UNIX
	dir, err := cliconfig.ConfigDir()
	if err != nil {
		return nil, err
	}
	machineDir := fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH)
	ret = append(ret, filepath.Join(dir, "plugins"))
	ret = append(ret, filepath.Join(dir, "plugins", machineDir))

	return ret, nil
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

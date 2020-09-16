/*
* CODE GENERATED AUTOMATICALLY WITH devops/internal/config
 */

package main

import (
	"os"
	"path"

	"github.com/docopt/docopt-go"
	"github.com/kyma-incubator/hydroform/function/pkg/resources/types"
	"github.com/kyma-incubator/hydroform/function/pkg/workspace"

	log "github.com/sirupsen/logrus"
)

const (
	usage = `init description

Usage:
	init --runtime=<RUNTIME> [--dir=<DIR>] [options]

Options:
	--debug                 Enable verbose output.
	-h --help               Show this screen.
	--version               Show version.`

	version = "0.0.1"
)

type config struct {
	Name    string `docopt:"--name" json:"name"`
	Debug   bool   `docopt:"--debug" json:"debug"`
	Dir     string `docopt:"--dir"`
	Runtime string `docopt:"--runtime" json:"runtime"`
}

func newConfig() (*config, error) {
	arguments, err := docopt.ParseArgs(usage, nil, version)
	if err != nil {
		return nil, err
	}
	var cfg config
	if err = arguments.Bind(&cfg); err != nil {
		return nil, err
	}
	if cfg.Dir == "" {
		cfg.Dir, err = os.Getwd()
		if err != nil {
			return nil, err
		}
	}
	return &cfg, nil
}

func main() {
	// parse command arguments
	cfg, err := newConfig()
	if err != nil {
		log.Fatal(err)
	}

	if cfg.Debug {
		log.SetLevel(log.DebugLevel)
	}

	entry := log.WithField("runtime", cfg.Runtime)
	entry.Debug("initializing project")

	if cfg.Name == "" {
		cfg.Name = path.Base(cfg.Dir)
	}

	configuration := workspace.Cfg{
		Runtime:    types.Runtime(cfg.Runtime),
		Name:       cfg.Name,
		Namespace:  "default",
		SourcePath: cfg.Dir,
	}

	if err := workspace.Initialize(configuration, cfg.Dir); err != nil {
		entry.Fatal(err)
	}
	entry.Debug("workspace initialized")
}

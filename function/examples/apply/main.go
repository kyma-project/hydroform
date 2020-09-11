/*
* CODE GENERATED AUTOMATICALLY WITH devops/internal/config
 */

package main

import (
	"github.com/docopt/docopt-go"

	log "github.com/sirupsen/logrus"
)

const (
	usage = `apply description

Usage:
	apply [--dir=<DIR>] [options]

Options:
	--debug                 Enable verbose output.
	-h --help               Show this screen.
	--version               Show version.`

	version = "0.0.1"
)

type config struct {
	Name  string `docopt:"--name" json:"name"`
	Debug bool   `docopt:"--debug" json:"debug"`
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



}

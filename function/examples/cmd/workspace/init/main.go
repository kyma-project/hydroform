/*
* CODE GENERATED AUTOMATICALLY WITH devops/internal/config
 */

package main

import (
	"os"
	"path"

	"github.com/docopt/docopt-go"
	"github.com/kyma-project/hydroform/function/pkg/workspace"

	log "github.com/sirupsen/logrus"
)

const (
	usage = `init description

Usage:
	init --runtime=<RUNTIME> [ --url=<URL> ] [ --reference=<REFERENCE> ] [ --base-dir=<PATH> ] [ --dir=<DIR> ] [ options ]

Options:
	--debug                   Enable verbose output.
	-h --help                 Show this screen.
	--version                 Show version.`

	version = "0.0.1"
)

type config struct {
	Name      string `docopt:"--name" json:"name"`
	Debug     bool   `docopt:"--debug" json:"debug"`
	Dir       string `docopt:"--dir"`
	Runtime   string `docopt:"--runtime" json:"runtime"`
	URL       string `docopt:"--url" json:"url"`
	Reference string `docopt:"--reference" json:"reference"`
	BaseDir   string `docopt:"--base-dir" json:"baseDir"`
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

	entry := log.WithField("runtime", cfg.Runtime)
	entry.Debug("initializing project")

	if cfg.Name == "" {
		cfg.Name = path.Base(cfg.Dir)
	}

	outputPath, err := func() (string, error) {
		switch cfg.Dir {
		case "":
			return os.Getwd()
		default:
			return cfg.Dir, nil
		}
	}()

	if err != nil {
		log.Fatal(err)
	}

	var source workspace.Source
	if cfg.URL != "" {
		source = workspace.Source{
			Type: workspace.SourceTypeGit,
			SourceGit: workspace.SourceGit{
				URL:        cfg.URL,
				Repository: cfg.Name,
				Reference:  cfg.Reference,
				BaseDir:    cfg.BaseDir,
			},
		}
	} else {
		source = workspace.Source{
			Type: workspace.SourceTypeInline,
			SourceInline: workspace.SourceInline{
				SourcePath: outputPath,
			},
		}
	}

	configuration := workspace.Cfg{
		Name:      cfg.Name,
		Namespace: "default",
		Runtime:   cfg.Runtime,
		Source:    source,
	}

	if err := workspace.Initialize(configuration, outputPath); err != nil {
		entry.Fatal(err)
	}
	entry.Debug("workspace initialized")
}

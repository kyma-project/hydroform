package terraform

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/kyma-incubator/hydroform/types"
	"github.com/pkg/errors"

	"github.com/hashicorp/terraform/command"
	"github.com/hashicorp/terraform/states/statefile"
)

// Terraform is an Operator.
type Terraform struct {
	ops Options
}

// New creates a new Terraform operator with the given options
func New(ops ...Option) *Terraform {
	// silence the logs since terraform prints a lot of stuff
	log.SetOutput(ioutil.Discard)

	return &Terraform{
		ops: options(ops...),
	}
}

// Create creates a new cluster for a specific provider based on configuration details. It returns a ClusterInfo object with provider-related information, or an error if cluster provisioning failed.
func (t *Terraform) Create(p types.ProviderType, cfg map[string]interface{}) (*types.ClusterInfo, error) {
	// silence stdErr during terraform execution, plugins send debug and trace entries there
	stderr := os.Stderr
	os.Stderr, _ = os.Open(os.DevNull)
	defer func() { os.Stderr = stderr }()

	// init cluster files
	if !t.ops.Persistent {
		// remove all files if not persistent after running
		defer cleanup(t.ops.DataDir(), cfg["project"].(string), cfg["cluster_name"].(string), p)
	}
	if err := initClusterFiles(t.ops.DataDir(), p, cfg); err != nil {
		return nil, errors.Wrap(err, "Could not initialize cluster data")
	}

	clusterDir, err := clusterDir(t.ops.DataDir(), cfg["project"].(string), cfg["cluster_name"].(string), p)
	if err != nil {
		return nil, err
	}

	// INIT
	if p == types.Gardener {
		if err := initGardenerProvider(); err != nil {
			return nil, errors.Wrap(err, "could not initialize the gardener provider")
		}
	}
	i := &command.InitCommand{
		Meta: t.ops.Meta,
	}

	e := i.Run(t.initArgs(p, cfg, clusterDir))
	if e != 0 {
		return nil, t.checkUIErrors()
	}

	// APPLY
	a := &command.ApplyCommand{
		Meta: t.ops.Meta,
	}
	e = a.Run(t.applyArgs(p, cfg, clusterDir))
	if e != 0 {
		return nil, t.checkUIErrors()
	}

	return clusterInfoFromFile(t.ops.DataDir(), cfg["project"].(string), cfg["cluster_name"].(string), p)
}

// Status checks the current state of the cluster from the file
func (t *Terraform) Status(sf *statefile.File, p types.ProviderType, cfg map[string]interface{}) (*types.ClusterStatus, error) {
	cs := &types.ClusterStatus{
		Phase: types.Unknown,
	}
	var err error

	// if no state given, try the file system
	if sf == nil {
		sf, err = stateFromFile(t.ops.DataDir(), cfg["project"].(string), cfg["cluster_name"].(string), p)
		if err != nil {
			return cs, errors.Wrap(err, "no state provided, attempted to load from file")
		}
	}

	if sf.State.HasResources() {
		cs.Phase = types.Provisioned
	}

	return cs, nil
}

// Delete removes an existing cluster or returns an error if removing the cluster is not possible.
func (t *Terraform) Delete(sf *statefile.File, p types.ProviderType, cfg map[string]interface{}) error {
	// silence stdErr during terraform execution, plugins send debug and trace entries there
	stderr := os.Stderr
	os.Stderr, _ = os.Open(os.DevNull)
	defer func() { os.Stderr = stderr }()

	// init cluster files
	if !t.ops.Persistent {
		// remove all files if not persistent after running
		defer cleanup(t.ops.DataDir(), cfg["project"].(string), cfg["cluster_name"].(string), p)
	}
	if err := initClusterFiles(t.ops.DataDir(), p, cfg); err != nil {
		return errors.Wrap(err, "Could not initialize cluster data")
	}

	var err error
	// if no state given, check if it is already in the file system
	if sf == nil {
		_, err = stateFromFile(t.ops.DataDir(), cfg["project"].(string), cfg["cluster_name"].(string), p)
		if err != nil {
			return errors.Wrap(err, "no state provided, attempted to load from file")
		}
	} else {
		// otherwise save the state into a file so terraform can use it
		if err := stateToFile(sf, t.ops.DataDir(), cfg["project"].(string), cfg["cluster_name"].(string), p); err != nil {
			return errors.Wrap(err, "could not store state into file")
		}
	}

	clusterDir, err := clusterDir(t.ops.DataDir(), cfg["project"].(string), cfg["cluster_name"].(string), p)
	if err != nil {
		return err
	}

	// INIT
	if p == types.Gardener {
		if err := initGardenerProvider(); err != nil {
			return errors.Wrap(err, "could not initialize the gardener provider")
		}
	}
	i := &command.InitCommand{
		Meta: t.ops.Meta,
	}

	e := i.Run(t.initArgs(p, cfg, clusterDir))
	if e != 0 {
		return t.checkUIErrors()
	}

	// DESTROY
	a := &command.ApplyCommand{
		Meta:    t.ops.Meta,
		Destroy: true,
	}
	e = a.Run(t.applyArgs(p, cfg, clusterDir))
	if e != 0 {
		return t.checkUIErrors()
	}
	return nil
}

// initArgs generates the flag list for the terraform init command based on the operator configuration
func (t *Terraform) initArgs(p types.ProviderType, cfg map[string]interface{}, clusterDir string) []string {
	args := make([]string, 0)

	varsFile := filepath.Join(clusterDir, tfVarsFile)

	args = append(args, fmt.Sprintf("-var-file=%s", varsFile), clusterDir)

	return args
}

// applyArgs generates the flag list for the terraform apply command based on the operator configuration
func (t *Terraform) applyArgs(p types.ProviderType, cfg map[string]interface{}, clusterDir string) []string {
	args := make([]string, 0)

	stateFile := filepath.Join(clusterDir, tfStateFile)
	varsFile := filepath.Join(clusterDir, tfVarsFile)

	args = append(args,
		fmt.Sprintf("-state=%s", stateFile),
		fmt.Sprintf("-var-file=%s", varsFile),
		"-auto-approve",
		clusterDir)

	return args
}

func (t *Terraform) checkUIErrors() error {
	var errsum strings.Builder
	if h, ok := t.ops.Ui.(*HydroUI); ok {
		for _, e := range h.Errors() {
			if _, err := errsum.WriteString(e.Error()); err != nil {
				return errors.Wrap(err, "could not fetch errors from terraform")
			}
		}
	}

	if errsum.Len() != 0 {
		return errors.New(errsum.String())
	}

	return nil
}

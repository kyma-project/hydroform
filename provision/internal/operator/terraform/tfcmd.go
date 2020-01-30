package terraform

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	be_init "github.com/hashicorp/terraform/backend/init"
	"github.com/hashicorp/terraform/command"
	"github.com/kyma-incubator/hydroform/provision/types"
	hashiCli "github.com/mitchellh/cli"
	"github.com/pkg/errors"
)

// tfInit runs the 'terraform init' command with the specified options and config in the given working directory
func tfInit(ops Options, p types.ProviderType, cfg map[string]interface{}, dir string) error {
	// need to init all backends before we start
	be_init.Init(ops.Services)
	i := &command.InitCommand{
		Meta: ops.Meta,
	}

	if e := i.Run(initArgs(p, cfg, dir)); e != 0 {
		return checkUIErrors(ops.Ui)
	}
	return nil
}

// initArgs generates the flag list for the terraform init command based on the operator configuration
func initArgs(p types.ProviderType, cfg map[string]interface{}, clusterDir string) []string {
	args := make([]string, 0)

	// TODO remove this condition when fully migrated to modules
	if m := tfMod(p); m != "" {
		args = append(args, fmt.Sprintf("-from-module=%s", tfMod(p)))
	}
	args = append(args, clusterDir)

	return args
}

// tfMod returns the terraform module URL for the given provider or empty string if none avilable
func tfMod(p types.ProviderType) string {
	switch p {
	case types.Azure:
		return azureMod
	case types.AWS:
		return ""
	case types.GCP:
		return ""
	case types.Gardener:
		return ""
	default:
		return ""
	}
}

// tfApply runs a smart 'terraform apply' command with the specified options
// and config in the given working directory.
//
// The smart logic is as follows:
// - Apply is attempted regularly
// - If failed with error "already exists" it means that the cluster exists but
//   we do not have its state locally, so import the existing cluster and
//   refresh the local state.
// - if failed with error "not found" => probably state is corrupt => delete
//   the state and start over with apply.
func tfApply(ops Options, p types.ProviderType, cfg map[string]interface{}, dir string) error {
	a := &command.ApplyCommand{
		Meta: ops.Meta,
	}
	e := a.Run(applyArgs(p, cfg, dir))
	if e != 0 {
		errList := checkUIErrors(ops.Ui)

		// if cluster already exists import it and refresh the state
		if strings.Contains(strings.ToLower(errList.Error()), "already exists") {
			i := &command.ImportCommand{
				Meta: ops.Meta,
			}

			if e := i.Run(importArgs(p, cfg, dir)); e != 0 {
				return checkUIErrors(ops.Ui)
			}

			r := &command.RefreshCommand{
				Meta: ops.Meta,
			}

			if e := r.Run(refreshArgs(p, cfg, dir)); e != 0 {
				return checkUIErrors(ops.Ui)
			}
			return nil
		}

		// if cluster was not found, cluster got deeted on the remote or state is wrong, delete state and start over
		if strings.Contains(errList.Error(), "not found") {
			// delete the corrupt state file
			stateFile := filepath.Join(dir, tfStateFile)
			if err := os.Remove(stateFile); err != nil {
				return err
			}

			// try applying again
			return tfApply(ops, p, cfg, dir)
		}
		return errList
	}
	return nil
}

// tfDestroy runs the 'terraform destroy' command with the specified options and config in the given working directory
func tfDestroy(ops Options, p types.ProviderType, cfg map[string]interface{}, dir string) error {
	a := &command.ApplyCommand{
		Meta:    ops.Meta,
		Destroy: true,
	}
	if e := a.Run(applyArgs(p, cfg, dir)); e != 0 {
		return checkUIErrors(ops.Ui)
	}
	return nil
}

// applyArgs generates the flag list for the terraform apply command based on the operator configuration
func applyArgs(p types.ProviderType, cfg map[string]interface{}, clusterDir string) []string {
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

// importArgs generates the flag list for the terraform import command based on the operator configuration
func importArgs(p types.ProviderType, cfg map[string]interface{}, clusterDir string) []string {
	args := make([]string, 0)

	stateFile := filepath.Join(clusterDir, tfStateFile)
	varsFile := filepath.Join(clusterDir, tfVarsFile)

	args = append(args,
		fmt.Sprintf("-state=%s", stateFile),
		fmt.Sprintf("-state-out=%s", stateFile),
		fmt.Sprintf("-var-file=%s", varsFile),
		fmt.Sprintf("-config=%s", clusterDir),
		clusterResource(p), // cluster resource
		clusterID(p, cfg))  // cluster ID

	return args
}

// clusterResource returns the cluster resource type defined in the terraform module for the given provider.
func clusterResource(p types.ProviderType) string {
	switch p {
	case types.GCP:
		return "google_container_cluster.gke_cluster"
	case types.Azure:
		return "azurerm_kubernetes_cluster.azure_cluster"
	case types.Gardener:
		return "gardener_shoot.gardener_cluster"
	case types.AWS:
		return "not supported"
	}
	return ""
}

// clusterID generates a cluster ID based on the given config.
// Each provider has a different way of identifying clusters.
func clusterID(p types.ProviderType, cfg map[string]interface{}) string {
	switch p {
	case types.GCP:
		return fmt.Sprintf("%s/%s/%s", cfg["project"], cfg["location"], cfg["cluster_name"])
	case types.Gardener:
		return fmt.Sprintf("%s/%s", cfg["namespace"], cfg["cluster_name"])
	case types.AWS:
		return "not supported"
	}
	return ""
}

// refreshArgs generates the flag list for the terraform refresh command based on the operator configuration
func refreshArgs(p types.ProviderType, cfg map[string]interface{}, clusterDir string) []string {
	args := make([]string, 0)

	stateFile := filepath.Join(clusterDir, tfStateFile)
	varsFile := filepath.Join(clusterDir, tfVarsFile)

	args = append(args,
		fmt.Sprintf("-state=%s", stateFile),
		fmt.Sprintf("-var-file=%s", varsFile),
		clusterDir)

	return args
}

func checkUIErrors(ui hashiCli.Ui) error {
	var errsum strings.Builder
	if h, ok := ui.(*HydroUI); ok {
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

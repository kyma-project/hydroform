package terraform

import (
	"encoding/base64"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/hashicorp/terraform/command/cliconfig"
	"github.com/hashicorp/terraform/states/statefile"
	"github.com/kyma-incubator/hydroform/provision/types"
	"github.com/pkg/errors"
)

const (
	// file names for terraform
	tfStateFile          = "terraform.tfstate"
	tfModuleFile         = "terraform.tf"
	tfVarsFile           = "terraform.tfvars"
	awsClusterTemplate   = ``
	azureClusterTemplate = `
	variable "cluster_name"  			{}
	variable "project"       			{}
	variable "location" 				{}	
	variable "client_id" 	{}
	variable "client_secret" {}
	variable "machine_type"  			{}
	variable "kubernetes_version"   	{}
	variable "disk_size" 				{}
	variable "node_count" 				{}
	variable "create_timeout" 			{}
	variable "update_timeout" 			{}
	variable "delete_timeout" 			{}

	resource "azurerm_resource_group" "azure_cluster" {
		name     = "${var.project}"
		location = "${var.location}"
	}

	resource "azurerm_kubernetes_cluster" "azure_cluster" {
		name                = "${var.cluster_name}"
		location            = "${azurerm_resource_group.azure_cluster.location}"
		resource_group_name = "${azurerm_resource_group.azure_cluster.name}"
		dns_prefix          = "${var.cluster_name}"

		default_node_pool {
			name            = "default"
			node_count      = "${var.node_count}"
			vm_size         = "${var.machine_type}"
			os_disk_size_gb = "${var.disk_size}"
		}

		service_principal {
			client_id     = "${var.client_id}"
			client_secret = "${var.client_secret}"
		}

		role_based_access_control {
			enabled       = true
		}

		tags = {
			Environment = "Production"
		}
	}
	output "id" {
		value = "${azurerm_kubernetes_cluster.azure_cluster.id}"
	}

	output "kube_config" {
		value = "${azurerm_kubernetes_cluster.azure_cluster.kube_config_raw}"
	}

	output "client_key" {
		value = "${azurerm_kubernetes_cluster.azure_cluster.kube_config.0.client_key}"
	}

	output "client_certificate" {
		value = "${azurerm_kubernetes_cluster.azure_cluster.kube_config.0.client_certificate}"
	}

	output "cluster_ca_certificate" {
		value = "${azurerm_kubernetes_cluster.azure_cluster.kube_config.0.cluster_ca_certificate}"
	}

	output "endpoint" {
		value = "${azurerm_kubernetes_cluster.azure_cluster.kube_config.0.host}"
	}
`
	gcpClusterTemplate = `
  variable "node_count"    		{}
  variable "cluster_name"  		{}
  variable "credentials_file_path" 	{}
  variable "project"       		{}
  variable "location"      		{}
  variable "machine_type"  		{}
  variable "kubernetes_version"   	{}
  variable "disk_size" 			{}
  variable "create_timeout" 	{}
  variable "update_timeout" 	{}
  variable "delete_timeout" 	{}

  provider "google" {
    	credentials   = "${file("${var.credentials_file_path}")}"
		project       = "${var.project}"
  }

  resource "google_container_cluster" "gke_cluster" {
    	name               = "${var.cluster_name}"
    	location 	   = "${var.location}"
    	initial_node_count = "${var.node_count}"
    	min_master_version = "${var.kubernetes_version}"
    	node_version       = "${var.kubernetes_version}"
    
    node_config {
      	machine_type = "${var.machine_type}"
		disk_size_gb = "${var.disk_size}"
    }

	timeouts {
		create = "${var.create_timeout}"
		update = "${var.update_timeout}"
		delete = "${var.delete_timeout}"
	}

    maintenance_policy {
      	daily_maintenance_window {
        	start_time = "03:00"
      		}
    	}
  }

  output "endpoint" {
    value = "${google_container_cluster.gke_cluster.endpoint}"
  }

  output "cluster_ca_certificate" {
    value = "${google_container_cluster.gke_cluster.master_auth.0.cluster_ca_certificate}"
  }
`

	gardenerClusterTemplate = `
variable "target_provider"			{}
variable "target_profile"			{}
variable "target_secret"			{}
variable "node_count"    			{}
variable "cluster_name"  			{}
variable "credentials_file_path" 	{}
variable "project"					{}
variable "namespace"       			{}
variable "location"      			{}
{{ if not (eq (index . "target_provider") "azure") }}
variable "zone"      				{}
{{ end }}
variable "workercidr"      			{}
{{ if eq (index . "target_provider") "azure" }}
variable "vnetcidr"				{}
{{ end }}
{{ if eq (index . "target_provider") "aws" }}
variable "vpccidr" 					{}
variable "publicscidr" 				{}
variable "internalscidr" 			{}
{{ end }}
variable "machine_type"  			{}
variable "kubernetes_version"   	{}
variable "disk_size" 				{}
variable "disk_type" 				{}
variable "autoscaler_min" 			{}
variable "autoscaler_max" 			{}
variable "max_surge" 				{}
variable "max_unavailable" 			{}
variable "create_timeout" 			{}
variable "update_timeout" 			{}
variable "delete_timeout" 			{}

provider "gardener" {
	kube_file          = "${file("${var.credentials_file_path}")}"
}

resource "gardener_shoot" "gardener_cluster" {
	metadata {
	  name      = "${var.cluster_name}"
	  namespace = "${var.namespace}"
  
	}

	timeouts {
		create = "${var.create_timeout}"
		update = "${var.update_timeout}"
		delete = "${var.delete_timeout}"
	}

	spec {
	  cloud {
		profile = "${var.target_profile}"
		region  = "${var.location}"
		secret_binding_ref {
		  name = "${var.target_secret}"
		}

		{{ if eq (index . "target_provider") "gcp" }}
		gcp {  
          networks {
			workers = ["${var.workercidr}"]
		  }
		{{ end }}

		{{ if eq (index . "target_provider") "azure" }}
		azure {  
          networks {
			vnet {
				cidr = "${var.vnetcidr}"
			}
			workers = "${var.workercidr}"
		  }
		{{ end }}

		{{ if eq (index . "target_provider") "aws" }}
		aws {  
          networks {
			workers       = ["${var.workercidr}"]
			public		  = ["${var.publicscidr}"]
			internal	  = ["${var.internalscidr}"]
			vpc	{
				cidr = "${var.vpccidr}"
			}
		  }
		{{ end }}

		  {{range (seq (index . "node_count"))}}
		  worker {
			  name            = "cpu-worker-{{.}}"
			  machine_type    = "${var.machine_type}"
			  auto_scaler_min = "${var.autoscaler_min}"
			  auto_scaler_max = "${var.autoscaler_max}"
			  max_surge       = "${var.max_surge}"
			  max_unavailable = "${var.max_unavailable}"
			  volume_size     = "${var.disk_size}Gi"
			  volume_type     = "${var.disk_type}"
		  }
		  {{end}}
          {{ if not (eq (index . "target_provider") "azure") }}
		  zones = ["${var.zone}"]
          {{ end }}
		}
	  }
  
	  kubernetes {
		allow_privileged_containers = true
		version = "${var.kubernetes_version}"
	  }
	}
  }
`

	kindClusterTemplate = `
variable "project"					{}
variable "cluster_name"				{}
variable "node_image"				{}
variable "create_timeout" 			{}
variable "update_timeout" 			{}
variable "delete_timeout" 			{}

provider "kind" {
}
resource "kind" "kind-cluster" {
	name       = "${var.cluster_name}"
	node_image = "${var.node_image}"

	timeouts {
		create = "${var.create_timeout}"
		update = "${var.update_timeout}"
		delete = "${var.delete_timeout}"
	}
}
`
)

// initClusterFiles initializes all necessary files for a cluster in the given data directory
func initClusterFiles(dataDir string, p types.ProviderType, cfg map[string]interface{}) error {
	dir, err := clusterDir(dataDir, cfg["project"].(string), cfg["cluster_name"].(string), p)
	if err != nil {
		return err
	}

	// create module file
	var data []byte
	switch p {
	case types.GCP:
		data = []byte(gcpClusterTemplate)
	case types.Gardener:
		t, err := expandGardenerClusterTemplate(cfg)
		if err != nil {
			return err
		}
		data = []byte(t)
	case types.Azure:
		data = []byte(azureClusterTemplate)
	case types.AWS:
		data = []byte(awsClusterTemplate)
	case types.Kind:
		data = []byte(kindClusterTemplate)
	}
	if err := ioutil.WriteFile(filepath.Join(dir, tfModuleFile), data, 0700); err != nil {
		return err
	}

	// create vars file
	var vars strings.Builder
	for k, v := range cfg {
		switch t := v.(type) {
		case int:
			if _, err := vars.WriteString(fmt.Sprintf("%s = \"%d\"\n", k, t)); err != nil {
				return err
			}
		case string:
			if _, err := vars.WriteString(fmt.Sprintf("%s = \"%s\"\n", k, t)); err != nil {
				return err
			}
		case time.Duration:
			if _, err := vars.WriteString(fmt.Sprintf("%s = \"%s\"\n", k, t.String())); err != nil {
				return err
			}
		}

	}
	if err := ioutil.WriteFile(filepath.Join(dir, tfVarsFile), []byte(vars.String()), 0700); err != nil {
		return err
	}

	return nil
}

// stateFromFile loads the terraform state file for the given cluster
func stateFromFile(dataDir, project, cluster string, p types.ProviderType) (*statefile.File, error) {
	dir, err := clusterDir(dataDir, project, cluster, p)
	if err != nil {
		return nil, err
	}

	stateFilePath := filepath.Join(dir, tfStateFile)
	f, err := os.Open(stateFilePath)
	if err != nil {
		return nil, err
	}

	st, err := statefile.Read(f)
	if err != nil {
		return nil, err
	}
	return st, nil
}

// stateToFile saves the terraform state into its corresponding file
func stateToFile(state *statefile.File, dataDir, project, cluster string, p types.ProviderType) error {
	dir, err := clusterDir(dataDir, project, cluster, p)
	if err != nil {
		return err
	}

	stateFilePath := filepath.Join(dir, tfStateFile)
	f, err := os.Create(stateFilePath)
	if err != nil {
		return err
	}
	return statefile.Write(state, f)
}

func clusterInfoFromFile(dataDir, project, cluster string, p types.ProviderType) (*types.ClusterInfo, error) {
	sf, err := stateFromFile(dataDir, project, cluster, p)
	if err != nil {
		return nil, err
	}

	var certificateData []byte
	var endpoint string

	if len(sf.State.Modules) > 0 {
		if val, ok := sf.State.Modules[""].OutputValues["cluster_ca_certificate"]; ok {
			certificateData, err = base64.StdEncoding.DecodeString(val.Value.AsString())
			if err != nil {
				return &types.ClusterInfo{
					InternalState: &types.InternalState{TerraformState: sf},
					Status:        &types.ClusterStatus{Phase: types.Errored},
				}, errors.Wrap(err, "Unable to decode certificate data")
			}
		}
		if val, ok := sf.State.Modules[""].OutputValues["endpoint"]; ok {
			endpoint = val.Value.AsString()
		}
	}

	return &types.ClusterInfo{
		Endpoint:                 endpoint,
		CertificateAuthorityData: certificateData,
		InternalState:            &types.InternalState{TerraformState: sf},
		Status:                   &types.ClusterStatus{Phase: types.Provisioned},
	}, nil
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

// clusterDir either returns or creates the directory for a given cluster inside the given data directory.
// All state and configuration files needed by the operator will be stored in this directory.
func clusterDir(dataDir, project, cluster string, p types.ProviderType) (string, error) {
	clDir, err := filepath.Abs(filepath.Join(dataDir, "clusters", string(p), project, cluster))
	if err != nil {
		return "", err
	}

	if _, err := os.Stat(clDir); os.IsNotExist(err) {
		err = os.MkdirAll(clDir, 0700)
		if err != nil {
			return "", err
		}
	}

	if runtime.GOOS == "windows" {
		clDir = `\\?\` + clDir
	}

	return clDir, nil
}

func expandGardenerClusterTemplate(cfg map[string]interface{}) (string, error) {
	funcs := template.FuncMap{
		"seq": func(n int) []int {
			r := make([]int, n)

			for i := 0; i < n; i++ {
				r[i] = i
			}
			return r
		},
	}

	t := template.Must(template.New("gardenerCluster").Funcs(funcs).Parse(gardenerClusterTemplate))
	s := &strings.Builder{}
	if err := t.Execute(s, cfg); err != nil {
		return "", err
	}
	return s.String(), nil
}

// cleanup removes all terraform generated files for a given cluster
func cleanup(dataDir, project, cluster string, p types.ProviderType) error {
	d, err := clusterDir(dataDir, project, cluster, p)
	if err != nil {
		return err
	}

	return os.RemoveAll(d)
}

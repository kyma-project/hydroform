package terraform

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
	"time"

	"github.com/hashicorp/terraform/command/cliconfig"
	"github.com/hashicorp/terraform/states/statefile"
	"github.com/kyma-incubator/hydroform/provision/types"
	"github.com/pkg/errors"
)

const (
	// file names for terraform
	tfStateFile  = "terraform.tfstate"
	tfModuleFile = "terraform.tf"
	tfVarsFile   = "terraform.tfvars"
	// TODO release modules and do not use master as ref when stable
	azureMod = "git::https://github.com/kyma-incubator/terraform-modules//azurerm_kubernetes_cluster?ref=v0.0.3"

	// TODO remove hardcoded TF templates once modules work
	awsClusterTemplate = ``
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
    	location 	       = "${var.location}"
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
variable "vnetcidr"					{
	default = ""
}
variable "networking_nodes"			{
	default = ""
}
variable "networking_pods"			{
	default = ""
}
variable "networking_services"		{
	default = ""
}
variable "networking_type"			{}
variable "zones"      				{}
variable "workercidr"      			{
	default = ""
}
{{ if eq (index .Cfg "target_provider") "gcp" }}
variable "gcp_control_plane_zone"		{}
{{ end }}
{{ if eq (index .Cfg "target_provider") "azure" }}
variable "zoned"      				{}
variable "service_endpoints"		{}
{{ end }}
variable "machine_type"  			{}
variable "kubernetes_version"   	{}
variable "disk_size" 				{}
variable "disk_type" 				{}
variable "create_timeout" 			{}
variable "update_timeout" 			{}
variable "delete_timeout" 			{}
variable "worker_max_surge" 		{}
variable "worker_max_unavailable"	{}
variable "worker_maximum"			{}
variable "worker_minimum"			{}
variable "machine_image_name"		{}
variable "machine_image_version"	{}


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
       cloud_profile_name = "${var.target_profile}"
       region  = "${var.location}"
	   secret_binding_name = "${var.target_secret}"
       networking {
         nodes = "${var.networking_nodes}"
         pods = "${var.networking_pods}"
         services = "${var.networking_services}"
	     type = "${var.networking_type}"
       }
      maintenance {
        auto_update {
          kubernetes_version = "true"
          machine_image_version = "true"
        }
		time_window {
		  begin = "030000+0000"
          end = "040000+0000"
        }
      }
      provider {
        type = "${var.target_provider}"
		{{ if eq (index .Cfg "target_provider") "gcp" }}
			control_plane_config {
				gcp {
					zone = "${var.gcp_control_plane_zone}"
 				}
			}
		{{ end }}
        infrastructure_config {
           {{ if eq (index .Cfg "target_provider") "azure" }}
			  azure {
				zoned = "${var.zoned}"
                networks {
                  vnet {
					cidr = "${var.vnetcidr}"
                  }
				  workers = "${var.workercidr}"
                  service_endpoints = "${var.service_endpoints}"
                }
              }
           {{ end }}
		   {{ if eq (index .Cfg "target_provider") "gcp" }}
				gcp {
					networks {
						workers = "${var.workercidr}"
					}
				}
           {{ end }}
		   {{ if eq (index .Cfg "target_provider") "aws" }}
				aws {
					networks {
						vpc {
							cidr = "${var.vnetcidr}"
						}
						{{range $i, $z:= (index .Cfg "zones")}}
						zones {
							name = "{{ $z }}"
							workers = "{{index $.WorkerNets $i}}"
							public = "{{index $.PublicNets $i}}"
							internal = "{{index $.InternalNets $i}}"
						}
						{{end}}
					}
				}
		   {{ end }}
        }
        worker {
         name = "cpu-worker"
		 zones = "${var.zones}"
         max_surge = "${var.worker_max_surge}"
		 max_unavailable = "${var.worker_max_unavailable}"
		 maximum = "${var.worker_maximum}"
         minimum = "${var.worker_minimum}"
		 volume {
           size = "${var.disk_size}Gi"
  		   type = "${var.disk_type}"
         }
		 machine {
		   image {
 			 name = "${var.machine_image_name}"
			 version = "${var.machine_image_version}"
		   }
           type = "${var.machine_type}"
		 }
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
variable "project"				{}
variable "cluster_name"			{}
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

	// create module file for providers that are not using modules
	// TODO delete this when all providers have downloadable modules
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
		break
	case types.AWS:
		data = []byte(awsClusterTemplate)
	case types.Kind:
		data = []byte(kindClusterTemplate)
	}

	if len(data) > 0 {
		if err := ioutil.WriteFile(filepath.Join(dir, tfModuleFile), data, 0700); err != nil {
			return err
		}
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
		case []string:
			var a []string
			for _, v := range t {
				x := fmt.Sprintf("\"%s\"", v)
				a = append(a, x)
			}
			b := strings.Join(a, ",")
			if _, err := vars.WriteString(fmt.Sprintf("%s = [%s]\n", k, b)); err != nil {
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
	tmpCfg := struct {
		WorkerNets   []string
		PublicNets   []string
		InternalNets []string
		Cfg          map[string]interface{}
	}{}

	tmpCfg.Cfg = cfg

	funcs := template.FuncMap{
		"seq": func(n int) []int {
			r := make([]int, n)

			for i := 0; i < n; i++ {
				r[i] = i
			}
			return r
		},
	}

	if cfg["target_provider"] == string(types.AWS) {
		// subnets for zones
		var err error
		tmpCfg.WorkerNets, tmpCfg.PublicNets, tmpCfg.InternalNets, err = generateGardenerAWSSubnets(cfg["vnetcidr"].(string), len(cfg["zones"].([]string)))
		if err != nil {
			return "", errors.Wrap(err, "Error generating subnets for AWS zones")
		}
	}

	t := template.Must(template.New("gardenerCluster").Funcs(funcs).Parse(gardenerClusterTemplate))
	s := &strings.Builder{}
	if err := t.Execute(s, tmpCfg); err != nil {
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

// isEmptyDir returns true if the given path contains no files or subdirectories, false otherwise.
func isEmptyDir(path string) (bool, error) {
	entries, err := ioutil.ReadDir(path)
	if err != nil {
		return false, err
	}
	return len(entries) == 0, nil
}

func generateGardenerAWSSubnets(baseNet string, zoneCount int) (workerNets, publicNets, internalNets []string, err error) {
	_, cidr, err := net.ParseCIDR(baseNet)
	if err != nil {
		return
	}
	if zoneCount < 1 {
		err = errors.New("There must be at least 1 zone defined.")
	}

	// each zone gets its own subnet
	const subnetSize = 64
	for i := 0; i < zoneCount; i++ {
		// workers subnet
		cidr.IP[2] = byte(i * subnetSize)
		cidr.Mask = net.CIDRMask(19, 8*net.IPv4len)
		workerNets = append(workerNets, cidr.String())

		// public and internal share the subnet and divide it further
		cidr.Mask = net.CIDRMask(20, 8*net.IPv4len)
		cidr.IP[2] = byte(i*subnetSize + subnetSize/2) // first half of the subnet after worker
		publicNets = append(publicNets, cidr.String())
		cidr.IP[2] = byte(int(cidr.IP[2]) + subnetSize/4) // second half of the subnet after worker
		internalNets = append(internalNets, cidr.String())
	}
	return
}

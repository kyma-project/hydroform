package operator

import (
	"strings"

	"github.com/pkg/errors"

	"github.com/hashicorp/terraform/terraform"
	tApi "github.com/kyma-incubator/hydroform/api/terraform"
	"github.com/terraform-providers/terraform-provider-google/google"
)

const awsClusterTemplate string = ``
const azureClusterTemplate string = ``
const googleClusterTemplate string = `
  variable "node_count"    		{}
  variable "cluster_name"  		{}
  variable "credentials_file_path" 	{}
  variable "project"       		{}
  variable "location"      		{}
  variable "machine_type"  		{}
  variable "kubernetes_version"   	{}
  variable "disk_size" 			{}

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

    maintenance_policy {
      	daily_maintenance_window {
        	start_time = "03:00"
      		}
    	}
  }
`

type Terraform struct {
}

func (t *Terraform) Create(provider string,
	configuration map[string]interface{}) error {

	var resourceProvider terraform.ResourceProvider
	var clusterTemplate string
	pr := strings.ToLower(provider)
	switch pr {
	case "google":
		resourceProvider = google.Provider()
		clusterTemplate = googleClusterTemplate
	case "aws":
		//resourceProvider = aws.Provider()
		//clusterTemplate = awsClusterTemplate
		return errors.New("aws not supported yet")
	case "azure":
		//resourceProvider = azure.Provider()
		//clusterTemplate = azureClusterTemplate
		return errors.New("azure not supported yet")
	default:
		return errors.New("unknown provider")
	}

	platform := tApi.NewPlatform(clusterTemplate)
	platform.AddProvider(pr, resourceProvider)
	for k, v := range configuration {
		platform.Var(k, v)
	}
	err := platform.Apply(false)
	return errors.Wrap(err, "unable to provision cluster")
}

func (t *Terraform) Delete(provider string, configuration map[string]interface{}) error {
	return nil
}

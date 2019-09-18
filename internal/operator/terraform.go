package operator

import (
	"encoding/base64"
	"fmt"
	"log"

	"github.com/kyma-incubator/hydroform/types"
	"github.com/pkg/errors"

	"github.com/hashicorp/terraform/terraform"
	terraformClient "github.com/kyma-incubator/hydroform/internal/terraform"
	"github.com/terraform-providers/terraform-provider-google/google"
)

const awsClusterTemplate string = ``
const azureClusterTemplate string = ``
const gcpClusterTemplate string = `
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

  output "endpoint" {
    value = "${google_container_cluster.gke_cluster.endpoint}"
  }

  output "cluster_ca_certificate" {
    value = "${google_container_cluster.gke_cluster.master_auth.0.cluster_ca_certificate}"
  }
`

type Terraform struct {
}

func (t *Terraform) Create(providerType types.ProviderType, configuration map[string]interface{}) (*types.ClusterInfo, error) {
	platform, err := t.newPlatform(providerType, configuration)
	if err != nil {
		return nil, err
	}

	state, err := platform.Apply(terraformClient.NewState(), false)
	if err != nil {
		return &types.ClusterInfo{
			InternalState: &types.InternalState{TerraformState: state},
			Status:        &types.ClusterStatus{Phase: types.Errored},
		}, errors.Wrap(err, "unable to provision cluster")
	}

	var certificateData []byte
	var endpoint string
	if len(state.Modules) > 0 {
		if val, ok := state.Modules[0].Outputs["cluster_ca_certificate"]; ok {
			certificateData, err = base64.StdEncoding.DecodeString(fmt.Sprintf("%v", val.Value))
			if err != nil {
				log.Printf("[ERROR] unable to decode certificate data: %v", err)
			}
		}
		if val, ok := state.Modules[0].Outputs["endpoint"]; ok {
			endpoint = fmt.Sprintf("%v", val.Value)
		}
	}

	return &types.ClusterInfo{
		Endpoint:                 endpoint,
		CertificateAuthorityData: certificateData,
		InternalState:            &types.InternalState{TerraformState: state},
		Status:                   &types.ClusterStatus{Phase: types.Provisioned},
	}, nil
}

func (t *Terraform) Delete(state *types.InternalState, providerType types.ProviderType, configuration map[string]interface{}) error {
	platform, err := t.newPlatform(providerType, configuration)
	if err != nil {
		return err
	}

	_, err = platform.Apply(state.TerraformState, true)
	return errors.Wrap(err, "unable to deprovision cluster")
}

func (t *Terraform) newPlatform(providerType types.ProviderType, configuration map[string]interface{}) (*terraformClient.Platform, error) {
	var resourceProvider terraform.ResourceProvider
	var clusterTemplate string
	//providerName must match the name of the provider on the templates
	var providerName string

	switch providerType {
	case types.GCP:
		resourceProvider = google.Provider()
		clusterTemplate = gcpClusterTemplate
		providerName = "google"
	case types.AWS:
		//resourceProvider = aws.Provider()
		//clusterTemplate = awsClusterTemplate
		//providerName = "aws"
		return nil, errors.New("aws not supported yet")
	case types.Azure:
		//resourceProvider = azure.Provider()
		//clusterTemplate = azureClusterTemplate
		//providerName = "azure"
		return nil, errors.New("azure not supported yet")
	default:
		return nil, errors.New("unknown provider")
	}

	platform := terraformClient.NewPlatform(clusterTemplate)
	platform.AddProvider(providerName, resourceProvider)
	for k, v := range configuration {
		platform.Var(k, v)
	}

	return platform, nil
}

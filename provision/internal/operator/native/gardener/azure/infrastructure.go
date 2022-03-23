package azure

import (
	"encoding/json"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const infrastructureConfigKind = "InfrastructureConfig"

func InfraConfig(cfg map[string]interface{}) (*runtime.RawExtension, error) {
	infra := InfrastructureConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       infrastructureConfigKind,
			APIVersion: azureAPIVersion,
		},
		Networks: Networks{},
	}
	if zones, ok := cfg["zones"].([]string); ok && len(zones) > 0 {
		infra.Zoned = true
	}
	if v, ok := cfg["workercidr"].(string); ok && len(v) > 0 {
		infra.Networks.Workers = v
	}
	if v, ok := cfg["vnetcidr"].(string); ok && len(v) > 0 {
		infra.Networks.VNet = VNet{
			CIDR: &v,
		}
	}

	data, err := json.Marshal(infra)

	return &runtime.RawExtension{
		Raw: data,
	}, err
}

// InfrastructureConfig infrastructure configuration resource
type InfrastructureConfig struct {
	metav1.TypeMeta
	// ResourceGroup is azure resource group
	ResourceGroup *ResourceGroup `json:"resourceGroup,omitempty"`
	// Networks is the network configuration (VNets, subnets, etc.)
	Networks Networks `json:"networks"`
	// Zoned indicates whether the cluster uses zones
	Zoned bool `json:"zoned"`
}

// ResourceGroup is azure resource group
type ResourceGroup struct {
	// Name is the name of the resource group
	Name string `json:"name"`
}

// Networks holds information about the Kubernetes and infrastructure networks.
type Networks struct {
	// VNet indicates whether to use an existing VNet or create a new one.
	VNet VNet `json:"vnet"`
	// Workers is the worker subnet range to create (used for the VMs).
	Workers string `json:"workers"`
	// ServiceEndpoints is a list of Azure ServiceEndpoints which should be associated with the worker subnet.
	ServiceEndpoints []string `json:"serviceEndpoints,omitempty"`
}

// VNet contains information about the VNet and some related resources.
type VNet struct {
	// Name is the VNet name.
	Name *string `json:"name,omitempty"`
	// ResourceGroup is the resource group where the existing vNet belongs to.
	ResourceGroup *string `json:"resourceGroup,omitempty"`
	// CIDR is the VNet CIDR
	CIDR *string `json:"cidr,omitempty"`
}

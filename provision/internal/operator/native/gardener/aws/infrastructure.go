package aws

import (
	"encoding/json"
	"errors"
	"net"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const infrastructureConfigKind = "InfrastructureConfig"

func InfraConfig(cfg map[string]interface{}) (*runtime.RawExtension, error) {
	infra := InfrastructureConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       infrastructureConfigKind,
			APIVersion: awsAPIVersion,
		},
		Networks: Networks{},
	}

	if v, ok := cfg["vnetcidr"].(string); ok && len(v) > 0 {
		infra.Networks.VPC = VPC{
			CIDR: &v,
		}
	} else {
		return nil, errors.New("Could not generate AWS virtual network, vnetcidr not provided")
	}

	if zones, ok := cfg["zones"].([]string); ok && len(zones) > 0 {
		workerNets, publicNets, internalNets, err := generateGardenerAWSSubnets(cfg["vnetcidr"].(string), len(zones))
		if err != nil {
			return nil, err
		}

		for i := range zones {
			z := Zone{
				Name:     zones[i],
				Internal: internalNets[i],
				Public:   publicNets[i],
				Workers:  workerNets[i],
			}
			infra.Networks.Zones = append(infra.Networks.Zones, z)
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

	// EnableECRAccess specifies whether the IAM role policy for the worker nodes shall contain
	// permissions to access the ECR.
	// default: true
	EnableECRAccess *bool `json:"enableECRAccess,omitempty"`

	// Networks is the AWS specific network configuration (VPC, subnets, etc.)
	Networks Networks `json:"networks"`
}

// Networks holds information about the Kubernetes and infrastructure networks.
type Networks struct {
	// VPC indicates whether to use an existing VPC or create a new one.
	VPC VPC `json:"vpc"`
	// Zones belonging to the same region
	Zones []Zone `json:"zones"`
}

// Zone describes the properties of a zone
type Zone struct {
	// Name is the name for this zone.
	Name string `json:"name"`
	// Internal is the private subnet range to create (used for internal load balancers).
	Internal string `json:"internal"`
	// Public is the public subnet range to create (used for bastion and load balancers).
	Public string `json:"public"`
	// Workers isis the workers subnet range to create  (used for the VMs).
	Workers string `json:"workers"`
}

// VPC contains information about the AWS VPC and some related resources.
type VPC struct {
	// ID is the VPC id.
	ID *string `json:"id,omitempty"`
	// CIDR is the VPC CIDR.
	CIDR *string `json:"cidr,omitempty"`
}

func generateGardenerAWSSubnets(baseNet string, zoneCount int) (workerNets, publicNets, internalNets []string, err error) {
	_, cidr, err := net.ParseCIDR(baseNet)
	if err != nil {
		return
	}
	if zoneCount < 1 {
		err = errors.New("there must be at least 1 zone defined")
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

package gcp

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
			APIVersion: gcpAPIVersion,
		},
		Networks: Networks{},
	}
	if v, ok := cfg["workercidr"].(string); ok && len(v) > 0 {
		infra.Networks.Worker = v
		infra.Networks.Workers = &v
	}

	data, err := json.Marshal(infra)

	return &runtime.RawExtension{
		Raw: data,
	}, err
}

// InfrastructureConfig infrastructure configuration resource
type InfrastructureConfig struct {
	metav1.TypeMeta

	// Networks is the network configuration (VPC, subnets, etc.)
	Networks Networks `json:"networks"`
}

// Networks holds information about the Kubernetes and infrastructure networks.
type Networks struct {
	// VPC indicates whether to use an existing VPC or create a new one.
	VPC *VPC `json:"vpc,omitempty"`
	// CloudNAT contains configation about the the CloudNAT resource
	CloudNAT *CloudNAT `json:"cloudNat,omitempty"`
	// Internal is a private subnet (used for internal load balancers).
	Internal *string `json:"internal,omitempty"`
	// Worker is the worker subnet range to create (used for the VMs).
	// Deprecated - use `workers` instead.
	Worker string `json:"worker"`
	// Workers is the worker subnet range to create (used for the VMs).
	Workers *string `json:"workers,omitempty"`
	// FlowLogs contains the flow log configuration for the subnet.
	FlowLogs *FlowLogs `json:"flowLogs,omitempty"`
}

// VPC contains information about the VPC and some related resources.
type VPC struct {
	// Name is the VPC name.
	Name string `json:"name"`
	// CloudRouter indicates whether to use an existing CloudRouter or create a new one
	CloudRouter *CloudRouter `json:"cloudRouter,omitempty"`
}

// CloudRouter contains information about the the CloudRouter configuration
type CloudRouter struct {
	// Name is the CloudRouter name.
	Name string `json:"name"`
}

// CloudNAT contains information about the the CloudNAT configuration
type CloudNAT struct {
	// MinPortsPerVM is the minimum number of ports allocated to a VM in the NAT config.
	// The default value is 2048 ports.
	MinPortsPerVM *int32 `json:"minPortsPerVM,omitempty"`
}

// FlowLogs contains the configuration options for the vpc flow logs.
type FlowLogs struct {
	// AggregationInterval for collecting flow logs.
	AggregationInterval *string `json:"aggregationInterval,omitempty"`
	// FlowSampling sets the sampling rate of VPC flow logs within the subnetwork where 1.0 means all collected logs are reported and 0.0 means no logs are reported.
	FlowSampling *float32 `json:"flowSampling,omitempty"`
	// Metadata configures whether metadata fields should be added to the reported VPC flow logs.
	Metadata *string `json:"metadata,omitempty"`
}

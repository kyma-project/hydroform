package types

import "github.com/kyma-incubator/hydroform/internal/terraform"

type Cluster struct {
	Name              string
	KubernetesVersion string
	CPU               string
	DiskSizeGB        int
	NodeCount         int
	MachineType       string
	Location          string
	ClusterInfo       *ClusterInfo
}

type ClusterInfo struct {
	Endpoint                 string
	CertificateAuthorityData []byte
	InternalState            *InternalState
	Status                   *ClusterStatus
}

type ClusterStatus struct {
	Phase Phase
}

type Phase string

const (
	Pending      Phase = "Pending"
	Provisioning Phase = "Provisioning"
	Provisioned  Phase = "Provisioned"
	Errored      Phase = "Errored"
	Stopping     Phase = "Stopping"
	Unknown      Phase = "Unknown"
)

type InternalState struct {
	TerraformState *terraform.State
}

package types

import "github.com/kyma-incubator/hydroform/internal/terraform"

// Cluster holds the detailed information for the type of the requested cluster.
type Cluster struct {
	Name              string       `json:"name"`
	KubernetesVersion string       `json:"kubernetesVersion"`
	CPU               string       `json:"cpu"`
	DiskSizeGB        int          `json:"diskSizeGB"`
	NodeCount         int          `json:"nodeCount"`
	MachineType       string       `json:"machineType"`
	Location          string       `json:"location"`
	ClusterInfo       *ClusterInfo `json:"clusterInfo"`
}

// ClusterInfo holds the resulting information after a attempt has been made to provision a cluster.
type ClusterInfo struct {
	Endpoint                 string         `json:"endpoint"`
	CertificateAuthorityData []byte         `json:"certificateAuthorityData"`
	InternalState            *InternalState `json:"internalState"`
	Status                   *ClusterStatus `json:"status"`
}

// ClusterStatus holds the fields to demonstrate the status of the cluster.
type ClusterStatus struct {
	Phase Phase `json:"phase"`
}

// Phase points out the current status of the cluster.
type Phase string

const (
	// Pending is a possible value for Phase, indicating that some work is actively being done on the cluster,
	// such as upgrading the master or node software
	Pending Phase = "Pending"
	// Provisioning is a possible value for Phase, indicating the cluster is being created.
	Provisioning Phase = "Provisioning"
	// Provisioned is a possible value for Phase, indicating the cluster has been created and is fully usable.
	Provisioned Phase = "Provisioned"
	// Errored is a possible value for Phase, indicating the cluster may be unusable.
	Errored Phase = "Errored"
	// Stopping is a possible value for Phase, indicating the cluster is being deleted.
	Stopping Phase = "Stopping"
	// Unknown is a possible value for Phase, indicating the status is not known.
	Unknown Phase = "Unknown"
)

// InternalState holds the state information for the internal operator in use.
type InternalState struct {
	TerraformState *terraform.State
}

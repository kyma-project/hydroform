package types

import "github.com/hashicorp/terraform/states"

// Cluster contains detailed cluster specification and properties.
type Cluster struct {
	// Name specifies the unique name used to identify the cluster.
	Name string `json:"name"`
	// KubernetesVersion specifies the Kubernetes version used.
	KubernetesVersion string `json:"kubernetesVersion"`
	// CPU specifies the number of CPUs available in the cluster.
	CPU int `json:"cpu"`
	// DiskSizeGB indicates the disk size available in the cluster.
	DiskSizeGB int `json:"diskSizeGB"`
	// NodeCount specifies the number of nodes available in the cluster.
	NodeCount int `json:"nodeCount"`
	// MachineType specifies the hardware cluster is provisioned on.
	MachineType string `json:"machineType"`
	// Location specifies the location of the actual cluster.
	Location    string       `json:"location"`
	ClusterInfo *ClusterInfo `json:"clusterInfo"`
}

// ClusterInfo contains the actual provider-related cluster details retrieved after the cluster was provisioned.
type ClusterInfo struct {
	// Endpoint specifies the URL at which you can reach the cluster.
	Endpoint string `json:"endpoint"`
	// CertificateAuthorityData contains certificates required to access the cluster.
	CertificateAuthorityData []byte `json:"certificateAuthorityData"`
	// InternalState contains the Hydroform-specific information used to manage the cluster.
	InternalState *InternalState `json:"internalState"`
	Status        *ClusterStatus `json:"status"`
}

// ClusterStatus contains possible values used to indicate the current cluster status.
type ClusterStatus struct {
	Phase Phase `json:"phase"`
}

// Phase indicates the current status of the cluster.
type Phase string

const (
	// Provisioned indicates that the cluster has been created and is fully usable.
	Provisioned Phase = "Provisioned"
	// Errored indicates that the cluster may be unusable due to errors.
	Errored Phase = "Errored"
	// Unknown indicates that the cluster status is not known.
	Unknown Phase = "Unknown"
)

// InternalState holds the state information of the internal operator which is currently in use. Hydroform uses this information for internal purposes only.
type InternalState struct {
	TerraformState *states.State
}

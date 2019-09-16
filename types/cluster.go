package types

type Cluster struct {
	Name              string
	KubernetesVersion string
	CPU               string
	DiskSizeGB        int
	NodeCount         int
	MachineType       string
	Location          string
}

type ClusterInfo struct {
	Status ClusterStatus
	IP     string
}

type ClusterStatus struct {
	Phase string
}

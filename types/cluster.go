package types

type Cluster struct {
	Name              string
	KubernetesVersion string
	CPU               string
	DiskSizeGB        uint32
}

type ClusterInfo struct {
	Status        string
	IP            string
	CloudPlatform string
}

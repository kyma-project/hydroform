package types

type Cluster struct {
	Name              string
	KubernetesVersion string
	CPU               string
	Memory            string
}

type ClusterInfo struct {
	Status string
}

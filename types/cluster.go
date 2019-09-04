package types

type Cluster struct {
	Configuration     map[string]interface{}
	Name              string
	NodesCount        uint32
	KubernetesVersion string
	CPU               string
	Memory            string
}

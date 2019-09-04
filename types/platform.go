package types

type Platform struct {
	Configuration map[string]interface{}
	NodesCount    uint32
	Location      string
	MachineType   string
	ProjectName   string
}

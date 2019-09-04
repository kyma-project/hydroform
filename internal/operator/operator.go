package operator

type Cluster interface {
	Create(provider string, configuration map[string]interface{}) error
	Delete(provider string, configuration map[string]interface{}) error
}

func NewTerraform() Cluster {
	return &Terraform{}
}

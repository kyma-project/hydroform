package errs

const (
	CannotBeEmpty    = "\n - %s cannot be empty"
	CannotBeLess     = "\n - %s cannot be less than %v"
	Custom           = "\n - %v"
	EmptyClusterInfo = "Cluster.ClusterInfo cannot be empty. Please provide the Cluster object returned from the Provision function."
)

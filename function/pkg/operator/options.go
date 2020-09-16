package operator

import v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type ApplyOptions struct {
	DryRun          []string
	OwnerReferences []v1.OwnerReference
	Labels          map[string]string
}

type DeleteOptions struct {
	DryRun []string
	v1.DeletionPropagation
}

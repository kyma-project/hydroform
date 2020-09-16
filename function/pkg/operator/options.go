package operator

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ApplyOptions struct {
	DryRun          []string
	OwnerReferences []v1.OwnerReference
}

type DeleteOptions struct {
	v1.DeletionPropagation
	DryRun   []string
}

package operator

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Callbacks struct {
	Pre  []Callback
	Post []Callback
}

type ApplyOptions struct {
	DryRun          []string
	OwnerReferences []v1.OwnerReference
	Callbacks
}

type DeleteOptions struct {
	v1.DeletionPropagation
	DryRun []string
	Callbacks
}

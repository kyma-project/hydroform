package operator

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Callbacks struct {
	Pre  []Callback
	Post []Callback
}

type Options struct {
	Callbacks
	DryRun []string
}

type ApplyOptions struct {
	Options
	OwnerReferences []v1.OwnerReference
}

type DeleteOptions struct {
	v1.DeletionPropagation
	Options
}

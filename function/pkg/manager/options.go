package manager

import "github.com/kyma-incubator/hydroform/function/pkg/operator"

type OnError int

const (
	NothingOnError OnError = iota
	PurgeOnError
)

type Options struct {
	Callbacks          operator.Callbacks
	OnError            OnError
	DryRun             bool
	SetOwnerReferences bool
	WaitForApply       bool
}

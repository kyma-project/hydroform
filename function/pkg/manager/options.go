package manager

import "github.com/kyma-incubator/hydroform/function/pkg/operator"

type OnError int

const (
	NothingOnError OnError = iota
	PurgeOnError
)

type ManagerOptions struct {
	Callbacks          operator.Callbacks
	OnError            OnError
	DryRun             bool
	SetOwnerReferences bool
}


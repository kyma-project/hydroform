package deployment

import (
	"fmt"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/components"
)

// ProcessEvent represents an event fired during process executing
type ProcessEvent string

const (
	// ProcessStart is set when main process gets started
	ProcessStart ProcessEvent = "ProcessStart"
	// ProcessRunning is indicating a running main process
	ProcessRunning ProcessEvent = "ProcessRunning"
	// ProcessFinished indicates a successfully finished main process
	ProcessFinished ProcessEvent = "ProcessFinished"
	// ProcessExecutionFailure indicates a failure during the execution (install/uninstall of a component failed)
	ProcessExecutionFailure ProcessEvent = "ProcessExecutionFailure"
	// ProcessTimeoutFailure indicates an exceeded timeout
	ProcessTimeoutFailure ProcessEvent = "ProcessTimeoutFailure"
	// ProcessForceQuitFailure indicates an cancelled main process
	ProcessForceQuitFailure ProcessEvent = "ProcessForceQuitFailure"
)

// InstallationPhase represents the current installation phase
type InstallationPhase string

const (
	// InstallPreRequisites indicates the main process is installing pre-requisites
	InstallPreRequisites InstallationPhase = "InstallPreRequisites"
	// UninstallPreRequisites indicates the main process is removing pre-requisites
	UninstallPreRequisites InstallationPhase = "UninstallPreRequisites"
	// InstallComponents indicates the main process is installing components
	InstallComponents InstallationPhase = "InstallComponents"
	// UninstallComponents indicates the main process is removing components
	UninstallComponents InstallationPhase = "UninstallComponents"
)

// ProcessUpdate is an update of the main process
type ProcessUpdate struct {
	Event ProcessEvent
	Phase InstallationPhase
	Error error
	//Component is only set during the component install/uninstall phase
	Component components.KymaComponent
}

func (pu *ProcessUpdate) IsComponentUpdate() bool {
	return pu.Component.Name != ""
}

func (pu ProcessUpdate) String() string {
	return fmt.Sprintf("[ProcessUpdateEvent: event=%s | InstallationPhase=%s | Error=%v | Component=%v]",
		pu.Event, pu.Phase, pu.Error, pu.Component)
}

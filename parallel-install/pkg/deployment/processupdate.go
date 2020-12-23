package deployment

import "github.com/kyma-incubator/hydroform/parallel-install/pkg/components"

// ProcessEvent represents an event fired during process executing
type ProcessEvent uint

const (
	// ProcessStart is set when main process gets started
	ProcessStart ProcessEvent = iota
	// ProcessRunning is indicating a running main process
	ProcessRunning ProcessEvent = iota
	// ProcessFinished indicates a successfully finished main process
	ProcessFinished ProcessEvent = iota
	// ProcessExecutionFailure indicates a failure during the execution (install/uninstall of a component failed)
	ProcessExecutionFailure ProcessEvent = iota
	// ProcessTimeoutFailure indicates an exceeded timeout
	ProcessTimeoutFailure ProcessEvent = iota
	// ProcessForceQuitFailure indicates an cancelled main process
	ProcessForceQuitFailure ProcessEvent = iota
)

// InstallationPhase represents the current installation phase
type InstallationPhase uint

const (
	// InstallPreRequisites indicates the main process is installing pre-requisites
	InstallPreRequisites InstallationPhase = iota
	// UninstallPreRequisites indicates the main process is removing pre-requisites
	UninstallPreRequisites InstallationPhase = iota
	// InstallComponents indicates the main process is installing components
	InstallComponents InstallationPhase = iota
	// UninstallComponents indicates the main process is removing components
	UninstallComponents InstallationPhase = iota
)

// ProcessUpdate is an update of the main process
type ProcessUpdate struct {
	Event ProcessEvent
	Phase InstallationPhase
	//Component is only set during the component install/uninstall phase
	Component components.KymaComponent
}

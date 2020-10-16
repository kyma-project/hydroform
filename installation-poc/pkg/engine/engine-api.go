package engine

import (
	"log"
	"time"
)

// Overrides are used to pass helm values to an Installation/Upgrade operation.
// It's a generic structure, where the value of the map
// can be any valid YAML type, including nested "Overrides" map.
type Overrides map[string]interface{}

// ComponentSource defines component's Helm chart location
type ComponentSource struct {
	URL string `json:"url"`
}

// Allows to uniquely identify the Component
type ComponentMeta struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// KymaComponent is an installation/upgrade unit
// If Source is nil, pre-cofigured location for Sources is used for lookup.
type Component struct {
	ComponentMeta
	Source *ComponentSource `json:"source,omitempty"`
}

// A Set of components. Pointer is used because pointers can be map keys.
type ComponentsSet map[*Component]struct{}

// Common configuration for all operations
type CommonConfig struct {
	RetryCount     int `json:"retryCount"`
	TimeoutSeconds int `json:"timeoutSeconds"`
}

// Install operation configuration
type InstallationConfig struct {
	//Configuration common to all components
	CommonConfig

	//Default location for component sources.
	//If the component does not declare it's "Source" property, this value
	//is extended with a single slash and then a component's name,
	//and such URL is used for lookup.
	CommonSourceDir *ComponentSource `json:"source,omitempty"`

	//Logger used for the operation, so that users can control the output. TODO: Figure out the best interface for this
	Logger log.Logger
}

// Uninstall operation configuration
type UninstallationConfig struct {
	//Configuration common to all components
	CommonConfig

	//Logger used for the operation, so that users can control the output. TODO: Figure out the best interface for this
	Logger log.Logger
}

// Generic status object.
// Timestamp denotes the moment the status was reached.
type Status struct {
	//TODO: We need a dictionary here, eg: "Pending", "Installing", "Installed", "Error"
	value      string
	errorValue error //error object, if status is "Error"
	timestamp  time.Time
}

// Component status, describing what happened during operation
type ComponentStatus struct {
	Meta           ComponentMeta
	RetryCount     int //Current RetryCount for component operation. Zero if not retried
	CurrentStatus  Status
	PreviousStatus *Status //status before CurrentStatus was set, if any
}

type OperationStatus struct {
	CurrentStatus  Status
	PreviousStatus *Status //status before CurrentStatus was set, if any
	OperationError error   //Operation error, if any
}

// Registers notification function for component status changes
type RegisterComponentStatusChangeNotification func(component ComponentMeta, onChange OnComponentStatusChangeFunc) error

// Registers notification function for operation status changes
type RegisterOperationStatusChangeNotification func(onChange OnOperationStatusChangeFunc) error

// Notification Function called by the engine when component status changes
type OnComponentStatusChangeFunc func(operationId string, component ComponentMeta)

// Notification Function called by the engine when operation status changes
type OnOperationStatusChangeFunc func(operationId string)

type RunFunction func() error

// Used to manage operations. Allows to register status notifications
// TODO: Think about concurrency here. Should "Run" be blocking or non-blocking?
//       What about notifications? Looks like Multiple goroutines should be used.
type Operation struct {
	Id                            string
	RegisterComponentNotification RegisterComponentStatusChangeNotification
	RegisterOperationNotification RegisterOperationStatusChangeNotification
	Run                           func() error //Runs the operation.
	//Stop() do we need such function? Maybe there should be
	//        =some channel-based API for stopping?
}

// Main interface
type KymaOperation interface {
	// Returns immediately. Returned *Operation is used to actually start the process.
	PrepareInstalllation(components ComponentsSet, overrides Overrides, config *InstallationConfig) (*Operation, error)
	// Returns immediately. Returned *Operation is used to actually start the process.
	PrepareUninstallation(components ComponentsSet, config *UninstallationConfig) (*Operation, error)

	// Returns
	GetComponentStatus(operationId string, component ComponentMeta) (ComponentStatus, error)
	GetOperationStatus(operationId string) (OperationStatus, error)
}

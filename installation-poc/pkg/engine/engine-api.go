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

// Installation configuration
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

// Uninstallation configuration
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
}

// Registers notification function for component status changes
type RegisterComponentStatusChangeNotification func(onChange OnComponentStatusChangeFunc) error

// Registers notification function for operation status changes
type RegisterOperationStatusChangeNotification func(onChange OnOperationStatusChangeFunc) error

// Notification Function called by the engine when component status changes
type OnComponentStatusChangeFunc func(component ComponentMeta)

// Notification Function called by the engine when operation status changes
type OnOperationStatusChangeFunc func()

// Main interface
type Installation interface {
	// Returns immediately. Returned *Operation is used to actually start the process.
	PrepareInstalllation(components ComponentsSet, overrides Overrides, config *InstallationConfig) (*Operation, error)
}

type Uninstallation interface {
	// Returns immediately. Returned *Operation is used to actually start the process.
	PrepareUninstallation(components ComponentsSet, config *UninstallationConfig) (*Operation, error)
}

//// Engine ////////////////////////////////////////////////////////////////////
//
type Engine struct {
}

//Returns new Engine
func New() (*Engine, error) {
	//TODO: Implement
	res := Engine{}
	return &res, nil
}

func (e *Engine) Initialize() *Bootstrap {
	//TODO: Implement
	res := Bootstrap{
		registerComponentNotification: nil,
		registerOperationNotification: nil,
		start:                         nil,
	}

	return &res
}

//// Bootstrap /////////////////////////////////////////////////////////////////
//
// Allows to register status notifications and start the operation
type Bootstrap struct {
	registerComponentNotification RegisterComponentStatusChangeNotification
	registerOperationNotification RegisterOperationStatusChangeNotification
	start                         func() (*Operation, error) //Runs the operation.
}

func (b *Bootstrap) RegisterOperationNotification(onChange OnOperationStatusChangeFunc) {
	b.registerOperationNotification(onChange)
}

func (b *Bootstrap) RegisterComponentNotification(onChange OnComponentStatusChangeFunc) {
	b.registerComponentNotification(onChange)
}

// Starts the operation. Non-blocking.
func (b *Bootstrap) Start() (*Operation, error) {
	return b.start()
}

//// Operation /////////////////////////////////////////////////////////////////
//
type Operation struct {
	getOperationStatus func() (OperationStatus, error)
	getComponentStatus func(component ComponentMeta) (ComponentStatus, error)
	cancel             func() error
}

func (o *Operation) GetOperationStatus() (OperationStatus, error) {
	return o.getOperationStatus()
}

func (o *Operation) GetComponentStatus(component ComponentMeta) (ComponentStatus, error) {
	return o.getComponentStatus(component)
}

func (o *Operation) Cancel() error {
	//TODO: Implement
	return nil
}

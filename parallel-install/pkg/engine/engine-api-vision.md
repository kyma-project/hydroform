package engine

import (
	"log"
	"time"
)

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

// Overrides are used to pass helm values to an Installation/Upgrade operation.
// It's a generic structure, where the value of the map
// can be any valid YAML type, including nested "Overrides" map.
type Overrides map[string]interface{}

type OverridesProvider interface {
	OverridesFor(componentMeta ComponentMeta) Overrides
}

// Common configuration for all operations
type CommonConfig struct {
	RetryCount     int  `json:"retryCount"`
	TimeoutSeconds int  `json:"timeoutSeconds"`
	DryRun         bool `json:"dryRun"`
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
	//TODO: We need a dictionary here, eg: "Scheduled", "Installing", "Installed", "Upgrading", "Upgraded", "Error"
	Value      string
	ErrorValue error //error object, if status is "Error"
	Timestamp  time.Time
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

//// Engine ////////////////////////////////////////////////////////////////////
//
//The internal functions are defined only to show engine's responsibilities.
type Engine struct {
	overridesProvider             OverridesProvider
	registerComponentNotification func(onChange OnComponentStatusChangeFunc) error
	registerOperationNotification func(onChange OnOperationStatusChangeFunc) error
	startInstallation             func(components ComponentsSet, config *InstallationConfig) (*Operation, error)
	startUninstallation           func(components ComponentsSet, config *UninstallationConfig) (*Operation, error)
}

//Creates new Engine
func New(ovp OverridesProvider) (*Engine, error) {
	//TODO: Implement
	res := Engine{
		overridesProvider:             ovp,
		registerComponentNotification: nil,
		registerOperationNotification: nil,
		startInstallation:             nil,
		startUninstallation:           nil,
	}
	return &res, nil
}

// TODO: Consider if channel-based notifications are not better.
func (e *Engine) RegisterOperationNotification(onChange OnOperationStatusChangeFunc) {
	e.registerOperationNotification(onChange)
}

func (e *Engine) RegisterComponentNotification(onChange OnComponentStatusChangeFunc) {
	e.registerComponentNotification(onChange)
}

// Does not block
func (e *Engine) StartInstallation(components ComponentsSet, config *InstallationConfig) (*Operation, error) {
	//TODO: Implement
	return nil, nil
}

// Does not block
func (e *Engine) StartUninstallation(components ComponentsSet, config *UninstallationConfig) (*Operation, error) {
	//TODO: Implement
	return nil, nil
}

//// Operation /////////////////////////////////////////////////////////////////
//
type Operation struct {
	getOperationStatus func() (OperationStatus, error)
	getComponentStatus func(component ComponentMeta) (ComponentStatus, error)
}

func NewOperation() (*Operation, error) {
	//TODO: Implement
	res := Operation{}
	return &res, nil
}

func (o *Operation) GetOperationStatus() (OperationStatus, error) {
	return o.getOperationStatus()
}

func (o *Operation) GetComponentStatus(component ComponentMeta) (ComponentStatus, error) {
	return o.getComponentStatus(component)
}

//// Notifications /////////////////////////////////////////////////////////////
//
// Notification Function called by the engine when component status changes
type OnComponentStatusChangeFunc func(component ComponentMeta)

// Notification Function called by the engine when operation status changes
type OnOperationStatusChangeFunc func()

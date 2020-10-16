package engine

import "log"

// A Set
type ComponentsSet map[*Component]struct{}

// ComponentSource defines component's Helm chart location
type ComponentSource struct {
	URL string `json:"url"`
}

// Allows to uniquely identify the Component
type ComponentMeta struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// KymaComponent is an installation/uninstallation unit
type Component struct {
	ComponentMeta
	Source *ComponentSource `json:"source,omitempty"`
}

// Common operation configuration
type CommonConfig struct {
	RetryCount     int `json:"retryCount"`
	TimeoutSeconds int `json:"timeoutSeconds"`
}

// Install operation configuration
type InstallConfig struct {
	CommonConfig
	Logger log.Logger //Logger used for the operation, so that users can control the output. TODO: Figure out the best interface for this
}

// Uninstall operation configuration
type UninstallConfig struct {
	CommonConfig
	Logger log.Logger //Logger used for the operation, so that users can control the output. TODO: Figure out the best interface for this
}

// Component status, describing what happened during operation
type ComponentStatus struct {
	ComponentMeta  ComponentMeta
	CurrentStatus  string //We need a dictionary here, eg: "Pending", "Installing", "Installed", "Error"
	PreviousStatus string //Previous value of "Status", set upon Status change
	RetryCount     int    //Current RetryCount, if retried
	LastError      error  //Last observed error for the component, if any
	PreviousError  error  //Prevous error (the one before LastError) for the component, if any
}

// Notification Function called by the engine when component status changes
type OnStatusChangeFunc func(status *ComponentStatus)

// Notification Function called by the engine when component status changes
type RegisterStatusChangeNotification func(component ComponentMeta, onChange OnStatusChangeFunc) error

// Used to distinguish operations. Allows to register a notifier
type Operation struct {
	Id                   string
	RegisterNotification RegisterStatusChangeNotification
}

// Main interface
type KymaOperation interface {
	InstallComponents(components ComponentsSet, config *InstallConfig) (*Operation, error)
	UninstallKyma(components ComponentsSet, config *UninstallConfig) (*Operation, error)
	GetComponentStatus(*Operation, ComponentMeta) (ComponentStatus, error)
}

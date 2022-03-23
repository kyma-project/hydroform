package gcp

import (
	"encoding/json"
	"errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	controlPlaneConfigKind = "ControlPlaneConfig"
	gcpAPIVersion          = "gcp.provider.extensions.gardener.cloud/v1alpha1"
)

func ControlPlaneConfig(cfg map[string]interface{}) (*runtime.RawExtension, error) {
	cp := ControlPlane{
		TypeMeta: metav1.TypeMeta{
			Kind:       controlPlaneConfigKind,
			APIVersion: gcpAPIVersion,
		},
	}

	if zones, ok := cfg["zones"].([]string); ok && len(zones) > 0 {
		cp.Zone = zones[0]
	} else {
		return nil, errors.New("failed creating GCP Control Plane config, no zones available")
	}

	data, err := json.Marshal(cp)

	return &runtime.RawExtension{
		Raw: data,
	}, err
}

// ControlPlane contains configuration settings for the control plane.
type ControlPlane struct {
	metav1.TypeMeta

	// Zones are the GCP zones.
	Zone string `json:"zone"`

	// CloudControllerManager contains configuration settings for the cloud-controller-manager.
	CloudControllerManager *CloudControllerManager `json:"cloudControllerManager,omitempty"`
}

// CloudControllerManagerConfig contains configuration settings for the cloud-controller-manager.
type CloudControllerManager struct {
	// FeatureGates contains information about enabled feature gates.
	FeatureGates map[string]bool
}

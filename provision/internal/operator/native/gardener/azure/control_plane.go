package azure

import (
	"encoding/json"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	controlPlaneConfigKind = "ControlPlaneConfig"

	azureAPIVersion = "azure.provider.extensions.gardener.cloud/v1alpha1"
)

func ControlPlaneConfig(cfg map[string]interface{}) (*runtime.RawExtension, error) {
	cp := ControlPlane{
		TypeMeta: metav1.TypeMeta{
			Kind:       controlPlaneConfigKind,
			APIVersion: azureAPIVersion,
		},
	}

	data, err := json.Marshal(cp)

	return &runtime.RawExtension{
		Raw: data,
	}, err
}

// ControlPlane contains configuration settings for the control plane for Azure and AWS.
type ControlPlane struct {
	metav1.TypeMeta

	// CloudControllerManager contains configuration settings for the cloud-controller-manager.
	// +optional
	CloudControllerManager *CloudControllerManager `json:"cloudControllerManager,omitempty"`
}

// CloudControllerManager contains configuration settings for the cloud-controller-manager.
type CloudControllerManager struct {
	// FeatureGates contains information about enabled feature gates.
	FeatureGates map[string]bool
}

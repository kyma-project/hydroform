package installation

import (
	"fmt"
	"github.com/kyma-incubator/hydroform/install/k8s"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type Configuration struct {
	// Configuration specifies a configuration for all components
	Configuration ConfigEntries
	// ComponentConfiguration specifies configuration for individual components
	ComponentConfiguration []ComponentConfiguration
}

type ComponentConfiguration struct {
	// Component specifies the name of the component for which the configuration will be used
	Component string
	// Configuration specifies configuration for the component
	Configuration ConfigEntries
}

type ConfigEntry struct {
	Key    string
	Value  string
	Secret bool
}

type ConfigEntries []ConfigEntry

func (c ConfigEntries) Get(key string) (ConfigEntry, bool) {
	for _, entry := range c {
		if entry.Key == key {
			return entry, true
		}
	}

	return ConfigEntry{}, false
}

func (c *ConfigEntries) Set(key, value string, secret bool) {
	for i, entry := range *c {
		if entry.Key == key {
			(*c)[i].Value = value
			(*c)[i].Secret = secret

			return
		}
	}

	*c = append(*c, ConfigEntry{Key: key, Value: value, Secret: secret})
}

func configurationToK8sResources(configuration Configuration) ([]*unstructured.Unstructured, error) {
	var unstructuredArray []*unstructured.Unstructured

	k8sResources, err := k8sResourcesFromConfiguration("global", "", configuration.Configuration)
	if err != nil {
		return nil, err
	}

	unstructuredArray = append(unstructuredArray, k8sResources...)

	for _, configs := range configuration.ComponentConfiguration {
		k8sResources, err := k8sResourcesFromConfiguration(configs.Component, configs.Component, configs.Configuration)
		if err != nil {
			return nil, err
		}
		unstructuredArray = append(unstructuredArray, k8sResources...)
	}

	return unstructuredArray, nil
}

func k8sResourcesFromConfiguration(namePrefix, component string, configuration []ConfigEntry) ([]*unstructured.Unstructured, error) {
	configMap := newOverridesConfigMap(namePrefix, component)
	secret := newOverridesSecret(namePrefix, component)

	for _, confEntry := range configuration {
		if confEntry.Secret {
			secret.Data[confEntry.Key] = []byte(confEntry.Value)
			continue
		}
		configMap.Data[confEntry.Key] = confEntry.Value
	}

	var output []*unstructured.Unstructured

	output, err := appendUnstructured(output, configMap)
	if err != nil {
		return nil, err
	}

	output, err = appendUnstructured(output, secret)
	if err != nil {
		return nil, err
	}

	return output, nil
}

func newOverridesConfigMap(namePrefix, component string) *corev1.ConfigMap {
	configMap := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-installer-config", namePrefix),
			Namespace: kymaInstallerNamespace,
			Labels: map[string]string{
				installerOverridesLabelKey: installerOverridesLabelVal,
			},
		},
		Data: map[string]string{},
	}

	if component != "" {
		configMap.Labels[ComponentOverridesLabelKey] = component
	}

	return configMap
}

func appendUnstructured(items []*unstructured.Unstructured, item metav1.Object) ([]*unstructured.Unstructured, error) {
	unstructuredArray, err := k8s.ToUnstructured(item)
	if err != nil {
		return nil, err
	}
	items = append(items, unstructuredArray)
	return items, nil
}

func newOverridesSecret(namePrefix, component string) *corev1.Secret {
	secret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-installer-config", namePrefix),
			Namespace: kymaInstallerNamespace,
			Labels: map[string]string{
				installerOverridesLabelKey: installerOverridesLabelVal,
				ComponentOverridesLabelKey: component,
			},
		},
		Data: map[string][]byte{},
	}

	if component != "" {
		secret.Labels[ComponentOverridesLabelKey] = component
	}

	return secret
}

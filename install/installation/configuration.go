package installation

import (
	"fmt"
	"github.com/kyma-incubator/hydroform/install/k8s"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Configuration struct {
	// Configuration specifies a configuration for all components
	Configuration ConfigEntries
	// ComponentConfiguration specifies configuration for individual components
	ComponentConfiguration []ComponentConfiguration
	// ConflictStrategy specifies system behaviour when global config is already defined
	ConflictStrategy string
}

type ComponentConfiguration struct {
	// Component specifies the name of the component for which the configuration will be used
	Component string
	// Configuration specifies configuration for the component
	Configuration ConfigEntries
	// ConflictStrategy specifies system behaviour when config is already defined
	ConflictStrategy string
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

func configurationToK8sResources(configuration Configuration) ([]*corev1.ConfigMap, []*corev1.Secret) {
	configMaps := make([]*corev1.ConfigMap, 0, len(configuration.ComponentConfiguration)+1)
	secrets := make([]*corev1.Secret, 0, len(configuration.ComponentConfiguration)+1)

	configMap, secret := k8sResourcesFromConfiguration("global", "", configuration.Configuration)

	configMap.ObjectMeta.Labels[k8s.OnConflictLabel] = configuration.ConflictStrategy
	secret.ObjectMeta.Labels[k8s.OnConflictLabel] = configuration.ConflictStrategy

	configMaps = append(configMaps, configMap)
	secrets = append(secrets, secret)

	for _, configs := range configuration.ComponentConfiguration {
		configMap, secret := k8sResourcesFromConfiguration(configs.Component, configs.Component, configs.Configuration)

		configMap.ObjectMeta.Labels[k8s.OnConflictLabel] = configs.ConflictStrategy
		secret.ObjectMeta.Labels[k8s.OnConflictLabel] = configs.ConflictStrategy

		configMaps = append(configMaps, configMap)
		secrets = append(secrets, secret)
	}

	return configMaps, secrets
}

func k8sResourcesFromConfiguration(namePrefix, component string, configuration []ConfigEntry) (*corev1.ConfigMap, *corev1.Secret) {
	configMap := newOverridesConfigMap(namePrefix, component)
	secret := newOverridesSecret(namePrefix, component)

	for _, confEntry := range configuration {
		if confEntry.Secret {
			secret.Data[confEntry.Key] = []byte(confEntry.Value)
			continue
		}
		configMap.Data[confEntry.Key] = confEntry.Value
	}

	return configMap, secret
}

func newOverridesConfigMap(namePrefix, component string) *corev1.ConfigMap {
	configMap := &corev1.ConfigMap{
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

func newOverridesSecret(namePrefix, component string) *corev1.Secret {
	secret := &corev1.Secret{
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

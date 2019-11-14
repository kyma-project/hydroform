package installation

import (
	"fmt"

	"github.com/pkg/errors"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Configuration struct {
	Configuration          []ConfigEntry
	ComponentConfiguration []ComponentConfiguration
}

type ComponentConfiguration struct {
	Component     string
	Configuration []ConfigEntry
}

type ConfigEntry struct {
	Key    string
	Value  string
	Secret bool
}

func YamlToConfiguration() {
	// TODO - implement helper method
}

func (k KymaInstaller) applyConfiguration(configuration Configuration) error {
	configMaps := make([]*corev1.ConfigMap, 0, len(configuration.ComponentConfiguration)+1)
	secrets := make([]*corev1.Secret, 0, len(configuration.ComponentConfiguration)+1)

	configMap, secret := k8sResourcesFromConfiguration("global", "", configuration.Configuration)
	configMaps = append(configMaps, configMap)
	secrets = append(secrets, secret)

	for _, configs := range configuration.ComponentConfiguration {
		configMap, secret := k8sResourcesFromConfiguration(configs.Component, configs.Component, configs.Configuration)
		configMaps = append(configMaps, configMap)
		secrets = append(secrets, secret)
	}

	err := k.k8sGenericClient.ApplyConfigMaps(configMaps, kymaInstallerNamespace)
	if err != nil {
		return errors.Wrap(err, "failed to create configuration config maps")
	}

	err = k.k8sGenericClient.ApplySecrets(secrets, kymaInstallerNamespace)
	if err != nil {
		return errors.Wrap(err, "failed to create configuration secrets")
	}

	return nil
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
		configMap.Labels[componentOverridesLabelKey] = component
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
				componentOverridesLabelKey: component,
			},
		},
		Data: map[string][]byte{},
	}

	if component != "" {
		secret.Labels[componentOverridesLabelKey] = component
	}

	return secret
}

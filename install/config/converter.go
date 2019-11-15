package config

import (
	"fmt"

	"github.com/kyma-incubator/hydroform/install/k8s"

	"github.com/kyma-incubator/hydroform/install/installation"

	"k8s.io/apimachinery/pkg/runtime"

	corev1 "k8s.io/api/core/v1"
)

const (
	configMapKind = "ConfigMap"
	secretKind    = "Secret"
)

// YAMLToConfiguration converts yaml content containing ConfigMaps and Secrets to installation Configuration
func YAMLToConfiguration(decoder runtime.Decoder, yamlContent string) (installation.Configuration, error) {
	k8sObjects, err := k8s.ParseYamlToK8sObjects(decoder, yamlContent)
	if err != nil {
		return installation.Configuration{}, fmt.Errorf("failed to convert yaml to configuration: %s", err.Error())
	}

	configMaps, secrets, err := parseToConfigMapsAndSecretsObjects(k8sObjects)
	if err != nil {
		return installation.Configuration{}, err
	}

	configuration := installation.Configuration{
		ComponentConfiguration: []installation.ComponentConfiguration{},
	}

	for _, cm := range configMaps {
		component, found := cm.Labels[installation.ComponentOverridesLabelKey]
		if !found {
			configuration.Configuration = addEntriesFromConfigMap(configuration.Configuration, cm.Data)
			continue
		}

		componentConfig := getOrNewComponentConfig(configuration.ComponentConfiguration, component)
		componentConfig.Configuration = addEntriesFromConfigMap(componentConfig.Configuration, cm.Data)

		setComponentConfig(&configuration, componentConfig)
	}

	for _, secret := range secrets {
		component, found := secret.Labels[installation.ComponentOverridesLabelKey]
		if !found {
			configuration.Configuration = addEntriesFromSecrets(configuration.Configuration, secret.Data)
			continue
		}

		componentConfig := getOrNewComponentConfig(configuration.ComponentConfiguration, component)
		componentConfig.Configuration = addEntriesFromSecrets(componentConfig.Configuration, secret.Data)

		setComponentConfig(&configuration, componentConfig)
	}

	return configuration, nil
}

func getOrNewComponentConfig(configs []installation.ComponentConfiguration, component string) installation.ComponentConfiguration {
	for _, conf := range configs {
		if conf.Component == component {
			return conf
		}
	}

	return installation.ComponentConfiguration{Component: component}
}

func setComponentConfig(configuration *installation.Configuration, config installation.ComponentConfiguration) {
	for i, compConf := range configuration.ComponentConfiguration {
		if compConf.Component == config.Component {
			configuration.ComponentConfiguration[i] = config
			return
		}
	}

	configuration.ComponentConfiguration = append(configuration.ComponentConfiguration, config)
}

func addEntriesFromConfigMap(existing []installation.ConfigEntry, newEntries map[string]string) []installation.ConfigEntry {
	if existing == nil {
		existing = make([]installation.ConfigEntry, 0, len(newEntries))
	}

	for key, val := range newEntries {
		existing = append(existing, installation.ConfigEntry{Key: key, Value: val, Secret: false})
	}

	return existing
}

func addEntriesFromSecrets(existing []installation.ConfigEntry, newEntries map[string][]byte) []installation.ConfigEntry {
	if existing == nil {
		existing = make([]installation.ConfigEntry, 0, len(newEntries))
	}

	for key, val := range newEntries {
		existing = append(existing, installation.ConfigEntry{Key: key, Value: string(val), Secret: true})
	}

	return existing
}

func parseToConfigMapsAndSecretsObjects(k8sObjects []k8s.K8sObject) ([]corev1.ConfigMap, []corev1.Secret, error) {
	configMaps := make([]corev1.ConfigMap, 0)
	secrets := make([]corev1.Secret, 0)

	for _, object := range k8sObjects {
		switch object.GVK.Kind {
		case configMapKind:
			cm, ok := object.Object.(*corev1.ConfigMap)
			if !ok {
				return nil, nil, fmt.Errorf("invalid type of object of kind ConfigMap, expected *ConfigMap, actual %T", object.Object)
			}
			configMaps = append(configMaps, *cm)
		case secretKind:
			secret, ok := object.Object.(*corev1.Secret)
			if !ok {
				return nil, nil, fmt.Errorf("invalid type of object of kind Secret, expected *Secret, actual %T", object.Object)
			}
			secrets = append(secrets, *secret)
		default:
			return nil, nil, fmt.Errorf("unexpected object kind %s, expected %s or %s", object.GVK.Kind, configMapKind, secretKind)
		}
	}

	return configMaps, secrets, nil
}

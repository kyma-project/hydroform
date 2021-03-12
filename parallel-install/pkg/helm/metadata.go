package helm

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/fatih/structs"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage"
	"k8s.io/apimachinery/pkg/api/errors"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	KymaLabelPrefix = "kyma"
)

type helmReleaseNotFoundError struct {
	name string
}

func (err *helmReleaseNotFoundError) Error() string {
	return fmt.Sprintf("No installed Helm release found for component '%s'", err.name)
}

type helmSecretNameInvalidError struct {
	namespace string
	secret    string
}

func (err *helmSecretNameInvalidError) Error() string {
	return fmt.Sprintf("Could not resolve version of Helm secret '%s' (namespace '%s')", err.secret, err.namespace)
}

type kymaMetadataUnavailableError struct {
	secret string
}

func (err *kymaMetadataUnavailableError) Error() string {
	return fmt.Sprintf("Kyma metadata not found in HELM secret '%s'", err.secret)
}

type kymaMetadataFieldUnknownError struct {
	field string
}

func (err *kymaMetadataFieldUnknownError) Error() string {
	return fmt.Sprintf("Kyma metadata struct does not contain a field '%s'", err.field)
}

type KymaMetadata struct {
	Profile   string
	Version   string
	Component string
}

func (km *KymaMetadata) isValid() bool {
	//check whether all mandatory fields are defined
	return km.Version != "" && km.Component != ""
}

type KymaVersion struct {
	Version    string
	Components []string
}

type KymaMetadataProvider struct {
	kubeClient kubernetes.Interface
}

func NewKymaMetadataProvider(client kubernetes.Interface) *KymaMetadataProvider {
	return &KymaMetadataProvider{
		kubeClient: client,
	}
}

func (mp *KymaMetadataProvider) Version() ([]*KymaVersion, error) {
	compField, err := mp.structField("Component")
	if err != nil {
		return nil, err
	}

	//get all secrets which are labeled as Kyma component
	options := metaV1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=true", compField.Name()),
	}
	secrets, err := mp.kubeClient.CoreV1().Secrets("").List(context.Background(), options)
	if err != nil {
		return nil, err
	}

	//group secrets by component-name (required to find latest secret)
	secretsPerComp := make(map[string][]v1.Secret)
	for _, secret := range secrets.Items {
		if name, ok := secret.Labels["name"]; ok {
			secretsPerComp[name] = append(secretsPerComp[name], secret)
		}
	}

	return mp.resolveKymaVersions(secretsPerComp)
}

func (mp *KymaMetadataProvider) resolveKymaVersions(secretsPerComp map[string][]v1.Secret) ([]*KymaVersion, error) {
	//collect all Kyma versions in cluster
	versionField, err := mp.structField("Version")
	if err != nil {
		return nil, err
	}

	versions := make(map[string]*KymaVersion)
	for compName, secrets := range secretsPerComp {
		latestSecret, err := mp.findLatestSecret(compName, secrets)
		if err != nil {
			return nil, err
		}
		kymaVersionOfComp := latestSecret.Labels[mp.labelName(versionField)]
		if kymaVersion, ok := versions[kymaVersionOfComp]; !ok {
			versions[kymaVersionOfComp] = &KymaVersion{
				Version: kymaVersionOfComp,
			}
		} else {
			kymaVersion.Components = append(kymaVersion.Components, compName)
		}
	}

	return mp.versionFromMap(versions), nil
}

func (mp *KymaMetadataProvider) structField(fieldName string) (*structs.Field, error) {
	field := structs.New(KymaMetadata{}).Field(fieldName)
	if field == nil {
		return nil, &kymaMetadataFieldUnknownError{field: fieldName}
	}
	return field, nil
}

func (mp *KymaMetadataProvider) versionFromMap(versionMap map[string]*KymaVersion) []*KymaVersion {
	versions := make([]*KymaVersion, 0, len(versionMap))
	for _, v := range versionMap {
		versions = append(versions, v)
	}
	return versions
}

func (mp *KymaMetadataProvider) findLatestSecret(name string, secrets []v1.Secret) (*v1.Secret, error) {
	var latestSecret v1.Secret

	//find latest Helm secret
	latestChartVersion := -1
	secretPrefix := mp.secretPrefix(name)
	for _, secret := range secrets {
		if strings.HasPrefix(secret.Name, secretPrefix) {
			currChartVersion, err := strconv.Atoi(strings.Replace(secret.Name, secretPrefix, "", 1))
			if err != nil {
				return nil, &helmSecretNameInvalidError{
					secret:    secret.Name,
					namespace: secret.Namespace,
				}
			}
			if currChartVersion > latestChartVersion {
				latestChartVersion = currChartVersion
				latestSecret = secret
			}
		}
	}
	if latestChartVersion == -1 {
		return nil, &helmReleaseNotFoundError{name: name}
	}
	return &latestSecret, nil
}

func (mp *KymaMetadataProvider) Set(release *release.Release, metadata *KymaMetadata) error {
	if metadata == nil {
		return fmt.Errorf("No Kyma metadata provided for Helm release '%s' (namespace '%s')", release.Name, release.Namespace)
	}

	secretName := mp.secretName(release.Name, release.Version)
	//get existing secret
	secret, err := mp.kubeClient.CoreV1().Secrets(release.Namespace).Get(context.Background(), secretName, metaV1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return &helmReleaseNotFoundError{name: secretName}
		}
		return err
	}

	//update secret
	mp.marshalMetadata(secret, metadata)
	_, err = mp.kubeClient.CoreV1().Secrets(release.Namespace).Update(context.Background(), secret, metaV1.UpdateOptions{})
	return err
}

func (mp *KymaMetadataProvider) Get(name string) (*KymaMetadata, error) {
	secret, err := mp.latestSecret(name, "")
	if err != nil {
		return &KymaMetadata{}, err
	}
	return mp.unmarshalMetadata(secret)
}

func (mp *KymaMetadataProvider) latestSecret(name, namespace string) (*v1.Secret, error) {
	secrets, err := mp.kubeClient.CoreV1().Secrets(namespace).List(context.Background(), metaV1.ListOptions{})
	if err != nil {
		return nil, err
	}

	latestSecret, err := mp.findLatestSecret(name, secrets.Items)
	if err != nil {
		return nil, err
	}
	if latestSecret == nil {
		return nil, &helmReleaseNotFoundError{name: name}
	}

	return latestSecret, nil
}

func (mp *KymaMetadataProvider) marshalMetadata(secret *v1.Secret, metadata *KymaMetadata) {
	if secret.Labels == nil {
		secret.Labels = make(map[string]string)
	}
	for _, field := range structs.New(metadata).Fields() {
		secret.Labels[mp.labelName(field)] = field.Value().(string)
	}
}

func (mp *KymaMetadataProvider) unmarshalMetadata(secret *v1.Secret) (*KymaMetadata, error) {
	var metadata *KymaMetadata = &KymaMetadata{}
	for _, field := range structs.New(metadata).Fields() {
		if value, ok := secret.Labels[mp.labelName(field)]; ok {
			if err := field.Set(value); err != nil {
				return nil, err
			}
		}
	}
	if metadata.isValid() {
		return metadata, nil
	}
	return nil, &kymaMetadataUnavailableError{secret: secret.Name}
}

func (mp *KymaMetadataProvider) labelName(field *structs.Field) string {
	return fmt.Sprintf("%s%s", KymaLabelPrefix, field.Name())
}

func (mp *KymaMetadataProvider) secretPrefix(name string) string {
	return fmt.Sprintf("%s.%s.v", storage.HelmStorageType, name)
}

func (mp *KymaMetadataProvider) secretName(name string, version int) string {
	return fmt.Sprintf("%s%d", mp.secretPrefix(name), version)
}

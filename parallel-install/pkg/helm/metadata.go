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

func (rErr *helmReleaseNotFoundError) Error() string {
	return fmt.Sprintf("No installed Helm release found for component '%s'", rErr.name)
}

type helmSecretNameInvalidError struct {
	namespace string
	secret    string
}

func (rErr *helmSecretNameInvalidError) Error() string {
	return fmt.Sprintf("Could not resolve version of Helm secret '%s' (namespace '%s')", rErr.secret, rErr.namespace)
}

type kymaMetadataUnavailableError struct {
	secret string
}

func (rErr *kymaMetadataUnavailableError) Error() string {
	return fmt.Sprintf("Kyma metadata not found in HELM secret '%s'", rErr.secret)
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

type KymaMetadataProvider struct {
	kubeClient kubernetes.Interface
}

func NewKymaMetadataProvider(client kubernetes.Interface) *KymaMetadataProvider {
	return &KymaMetadataProvider{
		kubeClient: client,
	}
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
	var latestSecret *v1.Secret

	secrets, err := mp.kubeClient.CoreV1().Secrets(namespace).List(context.Background(), metaV1.ListOptions{})
	if err != nil {
		return latestSecret, err
	}

	//find latest Helm secret
	latestChartVersion := -1
	secretPrefix := mp.secretPrefix(name)
	for _, secret := range secrets.Items {
		if strings.HasPrefix(secret.Name, secretPrefix) {
			currChartVersion, err := strconv.Atoi(strings.Replace(secret.Name, secretPrefix, "", 1))
			if err != nil {
				return latestSecret, &helmSecretNameInvalidError{
					secret:    secret.Name,
					namespace: secret.Namespace,
				}
			}
			if currChartVersion > latestChartVersion {
				latestChartVersion = currChartVersion
				latestSecret = &secret
			}
		}
	}

	if latestChartVersion < 0 {
		return latestSecret, &helmReleaseNotFoundError{name: name}
	}
	return latestSecret, nil
}

func (mp *KymaMetadataProvider) marshalMetadata(secret *v1.Secret, metadata *KymaMetadata) {
	if secret.Labels == nil {
		secret.Labels = make(map[string]string)
	}
	s := structs.New(metadata)
	for _, field := range s.Fields() {
		secret.Labels[mp.labelName(field)] = field.Value().(string)
	}
}

func (mp *KymaMetadataProvider) unmarshalMetadata(secret *v1.Secret) (*KymaMetadata, error) {
	var metadata *KymaMetadata = &KymaMetadata{}
	s := structs.New(metadata)
	for _, field := range s.Fields() {
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

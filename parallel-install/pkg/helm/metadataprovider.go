package helm

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unicode"

	"github.com/fatih/structs"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/config"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	KymaLabelPrefix = "kyma-project.io/install." //label prefix used to distinguish Kyma labels from Helm labels in the Helm secrets
)

//helmReleaseNotFoundError is fired when a release could not be found in the cluster
type helmReleaseNotFoundError struct {
	name string
}

func (err *helmReleaseNotFoundError) Error() string {
	return fmt.Sprintf("No installed Helm release found for component '%s'", err.name)
}

//helmSecretNameInvalidError is fired when the requested secret name became invalid
type helmSecretNameInvalidError struct {
	namespace string
	secret    string
}

func (err *helmSecretNameInvalidError) Error() string {
	return fmt.Sprintf("Could not resolve version of Helm secret '%s' (namespace '%s')", err.secret, err.namespace)
}

//kymaMetadataUnavailableError is fired when mandatory metadata labels were missing in the Helm secret
type kymaMetadataUnavailableError struct {
	secret string
	err    error
}

func (err *kymaMetadataUnavailableError) Error() string {
	return fmt.Sprintf("Kyma component metadata not found or incomplete in HELM secret '%s': %s", err.secret, err.err.Error())
}

//kymaMetadataFieldUnknownError is fired when a field of the metadata object doesn't exist (internal error, should never be fired)
type kymaMetadataFieldUnknownError struct {
	field string
}

func (err *kymaMetadataFieldUnknownError) Error() string {
	return fmt.Sprintf("Kyma metadata struct does not contain a field '%s'", err.field)
}

//KymaMetadataProvider enables access to Kyma component metadata and version information
type KymaMetadataProvider struct {
	kubeClient kubernetes.Interface
}

//NewKymaMetadataProvider creates a new KymaMetadataProvider
func NewKymaMetadataProvider(kubeconfigPath, kubeconfigRaw string) (*KymaMetadataProvider, error) {
	manager, err := config.NewKubeConfigManager(&kubeconfigPath, &kubeconfigRaw)
	if err != nil {
		return nil, err
	}

	restConfig, err := manager.Config()
	if err != nil {
		return nil, err
	}

	kubeClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	return &KymaMetadataProvider{
		kubeClient: kubeClient,
	}, nil
}

//Versions returns the set of installed Kyma versions
func (mp *KymaMetadataProvider) Versions() (*KymaVersionSet, error) {
	//get all secrets which are labeled as Kyma component
	compField, err := mp.structField("Component")
	if err != nil {
		return nil, err
	}
	options := metaV1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=true", mp.labelName(compField)),
	}
	secrets, err := mp.kubeClient.CoreV1().Secrets("").List(context.Background(), options)
	if err != nil {
		return nil, err
	}

	//group secrets by component-name (required to find latest secret)
	nameField, err := mp.structField("Name")
	if err != nil {
		return nil, err
	}
	secretsPerComp := make(map[string][]v1.Secret)
	for _, secret := range secrets.Items {
		if name, ok := secret.Labels[mp.labelName(nameField)]; ok {
			secretsPerComp[name] = append(secretsPerComp[name], secret)
		}
	}

	versions, err := mp.resolveKymaVersions(secretsPerComp)
	if err != nil {
		return nil, err
	}
	return &KymaVersionSet{
		Versions: versions,
	}, nil
}

//resolveKymaVersions creates KymaVersion instances from Helm Secret labels
func (mp *KymaMetadataProvider) resolveKymaVersions(secretsPerComp map[string][]v1.Secret) ([]*KymaVersion, error) {
	versions := make(map[string]*KymaVersion) //we se the opsID as differentiator between the different versions
	for compName, secrets := range secretsPerComp {
		latestSecret, err := mp.findLatestSecret(compName, secrets)
		if err != nil {
			return nil, err
		}
		compMeta, err := mp.unmarshalMetadata(latestSecret)
		if err != nil {
			return nil, err
		}
		kymaVersion, ok := versions[compMeta.OperationID]
		if !ok {
			//create version instance if missing
			kymaVersion = &KymaVersion{
				Version:      compMeta.Version,
				Profile:      compMeta.Profile,
				OperationID:  compMeta.OperationID,
				CreationTime: compMeta.CreationTime,
			}
			versions[compMeta.OperationID] = kymaVersion
		}
		//add component to version
		kymaVersion.Components = append(kymaVersion.Components, compMeta)
	}

	return mp.versionFromMap(versions), nil
}

//structField returns a structField from a KymaComponentMetadata object
func (mp *KymaMetadataProvider) structField(fieldName string) (*structs.Field, error) {
	field := structs.New(KymaComponentMetadata{}).Field(fieldName)
	if field == nil {
		return nil, &kymaMetadataFieldUnknownError{field: fieldName}
	}
	return field, nil
}

//versionFromMap returns the version instances tracked in the map
func (mp *KymaMetadataProvider) versionFromMap(versionMap map[string]*KymaVersion) []*KymaVersion {
	versions := make([]*KymaVersion, 0, len(versionMap))
	for _, v := range versionMap {
		versions = append(versions, v)
	}
	return versions
}

//findLatestSecret returns the latest Helm secret of a component
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

//Set adds Kyma metadata labels to a Helm secret
func (mp *KymaMetadataProvider) Set(release *release.Release, compMetaTpl *KymaComponentMetadataTemplate) error {
	if compMetaTpl == nil {
		return fmt.Errorf("No Kyma metadata factory provided for Helm release '%s' (namespace '%s')", release.Name, release.Namespace)
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
	metadata, err := compMetaTpl.Build(release.Namespace, release.Name)
	if err != nil {
		return err
	}
	mp.marshalMetadata(secret, metadata)
	_, err = mp.kubeClient.CoreV1().Secrets(release.Namespace).Update(context.Background(), secret, metaV1.UpdateOptions{})
	return err
}

//Get returns Kyma metadata of an installed component
func (mp *KymaMetadataProvider) Get(name string) (*KymaComponentMetadata, error) {
	secret, err := mp.latestSecret(name, "")
	if err != nil {
		return nil, err
	}
	return mp.unmarshalMetadata(secret)
}

//latestSecret returns the latest Helm secret of a component
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

//marshalMetadata creates a KymaComponentMetadata from secret labels
func (mp *KymaMetadataProvider) marshalMetadata(secret *v1.Secret, metadata *KymaComponentMetadata) {
	if secret.Labels == nil {
		secret.Labels = make(map[string]string)
	}
	var labelValue string
	for _, field := range structs.New(metadata).Fields() {
		switch field.Kind() {
		case reflect.Bool:
			labelValue = fmt.Sprintf("%t", field.Value())
		case reflect.String:
			labelValue = field.Value().(string)
		case reflect.Int64:
			labelValue = fmt.Sprintf("%d", field.Value())
		default:
		}
		secret.Labels[mp.labelName(field)] = labelValue
	}
}

//unmarshalMetadata converts a KymaComponentMetadata to secret labels
func (mp *KymaMetadataProvider) unmarshalMetadata(secret *v1.Secret) (*KymaComponentMetadata, error) {
	var metadata *KymaComponentMetadata = &KymaComponentMetadata{}
	var typedValue interface{}
	var err error
	for _, field := range structs.New(metadata).Fields() {
		if value, ok := secret.Labels[mp.labelName(field)]; ok {
			//convert label-values to typed values
			switch field.Kind() {
			case reflect.Bool:
				typedValue, err = strconv.ParseBool(value)
				if err != nil {
					return nil, fmt.Errorf("Cannot unmarshal KymaComponentMetadata field '%s' "+
						"because value '%s' cannot be converted to bool", field.Name(), value)
				}
			case reflect.String:
				typedValue = value
			case reflect.Int64:
				typedValue, err = strconv.ParseInt(value, 10, 64)
				if err != nil {
					return nil, fmt.Errorf("Cannot unmarshal KymaComponentMetadata field '%s' "+
						"because value '%s' cannot be converted to int64", field.Name(), value)
				}
			default:
				return nil, fmt.Errorf("Cannot unmarshal KymaComponentMetadata field '%s' "+
					"because kind '%s' is not supported yet", field.Name(), field.Kind().String())
			}
			//set the typed value to the field
			if err := field.Set(typedValue); err != nil {
				return nil, err
			}
		}
	}
	if err := metadata.isValid(); err != nil {
		return nil, &kymaMetadataUnavailableError{secret: secret.Name, err: err}
	}
	return metadata, nil
}

func (mp *KymaMetadataProvider) labelName(field *structs.Field) string {
	return fmt.Sprintf("%s%s", KymaLabelPrefix, lowercaseFirst(field.Name()))
}

func lowercaseFirst(str string) string {
	for i, v := range str {
		return string(unicode.ToLower(v)) + str[i+1:]
	}
	return ""
}

func (mp *KymaMetadataProvider) secretPrefix(name string) string {
	return fmt.Sprintf("%s.%s.v", storage.HelmStorageType, name)
}

func (mp *KymaMetadataProvider) secretName(name string, version int) string {
	return fmt.Sprintf("%s%d", mp.secretPrefix(name), version)
}

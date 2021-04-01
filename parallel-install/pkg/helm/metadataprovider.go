package helm

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"unicode"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/fatih/structs"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage"
	"helm.sh/helm/v3/pkg/storage/driver"
	"k8s.io/apimachinery/pkg/api/errors"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

//KymaMetadataProvider enables access to Kyma component metadata and version information
type KymaMetadataProvider struct {
	kubeClient kubernetes.Interface //TODO: remove this field as soon Helm supports to labels on releases (see)
	storage    *storage.Storage
}

//NewKymaMetadataProvider creates a new KymaMetadataProvider
func NewKymaMetadataProvider(client kubernetes.Interface) *KymaMetadataProvider {
	return &KymaMetadataProvider{
		kubeClient: client,
		storage:    storage.Init(driver.NewSecrets(client.CoreV1().Secrets(""))),
	}
}

//Versions returns the set of installed Kyma versions
func (mp *KymaMetadataProvider) Versions() (*KymaVersionSet, error) {
	releases, err := mp.storage.ListDeployed()
	if err != nil {
		return nil, err
	}
	versions, err := mp.resolveKymaVersions(releases)
	if err != nil {
		return nil, err
	}
	return &KymaVersionSet{
		Versions: versions,
	}, nil
}

//resolveKymaVersions creates KymaVersion instances from Helm Secret labels
func (mp *KymaMetadataProvider) resolveKymaVersions(releases []*release.Release) ([]*KymaVersion, error) {
	versions := make(map[string]*KymaVersion) //we se the opsID as differentiator between the different versions
	for _, release := range releases {
		compMeta, err := mp.unmarshalMetadata(release)
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

//versionFromMap returns the version instances tracked in the map
func (mp *KymaMetadataProvider) versionFromMap(versionMap map[string]*KymaVersion) []*KymaVersion {
	versions := make([]*KymaVersion, 0, len(versionMap))
	for _, v := range versionMap {
		versions = append(versions, v)
	}
	return versions
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
	release, err := mp.storage.Last(name)
	if err != nil {
		return nil, err
	}
	return mp.unmarshalMetadata(release)
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
func (mp *KymaMetadataProvider) unmarshalMetadata(release *release.Release) (*KymaComponentMetadata, error) {
	var metadata *KymaComponentMetadata = &KymaComponentMetadata{}
	var typedValue interface{}
	var err error
	for _, field := range structs.New(metadata).Fields() {
		if value, ok := release.Labels[mp.labelName(field)]; ok {
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
		return nil, &kymaMetadataUnavailableError{secret: release.Name, err: err}
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

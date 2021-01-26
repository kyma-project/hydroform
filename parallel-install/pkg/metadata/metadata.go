package metadata

import (
	"context"
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/cenkalti/backoff/v4"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var metadataName = "kyma"
var metadataNamespace = "kyma-system"
var initialInterval = time.Duration(3) * time.Second
var maxElapsedTime = time.Duration(20) * time.Second

// StatusEnum describes deployment / uninstallation status
type StatusEnum string

const (
	//DeploymentInProgress means deployment of kyma is in progress
	DeploymentInProgress StatusEnum = "DeploymentInProgress"

	//UninstallationInProgress means uninstallation of kyma is in progress
	UninstallationInProgress StatusEnum = "Uninstallation in progress"

	//DeploymentError means error occurred during kyma deployment
	DeploymentError StatusEnum = "DeploymentError"

	//UninstallationError means error occurred during kyma uninstallation
	UninstallationError StatusEnum = "UninstallationError"

	//Deployed means kyma deployed successfuly
	Deployed StatusEnum = "Deployed"
)

//Attributes represents common metadata attributes
type Attributes struct {
	profile             string
	version             string
	componentListData   []byte
	componentListFormat string
}

func NewAttributes(profile string, version string, componentListFile string) (*Attributes, error) {
	compListData, err := ioutil.ReadFile(componentListFile)
	if err != nil {
		return nil, err
	}
	return &Attributes{
		profile:             profile,
		version:             version,
		componentListData:   compListData,
		componentListFormat: filepath.Ext(componentListFile),
	}, nil
}

type MetadataProvider interface {
	ReadKymaMetadata() (*KymaMetadata, error)
	WriteKymaDeploymentInProgress(attr *Attributes) error
	WriteKymaDeploymentError(attr *Attributes, reason string) error
	WriteKymaDeployed(attr *Attributes) error
	WriteKymaUninstallationInProgress(attr *Attributes) error
	WriteKymaUninstallationError(attr *Attributes, reason string) error
}

type KymaMetadata struct {
	Profile             string
	Version             string
	ComponentListData   []byte
	ComponentListFormat string
	Status              StatusEnum
	Reason              string
}

type Provider struct {
	kubeClient kubernetes.Interface
}

func New(client kubernetes.Interface) MetadataProvider {
	return &Provider{
		kubeClient: client,
	}
}

func (p *Provider) ReadKymaMetadata() (*KymaMetadata, error) {
	kymaMetadataCM := &v1.ConfigMap{}
	var err error

	exponentialBackoff := backoff.NewExponentialBackOff()
	exponentialBackoff.InitialInterval = initialInterval
	exponentialBackoff.MaxElapsedTime = maxElapsedTime

	retryErr := retryOperation(func() error {
		kymaMetadataCM, err = p.kubeClient.CoreV1().ConfigMaps(metadataNamespace).Get(context.TODO(), metadataName, metav1.GetOptions{})
		if err != nil {
			if !errors.IsNotFound(err) {
				return err
			}
		}

		return nil
	})
	if retryErr != nil {
		return nil, retryErr
	}

	kymaMetaData := cmToMetadata(kymaMetadataCM.Data)

	return kymaMetaData, nil
}

func (p *Provider) WriteKymaDeploymentInProgress(attr *Attributes) error {
	meta := &KymaMetadata{
		Version:             attr.version,
		Profile:             attr.profile,
		ComponentListData:   attr.componentListData,
		ComponentListFormat: attr.componentListFormat,
		Status:              DeploymentInProgress,
	}

	return retryOperation(func() error {
		return p.writeKymaMetadata(meta)
	})
}

func (p *Provider) WriteKymaUninstallationInProgress(attr *Attributes) error {
	meta := &KymaMetadata{
		Version:             attr.version,
		Profile:             attr.profile,
		ComponentListData:   attr.componentListData,
		ComponentListFormat: attr.componentListFormat,
		Status:              UninstallationInProgress,
	}

	return retryOperation(func() error {
		return p.writeKymaMetadata(meta)
	})
}

func (p *Provider) WriteKymaDeploymentError(attr *Attributes, reason string) error {
	meta := &KymaMetadata{
		Version:             attr.version,
		Profile:             attr.profile,
		ComponentListData:   attr.componentListData,
		ComponentListFormat: attr.componentListFormat,
		Status:              DeploymentError,
		Reason:              reason,
	}

	return retryOperation(func() error {
		return p.writeKymaMetadata(meta)
	})
}

func (p *Provider) WriteKymaUninstallationError(attr *Attributes, reason string) error {
	meta := &KymaMetadata{
		Version:             attr.version,
		Profile:             attr.profile,
		ComponentListData:   attr.componentListData,
		ComponentListFormat: attr.componentListFormat,
		Status:              UninstallationError,
		Reason:              reason,
	}

	return retryOperation(func() error {
		return p.writeKymaMetadata(meta)
	})
}

func (p *Provider) WriteKymaDeployed(attr *Attributes) error {
	meta := &KymaMetadata{
		Version:             attr.version,
		Profile:             attr.profile,
		ComponentListData:   attr.componentListData,
		ComponentListFormat: attr.componentListFormat,
		Status:              Deployed,
	}

	return retryOperation(func() error {
		return p.writeKymaMetadata(meta)
	})
}

func (p *Provider) writeKymaMetadata(data *KymaMetadata) error {
	cmData := metadataToCM(data)

	kymaMetadataCM, err := p.kubeClient.CoreV1().ConfigMaps(metadataNamespace).Get(context.TODO(), metadataName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			cmToSave := &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      metadataName,
					Namespace: metadataNamespace,
				},
				Data: cmData,
			}

			_, err := p.kubeClient.CoreV1().ConfigMaps(metadataNamespace).Create(context.TODO(), cmToSave, metav1.CreateOptions{})
			if err != nil {
				return err
			}
			return nil
		}
		return err
	}

	kymaMetadataCM.Data = cmData
	_, err = p.kubeClient.CoreV1().ConfigMaps(metadataNamespace).Update(context.TODO(), kymaMetadataCM, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func metadataToCM(data *KymaMetadata) map[string]string {
	CMData := make(map[string]string)
	CMData["profile"] = data.Profile
	CMData["version"] = data.Version
	CMData["clData"] = string(data.ComponentListData)
	CMData["clFormat"] = data.ComponentListFormat
	CMData["status"] = string(data.Status)
	CMData["reason"] = data.Reason

	return CMData
}

func cmToMetadata(data map[string]string) *KymaMetadata {
	return &KymaMetadata{
		Profile:             data["profile"],
		Version:             data["version"],
		ComponentListData:   []byte(data["clData"]),
		ComponentListFormat: data["clFormat"],
		Status:              StatusEnum(data["status"]),
		Reason:              data["reason"],
	}
}

func retryOperation(operation func() error) error {
	exponentialBackoff := backoff.NewExponentialBackOff()
	exponentialBackoff.InitialInterval = initialInterval
	exponentialBackoff.MaxElapsedTime = maxElapsedTime

	return backoff.Retry(operation, exponentialBackoff)
}

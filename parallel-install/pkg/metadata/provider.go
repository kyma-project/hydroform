package metadata

import (
	"context"
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

type MetadataProvider interface {
	ReadKymaMetadata() (*KymaMetadata, error)
	WriteKymaDeploymentInProgress(attr *Attributes) error
	WriteKymaDeploymentError(attr *Attributes, reason string) error
	WriteKymaDeployed(attr *Attributes) error
	WriteKymaUninstallationInProgress(attr *Attributes) error
	WriteKymaUninstallationError(attr *Attributes, reason string) error
	DeleteKymaMetadata() error
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
	meta := (&KymaMetadata{}).withAttributes(attr).withStatus(DeploymentInProgress)

	return retryOperation(func() error {
		return p.writeKymaMetadata(meta)
	})
}

func (p *Provider) WriteKymaUninstallationInProgress(attr *Attributes) error {
	meta := (&KymaMetadata{}).withAttributes(attr).withStatus(UninstallationInProgress)

	return retryOperation(func() error {
		return p.writeKymaMetadata(meta)
	})
}

func (p *Provider) WriteKymaDeploymentError(attr *Attributes, reason string) error {
	meta := (&KymaMetadata{}).withAttributes(attr).withError(DeploymentError, reason)

	return retryOperation(func() error {
		return p.writeKymaMetadata(meta)
	})
}

func (p *Provider) WriteKymaUninstallationError(attr *Attributes, reason string) error {
	meta := (&KymaMetadata{}).withAttributes(attr).withError(UninstallationError, reason)

	return retryOperation(func() error {
		return p.writeKymaMetadata(meta)
	})
}

func (p *Provider) WriteKymaDeployed(attr *Attributes) error {
	meta := (&KymaMetadata{}).withAttributes(attr).withStatus(Deployed)

	return retryOperation(func() error {
		return p.writeKymaMetadata(meta)
	})
}

//DeleteKymaMetadata removes Kyma metadata from the cluster
func (p *Provider) DeleteKymaMetadata() error {
	err := p.kubeClient.CoreV1().ConfigMaps(metadataNamespace).Delete(context.TODO(), metadataName, metav1.DeleteOptions{})
	return err
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
	CMData["clFile"] = data.ComponentListFile
	CMData["status"] = string(data.Status)
	CMData["reason"] = data.Reason

	return CMData
}

func cmToMetadata(data map[string]string) *KymaMetadata {
	return &KymaMetadata{
		Profile:           data["profile"],
		Version:           data["version"],
		ComponentListData: []byte(data["clData"]),
		ComponentListFile: data["clFile"],
		Status:            StatusEnum(data["status"]),
		Reason:            data["reason"],
	}
}

func retryOperation(operation func() error) error {
	exponentialBackoff := backoff.NewExponentialBackOff()
	exponentialBackoff.InitialInterval = initialInterval
	exponentialBackoff.MaxElapsedTime = maxElapsedTime

	return backoff.Retry(operation, exponentialBackoff)
}

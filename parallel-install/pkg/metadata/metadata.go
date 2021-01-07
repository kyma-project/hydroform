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

//TODO it's in default ns cause at starting installation kyma-installer ns doesn't exist, this needs to be changed and metadata should be written to protected ns
var metadataNamespace = "default"

var initialInterval = time.Duration(3) * time.Second
var maxElapsedTime = time.Duration(30) * time.Second

type MetadataProvider interface {
	ReadKymaMetadata() (*KymaMetadata, error)
	WriteKymaDeploymentInProgress() error
	WriteKymaDeploymentError(error string) error
	WriteKymaDeployed() error
	WriteKymaUninstallationInProgress() error
	WriteKymaUninstallationError(error string) error
	DeleteKymaMetadata() error
}

type KymaMetadata struct {
	Profile string
	Version string
	//TODO enum needed
	Status string
	Reason string
}

type Provider struct {
	kubeClient kubernetes.Interface
	profile    string
	version    string
}

func New(client kubernetes.Interface, profile, version string) MetadataProvider {
	return &Provider{
		kubeClient: client,
		profile:    profile,
		version:    version,
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

func (p *Provider) WriteKymaDeploymentInProgress() error {
	meta := &KymaMetadata{
		Version: p.version,
		Profile: p.profile,
		Status:  "Deployment in progress",
	}

	return retryOperation(func() error {
		return p.writeKymaMetadata(meta)
	})
}

func (p *Provider) WriteKymaUninstallationInProgress() error {
	meta := &KymaMetadata{
		Version: p.version,
		Profile: p.profile,
		Status:  "Uninstallation in progress",
	}

	return retryOperation(func() error {
		return p.writeKymaMetadata(meta)
	})
}

func (p *Provider) WriteKymaDeploymentError(reason string) error {
	meta := &KymaMetadata{
		Version: p.version,
		Profile: p.profile,
		Status:  "Deployment error",
		Reason:  reason,
	}

	return retryOperation(func() error {
		return p.writeKymaMetadata(meta)
	})
}

func (p *Provider) WriteKymaUninstallationError(reason string) error {
	meta := &KymaMetadata{
		Version: p.version,
		Profile: p.profile,
		Status:  "Uninstallation error",
		Reason:  reason,
	}

	return retryOperation(func() error {
		return p.writeKymaMetadata(meta)
	})
}

func (p *Provider) WriteKymaDeployed() error {
	meta := &KymaMetadata{
		Version: p.version,
		Profile: p.profile,
		Status:  "Deployed",
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

func (p *Provider) DeleteKymaMetadata() error {
	return retryOperation(func() error {
		err := p.kubeClient.CoreV1().ConfigMaps(metadataNamespace).Delete(context.TODO(), metadataName, metav1.DeleteOptions{})
		if err != nil {
			return err
		}
		return nil
	})
}

func metadataToCM(data *KymaMetadata) map[string]string {
	CMData := make(map[string]string)
	CMData["profile"] = data.Profile
	CMData["version"] = data.Version
	CMData["status"] = data.Status
	CMData["reason"] = data.Reason

	return CMData
}

func cmToMetadata(data map[string]string) *KymaMetadata {
	return &KymaMetadata{
		Profile: data["profile"],
		Version: data["version"],
		Status:  data["status"],
		Reason:  data["reason"],
	}
}

func retryOperation(operation func() error) error {
	exponentialBackoff := backoff.NewExponentialBackOff()
	exponentialBackoff.InitialInterval = initialInterval
	exponentialBackoff.MaxElapsedTime = maxElapsedTime

	return backoff.Retry(operation, exponentialBackoff)
}

package metadata

import (
	"context"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var metadataName = "kyma"
//TODO it's in default ns cause at starting installation kyma-installer ns doesn't exist, this needs to be changed and metadata should be written to protected ns
var metadataNamespace = "default"

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
	//TODO retries
	kymaMetadataCM, err := p.kubeClient.CoreV1().ConfigMaps(metadataNamespace).Get(context.TODO(), metadataName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return &KymaMetadata{}, nil
		}
		return nil, err
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

	return p.writeKymaMetadata(meta)
}

func (p *Provider) WriteKymaUninstallationInProgress() error {
	meta := &KymaMetadata{
		Version: p.version,
		Profile: p.profile,
		Status:  "Uninstallation in progress",
	}

	return p.writeKymaMetadata(meta)
}

func (p *Provider) WriteKymaDeploymentError(error string) error {
	meta := &KymaMetadata{
		Version: p.version,
		Profile: p.profile,
		Status:  "Deployment error",
		Reason:  error,
	}

	return p.writeKymaMetadata(meta)
}

func (p *Provider) WriteKymaUninstallationError(error string) error {
	meta := &KymaMetadata{
		Version: p.version,
		Profile: p.profile,
		Status:  "Uninstallation error",
		Reason:  error,
	}

	return p.writeKymaMetadata(meta)
}

func (p *Provider) WriteKymaDeployed() error {
	meta := &KymaMetadata{
		Version: p.version,
		Profile: p.profile,
		Status:  "Deployed",
	}

	return p.writeKymaMetadata(meta)
}

func (p *Provider) writeKymaMetadata(data *KymaMetadata) error {
	cmData := metadataToCM(data)

	//TODO retries
	kymaMetadataCM, err := p.kubeClient.CoreV1().ConfigMaps(metadataNamespace).Get(context.TODO(), metadataName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			//TODO save CM
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

	//TODO update CM
	kymaMetadataCM.Data = cmData
	_, err = p.kubeClient.CoreV1().ConfigMaps(metadataNamespace).Update(context.TODO(), kymaMetadataCM, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (p *Provider) DeleteKymaMetadata() error {
	//TODO retries
	err := p.kubeClient.CoreV1().ConfigMaps(metadataNamespace).Delete(context.TODO(), metadataName, metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	return nil
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

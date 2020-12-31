package metadata

import (
	"context"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type MetadataProvider interface {
	ReadKymaMetadata() (*KymaMetadata, error)
	WriteKKymaMetadata(data *KymaMetadata) error
	DeleteKymaMetadata() error
}

type KymaMetadata struct {
	Profile string
	Version string
	//TODO enum needed
	Status  string
	Reason  string
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
	//TODO retries
	//TODO it's in default ns cause at starting installation kyma-installer ns doesn't exist, this needs to be changed and metadata should be written to protected ns
	kymaMetadataCM, err := p.kubeClient.CoreV1().ConfigMaps("default").Get(context.TODO(), "kyma", metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return &KymaMetadata{}, nil
		}
		return nil, err
	}

	kymaMetaData := cmToMetadata(kymaMetadataCM.Data)

	return kymaMetaData, nil
}

func (p *Provider) WriteKKymaMetadata(data *KymaMetadata) error {
	cmData := metadataToCM(data)

	//TODO retries
	kymaMetadataCM, err := p.kubeClient.CoreV1().ConfigMaps("default").Get(context.TODO(), "kyma", metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			//TODO save CM
			cmToSave := &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "kyma",
					Namespace: "default",
				},
				Data: cmData,
			}

			_, err := p.kubeClient.CoreV1().ConfigMaps("default").Create(context.TODO(), cmToSave, metav1.CreateOptions{})
			if err != nil {
				return err
			}
			return nil
		}
		return err
	}

	//TODO update CM
	kymaMetadataCM.Data = cmData
	_, err = p.kubeClient.CoreV1().ConfigMaps("default").Update(context.TODO(), kymaMetadataCM, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (p *Provider) DeleteKymaMetadata() error {
	//TODO retries
	err := p.kubeClient.CoreV1().ConfigMaps("default").Delete(context.TODO(), "kyma", metav1.DeleteOptions{})
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

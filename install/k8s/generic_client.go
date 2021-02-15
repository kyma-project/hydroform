package k8s

import (
	"context"
	"fmt"
	"github.com/kyma-incubator/hydroform/install/util"

	"time"

	"k8s.io/apimachinery/pkg/api/meta"

	"k8s.io/client-go/kubernetes"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	corev1Client "k8s.io/client-go/kubernetes/typed/core/v1"
)

const (
	OnConflictLabel   string = "on-conflict"
	ReplaceOnConflict string = "replace"
	MergeOnConflict   string = "merge"
)

//go:generate mockery -name=RESTMapper
type RESTMapper interface {
	RESTMapping(gk schema.GroupKind, versions ...string) (*meta.RESTMapping, error)
}

func NewGenericClient(restMapper RESTMapper, dynamicClient dynamic.Interface, k8sClientSet kubernetes.Interface) *GenericClient {
	return &GenericClient{
		restMapper:    restMapper,
		k8sClientSet:  k8sClientSet,
		dynamicClient: dynamicClient,
		coreClient:    k8sClientSet.CoreV1(),
	}
}

type GenericClient struct {
	restMapper    RESTMapper
	k8sClientSet  kubernetes.Interface
	dynamicClient dynamic.Interface
	coreClient    corev1Client.CoreV1Interface
}

func (c GenericClient) WaitForPodByLabel(namespace, labelSelector string, desiredPhase corev1.PodPhase, timeout, checkInterval time.Duration) error {
	return util.WaitFor(checkInterval, timeout, func() (bool, error) {
		pods, err := c.coreClient.Pods(namespace).List(context.Background(), v1.ListOptions{LabelSelector: labelSelector})
		if err != nil {
			return false, err
		}

		// pod does not exists, retry
		if len(pods.Items) == 0 {
			return false, nil
		}

		ok := true
		for _, pod := range pods.Items {
			// if any pod is not in the desired status no need to check further
			if desiredPhase != pod.Status.Phase {
				ok = false
				break
			}
		}

		return ok, nil
	})
}

func (c GenericClient) ApplyConfigMaps(configMaps []*corev1.ConfigMap, namespace string) error {
	client := c.coreClient.ConfigMaps(namespace)

	for _, cm := range configMaps {
		_, err := client.Create(context.Background(), cm, v1.CreateOptions{})
		if err != nil {
			if k8serrors.IsAlreadyExists(err) {
				err = c.updateConfigMap(client, cm)
				if err != nil {
					return fmt.Errorf("config map %s already exists, failed to updated config map: %s", cm.Name, err.Error())
				}
				continue
			}
			return fmt.Errorf("failed to apply %s config map: %s", cm.Name, err.Error())
		}
	}
	return nil
}

func (c GenericClient) updateConfigMap(client corev1Client.ConfigMapInterface, cm *corev1.ConfigMap) error {
	oldCM, err := client.Get(context.Background(), cm.Name, v1.GetOptions{})
	if err != nil {
		return err
	}

	if isMerge(cm.Labels) {
		cm.Data = util.MergeStringMaps(oldCM.Data, cm.Data)
	}

	_, err = client.Update(context.Background(), cm, v1.UpdateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (c GenericClient) ApplySecrets(secrets []*corev1.Secret, namespace string) error {
	client := c.coreClient.Secrets(namespace)

	for _, sec := range secrets {
		_, err := client.Create(context.Background(), sec, v1.CreateOptions{})
		if err != nil {
			if k8serrors.IsAlreadyExists(err) {
				err = c.updateSecret(client, sec)
				if err != nil {
					return fmt.Errorf("secret %s already exists, failed to updated secret: %s", sec.Name, err.Error())
				}
				continue
			}
			return fmt.Errorf("failed to apply %s secret: %s", sec.Name, err.Error())
		}
	}

	return nil
}

func (c GenericClient) updateSecret(client corev1Client.SecretInterface, cm *corev1.Secret) error {
	oldCM, err := client.Get(context.Background(), cm.Name, v1.GetOptions{})
	if err != nil {
		return err
	}

	if isMerge(cm.Labels) {
		cm.Data = util.MergeByteMaps(oldCM.Data, cm.Data)
	}

	_, err = client.Update(context.Background(), cm, v1.UpdateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func isMerge(labels map[string]string) bool {
	if labels == nil {
		return true
	}

	val, ok := labels[OnConflictLabel]
	return !ok || val != ReplaceOnConflict
}

func (c GenericClient) CreateResources(resources []K8sObject) ([]*unstructured.Unstructured, error) {
	return c.createResources(resources, c.createObject)
}

func (c GenericClient) ApplyResources(resources []K8sObject) ([]*unstructured.Unstructured, error) {
	return c.createResources(resources, c.applyObject)
}

func (c GenericClient) createResources(resources []K8sObject,
	createObjectFunction func(dynamic.ResourceInterface, *unstructured.Unstructured) (*unstructured.Unstructured, error)) ([]*unstructured.Unstructured, error) {
	var createdResources []*unstructured.Unstructured
	for _, resource := range resources {
		unstructuredObjRaw, err := runtime.DefaultUnstructuredConverter.ToUnstructured(resource.Object)
		if err != nil {
			return nil, err
		}

		unstructuredObject := &unstructured.Unstructured{Object: unstructuredObjRaw}

		client, err := c.clientForResource(unstructuredObject, resource.GVK)
		if err != nil {
			return nil, err
		}
		created, err := createObjectFunction(client, unstructuredObject)
		if err != nil {
			return nil, fmt.Errorf("failed to apply resource: %s", err.Error())
		}
		createdResources = append(createdResources, created)
	}

	return createdResources, nil
}

func (c GenericClient) createObject(client dynamic.ResourceInterface, unstructuredObject *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	created, err := client.Create(context.Background(), unstructuredObject, v1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create object %s of kind %s: %s", unstructuredObject.GetName(), unstructuredObject.GetKind(), err.Error())
	}

	return created, nil
}

func (c GenericClient) applyObject(client dynamic.ResourceInterface, unstructuredObject *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	created, err := client.Create(context.Background(), unstructuredObject, v1.CreateOptions{})
	if err != nil {
		if k8serrors.IsAlreadyExists(err) {
			updated, err := c.updateObject(client, unstructuredObject)
			if err != nil {
				return nil, fmt.Errorf("failed to create update %s of kind %s: %s", unstructuredObject.GetName(), unstructuredObject.GetKind(), err.Error())
			}
			return updated, nil
		}
		return nil, fmt.Errorf("failed to create object %s of kind %s: %s", unstructuredObject.GetName(), unstructuredObject.GetKind(), err.Error())
	}

	return created, nil
}

func (c GenericClient) updateObject(client dynamic.ResourceInterface, unstructuredObject *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	get, err := client.Get(context.Background(), unstructuredObject.GetName(), v1.GetOptions{})

	if err != nil {
		return nil, err
	}

	merged := util.MergeMaps(unstructuredObject.Object, get.Object)

	newObject := &unstructured.Unstructured{Object: merged}

	return client.Update(context.Background(), newObject, v1.UpdateOptions{})
}

func (c GenericClient) clientForResource(unstructuredObject *unstructured.Unstructured, gvk *schema.GroupVersionKind) (dynamic.ResourceInterface, error) {
	versionResource := schema.GroupVersionResource{Group: gvk.Group, Version: gvk.Version}

	restMapping, err := c.restMapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, fmt.Errorf("failed to map resource: %s", err.Error())
	}

	versionResource = restMapping.Resource

	namespace := unstructuredObject.GetNamespace()
	if namespace == "" {
		return c.dynamicClient.Resource(versionResource), nil
	}

	return c.dynamicClient.Resource(versionResource).Namespace(namespace), nil
}

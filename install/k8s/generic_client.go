package k8s

import (
	"fmt"

	"github.com/kyma-incubator/hydroform/install/util"

	"time"

	"k8s.io/apimachinery/pkg/api/meta"

	"k8s.io/client-go/kubernetes"

	installationClientset "github.com/kyma-project/kyma/components/kyma-operator/pkg/client/clientset/versioned"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	corev1Client "k8s.io/client-go/kubernetes/typed/core/v1"
)

//go:generate mockery -name=RESTMapper
type RESTMapper interface {
	RESTMapping(gk schema.GroupKind, versions ...string) (*meta.RESTMapping, error)
}

func NewGenericClient(restMapper RESTMapper, dynamicClient dynamic.Interface, k8sClientSet kubernetes.Interface, installationClientSet installationClientset.Interface) *GenericClient {
	return &GenericClient{
		restMapper:            restMapper,
		k8sClientSet:          k8sClientSet,
		installationClientSet: installationClientSet,
		dynamicClient:         dynamicClient,
		coreClient:            k8sClientSet.CoreV1(),
	}
}

type GenericClient struct {
	restMapper            RESTMapper
	k8sClientSet          kubernetes.Interface
	installationClientSet installationClientset.Interface
	dynamicClient         dynamic.Interface
	coreClient            corev1Client.CoreV1Interface
}

func (c GenericClient) WaitForPodByLabel(namespace, labelSelector string, desiredPhase corev1.PodPhase, timeout, checkInterval time.Duration) error {
	return util.WaitFor(checkInterval, timeout, func() (bool, error) {
		pods, err := c.coreClient.Pods(namespace).List(metav1.ListOptions{LabelSelector: labelSelector})
		if err != nil {
			return false, err
		}
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
		_, err := client.Create(cm)
		if err != nil {
			if k8serrors.IsAlreadyExists(err) {
				_, err := client.Update(cm)
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

func (c GenericClient) ApplySecrets(secrets []*corev1.Secret, namespace string) error {
	client := c.coreClient.Secrets(namespace)

	for _, sec := range secrets {
		_, err := client.Create(sec)
		if err != nil {
			if k8serrors.IsAlreadyExists(err) {
				_, err := client.Update(sec)
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

func (c GenericClient) ApplyResources(resources []K8sObject) error {
	for _, resource := range resources {
		unstructuredObjRaw, err := runtime.DefaultUnstructuredConverter.ToUnstructured(resource.Object)
		if err != nil {
			return err
		}

		unstructuredObject := &unstructured.Unstructured{Object: unstructuredObjRaw}

		client, err := c.clientForResource(unstructuredObject, resource.GVK)
		if err != nil {
			return err
		}
		err = c.applyObject(client, unstructuredObject)
		if err != nil {
			return fmt.Errorf("failed to apply resource: %s", err.Error())
		}
	}

	return nil
}

func (c GenericClient) applyObject(client dynamic.ResourceInterface, unstructuredObject *unstructured.Unstructured) error {
	_, err := client.Create(unstructuredObject, v1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create object %s of kind %s: %s", unstructuredObject.GetName(), unstructuredObject.GetKind(), err.Error())
	}

	return nil
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

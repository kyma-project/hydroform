package preinstaller

import (
	"context"
	"github.com/avast/retry-go"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/logger"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
)

//go:generate mockery --name ResourceManager

// ResourceManager manages resources on a k8s cluster.
type ResourceManager interface {
	// CreateResource of any type that matches the schema on k8s cluster.
	// Performs retries on unsuccessful resource creation action.
	CreateResource(resource *unstructured.Unstructured, resourceSchema schema.GroupVersionResource) error

	// GetResource of a given fileName from a k8s cluster, that matches the schema.
	// Performs retries on unsuccessful resource retrieval action.
	GetResource(resourceName string, resourceSchema schema.GroupVersionResource) (*unstructured.Unstructured, error)

	// UpdateResource of a given fileName from a k8s cluster, that matches the schema.
	// Performs retries on unsuccessful resource update action.
	UpdateResource(resource *unstructured.Unstructured, resourceSchema schema.GroupVersionResource) (*unstructured.Unstructured, error)
}

// DefaultResourceManager provides a default implementation of ResourceManager.
type DefaultResourceManager struct {
	dynamicClient dynamic.Interface
	log           logger.Interface
	retryOptions  []retry.Option
}

// NewResourceManager creates a new instance of ResourceManager.
func NewDefaultResourceManager(kubeconfigPath string, log logger.Interface, retryOptions []retry.Option) (*DefaultResourceManager, error) {
	// TODO
	//kubeConfigManager, err := config.NewKubeConfigManager(&kubeconfigPath, nil)
	//if err != nil {
	//	return nil, err
	//}

	restConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, err
	}

	dynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	return &DefaultResourceManager{
		dynamicClient: dynamicClient,
		log:           log,
		retryOptions:  retryOptions,
	}, nil
}

func (c *DefaultResourceManager) CreateResource(resource *unstructured.Unstructured, resourceSchema schema.GroupVersionResource) error {
	var err error
	err = retry.Do(func() error {
		if _, err = c.dynamicClient.Resource(resourceSchema).Create(context.TODO(), resource, metav1.CreateOptions{}); err != nil {
			c.log.Errorf("Error occurred during resource create: %s", err.Error())
			return err
		}

		return nil
	}, c.retryOptions...)

	if err != nil {
		return err
	}

	return nil
}

func (c *DefaultResourceManager) GetResource(resourceName string, resourceSchema schema.GroupVersionResource) (obj *unstructured.Unstructured, err error) {
	err = retry.Do(func() error {
		obj, err = c.getResource(resourceName, resourceSchema)
		if err != nil {
			if apierrors.IsNotFound(err) {
				c.log.Infof("Resource %s was not found.", resourceName)
				return nil
			}
			c.log.Errorf("Error occurred during resource get: %s", err.Error())
			return err
		}

		return err

	}, c.retryOptions...)

	if err != nil {
		return nil, err
	}

	return obj, nil
}

func (c *DefaultResourceManager) UpdateResource(resource *unstructured.Unstructured, resourceSchema schema.GroupVersionResource) (obj *unstructured.Unstructured, err error) {
	err = retry.Do(func() error {
		latestResource, err := c.getResource(resource.GetName(), resourceSchema)
		if err != nil {
			return err
		}

		resource.SetResourceVersion(latestResource.GetResourceVersion())
		obj, err = c.updateResource(resource, resourceSchema)
		if err != nil {
			c.log.Errorf("Error occurred during resource update: %s", err.Error())
			return err
		}

		return nil
	}, c.retryOptions...)

	if err != nil {
		return nil, err
	}

	return obj, nil
}

func (c *DefaultResourceManager) getResource(resourceName string, resourceSchema schema.GroupVersionResource) (*unstructured.Unstructured, error) {
	return c.dynamicClient.Resource(resourceSchema).Get(context.TODO(), resourceName, metav1.GetOptions{})
}

func (c *DefaultResourceManager) createResource(resource *unstructured.Unstructured, resourceSchema schema.GroupVersionResource) (*unstructured.Unstructured, error) {
	return c.dynamicClient.Resource(resourceSchema).Update(context.TODO(), resource, metav1.UpdateOptions{})
}

func (c *DefaultResourceManager) updateResource(resource *unstructured.Unstructured, resourceSchema schema.GroupVersionResource) (*unstructured.Unstructured, error) {
	return c.dynamicClient.Resource(resourceSchema).Update(context.TODO(), resource, metav1.UpdateOptions{})
}

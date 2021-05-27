package deployment

import (
	"context"
	"fmt"
	"strings"

	"github.com/avast/retry-go"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/config"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/logger"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

//go:generate mockery --name ResourceManager

// ResourceManager manages resources on a k8s cluster.
type ResourceManager interface {
	// CreateResource from a given object and schema on k8s cluster.
	// Performs retries on unsuccessful resource creation action.
	CreateResource(resource *unstructured.Unstructured, gvk schema.GroupVersionKind, opts metav1.CreateOptions) error

	// GetResource of a given name from a k8s cluster, that matches the schema.
	// Performs retries on unsuccessful resource retrieval action.
	GetResource(resourceName string, gvk schema.GroupVersionKind, opts metav1.GetOptions) (*unstructured.Unstructured, error)

	// UpdateResource on a k8s cluster, that matches the schema.
	// Performs retries on unsuccessful resource update action.
	UpdateResource(resource *unstructured.Unstructured, gvk schema.GroupVersionKind, opts metav1.UpdateOptions) (*unstructured.Unstructured, error)

	// DeleteCollectionOfResources from a k8s cluster, that match the schema, using query options for filtering them.
	// Performs retries on unsuccessful resources delete action.
	DeleteCollectionOfResources(gvk schema.GroupVersionKind, opts metav1.DeleteOptions, listOps metav1.ListOptions) error
}

// DefaultResourceManager provides a default implementation of ResourceManager.
type DefaultResourceManager struct {
	dynamicClient dynamic.Interface
	log           logger.Interface
	retryOptions  []retry.Option
}

// NewDefaultResourceManager creates a new instance of ResourceManager.
func NewDefaultResourceManager(kubeconfigSource config.KubeconfigSource, log logger.Interface, retryOptions []retry.Option) (*DefaultResourceManager, error) {
	restConfig, err := config.RestConfig(kubeconfigSource)
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

func (c *DefaultResourceManager) CreateResource(resource *unstructured.Unstructured, gvk schema.GroupVersionKind, opts metav1.CreateOptions) error {
	var err error
	err = retry.Do(func() error {
		if _, err = c.dynamicClient.Resource(retrieveGvrFrom(gvk)).Create(context.TODO(), resource, opts); err != nil {
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

func (c *DefaultResourceManager) GetResource(resourceName string, gvk schema.GroupVersionKind, opts metav1.GetOptions) (obj *unstructured.Unstructured, err error) {
	err = retry.Do(func() error {
		obj, err = c.dynamicClient.Resource(retrieveGvrFrom(gvk)).Get(context.TODO(), resourceName, opts)
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

func (c *DefaultResourceManager) UpdateResource(resource *unstructured.Unstructured, gvk schema.GroupVersionKind, opts metav1.UpdateOptions) (obj *unstructured.Unstructured, err error) {
	gvr := retrieveGvrFrom(gvk)

	err = retry.Do(func() error {
		latestResource, err := c.dynamicClient.Resource(gvr).Get(context.TODO(), resource.GetName(), metav1.GetOptions{})
		if err != nil {
			return err
		}

		resource.SetResourceVersion(latestResource.GetResourceVersion())
		obj, err = c.dynamicClient.Resource(gvr).Update(context.TODO(), resource, opts)
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

func (c *DefaultResourceManager) DeleteCollectionOfResources(gvk schema.GroupVersionKind, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	var err error
	err = retry.Do(func() error {
		if err = c.dynamicClient.Resource(retrieveGvrFrom(gvk)).DeleteCollection(context.TODO(), opts, listOpts); err != nil {
			c.log.Errorf("Error occurred during resources delete: %s", err.Error())
			return err
		}

		return nil
	}, c.retryOptions...)

	if err != nil {
		return err
	}

	return nil
}

func retrieveGvrFrom(gvk schema.GroupVersionKind) schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    gvk.Group,
		Version:  gvk.Version,
		Resource: pluralForm(gvk.Kind),
	}
}

func pluralForm(singular string) string {
	return fmt.Sprintf("%ss", strings.ToLower(singular))
}

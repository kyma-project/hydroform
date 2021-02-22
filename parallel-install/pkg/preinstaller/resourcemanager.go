package preinstaller

import (
	"context"
	"github.com/avast/retry-go"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

type resourceManager struct {
	dynamicClient dynamic.Interface
	retryOptions []retry.Option
}

func newResourceManager(dynamicClient dynamic.Interface, retryOptions []retry.Option) *resourceManager {
	return &resourceManager{
		dynamicClient: dynamicClient,
		retryOptions: retryOptions,
	}
}

func (c *resourceManager) createResource(resource *unstructured.Unstructured, resourceSchema schema.GroupVersionResource) error {
	var err error
	err = retry.Do(func() error {
		if _, err = c.dynamicClient.Resource(resourceSchema).Create(context.TODO(), resource, metav1.CreateOptions{}); err != nil {
			return err
		}

		return nil
	}, c.retryOptions...)

	return err
}

func (c *resourceManager) getResource(resourceName string, resourceSchema schema.GroupVersionResource) (*unstructured.Unstructured, error) {
	var obj *unstructured.Unstructured
	err := retry.Do(func() error {
		var err error

		if obj, err = c.dynamicClient.Resource(resourceSchema).Get(context.TODO(), resourceName, metav1.GetOptions{}); err != nil {
			return err
		}

		return nil
	}, c.retryOptions...)

	return obj, err
}

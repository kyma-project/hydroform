package preinstaller

import (
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/logger"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

//go:generate mockery --name ResourceApplier

// ResourceApplier creates a new resource from an object on k8s cluster.
type ResourceApplier interface {
	// Apply passed resource object on a k8s cluster.
	Apply(resource *unstructured.Unstructured) error
}

// GenericResourceApplier is a default implementation of ResourceApplier.
type GenericResourceApplier struct {
	log             logger.Interface
	resourceManager ResourceManager
}

// NewGenericResourceApplier returns a new instance of GenericResourceApplier.
func NewGenericResourceApplier(log logger.Interface, resourceManager ResourceManager) *GenericResourceApplier {
	return &GenericResourceApplier{
		log:             log,
		resourceManager: resourceManager,
	}
}

func (c *GenericResourceApplier) Apply(resource *unstructured.Unstructured) error {
	if resource == nil {
		return errors.New("Could not apply not existing resource")
	}

	gvk := resource.GroupVersionKind()
	resourceName := resource.GetName()
	obj, err := c.resourceManager.GetResource(resourceName, gvk)
	if err != nil {
		return err
	}

	if obj != nil {
		c.log.Infof("Resource: %s already exists. Performing update.", resourceName)

		_, err = c.resourceManager.UpdateResource(resource, gvk)
		if err != nil {
			return err
		}
	} else {
		c.log.Infof("Creating resource: %s.", resourceName)

		err = c.resourceManager.CreateResource(resource, gvk)
		if err != nil {
			return err
		}
	}

	return nil
}

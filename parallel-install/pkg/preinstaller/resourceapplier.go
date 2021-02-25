package preinstaller

import (
	"encoding/json"
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/logger"
	"github.com/pkg/errors"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"strings"
)

// ResourceApplier creates a new resource from manifest on k8s cluster.
type ResourceApplier interface {
	// Apply parses passed manifest and applies it on a k8s cluster.
	Apply(manifest string) error
}

// GenericResourceApplier is a default implementation of ResourceApplier.
type GenericResourceApplier struct {
	log             logger.Interface
	decoder         runtime.Decoder
	resourceManager ResourceManager
}

// NewGenericResourceApplier returns a new instance of GenericResourceApplier.
func NewGenericResourceApplier(log logger.Interface, resourceManager ResourceManager) *GenericResourceApplier {
	return &GenericResourceApplier{
		log:             log,
		decoder:         initializeDecoder(),
		resourceManager: resourceManager,
	}
}

func (c *GenericResourceApplier) Apply(manifest string) error {
	resource, err := c.parseResourceFrom(manifest)
	if err != nil {
		return err
	}

	gvk := resource.GroupVersionKind()
	resourceSchema := schema.GroupVersionResource{
		Group:    gvk.Group,
		Version:  gvk.Version,
		Resource: pluralForm(gvk.Kind),
	}

	resourceName := resource.GetName()
	obj, err := c.resourceManager.GetResource(resourceName, resourceSchema)
	if err != nil {
		return err
	}

	if obj != nil {
		c.log.Infof("Resource: %s already exists. Performing update.", resourceName)

		err = c.resourceManager.UpdateRefreshableResource(obj, resourceSchema)
		if err != nil {
			return err
		}
	} else {
		c.log.Infof("Creating resource: %s.", resourceName)

		err = c.resourceManager.CreateResource(resource, resourceSchema)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *GenericResourceApplier) parseResourceFrom(manifest string) (*unstructured.Unstructured, error) {
	var _, _, err = c.decoder.Decode([]byte(manifest), nil, nil)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Could not decode the resource file due to the following error: %s.", err.Error()))
	}

	converted, err := convertYamlToJson(manifest)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Could not convert the resource file to JSON due to the following error: %s.", err.Error()))
	}

	resource, err := parseManifest([]byte(converted))
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Could not parse the resource file due to the following error: %s.", err.Error()))
	}

	return resource, nil
}

func convertYamlToJson(input string) (string, error) {
	convertedInput, err := yaml.YAMLToJSON([]byte(input))
	if err != nil {
		return "", err
	}
	if string(convertedInput) != "null" {
		return string(convertedInput), nil
	}

	return "", err
}

func parseManifest(input []byte) (*unstructured.Unstructured, error) {
	var middleware map[string]interface{}
	err := json.Unmarshal(input, &middleware)
	if err != nil {
		return nil, err
	}

	resource := &unstructured.Unstructured{
		Object: middleware,
	}
	return resource, nil
}

func pluralForm(name string) string {
	return fmt.Sprintf("%ss", strings.ToLower(name))
}

func initializeDecoder() runtime.Decoder {
	sch := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(sch)
	_ = apiextv1.AddToScheme(sch)

	return serializer.NewCodecFactory(sch).UniversalDeserializer()
}

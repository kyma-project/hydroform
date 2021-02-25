package preinstaller

import (
	"encoding/json"
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/logger"
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
	Apply(manifest string) (bool, error)
}

type resourceType struct {
	name string
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

func (c *GenericResourceApplier) Apply(manifest string) (bool, error) {
	var _, grpVerKind, err = c.decoder.Decode([]byte(manifest), nil, nil)
	if err != nil {
		c.log.Warn(fmt.Sprintf("%s%s", "Could not decode the resource file due to the following error: %s. Skipping.", err.Error()))
		return false, nil
	}

	converted, err := convertYamlToJson(manifest)
	if err != nil {
		c.log.Warn(fmt.Sprintf("%s%s", "Could not convert the resource file to JSON due to the following error: %s. Skipping.", err.Error()))
		return false, nil
	}

	resource, err := parseManifest([]byte(converted))
	if err != nil {
		c.log.Warn(fmt.Sprintf("%s%s", "Could not parse the resource file due to the following error: %s. Skipping.", err.Error()))
		return false, nil
	}

	resourceSchema := schema.GroupVersionResource{
		Group:    grpVerKind.Group,
		Version:  grpVerKind.Version,
		Resource: pluralForm(grpVerKind.Kind),
	}

	resourceName := resource.GetName()
	obj, err := c.resourceManager.GetResource(resourceName, resourceSchema)
	if err != nil {
		return false, err
	}

	if obj != nil {
		c.log.Infof("Resource: %s already exists. Performing update.", resourceName)

		err = c.resourceManager.UpdateRefreshableResource(obj, resourceSchema)
		if err != nil {
			return false, err
		}
	} else {
		c.log.Infof("Creating resource: %s .", resourceName)

		err = c.resourceManager.CreateResource(resource, resourceSchema)
		if err != nil {
			return false, err
		}
	}

	return true, nil
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

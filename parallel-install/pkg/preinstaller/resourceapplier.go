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

type ResourceApplier interface {
	Apply(manifest string) error
}

type resourceType struct {
	name    string
}

type GenericResourceApplier struct {
	log             logger.Interface
	decoder         runtime.Decoder
	resourceManager ResourceManager
}

func NewGenericResourceApplier(log logger.Interface, resourceManager ResourceManager) *GenericResourceApplier {
	return &GenericResourceApplier{
		log:             log,
		decoder:        initializeDecoder(),
		resourceManager: resourceManager,
	}
}

func (c *GenericResourceApplier) Apply(manifest string) error {
	var _, grpVerKind, err = c.decoder.Decode([]byte(manifest), nil, nil)
	if err != nil {
		c.log.Warn("Could not parse the resource file. Skipping.")
		return nil
	}

	converted, err := convertYamlToJson(manifest)
	if err != nil {
		return err
	}

	resource, err := parseManifest([]byte(converted))
	if err != nil {
		return err
	}

	resourceSchema := schema.GroupVersionResource{
		Group:    grpVerKind.Group,
		Version:  grpVerKind.Version,
		Resource: pluralForm(grpVerKind.Kind),
	}

	resourceName := resource.GetName()
	obj, err := c.resourceManager.getResource(resourceName, resourceSchema)
	if err != nil {
		return err
	}

	if obj != nil {
		c.log.Infof("Resource: %s already exists. Skipping.", resourceName)
		return nil
	}

	c.log.Infof("Creating resource: %s .", resourceName)

	return c.resourceManager.createResource(resource, resourceSchema)
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

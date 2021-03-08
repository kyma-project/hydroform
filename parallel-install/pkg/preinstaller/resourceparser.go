package preinstaller

import (
	"encoding/json"
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"io/ioutil"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
)

// ResourceParser parses a resource from a given input.
type ResourceParser interface {
	// ParseUnstructuredResourceFrom given path and return it.
	ParseUnstructuredResourceFrom(path string) (*unstructured.Unstructured, error)
}

// GenericResourceParser is a default implementation of ResourceParser.
type GenericResourceParser struct {
	decoder         runtime.Decoder
}

// NewGenericResourceParser returns a new instance of GenericResourceParser.
func NewGenericResourceParser() *GenericResourceParser {
	return &GenericResourceParser{
		decoder:         initializeDecoder(),
	}
}

func (c *GenericResourceParser) ParseUnstructuredResourceFrom(path string) (obj *unstructured.Unstructured, err error) {
	manifest, err := ioutil.ReadFile(path)
	if err != nil {
		return obj, err
	}

	return c.parseResourceFrom(string(manifest))
}

func (c *GenericResourceParser) parseResourceFrom(manifest string) (*unstructured.Unstructured, error) {
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

func initializeDecoder() runtime.Decoder {
	sch := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(sch)
	_ = apiextv1.AddToScheme(sch)

	return serializer.NewCodecFactory(sch).UniversalDeserializer()
}

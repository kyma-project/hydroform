package preinstaller

import (
	"encoding/json"
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"io/ioutil"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
)

//go:generate mockery --name ResourceParser

// ResourceParser parses a resource from a given input.
type ResourceParser interface {
	// ParseFile from a given path and return it.
	ParseFile(path string) (*unstructured.Unstructured, error)
}

// GenericResourceParser is a default implementation of ResourceParser.
type GenericResourceParser struct{}

func (c *GenericResourceParser) ParseFile(path string) (obj *unstructured.Unstructured, err error) {
	manifest, err := ioutil.ReadFile(path)
	if err != nil {
		return obj, err
	}

	return c.parseResourceFrom(string(manifest))
}

func (c *GenericResourceParser) parseResourceFrom(manifest string) (resource *unstructured.Unstructured, err error) {
	decoder, err := initializeDefaultDecoder()
	if err != nil {
		return nil, errors.Wrap(err, "Could not initialize decoder.")
	}

	_, _, err = decoder.Decode([]byte(manifest), nil, nil)
	if err != nil {
		return nil, errors.Wrap(err, "Could not decode the resource file.")
	}

	converted, err := convertYamlToJson(manifest)
	if err != nil {
		return nil, errors.Wrap(err, "Could not convert the resource file to JSON")
	}

	resource, err = parseManifest([]byte(converted))
	if err != nil {
		return nil, errors.Wrap(err, "Could not parse the resource file.")
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

func initializeDefaultDecoder() (runtime.Decoder, error) {
	sch := runtime.NewScheme()
	err := clientgoscheme.AddToScheme(sch)
	if err != nil {
		return nil, err
	}

	err = apiextv1.AddToScheme(sch)
	if err != nil {
		return nil, err
	}

	if err := apiextv1beta1.AddToScheme(sch); err != nil {
		return nil, err
	}

	return serializer.NewCodecFactory(sch).UniversalDeserializer(), nil
}

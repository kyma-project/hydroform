package mocks

import (
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"regexp"
)

// AllowResourceParserMock is a mock type for the ResourceApplier type.
type AllowResourceParserMock struct{}

// ParseUnstructuredResourceFrom provides a mock function wih given fields: path.
func (c *AllowResourceParserMock) ParseUnstructuredResourceFrom(path string) (*unstructured.Unstructured, error) {
	return matchResourceFor(path)
}

// MixedResourceParserMock is a mock type for the ResourceApplier type.
type MixedResourceParserMock struct{}

// ParseUnstructuredResourceFrom provides a mock function wih given fields: path.
func (c *MixedResourceParserMock) ParseUnstructuredResourceFrom(path string) (*unstructured.Unstructured, error) {
	_, err := checkIncorrectResourceFor(path)
	if err != nil {
		return nil, err
	}

	return matchResourceFor(path)
}

// DenyResourceParserMock is a mock type for the ResourceApplier type.
type DenyResourceParserMock struct{}

// ParseUnstructuredResourceFrom provides a mock function wih given fields: path.
func (c *DenyResourceParserMock) ParseUnstructuredResourceFrom(path string) (*unstructured.Unstructured, error) {
	return nil, errors.New("Parser error")
}

func matchResourceFor(path string) (*unstructured.Unstructured, error) {
	ns, err := regexp.MatchString("ns.yaml", path)
	if err != nil {
		return nil, errors.New("Parser error")
	}
	if ns {
		return fixNamespaceResourceWith("Resource"), nil
	}

	crd, err := regexp.MatchString("crd.yaml", path)
	if err != nil {
		return nil, errors.New("Parser error")
	}
	if crd {
		return fixCrdResourceWith("Resource"), nil
	}

	return &unstructured.Unstructured{}, nil
}

func checkIncorrectResourceFor(path string) (*unstructured.Unstructured, error) {
	incorrect, err := regexp.MatchString("incorrect", path)
	if err != nil || incorrect {
		return nil, errors.New("Parser error")
	}

	return &unstructured.Unstructured{}, nil
}

func fixCrdResourceWith(name string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apiextensions.k8s.io/v1",
			"kind":       "CustomResourceDefinition",
			"metadata": map[string]interface{}{
				"name": name,
			},
			"spec": map[string]interface{}{
				"group": "group",
			},
		},
	}
}

func fixNamespaceResourceWith(name string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Namespace",
			"metadata": map[string]interface{}{
				"name": name,
			},
		},
	}
}
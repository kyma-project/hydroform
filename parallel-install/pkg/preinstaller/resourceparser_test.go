package preinstaller

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"regexp"
	"testing"
)

func TestResourceParser_ParseUnstructuredResourceFrom(t *testing.T) {

	t.Run("should correctly parse CRD resource", func(t *testing.T) {
		// given
		parser := NewGenericResourceParser()
		pathToFile := fmt.Sprintf("%s%s", getTestingResourcesDirectory(), "/generic/correct/crd.yaml")

		// when
		resource, err := parser.ParseUnstructuredResourceFrom(pathToFile)

		// then
		assert.NoError(t, err)
		assert.NotNil(t, resource)

		expectedGvk := schema.GroupVersionKind{
			Group:   "apiextensions.k8s.io",
			Version: "v1",
			Kind:    "CustomResourceDefinition",
		}
		assert.Equal(t, resource.GroupVersionKind(), expectedGvk)

		expectedName := "crontabs.stable.example.com"
		assert.Equal(t, resource.GetName(), expectedName)
	})

	t.Run("should correctly parse Namespace resource", func(t *testing.T) {
		// given
		parser := NewGenericResourceParser()
		pathToFile := fmt.Sprintf("%s%s", getTestingResourcesDirectory(), "/generic/correct/ns.yaml")

		// when
		resource, err := parser.ParseUnstructuredResourceFrom(pathToFile)

		// then
		assert.NoError(t, err)
		assert.NotNil(t, resource)

		expectedGvk := schema.GroupVersionKind{
			Group:   "",
			Version: "v1",
			Kind:    "Namespace",
		}
		assert.Equal(t, resource.GroupVersionKind(), expectedGvk)

		expectedName := "namespace"
		assert.Equal(t, resource.GetName(), expectedName)
	})

	t.Run("should not parse resource due to not registered kind", func(t *testing.T) {
		// given
		parser := NewGenericResourceParser()
		pathToFile := fmt.Sprintf("%s%s", getTestingResourcesDirectory(), "/generic/incorrect/notype.yaml")

		// when
		resource, err := parser.ParseUnstructuredResourceFrom(pathToFile)

		// then
		assert.Error(t, err)
		expectedError := "no kind \"OtherType\" is registered"
		receivedError := err.Error()
		matched, err := regexp.MatchString(expectedError, receivedError)
		assert.True(t, matched, fmt.Sprintf("Expected error message: %s but got: %s", expectedError, receivedError))
		assert.Nil(t, resource)
	})

}

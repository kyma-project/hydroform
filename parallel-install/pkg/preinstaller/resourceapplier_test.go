package preinstaller

import (
	"fmt"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/logger"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/preinstaller/mocks"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"regexp"
	"testing"
)

func TestResourceApplier_Apply(t *testing.T) {

	t.Run("should not apply resource", func(t *testing.T) {
		t.Run("due to not existing resource", func(t *testing.T) {
			// given
			manager := mocks.ValidResourceManager{}
			applier := NewGenericResourceApplier(logger.NewLogger(true), &manager)

			// when
			err := applier.Apply(nil)

			// then
			assert.Error(t, err)
			expectedError := "Could not apply not existing resource"
			receivedError := err.Error()
			matched, err := regexp.MatchString(expectedError, receivedError)
			assert.True(t, matched, fmt.Sprintf("Expected error message: %s but got: %s", expectedError, receivedError))
		})

		t.Run("due to get resource error", func(t *testing.T) {
			// given
			manager := mocks.GetErrorResourceManager{}
			applier := NewGenericResourceApplier(logger.NewLogger(true), &manager)
			resource := fixResourceWith("Resource")

			// when
			err := applier.Apply(resource)

			// then
			assert.Error(t, err)
			expectedError := "Get resource error"
			receivedError := err.Error()
			matched, err := regexp.MatchString(expectedError, receivedError)
			assert.True(t, matched, fmt.Sprintf("Expected error message: %s but got: %s", expectedError, receivedError))
		})

		t.Run("due to update resource error when resource existed on a cluster", func(t *testing.T) {
			// given
			manager := mocks.UpdateErrorResourceManager{}
			applier := NewGenericResourceApplier(logger.NewLogger(true), &manager)
			resource := fixResourceWith("Resource")

			// when
			err := applier.Apply(resource)

			// then
			assert.Error(t, err)
			expectedError := "Update resource error"
			receivedError := err.Error()
			matched, err := regexp.MatchString(expectedError, receivedError)
			assert.True(t, matched, fmt.Sprintf("Expected error message: %s but got: %s", expectedError, receivedError))
		})

		t.Run("due to creation error when resource did not exist on a cluster", func(t *testing.T) {
			// given
			manager := mocks.CreateErrorResourceManager{}
			applier := NewGenericResourceApplier(logger.NewLogger(true), &manager)
			resource := fixResourceWith("Resource")

			// when
			err := applier.Apply(resource)

			// then
			assert.Error(t, err)
			expectedError := "Create resource error"
			receivedError := err.Error()
			matched, err := regexp.MatchString(expectedError, receivedError)
			assert.True(t, matched, fmt.Sprintf("Expected error message: %s but got: %s", expectedError, receivedError))
		})
	})

	t.Run("should apply CRD", func(t *testing.T) {
		// given
		manager := mocks.ValidResourceManager{}
		applier := NewGenericResourceApplier(logger.NewLogger(true), &manager)
		resource := fixCrdResourceWith("Resource")

		// when
		err := applier.Apply(resource)

		// then
		assert.NoError(t, err)
	})

	t.Run("should create namespace", func(t *testing.T) {
		// given
		manager := mocks.ValidResourceManager{}
		applier := NewGenericResourceApplier(logger.NewLogger(true), &manager)
		resource := fixNamespaceResourceWith("Resource")

		// when
		err := applier.Apply(resource)

		// then
		assert.NoError(t, err)
	})
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

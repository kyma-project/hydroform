package preinstaller

import (
	"fmt"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/logger"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/preinstaller/mocks"
	"github.com/stretchr/testify/assert"
	"regexp"
	"testing"
)

func TestResourceApplier_Apply(t *testing.T) {

	t.Run("should not apply resource", func(t *testing.T) {
		t.Run("due to get resource error", func(t *testing.T) {
			// given
			manager := mocks.GetErrorResourceManager{}
			applier := NewGenericResourceApplier(logger.NewLogger(true), &manager)
			pathToFile := fmt.Sprintf("%s%s", getTestingResourcesDirectory(), "/generic/correct/crd.yaml")

			// when
			err := applier.Apply(pathToFile)

			// then
			expectedError := "Get resource error"
			receivedError := err.Error()
			matched, err := regexp.MatchString(expectedError, receivedError)
			assert.True(t, matched, fmt.Sprintf("Expected error message: %s but got: %s", expectedError, receivedError))
		})

		t.Run("due to update resource error when resource existed on a cluster", func(t *testing.T) {
			// given
			manager := mocks.UpdateErrorResourceManager{}
			applier := NewGenericResourceApplier(logger.NewLogger(true), &manager)
			pathToFile := fmt.Sprintf("%s%s", getTestingResourcesDirectory(), "/generic/correct/crd.yaml")

			// when
			err := applier.Apply(pathToFile)

			// then
			expectedError := "Update resource error"
			receivedError := err.Error()
			matched, err := regexp.MatchString(expectedError, receivedError)
			assert.True(t, matched, fmt.Sprintf("Expected error message: %s but got: %s", expectedError, receivedError))
		})

		t.Run("due to creation error when resource did not exist on a cluster", func(t *testing.T) {
			// given
			manager := mocks.CreateErrorResourceManager{}
			applier := NewGenericResourceApplier(logger.NewLogger(true), &manager)
			pathToFile := fmt.Sprintf("%s%s", getTestingResourcesDirectory(), "/generic/correct/crd.yaml")

			// when
			err := applier.Apply(pathToFile)

			// then
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
		pathToFile := fmt.Sprintf("%s%s", getTestingResourcesDirectory(), "/generic/correct/crd.yaml")

		// when
		err := applier.Apply(pathToFile)

		// then
		assert.NoError(t, err)
	})

	t.Run("should not apply CRD", func(t *testing.T) {
		// given
		manager := mocks.ValidResourceManager{}
		applier := NewGenericResourceApplier(logger.NewLogger(true), &manager)
		pathToFile := fmt.Sprintf("%s%s", getTestingResourcesDirectory(), "/generic/incorrect/crd.yaml")

		// when
		err := applier.Apply(pathToFile)

		// then
		expectedError := "Could not decode the resource file"
		receivedError := err.Error()
		matched, err := regexp.MatchString(expectedError, receivedError)
		assert.True(t, matched, fmt.Sprintf("Expected error message: %s but got: %s", expectedError, receivedError))
	})

	t.Run("should create namespace", func(t *testing.T) {
		// given
		manager := mocks.ValidResourceManager{}
		applier := NewGenericResourceApplier(logger.NewLogger(true), &manager)
		pathToFile := fmt.Sprintf("%s%s", getTestingResourcesDirectory(), "/generic/correct/ns.yaml")

		// when
		err := applier.Apply(pathToFile)

		// then
		assert.NoError(t, err)
	})

	t.Run("should not create namespace", func(t *testing.T) {
		// given
		manager := mocks.ValidResourceManager{}
		applier := NewGenericResourceApplier(logger.NewLogger(true), &manager)
		pathToFile := fmt.Sprintf("%s%s", getTestingResourcesDirectory(), "/generic/incorrect/ns.yaml")

		// when
		err := applier.Apply(pathToFile)

		// then
		expectedError := "Could not decode the resource file"
		receivedError := err.Error()
		matched, err := regexp.MatchString(expectedError, receivedError)
		assert.True(t, matched, fmt.Sprintf("Expected error message: %s but got: %s", expectedError, receivedError))
	})
}

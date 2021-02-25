package preinstaller

import (
	"fmt"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/logger"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/preinstaller/mocks"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"regexp"
	"testing"
)

func TestResourceApplier_Apply(t *testing.T) {

	t.Run("should not apply resource", func(t *testing.T) {
		t.Run("due to get resource error", func(t *testing.T) {
			// given
			pathToFile := fmt.Sprintf("%s%s", getTestingResourcesDirectory(), "/generic/correct/crd.yaml")
			manifest, err := getResourceTestingFileContentFrom(pathToFile)
			assert.NoError(t, err)

			manager := mocks.GetErrorResourceManager{}
			applier := NewGenericResourceApplier(logger.NewLogger(true), &manager)

			// when
			err = applier.Apply(manifest)

			// then
			expectedError := "Get resource error"
			receivedError := err.Error()
			matched, err := regexp.MatchString(expectedError, receivedError)
			assert.True(t, matched, fmt.Sprintf("Expected error message: %s but got: %s", expectedError, receivedError))
		})

		t.Run("due to update resource error when resource existed on a cluster", func(t *testing.T) {
			// given
			pathToFile := fmt.Sprintf("%s%s", getTestingResourcesDirectory(), "/generic/correct/crd.yaml")
			manifest, err := getResourceTestingFileContentFrom(pathToFile)
			assert.NoError(t, err)

			manager := mocks.UpdateErrorResourceManager{}
			applier := NewGenericResourceApplier(logger.NewLogger(true), &manager)

			// when
			err = applier.Apply(manifest)

			// then
			expectedError := "Update resource error"
			receivedError := err.Error()
			matched, err := regexp.MatchString(expectedError, receivedError)
			assert.True(t, matched, fmt.Sprintf("Expected error message: %s but got: %s", expectedError, receivedError))
		})

		t.Run("due to creation error when resource did not exist on a cluster", func(t *testing.T) {
			// given
			pathToFile := fmt.Sprintf("%s%s", getTestingResourcesDirectory(), "/generic/correct/crd.yaml")
			manifest, err := getResourceTestingFileContentFrom(pathToFile)
			assert.NoError(t, err)

			manager := mocks.CreateErrorResourceManager{}
			applier := NewGenericResourceApplier(logger.NewLogger(true), &manager)

			// when
			err = applier.Apply(manifest)

			// then
			expectedError := "Create resource error"
			receivedError := err.Error()
			matched, err := regexp.MatchString(expectedError, receivedError)
			assert.True(t, matched, fmt.Sprintf("Expected error message: %s but got: %s", expectedError, receivedError))
		})
	})

	t.Run("should apply CRD", func(t *testing.T) {
		// given
		pathToFile := fmt.Sprintf("%s%s", getTestingResourcesDirectory(), "/generic/correct/crd.yaml")
		manifest, err := getResourceTestingFileContentFrom(pathToFile)
		assert.NoError(t, err)

		manager := mocks.ValidResourceManager{}
		applier := NewGenericResourceApplier(logger.NewLogger(true), &manager)

		// when
		err = applier.Apply(manifest)

		// then
		assert.NoError(t, err)
	})

	t.Run("should not apply CRD", func(t *testing.T) {
		// given
		pathToFile := fmt.Sprintf("%s%s", getTestingResourcesDirectory(), "/generic/incorrect/crd.yaml")
		manifest, err := getResourceTestingFileContentFrom(pathToFile)
		assert.NoError(t, err)

		manager := mocks.ValidResourceManager{}
		applier := NewGenericResourceApplier(logger.NewLogger(true), &manager)

		// when
		err = applier.Apply(manifest)

		// then
		expectedError := "Could not decode the resource file"
		receivedError := err.Error()
		matched, err := regexp.MatchString(expectedError, receivedError)
		assert.True(t, matched, fmt.Sprintf("Expected error message: %s but got: %s", expectedError, receivedError))
	})

	t.Run("should create namespace", func(t *testing.T) {
		// given
		pathToFile := fmt.Sprintf("%s%s", getTestingResourcesDirectory(), "/generic/correct/ns.yaml")
		manifest, err := getResourceTestingFileContentFrom(pathToFile)
		assert.NoError(t, err)

		manager := mocks.ValidResourceManager{}
		applier := NewGenericResourceApplier(logger.NewLogger(true), &manager)

		// when
		err = applier.Apply(manifest)

		// then
		assert.NoError(t, err)
	})

	t.Run("should not create namespace", func(t *testing.T) {
		// given
		pathToFile := fmt.Sprintf("%s%s", getTestingResourcesDirectory(), "/generic/incorrect/ns.yaml")
		manifest, err := getResourceTestingFileContentFrom(pathToFile)
		assert.NoError(t, err)

		manager := mocks.ValidResourceManager{}
		applier := NewGenericResourceApplier(logger.NewLogger(true), &manager)

		// when
		err = applier.Apply(manifest)

		// then
		expectedError := "Could not decode the resource file"
		receivedError := err.Error()
		matched, err := regexp.MatchString(expectedError, receivedError)
		assert.True(t, matched, fmt.Sprintf("Expected error message: %s but got: %s", expectedError, receivedError))
	})
}

func getResourceTestingFileContentFrom(path string) (string, error) {
	resourceData, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}

	return string(resourceData), nil
}

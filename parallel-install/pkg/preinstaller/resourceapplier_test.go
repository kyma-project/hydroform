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

			manager := mocks.RetrievalErrorResourceManager{}
			applier := NewGenericResourceApplier(logger.NewLogger(true), &manager)

			// when
			applied, err := applier.Apply(manifest)

			// then
			expectedError := "Get resource error"
			receivedError := err.Error()
			matched, err := regexp.MatchString(expectedError, receivedError)
			assert.True(t, matched, fmt.Sprintf("Expected error message: %s but got: %s", expectedError, receivedError))
			assert.False(t, applied)
		})

		t.Run("due to get resource existing on a cluster", func(t *testing.T) {
			// given
			pathToFile := fmt.Sprintf("%s%s", getTestingResourcesDirectory(), "/generic/correct/crd.yaml")
			manifest, err := getResourceTestingFileContentFrom(pathToFile)
			assert.NoError(t, err)

			manager := mocks.ResourceExistedResourceManagerMock{}
			applier := NewGenericResourceApplier(logger.NewLogger(true), &manager)

			// when
			applied, err := applier.Apply(manifest)

			// then
			assert.NoError(t, err)
			assert.False(t, applied)
		})

		t.Run("due to creation error", func(t *testing.T) {
			// given
			pathToFile := fmt.Sprintf("%s%s", getTestingResourcesDirectory(), "/generic/correct/crd.yaml")
			manifest, err := getResourceTestingFileContentFrom(pathToFile)
			assert.NoError(t, err)

			manager := mocks.CreationErrorResourceManager{}
			applier := NewGenericResourceApplier(logger.NewLogger(true), &manager)

			// when
			applied, err := applier.Apply(manifest)

			// then
			expectedError := "Create resource error"
			receivedError := err.Error()
			matched, err := regexp.MatchString(expectedError, receivedError)
			assert.True(t, matched, fmt.Sprintf("Expected error message: %s but got: %s", expectedError, receivedError))
			assert.False(t, applied)
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
		applied, err := applier.Apply(manifest)

		// then
		assert.NoError(t, err)
		assert.True(t, applied)
	})

	t.Run("should not apply CRD", func(t *testing.T) {
		// given
		pathToFile := fmt.Sprintf("%s%s", getTestingResourcesDirectory(), "/generic/incorrect/crd.yaml")
		manifest, err := getResourceTestingFileContentFrom(pathToFile)
		assert.NoError(t, err)

		manager := mocks.ValidResourceManager{}
		applier := NewGenericResourceApplier(logger.NewLogger(true), &manager)

		// when
		applied, err := applier.Apply(manifest)

		// then
		assert.NoError(t, err)
		assert.False(t, applied)
	})

	t.Run("should create namespace", func(t *testing.T) {
		// given
		pathToFile := fmt.Sprintf("%s%s", getTestingResourcesDirectory(), "/generic/correct/ns.yaml")
		manifest, err := getResourceTestingFileContentFrom(pathToFile)
		assert.NoError(t, err)

		manager := mocks.ValidResourceManager{}
		applier := NewGenericResourceApplier(logger.NewLogger(true), &manager)

		// when
		applied, err := applier.Apply(manifest)

		// then
		assert.NoError(t, err)
		assert.True(t, applied)
	})

	t.Run("should not create namespace", func(t *testing.T) {
		// given
		pathToFile := fmt.Sprintf("%s%s", getTestingResourcesDirectory(), "/generic/incorrect/ns.yaml")
		manifest, err := getResourceTestingFileContentFrom(pathToFile)
		assert.NoError(t, err)

		manager := mocks.ValidResourceManager{}
		applier := NewGenericResourceApplier(logger.NewLogger(true), &manager)

		// when
		applied, err := applier.Apply(manifest)

		// then
		assert.NoError(t, err)
		assert.False(t, applied)
	})
}

func getResourceTestingFileContentFrom(path string) (string, error) {
	resourceData, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}

	return string(resourceData), nil
}

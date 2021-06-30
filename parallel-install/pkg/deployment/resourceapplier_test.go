package deployment

import (
	"fmt"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/deployment/mocks"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/logger"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"regexp"
	"testing"
)

func TestResourceApplier_Apply(t *testing.T) {

	resourceName := "name"

	t.Run("should not apply resource", func(t *testing.T) {
		t.Run("due to not existing resource", func(t *testing.T) {
			// given
			manager := &mocks.ResourceManager{}
			applier := NewGenericResourceApplier(logger.NewLogger(true), manager)

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
			resource := fixResourceWith(resourceName)
			manager := &mocks.ResourceManager{}
			manager.On("GetResource", resourceName, fixResourceGvkSchema(), mock.AnythingOfType("GetOptions")).Return(nil, errors.New("Get resource error"))
			applier := NewGenericResourceApplier(logger.NewLogger(true), manager)

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
			resource := fixResourceWith(resourceName)
			resourceSchema := fixResourceGvkSchema()
			manager := &mocks.ResourceManager{}
			manager.On("GetResource", resourceName, resourceSchema, mock.AnythingOfType("GetOptions")).Return(resource, nil)
			manager.On("UpdateResource", resource, resourceSchema, mock.AnythingOfType("UpdateOptions")).Return(nil, errors.New("Update resource error"))
			applier := NewGenericResourceApplier(logger.NewLogger(true), manager)

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
			resource := fixResourceWith(resourceName)
			resourceSchema := fixResourceGvkSchema()
			manager := &mocks.ResourceManager{}
			manager.On("GetResource", resourceName, resourceSchema, mock.AnythingOfType("GetOptions")).Return(nil, nil)
			manager.On("CreateResource", resource, resourceSchema, mock.AnythingOfType("CreateOptions")).Return(errors.New("Create resource error"))
			applier := NewGenericResourceApplier(logger.NewLogger(true), manager)

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

	t.Run("should correctly apply resource that did not exist on a cluster", func(t *testing.T) {
		// given
		resource := fixResourceWith(resourceName)
		resourceSchema := fixResourceGvkSchema()
		manager := &mocks.ResourceManager{}
		manager.On("GetResource", resourceName, resourceSchema, mock.AnythingOfType("GetOptions")).Return(nil, nil)
		manager.On("CreateResource", resource, resourceSchema, mock.AnythingOfType("CreateOptions")).Return(nil)
		applier := NewGenericResourceApplier(logger.NewLogger(true), manager)

		// when
		err := applier.Apply(resource)

		// then
		assert.NoError(t, err)
	})

	t.Run("should correctly apply resource that did existed on a cluster", func(t *testing.T) {
		// given
		resource := fixResourceWith(resourceName)
		resourceSchema := fixResourceGvkSchema()
		manager := &mocks.ResourceManager{}
		manager.On("GetResource", resourceName, resourceSchema, mock.AnythingOfType("GetOptions")).Return(resource, nil)
		manager.On("UpdateResource", resource, resourceSchema, mock.AnythingOfType("UpdateOptions")).Return(resource, nil)
		applier := NewGenericResourceApplier(logger.NewLogger(true), manager)

		// when
		err := applier.Apply(resource)

		// then
		assert.NoError(t, err)
	})
}

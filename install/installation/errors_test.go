package installation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrors(t *testing.T) {

	t.Run("should properly format installation error", func(t *testing.T) {
		installationError := InstallationError{
			ShortMessage: "error",
			ErrorEntries: []ErrorEntry{
				{
					Component:   "test",
					Log:         "logs",
					Occurrences: 2,
				},
				{
					Component:   "test2",
					Log:         "logs2",
					Occurrences: 1,
				},
			},
		}

		expectedDetails := `Installation errors: 
Component: test, Log: logs
Component: test2, Log: logs2`

		errMsg := installationError.Error()
		assert.Equal(t, "error", errMsg)

		details := installationError.Details()
		assert.Equal(t, expectedDetails, details)
	})

	t.Run("should properly format invalid installation state error", func(t *testing.T) {
		invalidInstallationStateErr := InvalidInstallationStateError{
			InstallationState:  "Invalid",
			InstallationStatus: "something is wrong",
		}

		errMsg := invalidInstallationStateErr.Error()

		assert.Equal(t, "invalid installation state: Invalid, status: something is wrong", errMsg)
	})
}

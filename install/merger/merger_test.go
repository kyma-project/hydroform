package merger

import (
	"github.com/kyma-incubator/hydroform/install/merger/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestMerger_PrepareInstallation(t *testing.T) {

	t.Run("should replace data if replace flag is specified", func(t *testing.T) {
		// given
		labels := map[string]string{
			OnConflictLabel: ReplaceOnConflict,
		}
		data := &mocks.Data{}
		data.On("Labels").Return(&labels)
		data.On("Update").Return(nil)

		// when
		err := Update(data)

		// then
		assert.NoError(t, err)
		mock.AssertExpectationsForObjects(t, data)
	})

	t.Run("should merge data if no flag at all", func(t *testing.T) {
		// given
		data := &mocks.Data{}
		data.On("Labels").Return(nil)
		data.On("Merge").Return(nil)
		data.On("Update").Return(nil)

		// when
		err := Update(data)

		// then
		assert.NoError(t, err)
		mock.AssertExpectationsForObjects(t, data)
	})

}

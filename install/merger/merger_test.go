package merger

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type DataImpl struct {
	labels      map[string]string
	updateCount int
	mergeCount  int
}

func NewData() *DataImpl {
	return &DataImpl{
		labels: map[string]string{
			"label": "value",
		},
	}
}

func (d *DataImpl) Labels() *map[string]string {
	return &d.labels
}

func (d *DataImpl) Update() error {
	d.updateCount += 1
	return nil
}

func (d *DataImpl) Merge() error {
	d.mergeCount += 1
	return nil
}

func TestMerger_PrepareInstallation(t *testing.T) {

	t.Run("should replace data if replace flag is specified", func(t *testing.T) {
		data := NewData()
		data.labels[OnConflictLabel] = ReplaceOnConflict
		err := data.Update()
		assert.NoError(t, err)
		assert.Equal(t, 0, data.mergeCount)
		assert.Equal(t, 1, data.updateCount)
	})

	t.Run("should merge data if replace flag is specified", func(t *testing.T) {
		data := NewData()
		err := Update(data)
		assert.NoError(t, err)
		assert.Equal(t, 1, data.mergeCount)
		assert.Equal(t, 1, data.updateCount)
	})

}

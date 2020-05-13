package k8s

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMergeMaps(t *testing.T) {
	t.Run("should merge two maps", func(t *testing.T) {
		//given
		oldMap := map[string]interface{}{
			"key1": map[string]interface{}{
				"key3": "terefere",
				"key4": map[string]interface{}{
					"key5": "bamboozled",
					"key6": "umbazled",
				},
			},
			"key2": map[string]interface{}{
				"key7": "kek",
			},
		}

		newMap := map[string]interface{}{
			"key1": map[string]interface{}{
				"key4": map[string]interface{}{
					"key5": "notbamboozled",
				},
			},
			"key2": map[string]interface{}{
				"key7": "lol",
				"key8": "yez",
			},
		}

		expectedMap := map[string]interface{}{
			"key1": map[string]interface{}{
				"key3": "terefere",
				"key4": map[string]interface{}{
					"key5": "notbamboozled",
					"key6": "umbazled",
				},
			},
			"key2": map[string]interface{}{
				"key7": "lol",
				"key8": "yez",
			},
		}

		//when
		actual := MergeMaps(newMap, oldMap)

		//then
		assert.Equal(t, expectedMap, actual)
	})
}

func TestMergeStringMaps(t *testing.T) {

	t.Run("should merge maps", func(t *testing.T) {
		//given
		oldMap := map[string]string{
			"key1": "old val 1",
			"key2": "old val 2",
		}

		newMap := map[string]string{
			"key1": "new val 1",
			"key3": "new val 3",
		}

		expectedMap := map[string]string{
			"key1": "new val 1",
			"key2": "old val 2",
			"key3": "new val 3",
		}

		// when
		merged := MergeStringMaps(oldMap, newMap)

		// then
		assert.Equal(t, expectedMap, merged)
	})

}

func TestMergeByteMaps(t *testing.T) {

	t.Run("should merge maps", func(t *testing.T) {
		//given
		oldMap := map[string][]byte{
			"key1": []byte("old val 1"),
			"key2": []byte("old val 2"),
		}

		newMap := map[string][]byte{
			"key1": []byte("new val 1"),
			"key3": []byte("new val 3"),
		}

		expectedMap := map[string][]byte{
			"key1": []byte("new val 1"),
			"key2": []byte("old val 2"),
			"key3": []byte("new val 3"),
		}

		// when
		merged := MergeByteMaps(oldMap, newMap)

		// then
		assert.Equal(t, expectedMap, merged)
	})

}

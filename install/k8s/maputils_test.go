package k8s

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
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
		assert.True(t, reflect.DeepEqual(expectedMap, actual))
	})
}

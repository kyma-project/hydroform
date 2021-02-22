package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMergeMaps(t *testing.T) {

	fixOldMap := func() map[string]interface{} {
		return map[string]interface{}{
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
	}

	fixNewMap := func() map[string]interface{} {
		return map[string]interface{}{
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
	}

	for _, testCase := range []struct {
		description string
		oldMap      map[string]interface{}
		newMap      map[string]interface{}
		expected    map[string]interface{}
	}{
		{
			description: "should merge non empty maps",
			oldMap:      fixOldMap(),
			newMap:      fixNewMap(),
			expected: map[string]interface{}{
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
			},
		},
		{
			description: "should merge with nil old map",
			oldMap:      nil,
			newMap:      fixNewMap(),
			expected:    fixNewMap(),
		},
		{
			description: "should merge with nil new map",
			oldMap:      fixOldMap(),
			newMap:      nil,
			expected:    fixOldMap(),
		},
		{
			description: "should merge two nil maps",
			oldMap:      nil,
			newMap:      nil,
			expected:    map[string]interface{}{},
		},
	} {
		t.Run(testCase.description, func(t *testing.T) {
			merged := MergeMaps(testCase.newMap, testCase.oldMap)

			assert.Equal(t, testCase.expected, merged)
		})
	}
}

func TestMergeStringMaps(t *testing.T) {

	fixOldMap := func() map[string]string {
		return map[string]string{
			"key1": "old val 1",
			"key2": "old val 2",
		}
	}

	fixNewMap := func() map[string]string {
		return map[string]string{
			"key1": "new val 1",
			"key3": "new val 3",
		}
	}

	for _, testCase := range []struct {
		description string
		oldMap      map[string]string
		newMap      map[string]string
		expected    map[string]string
	}{
		{
			description: "should merge non empty maps",
			oldMap:      fixOldMap(),
			newMap:      fixNewMap(),
			expected: map[string]string{
				"key1": "new val 1",
				"key2": "old val 2",
				"key3": "new val 3",
			},
		},
		{
			description: "should merge with nil old map",
			oldMap:      nil,
			newMap:      fixNewMap(),
			expected:    fixNewMap(),
		},
		{
			description: "should merge with nil new map",
			oldMap:      fixOldMap(),
			newMap:      nil,
			expected:    fixOldMap(),
		},
		{
			description: "should merge two nil maps",
			oldMap:      nil,
			newMap:      nil,
			expected:    map[string]string{},
		},
	} {
		t.Run(testCase.description, func(t *testing.T) {
			merged := MergeStringMaps(testCase.oldMap, testCase.newMap)

			assert.Equal(t, testCase.expected, merged)
		})
	}
}

func TestMergeByteMaps(t *testing.T) {

	fixOldMap := func() map[string][]byte {
		return map[string][]byte{
			"key1": []byte("old val 1"),
			"key2": []byte("old val 2"),
		}
	}

	fixNewMap := func() map[string][]byte {
		return map[string][]byte{
			"key1": []byte("new val 1"),
			"key3": []byte("new val 3"),
		}
	}

	for _, testCase := range []struct {
		description string
		oldMap      map[string][]byte
		newMap      map[string][]byte
		expected    map[string][]byte
	}{
		{
			description: "should merge non empty maps",
			oldMap:      fixOldMap(),
			newMap:      fixNewMap(),
			expected: map[string][]byte{
				"key1": []byte("new val 1"),
				"key2": []byte("old val 2"),
				"key3": []byte("new val 3"),
			},
		},
		{
			description: "should merge with nil old map",
			oldMap:      nil,
			newMap:      fixNewMap(),
			expected:    fixNewMap(),
		},
		{
			description: "should merge with nil new map",
			oldMap:      fixOldMap(),
			newMap:      nil,
			expected:    fixOldMap(),
		},
		{
			description: "should merge two nil maps",
			oldMap:      nil,
			newMap:      nil,
			expected:    map[string][]byte{},
		},
	} {
		t.Run(testCase.description, func(t *testing.T) {
			merged := MergeByteMaps(testCase.oldMap, testCase.newMap)

			assert.Equal(t, testCase.expected, merged)
		})
	}
}

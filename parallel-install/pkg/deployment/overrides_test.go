package deployment

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func Test_MergeOverrides(t *testing.T) {
	var err error

	overrides := Overrides{}
	overrides.AddFile("../test/data/deployment-overrides1.yaml")
	overrides.AddFile("../test/data/deployment-overrides2.json")

	override1 := make(map[string]interface{})
	override1["key4"] = "value4override1"
	overrides.AddOverrides(override1)

	override2 := make(map[string]interface{})
	override2["key5"] = "value5override2"
	overrides.AddOverrides(override2)

	// read expected result
	data, err := ioutil.ReadFile("../test/data/deployment-overrides-result.yaml")
	require.NoError(t, err)
	var expected map[string]interface{}
	err = yaml.Unmarshal(data, &expected)
	require.NoError(t, err)

	// verify merge result with expected data
	result, err := overrides.Merge()
	require.NoError(t, err)
	require.Equal(t, expected, result)
}

func Test_AddFile(t *testing.T) {
	var err error

	overrides := Overrides{}
	err = overrides.AddFile("../test/data/deployment-overrides1.yaml")
	require.NoError(t, err)
	err = overrides.AddFile("../test/data/deployment-overrides2.json")
	require.NoError(t, err)
	err = overrides.AddFile("../test/data/overrides.xml") // unsupported format
	require.Error(t, err)
}

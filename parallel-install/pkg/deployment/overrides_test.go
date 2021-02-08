package deployment

import (
	"fmt"
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
	err = overrides.AddOverrides("chart", override1)
	require.NoError(t, err)

	override2 := make(map[string]interface{})
	override2["key5"] = "value5override2"
	err = overrides.AddOverrides("chart", override2)
	require.NoError(t, err)

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

func Test_AddOverrides(t *testing.T) {
	var err error

	overrides := Overrides{}
	data := make(map[string]interface{})

	// invalid
	err = overrides.AddOverrides("", data)
	require.Error(t, err)

	//invalid
	err = overrides.AddOverrides("xyz", data)
	require.Error(t, err)

	//valid
	data["test"] = "abc"
	err = overrides.AddOverrides("xyz", data)
	require.NoError(t, err)
}

// interceptor which is replacing a value
type replaceOverrideInterceptor struct {
}

func (roi *replaceOverrideInterceptor) String(value interface{}) string {
	return fmt.Sprintf("%v", value)
}

func (roi *replaceOverrideInterceptor) Intercept(value interface{}) (interface{}, error) {
	return "intercepted", nil
}

// interceptor which is failing
type failingOverrideInterceptor struct {
}

func (roi *failingOverrideInterceptor) String(value interface{}) string {
	return fmt.Sprintf("%v", value)
}

func (roi *failingOverrideInterceptor) Intercept(value interface{}) (interface{}, error) {
	return nil, fmt.Errorf("Interceptor failed")
}

// interceptor which is failing
type stringerOverrideInterceptor struct {
}

func (roi *stringerOverrideInterceptor) String(value interface{}) string {
	return fmt.Sprintf("string-%v", value)
}

func (roi *stringerOverrideInterceptor) Intercept(value interface{}) (interface{}, error) {
	return nil, fmt.Errorf("Interceptor failed")
}

func Test_InterceptValue(t *testing.T) {
	t.Run("Test interceptor without failures", func(t *testing.T) {
		overrides := Overrides{}
		overrides.AddFile("../test/data/deployment-overrides-intercepted.yaml")
		overrides.AddInterceptor([]string{"chart.key2.key2.1", "chart.key4"}, &replaceOverrideInterceptor{})

		// read expected result
		data, err := ioutil.ReadFile("../test/data/deployment-overrides-intercepted-result.yaml")
		require.NoError(t, err)
		var expected map[string]interface{}
		err = yaml.Unmarshal(data, &expected)
		require.NoError(t, err)

		// verify merge result with expected data
		result, err := overrides.Merge()
		require.NoError(t, err)
		require.Equal(t, expected, result)
	})

	t.Run("Test interceptor with failure", func(t *testing.T) {
		overrides := Overrides{}
		overrides.AddFile("../test/data/deployment-overrides-intercepted.yaml")
		overrides.AddInterceptor([]string{"chart.key1"}, &failingOverrideInterceptor{})
		result, err := overrides.Merge()
		require.Empty(t, result)
		require.Error(t, err)
	})

	t.Run("String interceptor", func(t *testing.T) {
		overrides := Overrides{}
		overrides.AddFile("../test/data/deployment-overrides-intercepted.yaml")
		overrides.AddInterceptor([]string{"chart.key1", "chart.key3"}, &stringerOverrideInterceptor{})
		require.Equal(t,
			"map[chart:map[key1:string-value1yaml key2:map[key2.1:value2.1yaml key2.2:value2.2yaml] key3:string-value3yaml key4:value4yaml]]",
			fmt.Sprint(overrides))
	})
}

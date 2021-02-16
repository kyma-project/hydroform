package deployment

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// interceptor which is replacing a value
type replaceOverrideInterceptor struct {
}

func (roi *replaceOverrideInterceptor) String(value interface{}, key string) string {
	return fmt.Sprintf("%v", value)
}

func (roi *replaceOverrideInterceptor) Intercept(value interface{}, key string) (interface{}, error) {
	return "intercepted", nil
}

func (roi *replaceOverrideInterceptor) Undefined(overrides map[string]interface{}, key string) error {
	return nil
}

//stringerOverrideInterceptor hides the value of an override when the value is converted to a string
type stringerOverrideInterceptor struct {
}

func (i *stringerOverrideInterceptor) String(value interface{}, key string) string {
	return fmt.Sprintf("string-%v", value)
}

func (i *stringerOverrideInterceptor) Intercept(value interface{}, key string) (interface{}, error) {
	return value, nil
}

func (i *stringerOverrideInterceptor) Undefined(overrides map[string]interface{}, key string) error {
	return nil
}

// interceptor which is failing
type failingOverrideInterceptor struct {
}

func (roi *failingOverrideInterceptor) String(value interface{}, key string) string {
	return fmt.Sprintf("%v", value)
}

func (roi *failingOverrideInterceptor) Intercept(value interface{}, key string) (interface{}, error) {
	return nil, fmt.Errorf("Interceptor failed")
}

func (roi *failingOverrideInterceptor) Undefined(overrides map[string]interface{}, key string) error {
	return nil
}

// interceptor which is returning a value for an undefined key
type undefinedOverrideInterceptor struct {
}

func (roi *undefinedOverrideInterceptor) String(value interface{}, key string) string {
	return fmt.Sprintf("%v", value)
}

func (roi *undefinedOverrideInterceptor) Intercept(value interface{}, key string) (interface{}, error) {
	return value, nil
}

func (roi *undefinedOverrideInterceptor) Undefined(overrides map[string]interface{}, key string) error {
	return fmt.Errorf("This value was missing")
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
}

func Test_InterceptStringer(t *testing.T) {
	overrides := Overrides{}
	overrides.AddFile("../test/data/deployment-overrides-intercepted.yaml")
	overrides.AddInterceptor([]string{"chart.key1", "chart.key3"}, &stringerOverrideInterceptor{})
	require.Equal(t,
		"map[chart:map[key1:string- key2:map[key2.1:value2.1yaml key2.2:value2.2yaml] key3:<masked> key4:value4yaml]]",
		fmt.Sprint(overrides))
}

func Test_InterceptUndefined(t *testing.T) {
	overrides := Overrides{}
	overrides.AddFile("../test/data/deployment-overrides-intercepted.yaml")
	overrides.AddInterceptor([]string{"I.dont.exist"}, &undefinedOverrideInterceptor{})
	result, err := overrides.Merge()
	require.Empty(t, result)
	require.Error(t, err)
	require.Equal(t, "This value was missing", err.Error())
}

func Test_FallbackInterceptor(t *testing.T) {
	overrides := Overrides{}
	overrides.AddFile("../test/data/deployment-overrides-intercepted.yaml")

	t.Run("Test FallbackInterceptor happy path", func(t *testing.T) {
		overrides.AddInterceptor([]string{"I.dont.exist"}, NewFallbackOverrideInterceptor("I am the fallback"))
		result, err := overrides.Merge()
		require.NotEmpty(t, result)
		require.NoError(t, err)
		require.Equal(t, "I am the fallback", result["I"].(map[string]interface{})["dont"].(map[string]interface{})["exist"])
	})

	t.Run("Test FallbackInterceptor with sub-key which is not a map", func(t *testing.T) {
		overrides.AddInterceptor([]string{"chart.key3.xyz"}, NewFallbackOverrideInterceptor("Use me as fallback"))
		result, err := overrides.Merge()
		require.Empty(t, result)
		require.Error(t, err)
	})
}

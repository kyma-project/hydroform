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

func (roi *replaceOverrideInterceptor) String(o *Overrides, value interface{}) string {
	return fmt.Sprintf("%v", value)
}

func (roi *replaceOverrideInterceptor) Intercept(o *Overrides, value interface{}) (interface{}, error) {
	return "intercepted", nil
}

func (roi *replaceOverrideInterceptor) Undefined(o *Overrides, key string) error {
	return nil
}

// interceptor which is failing
type failingOverrideInterceptor struct {
}

func (roi *failingOverrideInterceptor) String(o *Overrides, value interface{}) string {
	return fmt.Sprintf("%v", value)
}

func (roi *failingOverrideInterceptor) Intercept(o *Overrides, value interface{}) (interface{}, error) {
	return nil, fmt.Errorf("Interceptor failed")
}

func (roi *failingOverrideInterceptor) Undefined(o *Overrides, key string) error {
	return nil
}

// interceptor which is returning a manipulated string
type stringerOverrideInterceptor struct {
}

func (roi *stringerOverrideInterceptor) String(o *Overrides, value interface{}) string {
	return fmt.Sprintf("string-%v", value)
}

func (roi *stringerOverrideInterceptor) Intercept(o *Overrides, value interface{}) (interface{}, error) {
	return nil, fmt.Errorf("Interceptor failed")
}

func (roi *stringerOverrideInterceptor) Undefined(o *Overrides, key string) error {
	return nil
}

// interceptor which is returning a value for an undefined key
type undefinedOverrideInterceptor struct {
}

func (roi *undefinedOverrideInterceptor) String(o *Overrides, value interface{}) string {
	return fmt.Sprintf("%v", value)
}

func (roi *undefinedOverrideInterceptor) Intercept(o *Overrides, value interface{}) (interface{}, error) {
	return value, nil
}

func (roi *undefinedOverrideInterceptor) Undefined(o *Overrides, key string) error {
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

func Test_InterceptString(t *testing.T) {
	overrides := Overrides{}
	overrides.AddFile("../test/data/deployment-overrides-intercepted.yaml")
	overrides.AddInterceptor([]string{"chart.key1", "chart.key3"}, &stringerOverrideInterceptor{})
	require.Equal(t,
		"map[chart:map[key1:string-value1yaml key2:map[key2.1:value2.1yaml key2.2:value2.2yaml] key3:string-value3yaml key4:value4yaml]]",
		fmt.Sprint(overrides))
}

func Test_InterceptUndefined(t *testing.T) {
	overrides := Overrides{}
	overrides.AddFile("../test/data/deployment-overrides-intercepted.yaml")
	overrides.AddInterceptor([]string{"chart.key1"}, &replaceOverrideInterceptor{})
	overrides.AddInterceptor([]string{"I.dont.exist"}, &undefinedOverrideInterceptor{})
	result, err := overrides.Merge()
	require.Empty(t, result)
	require.Error(t, err)
	require.Equal(t, "This value was missing", err.Error())
}

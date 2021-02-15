package deployment

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"
	"strings"

	"github.com/imdario/mergo"
	"gopkg.in/yaml.v3"
)

var (
	supportedFileExt   = []string{"yaml", "yml", "json"}
	defaultInterceptor = &defaultOverrideInterceptor{}
)

type interceptorOps string

const (
	interceptorOpsString    = "String"
	interceptorOpsIntercept = "Intercept"
)

// Overrides manages override merges
type Overrides struct {
	files              []string
	overrides          []map[string]interface{}
	interceptors       map[string]OverrideInterceptor
	processedOverrides []string // cache for processed override-keys
}

// Merge all provided overrides
func (o *Overrides) Merge() (map[string]interface{}, error) {
	merged, err := o.merge()
	if err != nil {
		return nil, err
	}
	return o.intercept(merged, interceptorOpsIntercept)
}

// String all provided overrides
func (o Overrides) String() string {
	merged, err := o.merge()
	if err != nil {
		return fmt.Sprintf("%v", err)
	}
	result, err := o.intercept(merged, interceptorOpsString)
	if err != nil {
		return fmt.Sprint(err)
	}
	return fmt.Sprintf("%v", result)
}

func (o *Overrides) merge() (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// merge files
	var fileOverrides map[string]interface{}
	for _, file := range o.files {
		// read data
		data, err := ioutil.ReadFile(file)
		if err != nil {
			return nil, err
		}
		// unmarshal
		if strings.HasSuffix(file, ".json") {
			err = json.Unmarshal(data, &fileOverrides)
		} else {
			err = yaml.Unmarshal(data, &fileOverrides)
		}
		if err != nil {
			return nil, err
		}
		// merge
		if err := mergo.Map(&result, fileOverrides, mergo.WithOverride); err != nil {
			return nil, err
		}
	}

	//merge overrides
	for _, override := range o.overrides {
		if err := mergo.Map(&result, override, mergo.WithOverride); err != nil {
			return nil, err
		}
	}

	return result, nil
}

func (o *Overrides) intercept(data map[string]interface{}, ops interceptorOps) (map[string]interface{}, error) {
	//reset processed overrides key-cache
	o.processedOverrides = make([]string, len(o.processedOverrides))

	result := make(map[string]interface{}, len(data))
	for key, value := range data {
		intercptVal, err := o.interceptValue(key, value, ops)
		if err != nil {
			return nil, err
		}
		result[key] = intercptVal
	}

	if err := o.interceptUndefined(); err != nil {
		return nil, err
	}

	return result, nil
}

func (o *Overrides) interceptValue(path string, value interface{}, ops interceptorOps) (interface{}, error) {
	o.processedOverrides = append(o.processedOverrides, path)
	if reflect.ValueOf(value).Kind() == reflect.Map {
		mapValue := value.(map[string]interface{})
		result := make(map[string]interface{}, len(mapValue))
		for key, value := range mapValue {
			var entryPath string
			if path == "" {
				entryPath = key
			} else {
				entryPath = fmt.Sprintf("%s.%s", path, key)
			}
			intercptVal, err := o.interceptValue(entryPath, value, ops)
			if err != nil {
				return nil, err
			}
			result[key] = intercptVal
		}
		return result, nil
	}
	var interceptor OverrideInterceptor
	interceptor, exists := o.interceptors[path]
	if !exists {
		interceptor = defaultInterceptor
	}
	//apply interceptor
	if ops == interceptorOpsString {
		return interceptor.String(o, value), nil
	}
	return interceptor.Intercept(o, value)
}

func (o *Overrides) interceptUndefined() error {
	// get list of intercepted override keys
	itercptOverrideKeys := make([]string, len(o.interceptors))
	i := 0
	for itercptOverrideKey := range o.interceptors {
		itercptOverrideKeys[i] = itercptOverrideKey
		i++
	}
	//retreive override keys which were not processed but have an interceptor registered
	undefinedOverrideKeys := diffSlices(itercptOverrideKeys, o.processedOverrides)
	for _, undefinedOverrideKey := range undefinedOverrideKeys {
		if err := o.interceptors[undefinedOverrideKey].Undefined(o, undefinedOverrideKey); err != nil {
			return err
		}
	}

	return nil
}

// diffSlices returns the elements in `a` that aren't in `b`.
func diffSlices(a, b []string) []string {
	cache := make(map[string]bool, len(b))
	for _, x := range b {
		cache[x] = true
	}
	var diff []string
	for _, x := range a {
		if _, found := cache[x]; !found {
			diff = append(diff, x)
		}
	}
	return diff
}

// AddFile adds overrides defined in a file
func (o *Overrides) AddFile(file string) error {
	for _, ext := range supportedFileExt {
		if strings.HasSuffix(file, fmt.Sprintf(".%s", ext)) {
			o.files = append(o.files, file)
			return nil
		}
	}
	return fmt.Errorf("Unsupported override file extension. Supported extensions are: %s", strings.Join(supportedFileExt, ", "))
}

// AddOverrides adds overrides for a chart
func (o *Overrides) AddOverrides(chart string, overrides map[string]interface{}) error {
	if chart == "" {
		return fmt.Errorf("Chart name cannot be empty when adding overrides")
	}
	if len(overrides) < 1 {
		return fmt.Errorf("Empty overrides map provided for chart '%s'", chart)
	}
	overridesMap := make(map[string]interface{})
	overridesMap[chart] = overrides
	o.overrides = append(o.overrides, overridesMap)
	return nil
}

// AddInterceptor registers an interceptor for a particular override keys
func (o *Overrides) AddInterceptor(overrideKeys []string, interceptor OverrideInterceptor) {
	if o.interceptors == nil {
		o.interceptors = make(map[string]OverrideInterceptor)
	}
	for _, overrideKey := range overrideKeys {
		o.interceptors[overrideKey] = interceptor
	}
}

package deployment

import (
	"fmt"
	"reflect"
	"strings"
)

// OverrideInterceptor is controlling access to override values
type OverrideInterceptor interface {
	//String shows the value of the override
	String(o *Overrides, value interface{}) string
	//Intercept is executed when the override is retrieved
	Intercept(o *Overrides, value interface{}) (interface{}, error)
	//Undefined is executed when the override is not defined
	Undefined(overrides map[string]interface{}, key string) error
}

type defaultOverrideInterceptor struct {
}

func (i *defaultOverrideInterceptor) String(o *Overrides, value interface{}) string {
	return fmt.Sprintf("%v", value)
}

func (i *defaultOverrideInterceptor) Intercept(o *Overrides, value interface{}) (interface{}, error) {
	return value, nil
}

func (i *defaultOverrideInterceptor) Undefined(overrides map[string]interface{}, key string) error {
	return nil
}

//MaskOverrideInterceptor hides the value of an override when the value is converted to a string
type MaskOverrideInterceptor struct {
}

func (i *MaskOverrideInterceptor) String(o *Overrides, value interface{}) string {
	return "<masked>"
}

func (i *MaskOverrideInterceptor) Intercept(o *Overrides, value interface{}) (interface{}, error) {
	return value, nil
}

func (i *MaskOverrideInterceptor) Undefined(overrides map[string]interface{}, key string) error {
	return nil
}

func NewFallbackOverrideInterceptor(fallback interface{}) *FallbackOverrideInterceptor {
	return &FallbackOverrideInterceptor{
		fallback: fallback,
	}
}

//FallbackOverrideInterceptor sets a default value for an undefined overwrite
type FallbackOverrideInterceptor struct {
	fallback interface{}
}

func (i *FallbackOverrideInterceptor) String(o *Overrides, value interface{}) string {
	return fmt.Sprintf("%v", value)
}

func (i *FallbackOverrideInterceptor) Intercept(o *Overrides, value interface{}) (interface{}, error) {
	return value, nil
}

func (i *FallbackOverrideInterceptor) Undefined(overrides map[string]interface{}, key string) error {
	subKeys := strings.Split(key, ".")
	maxDepth := len(subKeys)
	lastProcessedEntry := overrides

	for depth, subKey := range subKeys {
		if _, ok := lastProcessedEntry[subKey]; !ok {
			//sub-element does not exist - add map
			lastProcessedEntry[subKey] = make(map[string]interface{})
		} else if reflect.ValueOf(lastProcessedEntry[subKey]).Kind() != reflect.Map {
			//ensure existing sub-element is map otherwise fail
			return fmt.Errorf("Override '%s' cannot be set with default value as sub-key '%s' is not a map", key, strings.Join(subKeys[:depth+1], "."))
		}
		if depth == (maxDepth - 1) {
			//we are in the last loop, set default value
			lastProcessedEntry[subKey] = i.fallback
		} else {
			//continue processing the next sub-entry
			lastProcessedEntry = lastProcessedEntry[subKey].(map[string]interface{})
		}
	}

	return nil
}

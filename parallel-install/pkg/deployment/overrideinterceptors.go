package deployment

import "fmt"

// OverrideInterceptor is controlling access to override values
type OverrideInterceptor interface {
	//String shows the value of the override
	String(o *Overrides, value interface{}) string
	//Intercept is executed when the override is retrieved
	Intercept(o *Overrides, value interface{}) (interface{}, error)
	//Undefined is executed when the override is not defined
	Undefined(o *Overrides, key string) error
}

type defaultOverrideInterceptor struct {
}

func (i *defaultOverrideInterceptor) String(o *Overrides, value interface{}) string {
	return fmt.Sprintf("%v", value)
}

func (i *defaultOverrideInterceptor) Intercept(o *Overrides, value interface{}) (interface{}, error) {
	return value, nil
}

func (i *defaultOverrideInterceptor) Undefined(o *Overrides, key string) error {
	return nil
}

//MaskOverrideInterceptor does not show the value of an override
type MaskOverrideInterceptor struct {
}

func (i *MaskOverrideInterceptor) String(o *Overrides, value interface{}) string {
	return "<masked>"
}

func (i *MaskOverrideInterceptor) Intercept(o *Overrides, value interface{}) (interface{}, error) {
	return value, nil
}

func (i *MaskOverrideInterceptor) Undefined(o *Overrides, key string) error {
	return nil
}

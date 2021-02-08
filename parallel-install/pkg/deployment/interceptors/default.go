package interceptors

import "fmt"

type DefaultOverrideInterceptor struct {
}

func (doi *DefaultOverrideInterceptor) String(value interface{}) string {
	return fmt.Sprintf("%v", value)
}

func (doi *DefaultOverrideInterceptor) Intercept(value interface{}) (interface{}, error) {
	return value, nil
}

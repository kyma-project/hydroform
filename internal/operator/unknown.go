package operator

import "errors"

type Unknown struct {
}

func (u *Unknown) Create(provider string, configuration map[string]interface{}) error {
	return errors.New("unknown operator")
}

func (u *Unknown) Delete(provider string, configuration map[string]interface{}) error {
	return errors.New("unknown operator")
}

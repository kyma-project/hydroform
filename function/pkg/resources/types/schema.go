package types

import (
	"github.com/invopop/jsonschema"
)

func ReflectSchema() ([]byte, error) {
	schema := jsonschema.Reflect(&Function{})

	return schema.MarshalJSON()
}

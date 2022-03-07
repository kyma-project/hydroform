package workspace

import (
	"github.com/invopop/jsonschema"
)

func ReflectSchema() ([]byte, error) {
	schema := jsonschema.Reflect(&Cfg{})

	return schema.MarshalJSON()
}

package deployment

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/imdario/mergo"
	"gopkg.in/yaml.v3"
)

var (
	supportedFileExt = []string{"yaml", "yml", "json"}
)

// Overrides manages override merges
type Overrides struct {
	files     []string
	overrides []map[string]interface{}
}

// Merge all provided overrides
func (o *Overrides) Merge() (map[string]interface{}, error) {
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
func (o *Overrides) AddOverrides(chart string, overrides map[string]interface{}) {
	overridesMap := make(map[string]interface{})
	overridesMap[chart] = overrides
	o.overrides = append(o.overrides, overridesMap)
}

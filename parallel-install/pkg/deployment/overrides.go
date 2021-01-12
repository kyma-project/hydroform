package deployment

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

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
		result = o.mergeMaps(result, fileOverrides)
	}

	//merge overrides
	for _, override := range o.overrides {
		result = o.mergeMaps(result, override)
	}

	return result, nil
}

func (o *Overrides) mergeMaps(a, b map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(a))
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		if v, ok := v.(map[string]interface{}); ok {
			if bv, ok := out[k]; ok {
				if bv, ok := bv.(map[string]interface{}); ok {
					out[k] = o.mergeMaps(bv, v)
					continue
				}
			}
		}
		out[k] = v
	}
	return out
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

// AddOverrides adds another override
func (o *Overrides) AddOverrides(overrides map[string]interface{}) {
	o.overrides = append(o.overrides, overrides)
}

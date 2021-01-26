package components

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ComponentList collects component definitions
type ComponentList struct {
	Prerequisites []ComponentDefinition
	Components    []ComponentDefinition
}

// ComponentDefinition defines a component in components list
type ComponentDefinition struct {
	Name      string
	Namespace string
}

// ComponentListData is the raw component list
type ComponentListData struct {
	DefaultNamespace string `yaml:"defaultNamespace" json:"defaultNamespace"`
	Prerequisites    []ComponentDefinition
	Components       []ComponentDefinition
}

func (cld *ComponentListData) process() *ComponentList {
	compList := &ComponentList{}

	// read prerequisites
	for _, compDef := range cld.Prerequisites {
		if compDef.Namespace == "" {
			compDef.Namespace = cld.DefaultNamespace
		}
		compList.Prerequisites = append(compList.Prerequisites, compDef)
	}

	// read components
	for _, compDef := range cld.Components {
		if compDef.Namespace == "" {
			compDef.Namespace = cld.DefaultNamespace
		}
		compList.Components = append(compList.Components, compDef)
	}

	return compList
}

// NewComponentListFromFile creates a new component list
func NewComponentListFromFile(componentsListPath string) (*ComponentList, error) {
	if componentsListPath == "" {
		return nil, fmt.Errorf("Path to components list file is required")
	}
	if _, err := os.Stat(componentsListPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("Components list file '%s' not found", componentsListPath)
	}

	data, err := ioutil.ReadFile(componentsListPath)
	if err != nil {
		return nil, err
	}

	switch filepath.Ext(componentsListPath) {
	case ".json":
		return NewComponentListFromJSON(data)
	case ".yaml":
		return NewComponentListFromYAML(data)
	case ".yml":
		return NewComponentListFromYAML(data)
	default:
		return nil, fmt.Errorf("File format of components list is not supported")
	}
}

//NewComponentListFromYAML creates a component list object from YAML data
func NewComponentListFromYAML(data []byte) (*ComponentList, error) {
	var compListData *ComponentListData
	if err := yaml.Unmarshal(data, &compListData); err != nil {
		return nil, err
	}
	return compListData.process(), nil
}

//NewComponentListFromJSON creates a component list object from JSON data
func NewComponentListFromJSON(data []byte) (*ComponentList, error) {
	var compListData *ComponentListData
	if err := json.Unmarshal(data, &compListData); err != nil {
		return nil, err
	}
	return compListData.process(), nil
}

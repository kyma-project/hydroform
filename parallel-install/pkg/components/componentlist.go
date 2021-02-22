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

// NewComponentList creates a new component list
func NewComponentList(componentsListPath string) (*ComponentList, error) {
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

	var compListData *ComponentListData
	fileExt := filepath.Ext(componentsListPath)
	if fileExt == ".json" {
		if err := json.Unmarshal(data, &compListData); err != nil {
			return nil, err
		}
	} else if fileExt == ".yaml" || fileExt == ".yml" {
		if err := yaml.Unmarshal(data, &compListData); err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("File extension '%s' is not supported for component list files", fileExt)
	}

	return compListData.process(), nil
}

//Remove drops any component defintion with this particular name (independent whether it is listed as prequisite or component)
func (cl *ComponentList) Remove(compName string) {
	for idx, comp := range cl.Prerequisites {
		if comp.Name == compName {
			cl.Prerequisites = append(cl.Prerequisites[:idx], cl.Prerequisites[idx+1:]...)
		}
	}
	for idx, comp := range cl.Components {
		if comp.Name == compName {
			cl.Components = append(cl.Components[:idx], cl.Components[idx+1:]...)
		}
	}
}

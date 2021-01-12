package components

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// ComponentList collects component definitions
type ComponentList struct {
	prerequisites []ComponentDefinition
	components    []ComponentDefinition
}

// NewComponentList creates a new component list
func NewComponentList(componentsListPath string) (*ComponentList, error) {
	if componentsListPath == "" {
		return nil, fmt.Errorf("Path to components list file is required")
	}
	if _, err := os.Stat(componentsListPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("Components list file '%s' not found", componentsListPath)
	}

	clList := &ComponentList{}
	if err := clList.load(componentsListPath); err != nil {
		return nil, err
	}

	return clList, nil
}

func (cl *ComponentList) load(componentsListPath string) error {
	var componentList CompListFile

	// read file
	data, err := ioutil.ReadFile(componentsListPath)
	if err != nil {
		return err
	}

	if strings.HasSuffix(componentsListPath, ".json") {
		err = json.Unmarshal(data, &componentList)
	} else {
		err = yaml.Unmarshal(data, &componentList)
	}
	if err != nil {
		return err
	}

	// read prerequisites
	for _, compDef := range componentList.Prerequisites {
		if compDef.Namespace == "" {
			compDef.Namespace = componentList.DefaultNamespace
		}
		cl.prerequisites = append(cl.prerequisites, compDef)
	}

	// read components
	for _, compDef := range componentList.Components {
		if compDef.Namespace == "" {
			compDef.Namespace = componentList.DefaultNamespace
		}
		cl.components = append(cl.components, compDef)
	}

	return nil
}

// GetComponents returns all components on the list
func (cl *ComponentList) GetComponents() []ComponentDefinition {
	return cl.components
}

// GetPrerequisites returns all components on the list
func (cl *ComponentList) GetPrerequisites() []ComponentDefinition {
	return cl.prerequisites
}

// CompListFile is for components list marshalling
type CompListFile struct {
	DefaultNamespace string `yaml:"defaultNamespace" json:"defaultNamespace"`
	Prerequisites    []ComponentDefinition
	Components       []ComponentDefinition
}

// ComponentDefinition defines a component in components list
type ComponentDefinition struct {
	Name      string
	Namespace string
}

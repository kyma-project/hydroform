package k8s

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type K8sObject struct {
	Object runtime.Object
	GVK    *schema.GroupVersionKind
}

func ParseYamlToK8sObjects(decoder runtime.Decoder, yamlContent string) ([]K8sObject, error) {
	resources := strings.Split(yamlContent, "\n---\n")

	var objects = make([]K8sObject, 0, len(resources))
	for _, resource := range resources {
		if resource == "" {
			continue
		}

		object, groupVersionKind, err := decoder.Decode([]byte(resource), nil, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to decode resource: %w", err)
		}

		objects = append(objects, K8sObject{Object: object, GVK: groupVersionKind})
	}

	return objects, nil
}

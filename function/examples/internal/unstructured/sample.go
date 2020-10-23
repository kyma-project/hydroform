package unstructured

import "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

func NewSample(name, namespace string) unstructured.Unstructured {
	return unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "test.me.plz/v1alpha1",
			"kind":       "Sample",
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": namespace,
			},
		},
	}
}

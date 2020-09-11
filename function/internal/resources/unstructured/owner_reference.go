package unstructured

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func NewFunctionOwnerReference(uid, name string) unstructured.Unstructured {
	return NewOwnerReference("serverless.kyma-project.io", "Function", name, uid)
}

func NewOwnerReference(apiVersion, kind, name, uid string) unstructured.Unstructured {
	return unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": apiVersion,
			"kind":       kind,
			"name":       name,
			"uid":        uid,
		},
	}
}

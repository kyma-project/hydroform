package unstructured

import (
	"fmt"
	"github.com/kyma-incubator/hydroform/function/pkg/resources/types"
	"github.com/kyma-incubator/hydroform/function/pkg/workspace"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type Decorate = func(*unstructured.Unstructured) error

type Decorators = []Decorate

func withLabels(ls map[string]string) Decorate {
	return func(u *unstructured.Unstructured) (err error) {
		if u == nil {
			return fmt.Errorf("invalid value nil")
		}
		u.SetLabels(ls)
		return
	}
}

func decorateWithField(value interface{}, field string, fields ...string) Decorate {
	return func(u *unstructured.Unstructured) error {
		return unstructured.SetNestedField(u.Object, value, append([]string{field}, fields...)...)
	}
}

func withFunction(name, namespace string) Decorate {
	return func(u *unstructured.Unstructured) error {
		u.SetAPIVersion(functionApiVersion)
		u.SetKind("Function")
		u.SetName(name)
		u.SetNamespace(namespace)
		return nil
	}
}

func decorateWithMap(value map[string]interface{}, field string, fields ...string) Decorate {
	return func(u *unstructured.Unstructured) error {
		return unstructured.SetNestedMap(u.Object, value, append([]string{field}, fields...)...)
	}
}

func withLimits(limits workspace.ResourceList) Decorate {
	return decorateWithMap(limits, "spec", "resource", "limits")
}
func withRepository(value string) Decorate {
	return decorateWithField(value, "spec", "source")
}

func withRequests(requests workspace.ResourceList) Decorate {
	return decorateWithMap(requests, "spec", "resource", "requests")
}

func withRuntime(runtime types.Runtime) Decorate {
	return decorateWithField(runtime, "spec", "runtime")
}

func decorate(u *unstructured.Unstructured, ds Decorators) (err error) {
	for _, d := range ds {
		err = d(u)
		if err != nil {
			return
		}
	}
	return
}

package unstructured

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type Decorate = func(*unstructured.Unstructured) error

type Decorators = []Decorate

func decorateWithLabels(ls map[string]string) Decorate {
	return func(u *unstructured.Unstructured) (err error) {
		u.SetLabels(ls)
		return
	}
}

func decorateWithField(value interface{}, field string, fields ...string) Decorate {
	return func(u *unstructured.Unstructured) error {
		return unstructured.SetNestedField(u.Object, value, append([]string{field}, fields...)...)
	}
}

func decorateWithMetadata(name, namespace string) Decorate {
	return func(u *unstructured.Unstructured) error {
		u.SetName(name)
		u.SetNamespace(namespace)
		return nil
	}
}

var decorateWithFunction = func(u *unstructured.Unstructured) error {
	u.SetAPIVersion(functionApiVersion)
	u.SetKind("Function")
	return nil
}

var decorateWithGitRepository = func(u *unstructured.Unstructured) error {
	u.SetAPIVersion(gitRepositoryApiVersion)
	u.SetKind("GitRepository")
	return nil
}

func decorateWithMap(value map[string]interface{}, field string, fields ...string) Decorate {
	if len(value) == 0 {
		return func(u *unstructured.Unstructured) error {
			return nil
		}
	}
	return func(u *unstructured.Unstructured) error {
		return unstructured.SetNestedMap(u.Object, value, append([]string{field}, fields...)...)
	}
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

package unstructured

import (
	"context"
	"io"
	"strings"

	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

func FromString(ctx context.Context, v string) (unstructured.UnstructuredList, error) {
	r := strings.NewReader(v)
	return Load(ctx, r)
}

func Load(ctx context.Context, r io.Reader) (out unstructured.UnstructuredList, err error) {
	decoder := yaml.NewDecoder(r)
loop:
	for {
		select {
		case _, ok := <-ctx.Done():
			if !ok {
				break loop
			}
			err = ctx.Err()
			break loop
		default:
			var obj map[string]interface{}
			if err = decoder.Decode(&obj); err != nil {
				break loop
			}
			obj, err = runtime.DefaultUnstructuredConverter.ToUnstructured(&obj)
			if err != nil {
				break loop
			}
			u := unstructured.Unstructured{Object: obj}
			out.Items = append(out.Items, u)
		}
	}
	if err != nil && err != io.EOF {
		return
	}
	return out, nil
}

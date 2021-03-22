package client

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/apimachinery/pkg/watch"
)

type MapClient struct {
	Data       unstructured.UnstructuredList
	Namespace  string
	ApiVersion string
	Kind       string
	Resource   string
	Group      string
}

func (c *MapClient) groupResource() schema.GroupResource {
	return schema.GroupResource{
		Group:    c.Group,
		Resource: c.Resource,
	}
}

//TODO add dry run support and done support
func (c *MapClient) Create(ctx context.Context, obj *unstructured.Unstructured, options metav1.CreateOptions, subresources ...string) (*unstructured.Unstructured, error) {
	u, err := c.Get(ctx, obj.GetName(), metav1.GetOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}
	if u != nil {
		return nil, errors.NewAlreadyExists(c.groupResource(), obj.GetName())
	}

	copy := obj.DeepCopy()
	uid := uuid.NewUUID()
	copy.SetUID(uid)
	c.Data.Items = append(c.Data.Items, *copy)

	return copy, nil
}

func (c *MapClient) Update(ctx context.Context, obj *unstructured.Unstructured, options metav1.UpdateOptions, subresources ...string) (*unstructured.Unstructured, error) {
	panic("not implemented")
}

func (c *MapClient) UpdateStatus(ctx context.Context, obj *unstructured.Unstructured, options metav1.UpdateOptions) (*unstructured.Unstructured, error) {
	var err error = errors.NewNotFound(c.groupResource(), obj.GetName())
	out := unstructured.Unstructured{}
loop:
	for _, u := range c.Data.Items {
		select {
		case _, ok := <-ctx.Done():
			if !ok {
				break loop
			}
		default:
			if !c.isInScope(obj.GetAPIVersion(), obj.GetKind(), obj.GetNamespace()) || u.GetName() != obj.GetName() {
				break
			}
			var status map[string]interface{}
			var found bool
			status, found, err = unstructured.NestedMap(obj.Object, "status")
			if err == nil && found {
				err = unstructured.SetNestedMap(u.Object, status, "status")
			}
			break loop
		}
	}
	return &out, err
}

func (c *MapClient) Delete(ctx context.Context, name string, options metav1.DeleteOptions, subresources ...string) (err error) {
	err = errors.NewNotFound(c.groupResource(), name)
loop:
	for i, u := range c.Data.Items {
		select {
		case _, ok := <-ctx.Done():
			if !ok {
				err = nil
				break loop
			}
			return ctx.Err()
		default:
			if !c.isInScope(u.GetAPIVersion(), u.GetKind(), u.GetNamespace()) || name != u.GetName() {
				break
			}
			c.Data.Items = append(c.Data.Items[:i], c.Data.Items[i+1:]...)
			return nil
		}
	}
	return
}

func (c *MapClient) DeleteCollection(ctx context.Context, options metav1.DeleteOptions, listOptions metav1.ListOptions) error {
	panic("not implemented") // TODO: Implement
}

func (c *MapClient) get(name string, ul *unstructured.UnstructuredList) (*unstructured.Unstructured, error) {
	for _, u := range ul.Items {
		if u.GetName() == name {
			return &u, nil
		}
	}
	return nil, errors.NewNotFound(c.groupResource(), name)
}

func (c *MapClient) Get(ctx context.Context, name string, options metav1.GetOptions, subresources ...string) (*unstructured.Unstructured, error) {
	ul, err := c.List(ctx, metav1.ListOptions{
		TypeMeta: metav1.TypeMeta{
			APIVersion: c.ApiVersion,
			Kind:       c.Kind,
		},
	})
	if err != nil {
		return nil, err
	}
	u, err := c.get(name, ul)
	if err != nil {
		return nil, err
	}
	return u.DeepCopy(), nil
}

func (c *MapClient) isInScope(apiVersion, kind, namespace string) bool {
	return c.Namespace != namespace ||
		c.Kind != kind ||
		c.ApiVersion != apiVersion
}

func (c *MapClient) List(ctx context.Context, opts metav1.ListOptions) (*unstructured.UnstructuredList, error) {
	var out unstructured.UnstructuredList
	for _, u := range c.Data.Items {
		if c.Namespace != u.GetNamespace() ||
			c.Kind != u.GetKind() ||
			c.ApiVersion != u.GetAPIVersion() {
			continue
		}
		out.Items = append(out.Items, u)
	}
	return &out, nil
}

func (c *MapClient) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	panic("not supported")
}

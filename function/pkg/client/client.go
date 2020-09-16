package client

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type Client interface {
	Create(obj *unstructured.Unstructured, options v1.CreateOptions, subresources ...string) (*unstructured.Unstructured, error)
	Update(obj *unstructured.Unstructured, options v1.UpdateOptions, subresources ...string) (*unstructured.Unstructured, error)
	Delete(name string, options *v1.DeleteOptions, subresources ...string) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions, subresources ...string) (*unstructured.Unstructured, error)
	List(opts v1.ListOptions) (*unstructured.UnstructuredList, error)
}

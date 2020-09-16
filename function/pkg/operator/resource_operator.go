package operator

import (
	"github.com/kyma-incubator/hydroform/function/pkg/client"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type Resource interface {
	Apply(client.Client, ApplyOptions) (client.Status, error)
	GetGroupVersionResource() schema.GroupVersionResource
}

func NewResource(u unstructured.Unstructured, r schema.GroupVersionResource) Resource {
	return resource{
		Unstructured:         u,
		GroupVersionResource: r,
	}
}

func NewResourceTrigger(u unstructured.Unstructured) Resource {
	return NewResource(u, schema.GroupVersionResource{
		Group:    "eventing.knative.dev",
		Version:  "v1alpha1",
		Resource: "triggers",
	})
}

type resource struct {
	unstructured.Unstructured
	schema.GroupVersionResource
}

func (r resource) Apply(c client.Client, opts ApplyOptions) (client.Status, error) {
	r.Unstructured.SetOwnerReferences(opts.OwnerReferences)
	newLabels := mergeMap(r.Unstructured.GetLabels(), opts.Labels)
	r.Unstructured.SetLabels(newLabels)
	statusEntry, err := applyObject(c, r.Unstructured, opts.DryRun)
	return []client.StatusEntry{statusEntry}, err
}

func (r resource) GetGroupVersionResource() schema.GroupVersionResource {
	return r.GroupVersionResource
}

func mergeMap(labels map[string]string, extra map[string]string) map[string]string {
	for k, v := range extra {
		labels[k] = v
	}
	return labels
}

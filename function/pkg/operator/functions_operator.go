package operator

import (
	"github.com/kyma-incubator/hydroform/function/pkg/client"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	GVKFunction = schema.GroupVersionResource{
		Group:    "serverless.kyma-project.io",
		Version:  "v1alpha1",
		Resource: "functions",
	}
	GVKTriggers = schema.GroupVersionResource{
		Group:    "eventing.knative.dev",
		Version:  "v1alpha1",
		Resource: "triggers",
	}
)

type functionOperator struct {
	client.Client
	items []unstructured.Unstructured
}

func NewFunctionsOperator(c client.Client, u ...unstructured.Unstructured) Operator {
	return &functionOperator{
		Client: c,
		items:  u,
	}
}

func (p functionOperator) Apply(opts ApplyOptions) error {
	for _, u := range p.items {
		u.SetOwnerReferences(opts.OwnerReferences)
		// fire pre callbacks
		if err := fireCallbacks(&u, nil, opts.Pre...); err != nil {
			return err
		}
		new1, statusEntry, err := applyObject(p.Client, u, opts.DryRun)
		// fire post callbacks
		if err := fireCallbacks(statusEntry, err, opts.Callbacks.Post...); err != nil {
			return err
		}
		u.SetUnstructuredContent(new1.Object)
	}
	return nil
}

func (p functionOperator) Delete(opts DeleteOptions) error {
	for _, u := range p.items {
		// fire pre callbacks
		if err := fireCallbacks(&u, nil, opts.Pre...); err != nil {
			return err
		}
		status, err := deleteObject(p.Client, u, opts)
		// fire post callbacks
		if err := fireCallbacks(status, err, opts.Post...); err != nil {
			return err
		}
	}
	return nil
}

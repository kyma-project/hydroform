package operator

import (
	"context"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/kyma-project/hydroform/function/pkg/client"
)

type genericOperator struct {
	client.Client
	items []unstructured.Unstructured
}

func NewGenericOperator(c client.Client, u ...unstructured.Unstructured) Operator {
	return &genericOperator{
		Client: c,
		items:  u,
	}
}

func (p genericOperator) Apply(ctx context.Context, opts ApplyOptions) error {
	for i := range p.items {
		p.items[i].SetOwnerReferences(opts.OwnerReferences)
		// fire pre callbacks
		if err := fireCallbacks(&p.items[i], nil, opts.Pre...); err != nil {
			return err
		}

		applied, statusEntry, err := p.apply(ctx, p.items[i], opts)
		if err != nil {
			return err
		}

		// fire post callbacks
		if err := fireCallbacks(statusEntry, err, opts.Callbacks.Post...); err != nil {
			return err
		}
		p.items[i].SetUnstructuredContent(applied.Object)
	}
	return nil
}

func (p genericOperator) apply(ctx context.Context, item unstructured.Unstructured,
	opts ApplyOptions) (*unstructured.Unstructured, client.PostStatusEntry, error) {
	applied, statusEntry, err := applyObject(ctx, p.Client, item, opts.DryRun)
	if err != nil {
		return applied, statusEntry, err
	}
	if opts.WaitForApply {
		err = waitForObject(ctx, p.Client, *applied)
		if err != nil {
			return applied, statusEntry, err
		}
	}
	return applied, statusEntry, err
}

func (p genericOperator) Delete(ctx context.Context, opts DeleteOptions) error {
	for i := range p.items {
		// fire pre callbacks
		if err := fireCallbacks(&p.items[i], nil, opts.Pre...); err != nil {
			return err
		}
		status, err := deleteObject(ctx, p.Client, p.items[i], opts)
		if err != nil {
			return err
		}
		// fire post callbacks
		if err := fireCallbacks(status, err, opts.Post...); err != nil {
			return err
		}
	}
	return nil
}

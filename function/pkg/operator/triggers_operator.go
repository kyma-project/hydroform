package operator

import (
	"context"
	"fmt"

	"github.com/kyma-incubator/hydroform/function/pkg/client"
	"github.com/pkg/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const message = "ownerID"

type triggersOperator struct {
	items []unstructured.Unstructured
	client.Client
}

func NewTriggersOperator(c client.Client, u ...unstructured.Unstructured) Operator {
	return &triggersOperator{
		Client: c,
		items:  u,
	}
}

var errNotFound = errors.New("not found")

func (t triggersOperator) Apply(ctx context.Context, opts ApplyOptions) error {
	ownerID, found := findOwnerID(opts.OwnerReferences)
	if !found {
		return errors.Wrap(errNotFound, message)
	}
	if err := t.wipeRemoved(ctx, ownerID, opts); err != nil {
		return err
	}
	// apply all triggers
	for _, u := range t.items {
		u.SetOwnerReferences(opts.OwnerReferences)
		newLabels := mergeMap(u.GetLabels(), map[string]string{
			message: ownerID,
		})
		u.SetLabels(newLabels)
		// fire pre callbacks
		if err := fireCallbacks(&u, nil, opts.Pre...); err != nil {
			return err
		}
		new1, statusEntry, err := applyObject(ctx, t.Client, u, opts.DryRun)
		// fire post callbacks
		if err := fireCallbacks(statusEntry, err, opts.Post...); err != nil {
			return err
		}
		u.SetUnstructuredContent(new1.Object)
	}
	return nil
}

func (t triggersOperator) Delete(ctx context.Context, opts DeleteOptions) error {
	for _, u := range t.items {
		// fire pre callbacks
		if err := fireCallbacks(&u, nil, opts.Pre...); err != nil {
			return err
		}
		state, err := deleteObject(ctx, t.Client, u, opts)
		// fire post callbacks
		if err := fireCallbacks(state, err, opts.Post...); err != nil {
			return err
		}
	}
	return nil
}

func (t triggersOperator) wipeRemoved(ctx context.Context, ownerID string, opts ApplyOptions) error {
	list, err := t.Client.List(ctx, v1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", message, ownerID),
	})
	if err != nil {
		return err
	}

	policy := v1.DeletePropagationBackground

	// delete all removed triggers
	for _, item := range list.Items {
		if contains(t.items, item.GetName()) {
			continue
		}

		if err := fireCallbacks(&item, nil, opts.Pre...); err != nil {
			return err
		}
		// delete trigger, delegate flow ctrl to caller
		if err := t.Client.Delete(ctx, item.GetName(), v1.DeleteOptions{
			DryRun:            opts.DryRun,
			PropagationPolicy: &policy,
		}); err != nil {
			statusEntryFailed := client.NewPostStatusEntryFailed(item)
			if err := fireCallbacks(statusEntryFailed, err, opts.Post...); err != nil {
				return err
			}
		}
		statusEntryDeleted := client.NewPostStatusEntryDeleted(item)
		if err := fireCallbacks(statusEntryDeleted, nil, opts.Post...); err != nil {
			return err
		}
	}

	return nil
}

// seeks for uid of first Function kind or returns error
func findOwnerID(refs []v1.OwnerReference) (string, bool) {
	for _, ref := range refs {
		if ref.Kind == "Function" {
			return string(ref.UID), true
		}
	}
	return "", false
}

func contains(s []unstructured.Unstructured, name string) bool {
	for _, u := range s {
		if u.GetName() == name {
			return true
		}
	}
	return false
}

func mergeMap(l map[string]string, r map[string]string) map[string]string {
	if l == nil {
		return r
	}

	for k, v := range r {
		l[k] = v
	}
	return l
}

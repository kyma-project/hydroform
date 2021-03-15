package operator

import (
	"context"

	"github.com/kyma-incubator/hydroform/function/pkg/client"
	"github.com/kyma-incubator/hydroform/function/pkg/resources/types"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

type functionReference struct {
	name      string
	namespace string
}

type triggersOperator struct {
	fnRef functionReference
	items []unstructured.Unstructured
	client.Client
}

type predicate func(map[string]interface{}) (bool, error)

func NewTriggersOperator(c client.Client, fnName, fnNamespace string, u ...unstructured.Unstructured) Operator {
	return &triggersOperator{
		Client: c,
		items:  u,
		fnRef: functionReference{
			name:      fnName,
			namespace: fnNamespace,
		},
	}
}

// buildMatchRemovedTriggerPredicate - creates a predicate to match the triggers that should be deleted
func buildMatchRemovedTriggerPredicate(fnRef functionReference, items []unstructured.Unstructured) predicate {
	return func(obj map[string]interface{}) (bool, error) {
		var trigger types.Trigger
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj, &trigger); err != nil {
			return false, err
		}
		isRef := trigger.IsReference(fnRef.name, fnRef.namespace)
		if !isRef {
			return false, nil
		}

		containsTrigger := contains(items, trigger.ObjectMeta.Name)
		return !containsTrigger, nil
	}
}

var errNotFound = errors.New("not found")

func applyTriggers(ctx context.Context, c client.Client, p predicate, items []unstructured.Unstructured, opts ApplyOptions) error {
	if err := wipeRemoved(ctx, c, p, opts.Options); err != nil {
		return err
	}
	// apply all triggers
	for _, u := range items {
		u.SetOwnerReferences(opts.OwnerReferences)
		// fire pre callbacks
		if err := fireCallbacks(&u, nil, opts.Pre...); err != nil {
			return err
		}
		applied, statusEntry, err := applyObject(ctx, c, u, opts.DryRun)
		if opts.WaitForApply && applied != nil {
			err = waitForObject(ctx, c, *applied)
		}
		// fire post callbacks
		if err := fireCallbacks(statusEntry, err, opts.Post...); err != nil {
			return err
		}
		u.SetUnstructuredContent(applied.Object)
	}
	return nil
}

func (t triggersOperator) Apply(ctx context.Context, opts ApplyOptions) error {
	predicate := buildMatchRemovedTriggerPredicate(t.fnRef, t.items)
	return applyTriggers(ctx, t.Client, predicate, t.items, opts)
}

func deleteTriggers(ctx context.Context, c client.Client, items []unstructured.Unstructured, opts DeleteOptions) error {
	for _, u := range items {
		// fire pre callbacks
		if err := fireCallbacks(&u, nil, opts.Pre...); err != nil {
			return err
		}
		state, err := deleteObject(ctx, c, u, opts)
		// fire post callbacks
		if err := fireCallbacks(state, err, opts.Post...); err != nil {
			return err
		}
	}
	return nil
}

func (t triggersOperator) Delete(ctx context.Context, opts DeleteOptions) error {
	return deleteTriggers(ctx, t.Client, t.items, opts)
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

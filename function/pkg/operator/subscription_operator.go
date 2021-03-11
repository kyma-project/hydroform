package operator

import (
	"context"

	"github.com/kyma-incubator/hydroform/function/pkg/client"
	"github.com/kyma-incubator/hydroform/function/pkg/resources/types"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

type subscriptionOperator struct {
	fnRef functionReference
	items []unstructured.Unstructured
	client.Client
}

func NewSubscriptionOperator(c client.Client, fnName, fnNamespace string, u ...unstructured.Unstructured) Operator {
	return &subscriptionOperator{
		Client: c,
		items:  u,
		fnRef: functionReference{
			name:      fnName,
			namespace: fnNamespace,
		},
	}
}

func (t subscriptionOperator) Apply(ctx context.Context, opts ApplyOptions) error {
	predicate := buildMatchRemovedSubscriptionsPredicate(t.fnRef, t.items)

	if err := wipeRemoved(ctx, t.Client, predicate, opts.Options); err != nil {
		return err
	}
	// apply all subscriptions
	for _, u := range t.items {
		u.SetOwnerReferences(opts.OwnerReferences)
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

func (t subscriptionOperator) Delete(ctx context.Context, opts DeleteOptions) error {
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

// buildMatchRemovedSubscriptionsPredicate - creates a predicate to match the subscriptions that should be deleted
func buildMatchRemovedSubscriptionsPredicate(fnRef functionReference, items []unstructured.Unstructured) func(map[string]interface{}) (bool, error) {
	return func(obj map[string]interface{}) (bool, error) {
		var subscription types.Subscription
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj, &subscription); err != nil {
			return false, err
		}
		isRef := subscription.IsReference(fnRef.name, fnRef.namespace)
		if !isRef {
			return false, nil
		}

		containsSubscription := contains(items, subscription.ObjectMeta.Name)
		return !containsSubscription, nil
	}
}

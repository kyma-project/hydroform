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

func NewSubscriptionsOperator(c client.Client, fnName, fnNamespace string, u ...unstructured.Unstructured) Operator {
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
	return applyTriggers(ctx, t.Client, predicate, t.items, opts)
}

func (t subscriptionOperator) Delete(ctx context.Context, opts DeleteOptions) error {
	return deleteTriggers(ctx, t.Client, t.items, opts)
}

// buildMatchRemovedSubscriptionsPredicate - creates a predicate to match the subscriptions that should be deleted
func buildMatchRemovedSubscriptionsPredicate(fnRef functionReference, items []unstructured.Unstructured) predicate {
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

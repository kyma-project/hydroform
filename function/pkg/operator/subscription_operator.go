package operator

import (
	"context"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/kyma-project/hydroform/function/pkg/client"
	operator_types "github.com/kyma-project/hydroform/function/pkg/operator/types"
	"github.com/kyma-project/hydroform/function/pkg/resources/types"
	"github.com/kyma-project/hydroform/function/pkg/workspace"
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
	return applySubscriptions(ctx, t.Client, predicate, t.items, opts)
}

func (t subscriptionOperator) Delete(ctx context.Context, opts DeleteOptions) error {
	return deleteSubscriptions(ctx, t.Client, t.items, opts)
}

type functionReference struct {
	name      string
	namespace string
}

type predicate func(map[string]interface{}) (bool, error)

func deleteSubscriptions(ctx context.Context, c client.Client, items []unstructured.Unstructured,
	opts DeleteOptions) error {
	for i := range items {
		// fire pre callbacks
		if err := fireCallbacks(&items[i], nil, opts.Pre...); err != nil {
			return err
		}
		state, err := deleteObject(ctx, c, items[i], opts)
		// fire post callbacks
		if err := fireCallbacks(state, err, opts.Post...); err != nil {
			return err
		}
	}
	return nil
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

// buildMatchRemovedSubscriptionsPredicate - creates a predicate to match the subscriptions that should be deleted
func buildMatchRemovedSubscriptionsPredicate(fnRef functionReference, items []unstructured.Unstructured) predicate {
	return func(obj map[string]interface{}) (bool, error) {
		var subscription types.SubscriptionV1alpha1
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj, &subscription); err != nil {
			return false, err
		}
		isRef := subscription.IsReference(fnRef.name, fnRef.namespace)
		isOwnerRef := (len(subscription.OwnerReferences) == 0 || isOwnerReference(subscription.GetOwnerReferences(),
			fnRef.name))
		if !isRef || !isOwnerRef {
			return false, nil
		}

		containsSubscription := contains(items, subscription.ObjectMeta.Name)
		return !containsSubscription, nil
	}
}

func applySubscriptions(ctx context.Context, c client.Client, p predicate, items []unstructured.Unstructured,
	opts ApplyOptions) error {
	if err := wipeRemoved(ctx, c, p, opts.Options); err != nil {
		return err
	}
	// apply all subscriptions
	for i := range items {
		items[i].SetOwnerReferences(opts.OwnerReferences)
		// fire pre callbacks
		if err := fireCallbacks(&items[i], nil, opts.Pre...); err != nil {
			return err
		}
		applied, statusEntry, err := applyObject(ctx, c, items[i], opts.DryRun)
		if opts.WaitForApply && applied != nil {
			err = waitForObject(ctx, c, *applied)
			if err != nil {
				return err
			}
		}
		// fire post callbacks
		if err := fireCallbacks(statusEntry, err, opts.Post...); err != nil {
			return err
		}
		items[i].SetUnstructuredContent(applied.Object)
	}
	return nil
}

func SubscriptionGVR(subscription workspace.SchemaVersion) schema.GroupVersionResource {
	if subscription == workspace.SchemaVersionV0 {
		return operator_types.GVRSubscriptionV1alpha1
	}
	return operator_types.GVRSubscriptionV1alpha2
}

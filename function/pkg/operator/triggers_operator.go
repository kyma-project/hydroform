package operator

import (
	"context"

	"github.com/kyma-incubator/hydroform/function/pkg/client"
	"github.com/kyma-incubator/hydroform/function/pkg/resources/types"
	"github.com/pkg/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

const message = "ownerID"

type FnRef struct {
	name      string
	namespace string
}

type triggersOperator struct {
	fnRef FnRef
	items []unstructured.Unstructured
	client.Client
}

func NewTriggersOperator(c client.Client, fnName, fnNamespace string, u ...unstructured.Unstructured) Operator {
	return &triggersOperator{
		Client: c,
		items:  u,
		fnRef: FnRef{
			name:      fnName,
			namespace: fnNamespace,
		},
	}
}

var errNotFound = errors.New("not found")

func (t triggersOperator) Apply(ctx context.Context, opts ApplyOptions) error {
	ownerID, found := findOwnerID(opts.OwnerReferences)
	if !found {
		return errors.Wrap(errNotFound, message)
	}

	predicate := func(obj map[string]interface{}) (bool, error) {
		var trigger types.Trigger
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj, &trigger); err != nil {
			return false, err
		}
		isRef := trigger.IsReference(t.fnRef.name, t.fnRef.namespace)
		found := contains(t.items, trigger.Metadata.Name)
		ok := isRef && !found
		return ok, nil
	}

	if err := wipeRemoved(ctx, t.Client, predicate, opts.Options); err != nil {
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

package operator

import (
	"fmt"
	"github.com/kyma-incubator/hydroform/function/pkg/client"
	"github.com/pkg/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const message = "functionUID"

type triggersOperator struct {
	items []unstructured.Unstructured
	client.Client
}

func NewTriggersOperator(c client.Client, u ...unstructured.Unstructured) Operator {
	return triggersOperator{
		Client: c,
		items:  u,
	}
}

var errNotFound = errors.New("not found")

func (t triggersOperator) Apply(opts ApplyOptions, c ...Callback) error {
	functionUID, found := findFunctionUID(opts.OwnerReferences)
	if !found {
		return errors.Wrap(errNotFound, message)
	}
	if err := t.wipeRemoved(functionUID, opts, c...); err != nil {
		return err
	}
	// apply all triggers
	for _, u := range t.items {
		u.SetOwnerReferences(opts.OwnerReferences)
		newLabels := mergeMap(u.GetLabels(), map[string]string{
			message: functionUID,
		})
		u.SetLabels(newLabels)
		new1, statusEntry, err := applyObject(t.Client, u, opts.DryRun)
		if err := fireCallbacks(statusEntry, err, c); err != nil {
			return err
		}
		u.SetUnstructuredContent(new1.Object)
	}
	return nil
}

func (t triggersOperator) Delete(opts DeleteOptions, c ...Callback) error {
	for _, u := range t.items {
		state, err := deleteObject(t.Client, u, opts)
		if err := fireCallbacks(state, err, c); err != nil {
			return err
		}
	}
	return nil
}

func (t triggersOperator) wipeRemoved(functionUID string, opts ApplyOptions, c ...Callback) error {
	list, err := t.Client.List(v1.ListOptions{
		LabelSelector: fmt.Sprintf("functionUID=%s", functionUID),
	})
	if err != nil {
		return err
	}

	policy := v1.DeletePropagationForeground

	// delete all removed triggers
	for _, item := range list.Items {
		if contains(t.items, item.GetName()) {
			continue
		}
		// delete trigger, delegate flow ctrl to caller
		if err := t.Client.Delete(item.GetName(), &v1.DeleteOptions{
			DryRun:            opts.DryRun,
			PropagationPolicy: &policy,
		}); err != nil {
			statusEntryFailed := client.NewStatusEntryFailed(item)
			if err := fireCallbacks(statusEntryFailed, err, c); err != nil {
				return err
			}
		}
		statusEntryDeleted := client.NewStatusEntryDeleted(item)
		if err := fireCallbacks(statusEntryDeleted, nil, c); err != nil {
			return err
		}
	}

	return nil
}

// seeks for uid of first Function kind or returns error
func findFunctionUID(refs []v1.OwnerReference) (string, bool) {
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

func mergeMap(labels map[string]string, extra map[string]string) map[string]string {
	for k, v := range extra {
		labels[k] = v
	}
	return labels
}

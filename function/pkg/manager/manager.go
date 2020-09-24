package manager

import (
	"errors"
	"github.com/kyma-incubator/hydroform/function/pkg/client"
    "github.com/kyma-incubator/hydroform/function/pkg/operator"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type manager struct {
	operators map[operator.Operator][]operator.Operator
}

func NewManager(operators map[operator.Operator][]operator.Operator) manager {
	return manager{
		operators: operators,
	}
}

func (m manager) Do(options ManagerOptions) error {
	err := m.manageOperators(options)
	if err != nil {
		if options.OnError == PurgeOnError {
			m.purgeParents(options)
		}
		return err
	}
	return nil
}

func (m *manager) manageOperators(options ManagerOptions) error {
	for parent, children := range m.operators {
		references, err := m.useOperator(parent, options, nil)
		if err != nil {
			return err
		}

		for _, resource := range children {
			_, err := m.useOperator(resource, options, references)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *manager) useOperator(opr operator.Operator, options ManagerOptions, references []metav1.OwnerReference) ([]metav1.OwnerReference, error) {
	newRefs := &OwnerReferenceList{}
	if opr == nil {
		return newRefs.List, nil
	}

	applyOpts := operator.ApplyOptions{
		DryRun:          m.getDryRunFlag(options.DryRun),
		OwnerReferences: references,
		Callbacks:       m.ownerReferenceCallback(options.Callbacks, newRefs),
	}
	return newRefs.List, opr.Apply(applyOpts)
}

func (m *manager) purgeParents(options ManagerOptions) {
	deleteOptions := operator.DeleteOptions{
		DryRun:              m.getDryRunFlag(options.DryRun),
		DeletionPropagation: metav1.DeletePropagationForeground,
		Callbacks:           options.Callbacks,
	}

	for opr := range m.operators {
		if opr == nil {
			continue
		}
		_ = opr.Delete(deleteOptions)
	}
}

func (m *manager) getDryRunFlag(dryRun bool) []string {
	var flags []string
	if dryRun {
		flags = append(flags, metav1.DryRunAll)
	}
	return flags
}

type OwnerReferenceList struct {
	List []metav1.OwnerReference
}

func (m *manager) ownerReferenceCallback(callbacks operator.Callbacks,list *OwnerReferenceList) operator.Callbacks {
	ownerReferenceCallback := func(v interface{}, err error) error {
		entry, ok := v.(client.PostStatusEntry)
		if !ok {
			return errors.New("can't parse interface{} to StatusEntry interface")
		}
		if err == nil && entry.StatusType != client.StatusTypeFailed {
			list.List = append(list.List, metav1.OwnerReference{
				APIVersion: entry.GetAPIVersion(),
				Kind:       entry.GetKind(),
				Name:       entry.GetName(),
				UID:        entry.GetUID(),
			})
		}
		return err
	}

	if list != nil {
		callbacks.Post = append(callbacks.Post, ownerReferenceCallback)
	}
	return callbacks
}

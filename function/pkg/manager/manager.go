package manager

import (
	"context"
	"github.com/kyma-incubator/hydroform/function/pkg/client"
	"github.com/kyma-incubator/hydroform/function/pkg/operator"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Manager interface {
	Do(ctx context.Context, options Options) error
	AddParent(object operator.Operator, childes []operator.Operator)
}

type parent struct {
	object   operator.Operator
	children []operator.Operator
}

type manager struct {
	operators []parent
}

func NewManager() Manager {
	return &manager{}
}

func (m *manager) AddParent(object operator.Operator, childes []operator.Operator) {
	m.operators = append(m.operators, parent{
		object:   object,
		children: childes,
	})
}

func (m manager) Do(ctx context.Context, options Options) error {
	err := m.manageOperators(ctx, options)
	if err != nil {
		if options.OnError == PurgeOnError {
			m.purgeParents(options)
		}
		return err
	}
	return nil
}

func (m *manager) manageOperators(ctx context.Context, options Options) error {
	for _, parent := range m.operators {
		references, err := m.useOperator(ctx, parent.object, options, nil)
		if err != nil {
			return err
		}

		for _, resource := range parent.children {
			_, err := m.useOperator(ctx, resource, options, references)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *manager) useOperator(ctx context.Context, opr operator.Operator, options Options, references []metav1.OwnerReference) ([]metav1.OwnerReference, error) {
	newRefs := OwnerReferenceList{}
	if opr == nil {
		return newRefs, nil
	}

	callbacks := options.Callbacks
	if options.SetOwnerReferences {
		callbacks = m.ownerReferenceCallback(options.Callbacks, &newRefs)
	}
	applyOpts := operator.ApplyOptions{
		OwnerReferences: references,
		Options: operator.Options{
			DryRun:       m.getDryRunFlag(options.DryRun),
			Callbacks:    callbacks,
			WaitForApply: options.WaitForApply,
		},
	}
	return newRefs, opr.Apply(ctx, applyOpts)
}

func (m *manager) purgeParents(options Options) {
	deleteOptions := operator.DeleteOptions{
		DeletionPropagation: metav1.DeletePropagationForeground,
		Options: operator.Options{
			DryRun:    m.getDryRunFlag(options.DryRun),
			Callbacks: options.Callbacks,
		},
	}

	for _, parent := range m.operators {
		if parent.object == nil {
			continue
		}
		_ = parent.object.Delete(context.Background(), deleteOptions)
	}
}

func (m *manager) getDryRunFlag(dryRun bool) []string {
	var flags []string
	if dryRun {
		flags = append(flags, metav1.DryRunAll)
	}
	return flags
}

type OwnerReferenceList []metav1.OwnerReference

func (m *manager) ownerReferenceCallback(callbacks operator.Callbacks, list *OwnerReferenceList) operator.Callbacks {
	if list == nil {
		return callbacks
	}

	ownerReferenceCallback := func(v interface{}, err error) error {
		entry, ok := v.(client.PostStatusEntry)
		if !ok {
			return errors.New("can't parse interface{} to StatusEntry interface")
		}
		if err == nil && entry.StatusType != client.StatusTypeApplyFailed && entry.StatusType != client.StatusTypeDeleteFailed {
			*list = append(*list, metav1.OwnerReference{
				APIVersion: entry.GetAPIVersion(),
				Kind:       entry.GetKind(),
				Name:       entry.GetName(),
				UID:        entry.GetUID(),
			})
		}
		return err
	}

	callbacks.Post = append(callbacks.Post, ownerReferenceCallback)
	return callbacks
}

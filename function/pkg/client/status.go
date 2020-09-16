package client

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
)

type StatusType int

func (t StatusType) String() string {
	switch t {
	case StatusTypeDeleted:
		return "deleted"
	case StatusTypeSkipped:
		return "skipped"
	case StatusTypeFailed:
		return "failed"
	case StatusTypeCreated:
		return "created"
	case StatusTypeUpdated:
		return "updated"
	default:
		return "<unknown>"
	}
}

type StatusEntry struct {
	StatusType
	IdentifiedNamedKind
}

func (e StatusEntry) toOwnerReference() v1.OwnerReference {
	return v1.OwnerReference{
		APIVersion: e.GetAPIVersion(),
		Kind:       e.GetKind(),
		Name:       e.GetName(),
		UID:        e.GetUID(),
	}
}

func NewStatusEntryFailed(u unstructured.Unstructured) StatusEntry {
	return StatusEntry{
		StatusType:          StatusTypeFailed,
		IdentifiedNamedKind: &u,
	}
}

func NewStatusEntrySkipped(u unstructured.Unstructured) StatusEntry {
	return StatusEntry{
		StatusType:          StatusTypeSkipped,
		IdentifiedNamedKind: &u,
	}
}

func NewStatusEntryUpdated(u unstructured.Unstructured) StatusEntry {
	return StatusEntry{
		StatusType:          StatusTypeUpdated,
		IdentifiedNamedKind: &u,
	}
}

func NewStatusEntryCreated(u unstructured.Unstructured) StatusEntry {
	return StatusEntry{
		StatusType:          StatusTypeCreated,
		IdentifiedNamedKind: &u,
	}
}

func NewStatusEntryDeleted(u unstructured.Unstructured) StatusEntry {
	return StatusEntry{
		StatusType:          StatusTypeDeleted,
		IdentifiedNamedKind: &u,
	}
}

type IdentifiedNamedKind interface {
	GetKind() string
	GetName() string
	GetUID() types.UID
	GetAPIVersion() string
}

type Status []StatusEntry

func (s Status) GetOwnerReferences() []v1.OwnerReference {
	size := len(s)
	if size == 0 {
		return nil
	}
	var result []v1.OwnerReference
	for _, entry := range s {
		result = append(result, entry.toOwnerReference())
	}
	return result
}

const (
	StatusTypeCreated StatusType = iota
	StatusTypeUpdated
	StatusTypeSkipped
	StatusTypeFailed
	StatusTypeDeleted
)

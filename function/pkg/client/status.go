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
		return "unknown"
	}
}

type PostStatusEntry struct {
	StatusType
	unstructured.Unstructured
}

func (e PostStatusEntry) toOwnerReference() v1.OwnerReference {
	return v1.OwnerReference{
		APIVersion: e.Unstructured.GetAPIVersion(),
		Kind:       e.Unstructured.GetKind(),
		Name:       e.Unstructured.GetName(),
		UID:        e.Unstructured.GetUID(),
	}
}

func NewPostStatusEntryFailed(u unstructured.Unstructured) PostStatusEntry {
	return PostStatusEntry{
		StatusType:   StatusTypeFailed,
		Unstructured: u,
	}
}

func NewPostStatusEntrySkipped(u unstructured.Unstructured) PostStatusEntry {
	return PostStatusEntry{
		StatusType:   StatusTypeSkipped,
		Unstructured: u,
	}
}

func NewPostStatusEntryUpdated(u unstructured.Unstructured) PostStatusEntry {
	return PostStatusEntry{
		StatusType:   StatusTypeUpdated,
		Unstructured: u,
	}
}

func NewStatusEntryCreated(u unstructured.Unstructured) PostStatusEntry {
	return PostStatusEntry{
		StatusType:   StatusTypeCreated,
		Unstructured: u,
	}
}

func NewPostStatusEntryDeleted(u unstructured.Unstructured) PostStatusEntry {
	return PostStatusEntry{
		StatusType:   StatusTypeDeleted,
		Unstructured: u,
	}
}

type NamedKindVersion interface {
	GetKind() string
	GetName() string
	GetAPIVersion() string
}

type IdentifiedNamedKindVersion interface {
	NamedKindVersion
	GetUID() types.UID
}

type Status []PostStatusEntry

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

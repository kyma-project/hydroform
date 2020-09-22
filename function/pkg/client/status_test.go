package client

import (
	"reflect"
	"testing"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var testData = func() unstructured.Unstructured {
	result := unstructured.Unstructured{}
	result.SetName("test-name")
	result.SetAPIVersion("test-api-version")
	result.SetKind("test-kind")
	result.SetUID("test-uid")
	return result
}()

func TestNewStatusEntryCreated(t *testing.T) {

	type args struct {
		u unstructured.Unstructured
	}
	tests := []struct {
		name string
		args args
		want StatusEntry
	}{
		{
			name: "happy path",
			args: args{
				u: testData,
			},
			want: StatusEntry{
				StatusType:                 StatusTypeCreated,
				IdentifiedNamedKindVersion: &testData,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewStatusEntryCreated(tt.args.u); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewStatusEntryCreated() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewStatusEntryDeleted(t *testing.T) {
	type args struct {
		u unstructured.Unstructured
	}
	tests := []struct {
		name string
		args args
		want StatusEntry
	}{
		{
			name: "happy path",
			args: args{
				u: testData,
			},
			want: StatusEntry{
				StatusType:                 StatusTypeDeleted,
				IdentifiedNamedKindVersion: &testData,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewStatusEntryDeleted(tt.args.u); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewStatusEntryDeleted() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewStatusEntryFailed(t *testing.T) {
	type args struct {
		u unstructured.Unstructured
	}
	tests := []struct {
		name string
		args args
		want StatusEntry
	}{
		{
			name: "happy path",
			args: args{
				u: testData,
			},
			want: StatusEntry{
				StatusType:                 StatusTypeFailed,
				IdentifiedNamedKindVersion: &testData,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewStatusEntryFailed(tt.args.u); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewStatusEntryFailed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewStatusEntrySkipped(t *testing.T) {
	type args struct {
		u unstructured.Unstructured
	}
	tests := []struct {
		name string
		args args
		want StatusEntry
	}{
		{
			name: "happy path",
			args: args{
				u: testData,
			},
			want: StatusEntry{
				StatusType:                 StatusTypeSkipped,
				IdentifiedNamedKindVersion: &testData,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewStatusEntrySkipped(tt.args.u); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewStatusEntrySkipped() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewStatusEntryUpdated(t *testing.T) {
	type args struct {
		u unstructured.Unstructured
	}
	tests := []struct {
		name string
		args args
		want StatusEntry
	}{
		{
			name: "happy path",
			args: args{
				u: testData,
			},
			want: StatusEntry{
				StatusType:                 StatusTypeUpdated,
				IdentifiedNamedKindVersion: &testData,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewStatusEntryUpdated(tt.args.u); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewStatusEntryUpdated() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStatusEntry_toOwnerReference(t *testing.T) {
	type fields struct {
		StatusType                 StatusType
		IdentifiedNamedKindVersion IdentifiedNamedKindVersion
	}
	tests := []struct {
		name   string
		fields fields
		want   v1.OwnerReference
	}{
		{
			name: "happy path",
			fields: fields{
				StatusType:                 StatusTypeFailed,
				IdentifiedNamedKindVersion: &testData,
			},
			want: v1.OwnerReference{
				APIVersion: testData.GetAPIVersion(),
				Kind:       testData.GetKind(),
				Name:       testData.GetName(),
				UID:        testData.GetUID(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := StatusEntry{
				StatusType:                 tt.fields.StatusType,
				IdentifiedNamedKindVersion: tt.fields.IdentifiedNamedKindVersion,
			}
			if got := e.toOwnerReference(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toOwnerReference() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStatusType_String(t *testing.T) {
	tests := []struct {
		name string
		t    StatusType
		want string
	}{
		{
			name: "failed",
			t:    StatusTypeFailed,
			want: "failed",
		},
		{
			name: "updated",
			t:    StatusTypeUpdated,
			want: "updated",
		},
		{
			name: "skipped",
			t:    StatusTypeSkipped,
			want: "skipped",
		},
		{
			name: "created",
			t:    StatusTypeCreated,
			want: "created",
		},
		{
			name: "deleted",
			t:    StatusTypeDeleted,
			want: "deleted",
		},
		{
			name: "unknown",
			t:    StatusType(-1),
			want: "unknown",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.t.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStatus_GetOwnerReferences(t *testing.T) {
	tests := []struct {
		name string
		s    Status
		want []v1.OwnerReference
	}{
		{
			name: "nil",
			s:    nil,
			want: nil,
		},
		{
			name: "happy path",
			s:    []StatusEntry{NewStatusEntryUpdated(testData)},
			want: []v1.OwnerReference{
				{
					APIVersion: testData.GetAPIVersion(),
					Kind:       testData.GetKind(),
					Name:       testData.GetName(),
					UID:        testData.GetUID(),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.GetOwnerReferences(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetOwnerReferences() = %v, want %v", got, tt.want)
			}
		})
	}
}

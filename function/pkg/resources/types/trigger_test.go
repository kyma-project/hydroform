package types

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestTrigger_IsReference(t *testing.T) {
	type fields struct {
		ObjectMeta metav1.ObjectMeta
		Spec       TriggerSpec
	}
	type args struct {
		name      string
		namespace string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "is reference",
			fields: fields{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-trigger-name",
					Namespace: "test-trigger-namespace",
				},
				Spec: TriggerSpec{
					Subscriber: TriggerSubscriber{
						Reference: TriggerReference{
							Kind:      "Service",
							Name:      "test-function-name",
							Namespace: "test-function-namespace",
						},
					},
				},
			},
			args: args{
				name:      "test-function-name",
				namespace: "test-function-namespace",
			},
			want: true,
		},
		{
			name: "is not reference",
			fields: fields{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-trigger-name",
					Namespace: "test-trigger-namespace",
				},
				Spec: TriggerSpec{
					Subscriber: TriggerSubscriber{
						Reference: TriggerReference{
							Kind:      "Service",
							Name:      "test-function-name",
							Namespace: "test-other-function-namespace",
						},
					},
				},
			},
			args: args{
				name:      "test-function-name",
				namespace: "test-function-namespace",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trigger := Trigger{
				ObjectMeta: tt.fields.ObjectMeta,
				Spec:       tt.fields.Spec,
			}
			if got := trigger.IsReference(tt.args.name, tt.args.namespace); got != tt.want {
				t.Errorf("Trigger.IsReference() = %v, want %v", got, tt.want)
			}
		})
	}
}

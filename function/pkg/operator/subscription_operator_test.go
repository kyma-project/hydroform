package operator

import (
	"fmt"
	"testing"

	"github.com/kyma-incubator/hydroform/function/pkg/resources/types"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

func newTestSubscription(name, namespace string) (unstructured.Unstructured, error) {
	subscription := types.Subscription{
		TypeMeta: v1.TypeMeta{
			Kind:       "Subscription",
			APIVersion: GVRSubscription.Version,
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: types.SubscriptionSpec{
			Filter: types.Filter{
				Filters: []types.EventFilter{},
			},
			ID:       "",
			Protocol: "",
			ProtocolSettings: types.ProtocolSettings{
				ContentMode:     "",
				ExemptHandshake: false,
				Qos:             "",
				WebhookAuth: types.WebhookAuth{
					ClientID:     "",
					ClientSecret: "",
					GrantType:    "",
					Scope:        []string{},
					TokenURL:     "",
					Type:         "",
				},
			},
			Sink: fmt.Sprintf("%s.%s.svc.cluster.local", name, namespace),
		},
	}
	subscriptionObject, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&subscription)
	if err != nil {
		return unstructured.Unstructured{}, err
	}
	return unstructured.Unstructured{Object: subscriptionObject}, nil
}

func Test_buildMatchRemovedSubscriptionsPredicate(t *testing.T) {

	var subscription1 unstructured.Unstructured
	var subscription2 unstructured.Unstructured

	for i, s := range []*unstructured.Unstructured{
		&subscription1, &subscription2,
	} {
		var err error
		(*s), err = newTestSubscription(fmt.Sprintf("test-%d", i+1), "test-namespace")
		if err != nil {
			t.Fatal(err)
		}
	}

	type args struct {
		fnRef        functionReference
		items        []unstructured.Unstructured
		subscription unstructured.Unstructured
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "no match 1",
			args: args{
				items: []unstructured.Unstructured{subscription1, subscription2},
				fnRef: functionReference{
					name:      "test-1",
					namespace: "test-namespace",
				},
				subscription: subscription1,
			},
			want: false,
		},
		{
			name: "no match 2",
			args: args{
				items: []unstructured.Unstructured{subscription1, subscription2},
				fnRef: functionReference{
					name:      "test-me",
					namespace: "test-namespace",
				},
				subscription: subscription1,
			},
			want: false,
		},
		{
			name: "match",
			args: args{
				items: []unstructured.Unstructured{subscription2},
				fnRef: functionReference{
					name:      "test-1",
					namespace: "test-namespace",
				},
				subscription: subscription1,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			predicate := buildMatchRemovedSubscriptionsPredicate(tt.args.fnRef, tt.args.items)
			got, err := predicate(tt.args.subscription.Object)
			if (err != nil) != tt.wantErr {
				t.Errorf("buildMatchRemovedSubscriptionsPredicate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("predicate() bool = %v, want %v", got, tt.want)
			}
		})
	}
}

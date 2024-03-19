package unstructured

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/kyma-project/hydroform/function/pkg/resources/types"
	"github.com/kyma-project/hydroform/function/pkg/workspace"
	"github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var errExpectedError = fmt.Errorf("expected error")

func Test_newSubscriptionsV1alpha1(t *testing.T) {
	type args struct {
		cfg workspace.Cfg
		f   toUnstructured
	}
	tests := []struct {
		name    string
		args    args
		want    []unstructured.Unstructured
		wantErr bool
	}{
		{
			name: "Err",
			args: args{
				cfg: workspace.Cfg{
					Name:      "should-fail",
					Namespace: "failed-tests",
					Runtime:   "nodejs20",
					Subscriptions: []workspace.Subscription{
						{
							Name: "fixme",
							V0: &workspace.SubscriptionV0{
								Protocol: "fixme",
								Filter: workspace.Filter{
									Dialect: "fixme",
									Filters: []workspace.EventFilter{
										{
											EventSource: workspace.EventSource{
												Property: "source",
												Type:     "exact",
												Value:    "b",
											},
											EventType: workspace.EventType{
												Property: "type",
												Type:     "exact",
												Value:    "c.a",
											},
										},
									},
								},
							},
						},
					},
				},
				f: func(obj interface{}) (map[string]interface{}, error) {
					return nil, errExpectedError
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newSubscriptionsV1alpha1(tt.args.cfg, tt.args.f)
			if (err != nil) != tt.wantErr {
				t.Errorf("newSubscriptionsV1alpha1() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			g := gomega.NewWithT(t)
			g.Expect(got).To(gomega.Equal(tt.want))
		})
	}
}

func Test_newSubscriptionsV1alpha2(t *testing.T) {
	type args struct {
		cfg workspace.Cfg
		f   toUnstructured
	}
	tests := []struct {
		name    string
		args    args
		want    []unstructured.Unstructured
		wantErr bool
	}{
		{
			name: "Err",
			args: args{
				cfg: workspace.Cfg{
					Name:      "should-fail",
					Namespace: "failed-tests",
					Runtime:   "nodejs20",
					Subscriptions: []workspace.Subscription{
						{
							Name: "fixme",
							V1: &workspace.SubscriptionV1{
								TypeMatching: "matchingType",
								Source:       "source",
								Types:        []string{"type1", "type2", "type3"},
							},
						},
					},
				},
				f: func(obj interface{}) (map[string]interface{}, error) {
					return nil, errExpectedError
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newSubscriptionsV1alpha2(tt.args.cfg, tt.args.f)
			if (err != nil) != tt.wantErr {
				t.Errorf("newSubscriptionsV1alpha2() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			g := gomega.NewWithT(t)
			g.Expect(got).To(gomega.Equal(tt.want))
		})
	}
}

func Test_joinNonEmpty(t *testing.T) {
	type args struct {
		elems []string
		sep   string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "empty slice",
			args: args{
				elems: []string{},
				sep:   "!?",
			},
			want: "",
		},
		{
			name: "nil slice",
			args: args{
				elems: nil,
				sep:   "!?",
			},
			want: "",
		},
		{
			name: "just one",
			args: args{
				elems: []string{"one"},
				sep:   "!?",
			},
			want: "one",
		},
		{
			name: "multiple",
			args: args{
				elems: []string{"hello", "there"},
				sep:   "+",
			},
			want: "hello+there",
		},
		{
			name: "multiple with empty element",
			args: args{
				elems: []string{"test", "", "me"},
				sep:   "*",
			},
			want: "test*me",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := joinNonEmpty(tt.args.elems, tt.args.sep); got != tt.want {
				t.Errorf("joinNonEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewSubscriptions(t *testing.T) {
	type args struct {
		cfg workspace.Cfg
	}
	tests := []struct {
		name    string
		args    args
		want    []unstructured.Unstructured
		wantErr bool
	}{
		{
			name: "OK",
			args: args{
				cfg: workspace.Cfg{
					Name:      "test-name",
					Namespace: "test-namespace",
					Labels: map[string]string{
						"test": "me",
					},
					Runtime: types.Python39,
					Subscriptions: []workspace.Subscription{
						{
							V0: &workspace.SubscriptionV0{
								Protocol: "NATS",
								Filter: workspace.Filter{
									Dialect: "klingon",
									Filters: []workspace.EventFilter{
										{
											EventSource: workspace.EventSource{
												Property: "source",
												Type:     "exact",
												Value:    "",
											},
											EventType: workspace.EventType{
												Property: "type",
												Type:     "exact",
												Value:    "c",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			want: []unstructured.Unstructured{
				{
					Object: map[string]interface{}{
						"apiVersion": "eventing.kyma-project.io/v1alpha1",
						"kind":       "Subscription",
						"metadata": map[string]interface{}{
							"labels": map[string]interface{}{
								"test": "me",
							},
							"name":              "test-name-0",
							"namespace":         "test-namespace",
							"creationTimestamp": nil,
						},
						"spec": map[string]interface{}{
							"protocol": "NATS",
							"filter": map[string]interface{}{
								"dialect": "klingon",
								"filters": []interface{}{
									map[string]interface{}{
										"eventSource": map[string]interface{}{
											"property": "source",
											"type":     "exact",
											"value":    "",
										},
										"eventType": map[string]interface{}{
											"property": "type",
											"type":     "exact",
											"value":    "c",
										},
									},
								},
							},
							"protocolsettings": map[string]interface{}{},
							"sink":             "http://test-name.test-namespace.svc.cluster.local",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewSubscriptions(tt.args.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewSubscriptions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewSubscriptions() = %v, want %v", got, tt.want)
			}
		})
	}
}

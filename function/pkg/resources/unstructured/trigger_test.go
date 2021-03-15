package unstructured

import (
	"reflect"
	"testing"

	"github.com/kyma-incubator/hydroform/function/pkg/workspace"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestNewTriggers(t *testing.T) {
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
			name: "Ok",
			args: args{
				cfg: workspace.Cfg{
					Name:      "test-name",
					Namespace: "test-namespace",
					Labels: map[string]string{
						"test": "me",
					},
					Runtime: "python38",
					Triggers: []workspace.Trigger{
						{
							EventTypeVersion: "a",
							Source:           "b",
							Type:             "c",
						},
					},
				},
			},
			want: []unstructured.Unstructured{
				{
					Object: map[string]interface{}{
						"apiVersion": "eventing.knative.dev/v1alpha1",
						"kind":       "Trigger",
						"metadata": map[string]interface{}{
							"labels": map[string]interface{}{
								"test": "me",
							},
							"name":              "test-name-b",
							"namespace":         "test-namespace",
							"creationTimestamp": nil,
						},
						"spec": map[string]interface{}{
							"broker": "default",
							"filter": map[string]interface{}{
								"attributes": map[string]interface{}{
									"eventtypeversion": "a",
									"source":           "b",
									"type":             "c",
								},
							},
							"subscriber": map[string]interface{}{
								"ref": map[string]interface{}{
									"apiVersion": "v1",
									"kind":       "Service",
									"name":       "test-name",
									"namespace":  "test-namespace",
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewTriggers(tt.args.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewTriggers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewTriggers() got = %v, want %v", got, tt.want)
			}
		})
	}
}

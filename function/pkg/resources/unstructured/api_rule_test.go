package unstructured

import (
	"github.com/kyma-incubator/hydroform/function/pkg/workspace"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"reflect"
	"testing"
)

func TestNewApiRule(t *testing.T) {
	type args struct {
		cfg workspace.Cfg
	}
	tests := []struct {
		name    string
		args    args
		wantOut unstructured.Unstructured
		wantErr bool
	}{
		{
			name: "Simple APiRule",
			args: args{
				cfg: workspace.Cfg{
					Name:      "function-name",
					Namespace: "test-ns",
					Labels: map[string]string{
						"test": "me",
					},
					ApiRule: workspace.ApiRule{
						Host:    "test-host",
						Name:    "test-name",
						Port:    "80",
						Methods: []string{"POST"},
						Handler: "test-handler",
						Path:    "test-path",
					},
				},
			},
			wantOut: unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": apiRuleApiVersion,
					"kind":       apiRuleKind,
					"metadata": map[string]interface{}{
						"name":              "test-name",
						"namespace":         "test-ns",
						"creationTimestamp": nil,
						"labels": map[string]interface{}{
							"test": "me",
						},
					},
					"spec": map[string]interface{}{
						"service": map[string]interface{}{
							"host": "test-host",
							"name": "function-name",
							"port": "80",
						},
						"rules": []map[string]interface{}{
							{
								//			"accessStrategies": []map[string]interface{}{
								//				{
								//					//"config": "map[]",
								//					"handler": "test-handler",
								//				}},
								//			"methods": []string{"POST"},
								"path": "test-path",
							},
						},
						"gateway": apiRuleGateway,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOut, err := NewApiRule(tt.args.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewApiRule() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotOut, tt.wantOut) {
				t.Errorf("NewApiRule() gotOut = %v, want %v", gotOut, tt.wantOut)
			}
		})
	}
}

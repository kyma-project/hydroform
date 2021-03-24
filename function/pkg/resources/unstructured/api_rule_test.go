package unstructured

import (
	"github.com/kyma-incubator/hydroform/function/pkg/workspace"
	"github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"testing"
)

func TestNewApiRule(t *testing.T) {
	g := gomega.NewWithT(t)
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
						Host:           "test-host",
						Name:           "test-name",
						Port:           "80",
						Methods:        []string{"POST"},
						Handler:        "test-handler",
						Path:           "test-path",
						JwksUrls:       []string{"test"},
						TrustedIssuers: []string{"test"},
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
						"rules": []interface{}{
							map[string]interface{}{
								"methods": []interface{}{"POST"},
								"path":    "test-path",
								"accessStrategies": []interface{}{
									map[string]interface{}{
										"handler": "test-handler",
										"config": map[string]interface{}{
											"jwks_urls":       []interface{}{"test"},
											"trusted_issuers": []interface{}{"test"},
										},
									},
								},
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
			gotOut, _ := NewApiRule(tt.args.cfg)
			g.Expect(gotOut).To(gomega.Equal(tt.wantOut))
		})
	}
}

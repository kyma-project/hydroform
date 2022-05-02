package unstructured

import (
	"testing"

	"github.com/kyma-project/hydroform/function/pkg/workspace"
	"github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestNewApiRule(t *testing.T) {
	g := gomega.NewWithT(t)
	type args struct {
		cfg            workspace.Cfg
		defaultAddress string
	}
	tests := []struct {
		name    string
		args    args
		wantOut []unstructured.Unstructured
		wantErr gomega.OmegaMatcher
	}{
		{
			name: "Should return empty map",
			args: args{
				cfg: workspace.Cfg{
					Name:      "function-name",
					Namespace: "test-ns",
					Labels: map[string]string{
						"test": "me",
					},
				},
			},
			wantOut: nil,
			wantErr: gomega.BeNil(),
		},
		{
			name: "Should return defaulted ApiRule",
			args: args{
				defaultAddress: "kyma.local",
				cfg: workspace.Cfg{
					Name:      "test-name",
					Namespace: "test-ns",
					APIRules: []workspace.APIRule{
						{
							Rules: []workspace.Rule{
								{
									Methods: []string{"PUT"},
								},
							},
						},
					},
				},
			},
			wantOut: []unstructured.Unstructured{
				{
					Object: map[string]interface{}{
						"apiVersion": apiRuleAPIVersion,
						"kind":       apiRuleKind,
						"metadata": map[string]interface{}{
							"name":              "test-name",
							"namespace":         "test-ns",
							"creationTimestamp": nil,
						},
						"spec": map[string]interface{}{
							"service": map[string]interface{}{
								"host": "test-name.kyma.local",
								"name": "test-name",
								"port": int64(80),
							},
							"rules": []interface{}{
								map[string]interface{}{
									"methods": []interface{}{"PUT"},
									"path":    "/.*",
									"accessStrategies": []interface{}{
										map[string]interface{}{
											"handler": "allow",
										},
									},
								},
							},
							"gateway": "kyma-gateway.kyma-system.svc.cluster.local",
						},
					},
				},
			},
			wantErr: gomega.BeNil(),
		},
		{
			name: "Should return configurated ApiRule",
			args: args{
				cfg: workspace.Cfg{
					Name:      "function-name",
					Namespace: "test-ns",
					Labels: map[string]string{
						"test": "me",
					},
					APIRules: []workspace.APIRule{
						{
							Name:    "test-name",
							Gateway: "test-gateway",
							Service: workspace.Service{
								Host: "test-host",
								Port: 9090,
							},
							Rules: []workspace.Rule{
								{
									Path:    "test-path",
									Methods: []string{"POST"},
									AccessStrategies: []workspace.AccessStrategie{
										{
											Config: workspace.AccessStrategieConfig{
												JwksUrls:       []string{"test-jwks"},
												TrustedIssuers: []string{"test-trusted"},
												RequiredScope:  []string{"test-required"},
											},
											Handler: "test-handler",
										},
									},
								},
							},
						},
					},
				},
			},
			wantErr: gomega.BeNil(),
			wantOut: []unstructured.Unstructured{
				{
					Object: map[string]interface{}{
						"apiVersion": apiRuleAPIVersion,
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
								"port": int64(9090),
							},
							"rules": []interface{}{
								map[string]interface{}{
									"methods": []interface{}{"POST"},
									"path":    "test-path",
									"accessStrategies": []interface{}{
										map[string]interface{}{
											"handler": "test-handler",
											"config": map[string]interface{}{
												"jwks_urls":       []interface{}{"test-jwks"},
												"trusted_issuers": []interface{}{"test-trusted"},
												"required_scope":  []interface{}{"test-required"},
											},
										},
									},
								},
							},
							"gateway": "test-gateway",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			gotOut, err := NewAPIRule(tt.args.cfg, tt.args.defaultAddress)
			g.Expect(gotOut).To(gomega.Equal(tt.wantOut))
			g.Expect(err).To(tt.wantErr)
		})
	}
}

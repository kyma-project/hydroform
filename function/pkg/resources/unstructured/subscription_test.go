package unstructured

import (
	"testing"

	"github.com/kyma-incubator/hydroform/function/pkg/workspace"
	"github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

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
						"apiVersion": "eventing.kyma-project.io/v1alpha1",
						"kind":       "Subscription",
						"metadata": map[string]interface{}{
							"labels": map[string]interface{}{
								"test": "me",
							},
							"name":              "test-name-b",
							"namespace":         "test-namespace",
							"creationTimestamp": nil,
						},
						"spec": map[string]interface{}{
							"protocol": "NATS",
							"filter": map[string]interface{}{
								"filters": []interface{}{
									map[string]interface{}{
										"eventSource": map[string]interface{}{
											"property": "source",
											"type":     "exact",
											"value":    "b",
										},
										"eventType": map[string]interface{}{
											"property": "type",
											"type":     "exact",
											"value":    "c.a",
										},
									},
								},
							},
							"protocolsettings": map[string]interface{}{
								"exemptHandshake": true,
								"qos":             "AT-LEAST-ONCE",
								"webhookAuth": map[string]interface{}{
									"clientId":     "",
									"clientSecret": "",
									"grantType":    "",
									"tokenUrl":     "",
									"type":         "",
								},
							},
							"sink": "test-name.test-namespace.svc.cluster.local",
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
			g := gomega.NewWithT(t)
			g.Expect(got).To(gomega.Equal(tt.want))
		})
	}
}

package unstructured

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/kyma-project/hydroform/function/pkg/workspace"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestNewPublicGitRepository(t *testing.T) {
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
			name: "public repo",
			args: args{
				cfg: workspace.Cfg{
					Namespace: "test-ns",
					Source: workspace.Source{
						Type: workspace.SourceTypeGit,
						SourceGit: workspace.SourceGit{
							URL:                   "test-url",
							Repository:            "test-repository",
							CredentialsType:       "key",
							CredentialsSecretName: "test-secretname",
						},
					},
				},
			},
			wantOut: unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": gitRepositoryAPIVersion,
					"kind":       "GitRepository",
					"metadata": map[string]interface{}{
						"name":              "test-repository",
						"namespace":         "test-ns",
						"creationTimestamp": nil,
					},
					"spec": map[string]interface{}{
						"source": map[string]interface{}{
							"gitRepository": map[string]interface{}{
								"auth": map[string]interface{}{
									"type":       "key",
									"secretName": "test-secretname",
								},
								"url": "test-url",
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
			gotOut, err := NewPublicGitRepository(tt.args.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewPublicGitRepository() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			_ = json.NewEncoder(os.Stdout).Encode(&gotOut)
			if !reflect.DeepEqual(gotOut, tt.wantOut) {
				t.Errorf("NewPublicGitRepository() gotOut = %v, want %v", gotOut, tt.wantOut)
			}
		})
	}
}

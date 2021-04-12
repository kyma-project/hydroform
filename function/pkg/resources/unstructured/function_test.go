package unstructured

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/kyma-incubator/hydroform/function/pkg/resources/types"
	"github.com/kyma-incubator/hydroform/function/pkg/workspace"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func Test_newFunction(t *testing.T) {
	type args struct {
		cfg      workspace.Cfg
		readFile ReadFile
	}
	tests := []struct {
		name    string
		args    args
		wantOut unstructured.Unstructured
		wantErr bool
	}{
		{
			name: "inline - OK",
			args: args{
				readFile: func(filename string) ([]byte, error) {
					switch filename {
					case "/test/path/test.my.source":
						return []byte("test-source-content"), nil
					case "/test/path/test.my.deps":
						return []byte("test-deps-content"), nil
					default:
						return []byte{}, nil
					}
				},
				cfg: workspace.Cfg{
					Name:      "test-name",
					Namespace: "test-ns",
					Labels: map[string]string{
						"test": "me",
					},
					Resources: workspace.Resources{
						Limits: workspace.ResourceList{
							workspace.ResourceNameCPU:    "1",
							workspace.ResourceNameMemory: "10M",
						},
						Requests: workspace.ResourceList{
							workspace.ResourceNameCPU:    "1",
							workspace.ResourceNameMemory: "10M",
						},
					},
					Runtime: types.Python38,
					Source: workspace.Source{
						Type: workspace.SourceTypeGit,
						SourceInline: workspace.SourceInline{
							SourcePath:        "/test/path",
							SourceHandlerName: "test.my.source",
							DepsHandlerName:   "test.my.deps",
						},
					},
					Env: []workspace.EnvVar{
						{
							Name:  "TEST_ENV",
							Value: "test",
						},
						{
							Name: "TEST_ENV_SECRET",
							ValueFrom: &workspace.EnvVarSource{
								SecretKeyRef: &workspace.SecretKeySelector{
									Name: "secretName",
									Key:  "secretKey",
								},
							},
						},
						{
							Name: "TEST_ENV_CM",
							ValueFrom: &workspace.EnvVarSource{
								ConfigMapKeyRef: &workspace.ConfigMapKeySelector{
									Name: "configMapName",
									Key:  "configMapKey",
								},
							},
						},
					},
				},
			},
			wantOut: unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": functionAPIVersion,
					"kind":       "Function",
					"metadata": map[string]interface{}{
						"name":              "test-name",
						"namespace":         "test-ns",
						"creationTimestamp": nil,
					},
					"spec": map[string]interface{}{
						"runtime": "python38",
						"resources": map[string]interface{}{
							"limits": workspace.ResourceList{
								workspace.ResourceNameCPU:    "1",
								workspace.ResourceNameMemory: "10M",
							},
							"requests": workspace.ResourceList{
								workspace.ResourceNameCPU:    "1",
								workspace.ResourceNameMemory: "10M",
							},
						},
						"source": "test-source-content",
						"deps":   "test-deps-content",
						"labels": map[string]interface{}{
							"test": "me",
						},
						"env": []interface{}{
							map[string]interface{}{
								"name":  "TEST_ENV",
								"value": "test",
							},
							map[string]interface{}{
								"name": "TEST_ENV_SECRET",
								"valueFrom": map[string]interface{}{
									"secretKeyRef": map[string]interface{}{
										"name": "secretName",
										"key":  "secretKey",
									},
								},
							},
							map[string]interface{}{
								"name": "TEST_ENV_CM",
								"valueFrom": map[string]interface{}{
									"configMapKeyRef": map[string]interface{}{
										"name": "configMapName",
										"key":  "configMapKey",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "empty deps inline - OK",
			args: args{
				readFile: func(filename string) ([]byte, error) {
					switch filename {
					case "/test/path/test.my.source":
						return []byte("test-source-content"), nil
					default:
						return []byte{}, nil
					}
				},
				cfg: workspace.Cfg{
					Name:      "test-name",
					Namespace: "test-ns",
					Labels: map[string]string{
						"test": "me",
					},
					Resources: workspace.Resources{
						Limits: workspace.ResourceList{
							workspace.ResourceNameCPU:    "1",
							workspace.ResourceNameMemory: "10M",
						},
						Requests: workspace.ResourceList{
							workspace.ResourceNameCPU:    "1",
							workspace.ResourceNameMemory: "10M",
						},
					},
					Runtime: types.Python38,
					Source: workspace.Source{
						Type: workspace.SourceTypeGit,
						SourceInline: workspace.SourceInline{
							SourcePath:        "/test/path",
							SourceHandlerName: "test.my.source",
							DepsHandlerName:   "test.my.deps",
						},
					},
					Subscriptions: []workspace.Subscription{
						{
							Name:     "test",
							Protocol: "",
							Filter: workspace.Filter{
								Filters: []workspace.EventFilter{
									{
										EventSource: workspace.EventFilterProperty{
											Property: "type",
											Type:     "exact",
											Value:    "test-subscription-type.test-subscription-etv",
										},
										EventType: workspace.EventFilterProperty{
											Property: "source",
											Type:     "exact",
											Value:    "test-subscription-source",
										},
									},
								},
							},
						},
					},
					APIRules: []workspace.APIRule{
						{
							Name: "test-name",
							Service: workspace.Service{
								Host: "test-host",
								Port: 80,
							},
							Rules: []workspace.Rule{
								{
									Path:    "test-path",
									Methods: []string{"POST"},
									AccessStrategies: []workspace.AccessStrategie{
										{
											Handler: "test-handler",
										},
									},
								},
							},
						},
					},
				},
			},
			wantOut: unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": functionAPIVersion,
					"kind":       "Function",
					"metadata": map[string]interface{}{
						"name":              "test-name",
						"namespace":         "test-ns",
						"creationTimestamp": nil,
					},
					"spec": map[string]interface{}{
						"runtime": "python38",
						"resources": map[string]interface{}{
							"limits": workspace.ResourceList{
								workspace.ResourceNameCPU:    "1",
								workspace.ResourceNameMemory: "10M",
							},
							"requests": workspace.ResourceList{
								workspace.ResourceNameCPU:    "1",
								workspace.ResourceNameMemory: "10M",
							},
						},
						"source": "test-source-content",
						"labels": map[string]interface{}{
							"test": "me",
						},
					},
				},
			},
		},
		{
			name: "inline - minimal",
			args: args{
				readFile: func(filename string) ([]byte, error) {
					switch filename {
					case "/test/path/test.my.source":
						return []byte("test-source-content"), nil
					case "/test/path/test.my.deps":
						return []byte("test-deps-content"), nil
					default:
						return []byte{}, nil
					}
				},
				cfg: workspace.Cfg{
					Name:      "test-name",
					Namespace: "test-ns",
					Runtime:   types.Python38,
					Source: workspace.Source{
						Type: workspace.SourceTypeGit,
						SourceInline: workspace.SourceInline{
							SourcePath:        "/test/path",
							SourceHandlerName: "test.my.source",
							DepsHandlerName:   "test.my.deps",
						},
					},
				},
			},
			wantOut: unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": functionAPIVersion,
					"kind":       "Function",
					"metadata": map[string]interface{}{
						"name":              "test-name",
						"namespace":         "test-ns",
						"creationTimestamp": nil,
					},
					"spec": map[string]interface{}{
						"runtime": "python38",
						"source":  "test-source-content",
						"deps":    "test-deps-content",
					},
				},
			},
		},
		{
			name: "inline - only resources requests",
			args: args{
				readFile: func(filename string) ([]byte, error) {
					switch filename {
					case "/test/path/test.my.source":
						return []byte("test-source-content"), nil
					case "/test/path/test.my.deps":
						return []byte("test-deps-content"), nil
					default:
						return []byte{}, nil
					}
				},
				cfg: workspace.Cfg{
					Name:      "test-name",
					Namespace: "test-ns",
					Runtime:   types.Python38,
					Source: workspace.Source{
						Type: workspace.SourceTypeGit,
						SourceInline: workspace.SourceInline{
							SourcePath:        "/test/path",
							SourceHandlerName: "test.my.source",
							DepsHandlerName:   "test.my.deps",
						},
					},
					Resources: workspace.Resources{
						Requests: workspace.ResourceList{
							workspace.ResourceNameCPU:    "1",
							workspace.ResourceNameMemory: "10M",
						},
					},
				},
			},
			wantOut: unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": functionAPIVersion,
					"kind":       "Function",
					"metadata": map[string]interface{}{
						"name":              "test-name",
						"namespace":         "test-ns",
						"creationTimestamp": nil,
					},
					"spec": map[string]interface{}{
						"runtime": "python38",
						"source":  "test-source-content",
						"deps":    "test-deps-content",
						"resources": map[string]interface{}{
							"requests": workspace.ResourceList{
								workspace.ResourceNameCPU:    "1",
								workspace.ResourceNameMemory: "10M",
							},
						},
					},
				},
			},
		},
		{
			name: "inline - only resources limits",
			args: args{
				readFile: func(filename string) ([]byte, error) {
					switch filename {
					case "/test/path/test.my.source":
						return []byte("test-source-content"), nil
					case "/test/path/test.my.deps":
						return []byte("test-deps-content"), nil
					default:
						return []byte{}, nil
					}
				},
				cfg: workspace.Cfg{
					Name:      "test-name",
					Namespace: "test-ns",
					Runtime:   types.Python38,
					Source: workspace.Source{
						Type: workspace.SourceTypeGit,
						SourceInline: workspace.SourceInline{
							SourcePath:        "/test/path",
							SourceHandlerName: "test.my.source",
							DepsHandlerName:   "test.my.deps",
						},
					},
					Resources: workspace.Resources{
						Limits: workspace.ResourceList{
							workspace.ResourceNameCPU:    "1",
							workspace.ResourceNameMemory: "10M",
						},
					},
				},
			},
			wantOut: unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": functionAPIVersion,
					"kind":       "Function",
					"metadata": map[string]interface{}{
						"name":              "test-name",
						"namespace":         "test-ns",
						"creationTimestamp": nil,
					},
					"spec": map[string]interface{}{
						"runtime": "python38",
						"source":  "test-source-content",
						"deps":    "test-deps-content",
						"resources": map[string]interface{}{
							"limits": workspace.ResourceList{
								workspace.ResourceNameCPU:    "1",
								workspace.ResourceNameMemory: "10M",
							},
						},
					},
				},
			},
		},
		{
			name: "inline - only resources cpu",
			args: args{
				readFile: func(filename string) ([]byte, error) {
					switch filename {
					case "/test/path/test.my.source":
						return []byte("test-source-content"), nil
					case "/test/path/test.my.deps":
						return []byte("test-deps-content"), nil
					default:
						return []byte{}, nil
					}
				},
				cfg: workspace.Cfg{
					Name:      "test-name",
					Namespace: "test-ns",
					Runtime:   types.Python38,
					Source: workspace.Source{
						Type: workspace.SourceTypeGit,
						SourceInline: workspace.SourceInline{
							SourcePath:        "/test/path",
							SourceHandlerName: "test.my.source",
							DepsHandlerName:   "test.my.deps",
						},
					},
					Resources: workspace.Resources{
						Limits: workspace.ResourceList{
							workspace.ResourceNameCPU: "1",
						},
						Requests: workspace.ResourceList{
							workspace.ResourceNameCPU: "1",
						},
					},
				},
			},
			wantOut: unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": functionAPIVersion,
					"kind":       "Function",
					"metadata": map[string]interface{}{
						"name":              "test-name",
						"namespace":         "test-ns",
						"creationTimestamp": nil,
					},
					"spec": map[string]interface{}{
						"runtime": "python38",
						"source":  "test-source-content",
						"deps":    "test-deps-content",
						"resources": map[string]interface{}{
							"limits": workspace.ResourceList{
								workspace.ResourceNameCPU: "1",
							},
							"requests": workspace.ResourceList{
								workspace.ResourceNameCPU: "1",
							},
						},
					},
				},
			},
		},
		{
			name: "inline - read err",
			args: args{
				readFile: func(filename string) ([]byte, error) {
					return nil, fmt.Errorf("read error")
				},
				cfg: workspace.Cfg{
					Name:      "test-name",
					Namespace: "test-ns",
					Labels: map[string]string{
						"test": "me",
					},
					Runtime: types.Python38,
					Source: workspace.Source{
						Type: workspace.SourceTypeGit,
						SourceInline: workspace.SourceInline{
							SourcePath:        "/test/path",
							SourceHandlerName: "test.my.source",
							DepsHandlerName:   "test.my.deps",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "inline - only resources memory",
			args: args{
				readFile: func(filename string) ([]byte, error) {
					switch filename {
					case "/test/path/test.my.source":
						return []byte("test-source-content"), nil
					case "/test/path/test.my.deps":
						return []byte("test-deps-content"), nil
					default:
						return []byte{}, nil
					}
				},
				cfg: workspace.Cfg{
					Name:      "test-name",
					Namespace: "test-ns",
					Runtime:   types.Python38,
					Source: workspace.Source{
						Type: workspace.SourceTypeGit,
						SourceInline: workspace.SourceInline{
							SourcePath:        "/test/path",
							SourceHandlerName: "test.my.source",
							DepsHandlerName:   "test.my.deps",
						},
					},
					Resources: workspace.Resources{
						Limits: workspace.ResourceList{
							workspace.ResourceNameMemory: "10M",
						},
						Requests: workspace.ResourceList{
							workspace.ResourceNameMemory: "10M",
						},
					},
				},
			},
			wantOut: unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": functionAPIVersion,
					"kind":       "Function",
					"metadata": map[string]interface{}{
						"name":              "test-name",
						"namespace":         "test-ns",
						"creationTimestamp": nil,
					},
					"spec": map[string]interface{}{
						"runtime": "python38",
						"source":  "test-source-content",
						"deps":    "test-deps-content",
						"resources": map[string]interface{}{
							"limits": workspace.ResourceList{
								workspace.ResourceNameMemory: "10M",
							},
							"requests": workspace.ResourceList{
								workspace.ResourceNameMemory: "10M",
							},
						},
					},
				},
			},
		},
		{
			name: "inline - unknown runtime err",
			args: args{
				cfg: workspace.Cfg{
					Name:      "test-name",
					Namespace: "test-ns",
					Labels: map[string]string{
						"test": "me",
					},
					Runtime: "unknown",
					Source: workspace.Source{
						Type: workspace.SourceTypeGit,
						SourceInline: workspace.SourceInline{
							SourcePath:        "/test/path",
							SourceHandlerName: "test.my.source",
							DepsHandlerName:   "test.my.deps",
						},
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOut, err := newFunction(tt.args.cfg, tt.args.readFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("newFunction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotOut, tt.wantOut) {
				t.Errorf("newFunction() gotOut = %v, want %v", gotOut, tt.wantOut)
			}
		})
	}
}

func Test_newGitFunction(t *testing.T) {
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
			name: "git - OK",
			args: args{
				cfg: workspace.Cfg{
					Name:      "test-name",
					Namespace: "test-ns",
					Labels: map[string]string{
						"test": "me",
					},
					Resources: workspace.Resources{
						Limits: workspace.ResourceList{
							workspace.ResourceNameCPU:    "1",
							workspace.ResourceNameMemory: "10M",
						},
						Requests: workspace.ResourceList{
							workspace.ResourceNameCPU:    "1",
							workspace.ResourceNameMemory: "10M",
						},
					},
					Runtime: types.Python38,
					Source: workspace.Source{
						Type: workspace.SourceTypeGit,
						SourceGit: workspace.SourceGit{
							URL:       "test-url",
							Reference: "test-reference",
							BaseDir:   "test-base-dir",
						},
					},
					Subscriptions: []workspace.Subscription{
						{
							Name:     "fixme",
							Protocol: "fixme",
							Filter: workspace.Filter{
								Dialect: "fixme",
								Filters: []workspace.EventFilter{
									{
										EventSource: workspace.EventFilterProperty{
											Property: "source",
											Type:     "exact",
											Value:    "test-subscription-source",
										},
										EventType: workspace.EventFilterProperty{
											Property: "type",
											Type:     "exact",
											Value:    "test-subscription-type.test-subscription-etv",
										},
									},
								},
							},
						},
					},
					APIRules: []workspace.APIRule{
						{
							Name: "test-name",
							Service: workspace.Service{
								Host: "test-host",
								Port: 80,
							},
							Rules: []workspace.Rule{
								{
									Path:    "test-path",
									Methods: []string{"POST"},
									AccessStrategies: []workspace.AccessStrategie{
										{
											Handler: "test-handler",
										},
									},
								},
							},
						},
					},
					Env: []workspace.EnvVar{
						{
							Name:  "TEST_ENV",
							Value: "test",
						},
						{
							Name: "TEST_ENV_SECRET",
							ValueFrom: &workspace.EnvVarSource{
								SecretKeyRef: &workspace.SecretKeySelector{
									Name: "secretName",
									Key:  "secretKey",
								},
							},
						},
						{
							Name: "TEST_ENV_CM",
							ValueFrom: &workspace.EnvVarSource{
								ConfigMapKeyRef: &workspace.ConfigMapKeySelector{
									Name: "configMapName",
									Key:  "configMapKey",
								},
							},
						},
					},
				},
			},
			wantOut: unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": functionAPIVersion,
					"kind":       "Function",
					"metadata": map[string]interface{}{
						"name":              "test-name",
						"namespace":         "test-ns",
						"creationTimestamp": nil,
					},
					"spec": map[string]interface{}{
						"runtime": "python38",
						"resources": map[string]interface{}{
							"limits": workspace.ResourceList{
								workspace.ResourceNameCPU:    "1",
								workspace.ResourceNameMemory: "10M",
							},
							"requests": workspace.ResourceList{
								workspace.ResourceNameCPU:    "1",
								workspace.ResourceNameMemory: "10M",
							},
						},
						"source":    "test-name",
						"baseDir":   "test-base-dir",
						"reference": "test-reference",
						"type":      "git",
						"labels": map[string]interface{}{
							"test": "me",
						},
						"env": []interface{}{
							map[string]interface{}{
								"name":  "TEST_ENV",
								"value": "test",
							},
							map[string]interface{}{
								"name": "TEST_ENV_SECRET",
								"valueFrom": map[string]interface{}{
									"secretKeyRef": map[string]interface{}{
										"name": "secretName",
										"key":  "secretKey",
									},
								},
							},
							map[string]interface{}{
								"name": "TEST_ENV_CM",
								"valueFrom": map[string]interface{}{
									"configMapKeyRef": map[string]interface{}{
										"name": "configMapName",
										"key":  "configMapKey",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "override repository git - OK",
			args: args{
				cfg: workspace.Cfg{
					Name:      "test-name",
					Namespace: "test-ns",
					Labels: map[string]string{
						"test": "me",
					},
					Resources: workspace.Resources{
						Limits: workspace.ResourceList{
							workspace.ResourceNameCPU:    "1",
							workspace.ResourceNameMemory: "10M",
						},
						Requests: workspace.ResourceList{
							workspace.ResourceNameCPU:    "1",
							workspace.ResourceNameMemory: "10M",
						},
					},
					Runtime: types.Python38,
					Source: workspace.Source{
						Type: workspace.SourceTypeGit,
						SourceGit: workspace.SourceGit{
							URL:        "test-url",
							Repository: "test-repository",
							Reference:  "test-reference",
							BaseDir:    "test-base-dir",
						},
					},
					Subscriptions: []workspace.Subscription{
						{
							Name:     "fixmne",
							Protocol: "fixme",
							Filter: workspace.Filter{
								Dialect: "fixme",
								Filters: []workspace.EventFilter{
									{
										EventSource: workspace.EventFilterProperty{
											Property: "source",
											Type:     "exact",
											Value:    "test-subscription-source",
										},
										EventType: workspace.EventFilterProperty{
											Property: "type",
											Type:     "exact",
											Value:    "test-subscription-type.test-subscription-etv",
										},
									},
								},
							},
						},
					},
					APIRules: []workspace.APIRule{
						{
							Name: "test-name",
							Service: workspace.Service{
								Host: "test-host",
								Port: 80,
							},
							Rules: []workspace.Rule{
								{
									Path:    "test-path",
									Methods: []string{"POST"},
									AccessStrategies: []workspace.AccessStrategie{
										{
											Handler: "test-handler",
										},
									},
								},
							},
						},
					},
				},
			},
			wantOut: unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": functionAPIVersion,
					"kind":       "Function",
					"metadata": map[string]interface{}{
						"name":              "test-name",
						"namespace":         "test-ns",
						"creationTimestamp": nil,
					},
					"spec": map[string]interface{}{
						"runtime": "python38",
						"resources": map[string]interface{}{
							"limits": workspace.ResourceList{
								workspace.ResourceNameCPU:    "1",
								workspace.ResourceNameMemory: "10M",
							},
							"requests": workspace.ResourceList{
								workspace.ResourceNameCPU:    "1",
								workspace.ResourceNameMemory: "10M",
							},
						},
						"source":    "test-repository",
						"baseDir":   "test-base-dir",
						"reference": "test-reference",
						"type":      "git",
						"labels": map[string]interface{}{
							"test": "me",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOut, err := newGitFunction(tt.args.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("newGitFunction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotOut, tt.wantOut) {
				t.Errorf("newGitFunction() gotOut = %v, want %v", gotOut, tt.wantOut)
			}
		})
	}
}

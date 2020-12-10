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
					Triggers: []workspace.Trigger{
						{
							Version: "test-trigger-etv",
							Source:  "test-trigger-source",
							Type:    "test-trigger-type",
						},
					},
				},
			},
			wantOut: unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": functionApiVersion,
					"kind":       "Function",
					"metadata": map[string]interface{}{
						"name":      "test-name",
						"namespace": "test-ns",
						"labels": map[string]interface{}{
							"test": "me",
						},
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
					Triggers: []workspace.Trigger{
						{
							Version: "test-trigger-etv",
							Source:  "test-trigger-source",
							Type:    "test-trigger-type",
						},
					},
				},
			},
			wantOut: unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": functionApiVersion,
					"kind":       "Function",
					"metadata": map[string]interface{}{
						"name":      "test-name",
						"namespace": "test-ns",
						"labels": map[string]interface{}{
							"test": "me",
						},
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
					Triggers: []workspace.Trigger{
						{
							Version: "test-trigger-etv",
							Source:  "test-trigger-source",
							Type:    "test-trigger-type",
						},
					},
				},
			},
			wantOut: unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": functionApiVersion,
					"kind":       "Function",
					"metadata": map[string]interface{}{
						"name":      "test-name",
						"namespace": "test-ns",
						"labels": map[string]interface{}{
							"test": "me",
						},
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
					Triggers: []workspace.Trigger{
						{
							Version: "test-trigger-etv",
							Source:  "test-trigger-source",
							Type:    "test-trigger-type",
						},
					},
				},
			},
			wantOut: unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": functionApiVersion,
					"kind":       "Function",
					"metadata": map[string]interface{}{
						"name":      "test-name",
						"namespace": "test-ns",
						"labels": map[string]interface{}{
							"test": "me",
						},
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

package workspace

import (
	"bytes"
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"io"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/kyma-project/hydroform/function/pkg/client"
	mockclient "github.com/kyma-project/hydroform/function/pkg/client/automock"
	"github.com/kyma-project/hydroform/function/pkg/resources/types"
	"github.com/pkg/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type fileType string

const (
	fileTypeText fileType = "text"
	fileTypeJson fileType = "json"
	fileTypeCfg  fileType = "cfg"
)

func Test_Synchronise(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type args struct {
		ctx        context.Context
		cfg        Cfg
		outputPath string
		build      client.Build
	}
	type fileData struct {
		fileType fileType
		data     interface{}
	}

	name := "test"
	namespace := "test-ns"

	tests := []struct {
		name    string
		args    args
		want    workspace
		wantErr bool
		wantOut map[string]fileData
	}{
		{
			name:    "getting function should fail",
			wantErr: true,
			args: args{
				build: func(_ string, _ schema.GroupVersionResource) client.Client {
					result := mockclient.NewMockClient(ctrl)

					result.EXPECT().
						Get(nil, "", v1.GetOptions{}).
						Return(nil, errors.New("")).
						Times(1)

					return result
				},
			},
		},
		{
			name:    "getting subscriptions as unstructured list should fail",
			wantErr: true,
			args: args{
				cfg: Cfg{
					Name:      name,
					Namespace: namespace,
				},
				build: func() client.Build {

					result := mockclient.NewMockClient(ctrl)

					result.EXPECT().
						Get(gomock.Any(), name, v1.GetOptions{}).
						Return(&unstructured.Unstructured{Object: map[string]interface{}{"test": "test"}}, nil).
						Times(1)

					result.EXPECT().
						List(gomock.Any(), v1.ListOptions{}).
						Return(&unstructured.UnstructuredList{}, errors.New("the error")).
						Times(1)

					return func(_ string, _ schema.GroupVersionResource) client.Client {
						return result
					}
				}(),
				ctx: context.Background(),
			},
		},
		{
			name: "inline happy path with subscriptions and apirules",
			args: args{
				cfg: Cfg{
					Name:          name,
					Namespace:     namespace,
					SchemaVersion: SchemaVersionV0,
				},
				build: func() client.Build {
					c := inlineClientV0(ctrl, name, namespace)
					return func(_ string, _ schema.GroupVersionResource) client.Client {
						return c
					}
				}(),
			},
			wantErr: false,
			wantOut: map[string]fileData{
				"handler.js": {
					fileType: fileTypeText,
					data: `module.exports = {
							main: function (event, context) {
								return 'Hello Serverless'
							}
						}`,
				},
				"package.json": {
					fileType: fileTypeJson,
					data: `{
						  "name": "test",
						  "version": "0.0.1",
						  "dependencies": {}
						}`,
				},
				"config.yaml": {
					fileType: fileTypeCfg,
					data: Cfg{
						Name:      name,
						Namespace: namespace,
						Runtime:   types.Nodejs16,
						Source: Source{
							Type: SourceTypeInline,
						},
						Env: []EnvVar{
							{
								Name:  "TEST_ENV",
								Value: "test",
							},
							{
								Name: "TEST_ENV_SECRET",
								ValueFrom: &EnvVarSource{
									SecretKeyRef: &SecretKeySelector{
										Name: "secretName",
										Key:  "secretKey",
									},
								},
							},
							{
								Name: "TEST_ENV_CM",
								ValueFrom: &EnvVarSource{
									ConfigMapKeyRef: &ConfigMapKeySelector{
										Name: "configMapName",
										Key:  "configMapKey",
									},
								},
							},
						},
						Resources: Resources{
							Limits:   nil,
							Requests: nil,
						},
						Subscriptions: []Subscription{
							{
								Name: "subscription1",
								V0: &SubscriptionV0{
									Protocol: "",
									Filter: Filter{
										Dialect: "filter-dialect",
										Filters: []EventFilter{
											{
												EventSource: EventSource{
													Property: "source",
													Type:     "exact",
													Value:    "",
												},
												EventType: EventType{
													Property: "type",
													Type:     "exact",
													Value:    "t1.v1.0.0",
												},
											},
										},
									},
								},
							},
						},
						APIRules: []APIRule{
							{
								Name:    "test-name",
								Gateway: "test-gateway",
								Service: Service{
									Host: "test-host",
									Port: 9090,
								},
								Rules: []Rule{
									{
										Path:    "test-path",
										Methods: []string{"test-method"},
										AccessStrategies: []AccessStrategie{
											{
												Config: AccessStrategieConfig{
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
						SchemaVersion: SchemaVersionV0,
					},
				},
			},
		},
		{
			name: "inline happy path with subscriptions and apirules for v1alpha2",
			args: args{
				cfg: Cfg{
					Name:          name,
					Namespace:     namespace,
					SchemaVersion: SchemaVersionV1,
				},
				build: func() client.Build {
					c := inlineClientV1(ctrl, name, namespace)
					return func(_ string, _ schema.GroupVersionResource) client.Client {
						return c
					}
				}(),
			},
			wantErr: false,
			wantOut: map[string]fileData{
				"handler.js": {
					fileType: fileTypeText,
					data: `module.exports = {
							main: function (event, context) {
								return 'Hello Serverless'
							}
						}`,
				},
				"package.json": {
					fileType: fileTypeJson,
					data: `{
						  "name": "test",
						  "version": "0.0.1",
						  "dependencies": {}
						}`,
				},
				"config.yaml": {
					fileType: fileTypeCfg,
					data: Cfg{
						Name:      name,
						Namespace: namespace,
						Runtime:   types.Nodejs16,
						Source: Source{
							Type: SourceTypeInline,
						},
						Env: []EnvVar{
							{
								Name:  "TEST_ENV",
								Value: "test",
							},
							{
								Name: "TEST_ENV_SECRET",
								ValueFrom: &EnvVarSource{
									SecretKeyRef: &SecretKeySelector{
										Name: "secretName",
										Key:  "secretKey",
									},
								},
							},
							{
								Name: "TEST_ENV_CM",
								ValueFrom: &EnvVarSource{
									ConfigMapKeyRef: &ConfigMapKeySelector{
										Name: "configMapName",
										Key:  "configMapKey",
									},
								},
							},
						},
						Resources: Resources{
							Limits:   nil,
							Requests: nil,
						},
						Subscriptions: []Subscription{
							{
								Name: "subscription1",
								V1: &SubscriptionV1{
									TypeMatching: "standard",
									Source:       "commerce",
									Types: []string{
										"order.created.v1",
									},
								},
							},
						},
						APIRules: []APIRule{
							{
								Name:    "test-name",
								Gateway: "test-gateway",
								Service: Service{
									Host: "test-host",
									Port: 9090,
								},
								Rules: []Rule{
									{
										Path:    "test-path",
										Methods: []string{"test-method"},
										AccessStrategies: []AccessStrategie{
											{
												Config: AccessStrategieConfig{
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
						SchemaVersion: SchemaVersionV1,
					},
				},
			},
		},
		{
			name:    "getting apirules as unstructured list should fail",
			wantErr: true,
			args: args{
				cfg: Cfg{
					Name:          name,
					Namespace:     namespace,
					SchemaVersion: SchemaVersionV0,
				},
				build: func() client.Build {

					result := mockclient.NewMockClient(ctrl)

					result.EXPECT().
						Get(gomock.Any(), name, v1.GetOptions{}).
						Return(&unstructured.Unstructured{Object: map[string]interface{}{"test": "test"}}, nil).
						Times(1)

					result.EXPECT().
						List(gomock.Any(), v1.ListOptions{}).
						Return(&unstructured.UnstructuredList{}, nil).
						Times(1)

					result.EXPECT().List(gomock.Any(), v1.ListOptions{}).
						Return(&unstructured.UnstructuredList{}, errors.New("the error")).Times(1)

					return func(_ string, _ schema.GroupVersionResource) client.Client {
						return result
					}
				}(),
				ctx: context.Background(),
			},
		},
		{
			name:    "getting apirules as unstructured list should fail for v1alpha2",
			wantErr: true,
			args: args{
				cfg: Cfg{
					Name:          name,
					Namespace:     namespace,
					SchemaVersion: SchemaVersionV1,
				},
				build: func() client.Build {

					result := mockclient.NewMockClient(ctrl)

					result.EXPECT().
						Get(gomock.Any(), name, v1.GetOptions{}).
						Return(&unstructured.Unstructured{Object: map[string]interface{}{"test": "test"}}, nil).
						Times(1)

					result.EXPECT().
						List(gomock.Any(), v1.ListOptions{}).
						Return(&unstructured.UnstructuredList{}, nil).
						Times(1)

					result.EXPECT().List(gomock.Any(), v1.ListOptions{}).
						Return(&unstructured.UnstructuredList{}, errors.New("the error")).Times(1)

					return func(_ string, _ schema.GroupVersionResource) client.Client {
						return result
					}
				}(),
				ctx: context.Background(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//TODO: Please refactor this test because:
			// - the mocks or tests are wrongly configured as a result: the unit test together pass, but e.g.: "gitrepo happy path" apart fails.
			bp := BufferProvider{
				buffers: map[string]*bytes.Buffer{},
			}
			wp := func(path string) (io.Writer, Cancel, error) {
				b := bp.NewBuffer(path)
				return b, func() error { return nil }, nil
			}
			err := synchronise(tt.args.ctx, tt.args.cfg, tt.args.outputPath, tt.args.build, wp)
			if (err != nil) != tt.wantErr {
				t.Errorf("Synchronise() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				var gotFileNames []string
				for k := range bp.buffers {
					gotFileNames = append(gotFileNames, k)
				}
				var wantFileNames []string
				for k := range tt.wantOut {
					wantFileNames = append(wantFileNames, k)
				}
				require.ElementsMatch(t, gotFileNames, wantFileNames)

				for _, fileName := range wantFileNames {
					data := tt.wantOut[fileName]
					switch data.fileType {
					case fileTypeCfg:
						{
							var gotCfg Cfg
							if err := yaml.NewDecoder(bytes.NewReader(bp.buffers[fileName].Bytes())).Decode(&gotCfg); err != nil {
								t.Errorf("Synchronise() error while trying to decode output")
							}
							require.Equal(t, tt.wantOut[fileName].data.(Cfg), gotCfg)
						}
					case fileTypeJson:
						require.JSONEq(t, tt.wantOut[fileName].data.(string), bp.buffers[fileName].String())
					case fileTypeText:
						require.Equal(t,
							removeSpaces(tt.wantOut[fileName].data.(string)),
							removeSpaces(bp.buffers[fileName].String()))
					}
				}
			}
		})
	}
}

func removeSpaces(s string) string {
	return strings.Join(strings.Fields(s), "")
}

type BufferProvider struct {
	buffers map[string]*bytes.Buffer
}

func (p *BufferProvider) NewBuffer(path string) io.Writer {
	b := bytes.Buffer{}
	p.buffers[path] = &b
	return &b
}

func inlineClientV0(ctrl *gomock.Controller, name, namespace string) client.Client {
	result := mockclient.NewMockClient(ctrl)

	inlineClientGetFunction(result, name, namespace)
	inlineClientGetSubscriptionV1Alpha1(result, name, namespace)
	inlineClientListApiRules(result, name)

	return result
}

func inlineClientV1(ctrl *gomock.Controller, name, namespace string) client.Client {
	result := mockclient.NewMockClient(ctrl)

	inlineClientGetFunction(result, name, namespace)
	inlineClientGetSubscriptionV1Alpha2(result, name, namespace)
	inlineClientListApiRules(result, name)

	return result
}

func inlineClientListApiRules(result *mockclient.MockClient, name string) {
	result.EXPECT().List(gomock.Any(), v1.ListOptions{}).
		Return(&unstructured.UnstructuredList{
			Items: []unstructured.Unstructured{
				{
					Object: map[string]interface{}{
						"metadata": map[string]interface{}{
							"name": "test-name",
						},
						"spec": map[string]interface{}{
							"host": "test-host",
							"service": map[string]interface{}{
								"name": name,
								"port": int64(9090),
							},
							"rules": []interface{}{
								map[string]interface{}{
									"methods": []interface{}{"test-method"},
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
		}, nil).Times(1)
}

func inlineClientGetSubscriptionV1Alpha1(result *mockclient.MockClient, name string, namespace string) {
	result.EXPECT().
		List(gomock.Any(), v1.ListOptions{}).Return(&unstructured.UnstructuredList{
		Items: []unstructured.Unstructured{
			{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"name":      "subscription1",
						"namespace": "test-ns",
					},
					"spec": map[string]interface{}{
						"protocol": "",
						"sink":     fmt.Sprintf("http://%s.%s.svc.cluster.local", name, namespace),
						"filter": map[string]interface{}{
							"dialect": "filter-dialect",
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
										"value":    "t1.v1.0.0",
									},
								},
							},
						},
					},
				},
			},
		},
	}, nil).Times(1)
}

func inlineClientGetSubscriptionV1Alpha2(result *mockclient.MockClient, name string, namespace string) {
	result.EXPECT().
		List(gomock.Any(), v1.ListOptions{}).Return(&unstructured.UnstructuredList{
		Items: []unstructured.Unstructured{
			{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"name":      "subscription1",
						"namespace": "test-ns",
					},
					"spec": map[string]interface{}{
						"typeMatching": "standard",
						"source":       "commerce",
						"types": []string{
							"order.created.v1",
						},
						"sink": fmt.Sprintf("http://%s.%s.svc.cluster.local", name, namespace),
					},
				},
			},
		},
	}, nil).Times(1)
}

func inlineClientGetFunction(result *mockclient.MockClient, name string, namespace string) {
	result.EXPECT().
		Get(gomock.Any(), name, v1.GetOptions{}).
		Return(&unstructured.Unstructured{Object: map[string]interface{}{
			"apiVersion": "serverless.kyma-project.io/v1alpha2",
			"kind":       "Function",
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": namespace,
			},
			"spec": map[string]interface{}{
				"maxReplicas": 1,
				"minReplicas": 1,
				"resources": map[string]interface{}{
					"limits": map[string]interface{}{
						"cpu":    "100m",
						"memory": "128Mi",
					},
				},
				"runtime": "nodejs16",
				"source": map[string]interface{}{
					"inline": map[string]interface{}{
						"source":       handlerJs,
						"dependencies": packageJSON,
					},
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
		}}, nil).Times(1)
}

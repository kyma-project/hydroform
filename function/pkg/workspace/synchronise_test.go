package workspace

import (
	"context"
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

func Test_Synchronise(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type args struct {
		ctx        context.Context
		cfg        Cfg
		outputPath string
		build      client.Build
	}

	name := "test"
	namespace := "test-ns"

	tests := []struct {
		name    string
		args    args
		want    workspace
		wantErr bool
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
					Name:      name,
					Namespace: namespace,
					Runtime:   types.Nodejs16,
					Source: Source{
						Type: SourceTypeInline,
						SourceInline: SourceInline{
							SourcePath:        "./testdir/inline",
							SourceHandlerName: handlerJs,
							DepsHandlerName:   packageJSON,
						},
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
							Name:     "fixme",
							Protocol: "fixme",
							Filter: Filter{
								Dialect: "fixme",
								Filters: []EventFilter{
									{
										EventSource: EventSource{
											Property: "source",
											Type:     "exact",
											Value:    "the-source",
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
				},
				build: func() client.Build {
					c := inlineClient(ctrl, name, namespace)
					return func(_ string, _ schema.GroupVersionResource) client.Client {
						return c
					}
				}(),
			},
			wantErr: false,
		},
		{
			name: "gitrepo happy path",
			args: args{
				cfg: Cfg{
					Name:      name,
					Namespace: namespace,
					Runtime:   types.Nodejs16,
					Source: Source{
						SourceGit: SourceGit{
							URL:       "https://test.com",
							Reference: "master",
							BaseDir:   "/",
						},
					},
					Resources: Resources{
						Limits:   nil,
						Requests: nil,
					},
					Subscriptions: []Subscription{
						{
							Name:     "fixme",
							Protocol: "fixme",
							Filter: Filter{
								Dialect: "fixme",
								Filters: []EventFilter{
									{
										EventSource: EventSource{
											Property: "source",
											Type:     "exact",
											Value:    "the-source",
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
				},
				build: func() client.Build {
					c := gitClient(ctrl, namespace)
					return func(_ string, _ schema.GroupVersionResource) client.Client {
						return c
					}
				}(),
			},
			wantErr: false,
		},
		{
			name:    "getting apirules as unstructured list should fail",
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
			name: "inline happy path with subscriptions and apirules",
			args: args{
				cfg: Cfg{
					Name:      name,
					Namespace: namespace,
					Runtime:   types.Nodejs16,
					Source: Source{
						Type: SourceTypeInline,
						SourceInline: SourceInline{
							SourcePath:        "./testdir/inline",
							SourceHandlerName: handlerJs,
							DepsHandlerName:   packageJSON,
						},
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
					//Resources: Resources{
					//	Limits:   nil,
					//	Requests: nil,
					//},
					Subscriptions: []Subscription{
						{
							Name:     "fixme",
							Protocol: "fixme",
							Filter: Filter{
								Dialect: "fixme",
								Filters: []EventFilter{
									{
										EventSource: EventSource{
											Property: "source",
											Type:     "exact",
											Value:    "the-source",
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
				},
				build: func() client.Build {
					c := inlineClient(ctrl, name, namespace)
					return func(_ string, _ schema.GroupVersionResource) client.Client {
						return c
					}
				}(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//TODO: Please refactor this test because:
			// - it doesn't check the output config
			// - the mocks or tests are wrongly configured as a result: the unit test together pass, but e.g.: "gitrepo happy path" apart fails.
			err := synchronise(tt.args.ctx, tt.args.cfg, tt.args.outputPath, tt.args.build, newStrWriterProvider())
			if (err != nil) != tt.wantErr {
				t.Errorf("Synchronise() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

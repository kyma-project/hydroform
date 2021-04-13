package operator

import (
	"context"
	errs "errors"
	"fmt"
	"testing"

	"k8s.io/apimachinery/pkg/watch"

	"github.com/golang/mock/gomock"
	"github.com/kyma-incubator/hydroform/function/pkg/client"
	mockclient "github.com/kyma-incubator/hydroform/function/pkg/client/automock"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func Test_genericOperator_Apply(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type fields struct {
		Client client.Client
		items  []unstructured.Unstructured
	}
	type args struct {
		ctx  context.Context
		opts ApplyOptions
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "pre callback error",
			fields: fields{
				items: []unstructured.Unstructured{testObj},
			},
			args: args{
				opts: ApplyOptions{
					Options: Options{
						Callbacks: Callbacks{
							Pre: []Callback{
								func(_ interface{}, _ error) error {
									return fmt.Errorf("callback error")
								},
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "post callback error",
			fields: fields{
				Client: func() client.Client {
					result := mockclient.NewMockClient(ctrl)

					result.EXPECT().
						Get(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(nil, errors.NewNotFound(schema.GroupResource{}, "test error")).
						Times(1)

					result.EXPECT().
						Create(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(testObj.DeepCopy(), nil).
						Times(1)

					return result
				}(),
				items: []unstructured.Unstructured{testObj},
			},
			args: args{
				opts: ApplyOptions{
					Options: Options{
						Callbacks: Callbacks{
							Post: []Callback{
								func(_ interface{}, _ error) error {
									return fmt.Errorf("callback error")
								},
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "apply error",
			fields: fields{
				Client: func() client.Client {
					result := mockclient.NewMockClient(ctrl)

					result.EXPECT().
						Get(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(nil, errors.NewNotFound(schema.GroupResource{}, "test error")).
						Times(1)

					result.EXPECT().
						Create(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(nil, errs.New("test error")).
						Times(1)

					return result
				}(),
				items: []unstructured.Unstructured{testObj},
			},
			args: args{
				opts: ApplyOptions{},
			},
			wantErr: true,
		},
		{
			name: "apply",
			fields: fields{
				Client: func() client.Client {
					result := mockclient.NewMockClient(ctrl)

					result.EXPECT().
						Get(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(nil, errors.NewNotFound(schema.GroupResource{}, "test error")).
						Times(1)

					result.EXPECT().
						Create(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(testObj.DeepCopy(), nil).
						Times(1)

					fakeWatcher := watch.NewRaceFreeFake()
					testObject := fixUnstructured()
					fakeWatcher.Add(&testObject)

					result.EXPECT().
						Watch(gomock.Any(), gomock.Any()).
						Return(fakeWatcher, nil).
						Times(1)

					return result
				}(),
				items: []unstructured.Unstructured{testObj},
			},
			args: args{
				ctx: context.Background(),
				opts: ApplyOptions{
					Options: Options{
						WaitForApply: true,
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewGenericOperator(tt.fields.Client, tt.fields.items...)
			if err := p.Apply(tt.args.ctx, tt.args.opts); (err != nil) != tt.wantErr {
				t.Errorf("Apply() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_genericOperator_Delete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type fields struct {
		Client client.Client
		items  []unstructured.Unstructured
	}
	type args struct {
		ctx  context.Context
		opts DeleteOptions
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "pre callback error",
			fields: fields{
				items: []unstructured.Unstructured{testObj},
			},
			args: args{
				opts: DeleteOptions{
					DeletionPropagation: v1.DeletePropagationForeground,
					Options: Options{
						Callbacks: Callbacks{
							Pre: []Callback{
								func(_ interface{}, _ error) error {
									return fmt.Errorf("pre callback error")
								},
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "post callback error",
			fields: fields{
				Client: func() client.Client {
					result := mockclient.NewMockClient(ctrl)

					result.EXPECT().
						Delete(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(nil).
						Times(1)

					return result
				}(),
				items: []unstructured.Unstructured{testObj},
			},
			args: args{
				opts: DeleteOptions{
					DeletionPropagation: v1.DeletePropagationForeground,
					Options: Options{
						Callbacks: Callbacks{
							Post: []Callback{
								func(_ interface{}, _ error) error {
									return fmt.Errorf("post callback error")
								},
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "delete error",
			fields: fields{
				Client: func() client.Client {
					result := mockclient.NewMockClient(ctrl)

					result.EXPECT().
						Delete(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(fmt.Errorf("delete error")).
						Times(1)

					return result
				}(),
				items: []unstructured.Unstructured{testObj},
			},
			args: args{
				opts: DeleteOptions{
					DeletionPropagation: v1.DeletePropagationForeground,
				},
			},
			wantErr: true,
		},
		{
			name: "delete",
			fields: fields{
				Client: func() client.Client {
					result := mockclient.NewMockClient(ctrl)

					result.EXPECT().
						Delete(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(nil).
						Times(1)

					return result
				}(),
				items: []unstructured.Unstructured{testObj},
			},
			args: args{
				opts: DeleteOptions{
					DeletionPropagation: v1.DeletePropagationForeground,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewGenericOperator(tt.fields.Client, tt.fields.items...)
			if err := p.Delete(tt.args.ctx, tt.args.opts); (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

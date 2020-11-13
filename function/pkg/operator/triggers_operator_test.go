package operator

import (
	"context"
	"fmt"
	"github.com/kyma-incubator/hydroform/function/pkg/resources/types"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/kyma-incubator/hydroform/function/pkg/client"
	mockclient "github.com/kyma-incubator/hydroform/function/pkg/client/automock"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func Test_contains(t *testing.T) {
	type args struct {
		s    []unstructured.Unstructured
		name string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "nil check",
			args: args{
				s:    nil,
				name: "test-name",
			},
			want: false,
		},
		{
			name: "found",
			args: args{
				s:    []unstructured.Unstructured{testObj},
				name: "test-obj",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := contains(tt.args.s, tt.args.name); got != tt.want {
				t.Errorf("contains() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_findOwnerID(t *testing.T) {
	type args struct {
		refs []v1.OwnerReference
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 bool
	}{
		{
			name: "nil not found",
			args: args{
				refs: nil,
			},
			want:  "",
			want1: false,
		},
		{
			name: "not found",
			args: args{
				refs: []v1.OwnerReference{
					{
						Kind: "Function123",
						Name: "not-found",
					},
				},
			},
			want:  "",
			want1: false,
		},
		{
			name: "found",
			args: args{
				refs: []v1.OwnerReference{
					{
						Kind: "Function",
						Name: "test-obj",
					},
				},
			},
			want:  "",
			want1: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := findOwnerID(tt.args.refs)
			if got != tt.want {
				t.Errorf("findOwnerID() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("findOwnerID() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_mergeMap(t *testing.T) {
	type args struct {
		l map[string]string
		r map[string]string
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "nil check",
			args: args{
				l: nil,
				r: nil,
			},
			want: nil,
		},
		{
			name: "nil check #2",
			args: args{
				l: nil,
				r: map[string]string{
					"test": "me",
				},
			},
			want: map[string]string{
				"test": "me",
			},
		},
		{
			name: "override",
			args: args{
				l: map[string]string{"a": "a1", "b": "b1"},
				r: map[string]string{"a": "a2", "c": "c2"},
			},
			want: map[string]string{"a": "a2", "b": "b1", "c": "c2"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mergeMap(tt.args.l, tt.args.r); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mergeMap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_triggersOperator_Apply(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type fields struct {
		items  []unstructured.Unstructured
		Client client.Client
	}
	type args struct {
		opts ApplyOptions
		ctx  context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "ownerID not found",
			args: args{
				opts: ApplyOptions{
					OwnerReferences: []v1.OwnerReference{
						{
							Kind: "Function123",
							UID:  "you-shall-not-pass",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "wipe triggers error",
			args: args{
				opts: ApplyOptions{
					OwnerReferences: []v1.OwnerReference{
						{
							Kind: "Function",
							UID:  "123",
						},
					},
				},
			},
			fields: fields{
				items: []unstructured.Unstructured{testObj},
				Client: func() client.Client {
					result := mockclient.NewMockClient(ctrl)

					result.EXPECT().
						List(gomock.Any(), gomock.Any()).
						Return(nil, fmt.Errorf("list error")).
						Times(1)

					return result
				}(),
			},
			wantErr: true,
		},
		{
			name: "apply error",
			args: args{
				opts: ApplyOptions{
					OwnerReferences: []v1.OwnerReference{
						{
							Kind: "Function",
							UID:  "123",
						},
					},
				},
			},
			fields: fields{
				items: []unstructured.Unstructured{testObj},
				Client: func() client.Client {
					result := mockclient.NewMockClient(ctrl)

					result.EXPECT().
						List(gomock.Any(), gomock.Any()).
						Return(&unstructured.UnstructuredList{}, nil).
						Times(1)

					result.EXPECT().
						Get(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(nil, fmt.Errorf("get error")).
						Times(1)

					return result
				}(),
			},
			wantErr: true,
		},
		{
			name: "post callback error",
			args: args{
				opts: ApplyOptions{
					OwnerReferences: []v1.OwnerReference{
						{
							Kind: "Function",
							UID:  "123",
						},
					},
					Options: Options{
						Callbacks: Callbacks{
							Post: []Callback{
								func(_ interface{}, _ error) error {
									return fmt.Errorf("test error")
								},
							},
						},
					},
				},
			},
			fields: fields{
				items: []unstructured.Unstructured{testObj},
				Client: func() client.Client {
					result := mockclient.NewMockClient(ctrl)

					result.EXPECT().
						List(gomock.Any(), gomock.Any()).
						Return(&unstructured.UnstructuredList{}, nil).
						Times(1)

					result.EXPECT().
						Get(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(testObj.DeepCopy(), nil).
						Times(1)

					return result
				}(),
			},
			wantErr: true,
		},
		{
			name: "pre callback error",
			args: args{
				opts: ApplyOptions{
					OwnerReferences: []v1.OwnerReference{
						{
							Kind: "Function",
							UID:  "123",
						},
					},
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
			fields: fields{
				items: []unstructured.Unstructured{testObj},
				Client: func() client.Client {
					result := mockclient.NewMockClient(ctrl)

					result.EXPECT().
						List(gomock.Any(), gomock.Any()).
						Return(&unstructured.UnstructuredList{}, nil).
						Times(1)

					return result
				}(),
			},
			wantErr: true,
		},
		{
			name: "apply",
			args: args{
				opts: ApplyOptions{
					OwnerReferences: []v1.OwnerReference{
						{
							Kind: "Function",
							UID:  "123",
						},
					},
				},
			},
			fields: fields{
				items: []unstructured.Unstructured{testObj},
				Client: func() client.Client {
					result := mockclient.NewMockClient(ctrl)

					result.EXPECT().
						List(gomock.Any(), gomock.Any()).
						Return(&unstructured.UnstructuredList{}, nil).
						Times(1)

					result.EXPECT().
						Get(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(testObj.DeepCopy(), nil).
						Times(1)

					return result
				}(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t1 *testing.T) {
			t := NewTriggersOperator(tt.fields.Client, "test", "test-namespace", tt.fields.items...)
			if err := t.Apply(tt.args.ctx, tt.args.opts); (err != nil) != tt.wantErr {
				t1.Errorf("Apply() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_triggersOperator_Delete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	type fields struct {
		items  []unstructured.Unstructured
		Client client.Client
	}
	type args struct {
		opts DeleteOptions
		ctx  context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "error delete",
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
					DeletionPropagation: v1.DeletePropagationOrphan,
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
					DeletionPropagation: v1.DeletePropagationOrphan,
					Options: Options{
						Callbacks: Callbacks{
							Post: []Callback{
								func(_ interface{}, _ error) error {
									return fmt.Errorf("test error")
								},
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "pre callback error",
			fields: fields{
				items: []unstructured.Unstructured{testObj},
			},
			args: args{
				opts: DeleteOptions{
					DeletionPropagation: v1.DeletePropagationOrphan,
					Options: Options{
						Callbacks: Callbacks{
							Pre: []Callback{
								func(_ interface{}, _ error) error {
									return fmt.Errorf("test error")
								},
							},
						},
					},
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
					DeletionPropagation: v1.DeletePropagationOrphan,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t1 *testing.T) {
			t := NewTriggersOperator(tt.fields.Client, "test", "test-namespace", tt.fields.items...)
			if err := t.Delete(tt.args.ctx, tt.args.opts); (err != nil) != tt.wantErr {
				t1.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_triggersOperator_wipeRemoved(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	type args struct {
		ownerID string
		opts    ApplyOptions
		ctx     context.Context
		items   []unstructured.Unstructured
		Client  client.Client
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "list error",
			args: args{
				Client: func() client.Client {
					result := mockclient.NewMockClient(ctrl)

					result.EXPECT().
						List(gomock.Any(), gomock.Any()).
						Return(nil, fmt.Errorf("list error")).
						Times(1)

					return result
				}(),
				items: []unstructured.Unstructured{testObj},
			},
			wantErr: true,
		},
		{
			name: "delete err",
			args: args{
				ownerID: "test-id",
				opts:    ApplyOptions{},
				Client: func() client.Client {
					result := mockclient.NewMockClient(ctrl)

					result.EXPECT().
						List(gomock.Any(), gomock.Any()).
						Return(&unstructured.UnstructuredList{
							Items: []unstructured.Unstructured{
								testObj2,
							},
						}, nil).
						Times(1)

					result.EXPECT().
						Delete(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(fmt.Errorf("delete error")).
						Times(1)

					return result
				}(),
				items: []unstructured.Unstructured{testObj},
			},
			wantErr: true,
		},
		{
			name: "post callbacks error",
			args: args{
				Client: func() client.Client {
					result := mockclient.NewMockClient(ctrl)

					result.EXPECT().
						List(gomock.Any(), gomock.Any()).
						Return(&unstructured.UnstructuredList{
							Items: []unstructured.Unstructured{
								testObj2,
							},
						}, nil).
						Times(1)

					result.EXPECT().
						Delete(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(nil).
						Times(1)

					return result
				}(),
				items: []unstructured.Unstructured{testObj},
				opts: ApplyOptions{
					Options: Options{
						Callbacks: Callbacks{
							Post: []Callback{
								func(_ interface{}, _ error) error {
									panic("it's fine")
								},
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "pre callbacks error",
			args: args{
				Client: func() client.Client {
					result := mockclient.NewMockClient(ctrl)

					result.EXPECT().
						List(gomock.Any(), gomock.Any()).
						Return(&unstructured.UnstructuredList{
							Items: []unstructured.Unstructured{
								testObj2,
							},
						}, nil).
						Times(1)

					return result
				}(),
				items: []unstructured.Unstructured{testObj},
				opts: ApplyOptions{
					Options: Options{
						Callbacks: Callbacks{
							Pre: []Callback{
								func(_ interface{}, _ error) error {
									panic("it's fine")
								},
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "no wipe",
			args: args{
				Client: func() client.Client {
					result := mockclient.NewMockClient(ctrl)

					result.EXPECT().
						List(gomock.Any(), gomock.Any()).
						Return(&unstructured.UnstructuredList{
							Items: []unstructured.Unstructured{
								testObj,
							},
						}, nil).
						Times(1)

					return result
				}(),
				items:   []unstructured.Unstructured{testObj},
				ownerID: "test-id",
				opts:    ApplyOptions{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t1 *testing.T) {
			if err := wipeRemoved(tt.args.ctx, tt.args.Client, func(obj map[string]interface{}) (bool, error) {
				var trigger types.Trigger
				if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj, &trigger); err != nil {
					return false, err
				}

				isRef := trigger.IsReference("test-function-name", "test-namespace")
				isListed := contains(tt.args.items, trigger.Metadata.Name)

				return isRef && !isListed, nil
			}, tt.args.opts.Options); (err != nil) != tt.wantErr {
				t1.Errorf("wipeRemoved() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

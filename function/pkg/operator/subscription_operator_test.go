package operator

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/kyma-project/hydroform/function/pkg/client"
	mockclient "github.com/kyma-project/hydroform/function/pkg/client/automock"
	operator_types "github.com/kyma-project/hydroform/function/pkg/operator/types"
	"github.com/kyma-project/hydroform/function/pkg/resources/types"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
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

func Test_applySubscriptions(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testPredicate := func(map[string]interface{}) (bool, error) {
		return false, nil
	}

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
			name: "wipe subscriptions error",
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
				ctx: context.Background(),
				opts: ApplyOptions{
					Options: Options{
						WaitForApply: true,
					},
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

					fakeWatcher := watch.NewRaceFreeFake()
					testObject := fixUnstructured()
					fakeWatcher.Add(&testObject)

					result.EXPECT().
						Watch(gomock.Any(), gomock.Any()).
						Return(fakeWatcher, nil).
						Times(1)

					return result
				}(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t1 *testing.T) {
			if err := applySubscriptions(tt.args.ctx, tt.fields.Client, testPredicate, tt.fields.items, tt.args.opts); (err != nil) != tt.wantErr {
				t1.Errorf("Apply() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_deleteSubscriptions(t *testing.T) {
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
			if err := deleteSubscriptions(tt.args.ctx, tt.fields.Client, tt.fields.items, tt.args.opts); (err != nil) != tt.wantErr {
				t1.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_subscriptionsOperator_wipeRemoved(t *testing.T) {
	var subscription1 unstructured.Unstructured
	var subscription2 unstructured.Unstructured

	for i, s := range []*unstructured.Unstructured{
		&subscription1, &subscription2,
	} {
		var err error
		(*s), err = newTestSubscription(fmt.Sprintf("test-%d", i+1), "test-namespace")
		if err != nil {
			t.Fatal(err)
		}
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	type args struct {
		opts   ApplyOptions
		ctx    context.Context
		items  []unstructured.Unstructured
		Client client.Client
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
				items: []unstructured.Unstructured{subscription1},
			},
			wantErr: true,
		},
		{
			name: "delete err",
			args: args{
				opts: ApplyOptions{},
				Client: func() client.Client {
					result := mockclient.NewMockClient(ctrl)

					result.EXPECT().
						List(gomock.Any(), gomock.Any()).
						Return(&unstructured.UnstructuredList{
							Items: []unstructured.Unstructured{
								subscription2,
							},
						}, nil).
						Times(1)

					result.EXPECT().
						Delete(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(fmt.Errorf("delete error")).
						Times(1)

					return result
				}(),
				items: []unstructured.Unstructured{subscription1},
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
								subscription2,
							},
						}, nil).
						Times(1)

					result.EXPECT().
						Delete(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(nil).
						Times(1)

					return result
				}(),
				items: []unstructured.Unstructured{subscription1},
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
								subscription2,
							},
						}, nil).
						Times(1)

					return result
				}(),
				items: []unstructured.Unstructured{subscription1},
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
								subscription1,
							},
						}, nil).
						Times(1)

					return result
				}(),
				items: []unstructured.Unstructured{subscription1},
				opts:  ApplyOptions{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t1 *testing.T) {
			predicate := buildMatchRemovedSubscriptionsPredicate(functionReference{
				name:      "test-2",
				namespace: "test-namespace",
			}, tt.args.items)
			if err := wipeRemoved(tt.args.ctx, tt.args.Client, predicate, tt.args.opts.Options); (err != nil) != tt.wantErr {
				t1.Errorf("wipeRemoved() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func newTestSubscription(name, namespace string) (unstructured.Unstructured, error) {
	subscription := types.SubscriptionV1alpha1{
		TypeMeta: v1.TypeMeta{
			Kind:       "Subscription",
			APIVersion: operator_types.GVRSubscriptionV1alpha1.Version,
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: types.SubscriptionSpecV1alpha1{
			Sink: fmt.Sprintf("http://%s.%s.svc.cluster.local", name, namespace),
		},
	}
	subscriptionObject, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&subscription)
	if err != nil {
		return unstructured.Unstructured{}, err
	}
	return unstructured.Unstructured{Object: subscriptionObject}, nil
}

func Test_buildMatchRemovedSubscriptionsPredicate(t *testing.T) {
	var subscription1 unstructured.Unstructured
	var subscription2 unstructured.Unstructured
	var subscription3 unstructured.Unstructured

	for i, s := range []*unstructured.Unstructured{
		&subscription1, &subscription2, &subscription3,
	} {
		var err error
		(*s), err = newTestSubscription(fmt.Sprintf("test-%d", i+1), "test-namespace")
		if err != nil {
			t.Fatal(err)
		}
	}

	subscription3.SetOwnerReferences([]v1.OwnerReference{
		fixOwnerRef("test-me"),
	})

	type args struct {
		fnRef        functionReference
		items        []unstructured.Unstructured
		subscription unstructured.Unstructured
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "no match 1",
			args: args{
				items: []unstructured.Unstructured{subscription1, subscription2},
				fnRef: functionReference{
					name:      "test-1",
					namespace: "test-namespace",
				},
				subscription: subscription1,
			},
			want: false,
		},
		{
			name: "no match 2",
			args: args{
				items: []unstructured.Unstructured{subscription1, subscription2},
				fnRef: functionReference{
					name:      "test-me",
					namespace: "test-namespace",
				},
				subscription: subscription1,
			},
			want: false,
		},
		{
			name: "no match 3",
			args: args{
				items: []unstructured.Unstructured{subscription2},
				fnRef: functionReference{
					name:      "test-3",
					namespace: "test-namespace",
				},
				subscription: subscription3,
			},
			want: false,
		},
		{
			name: "match",
			args: args{
				items: []unstructured.Unstructured{subscription2},
				fnRef: functionReference{
					name:      "test-1",
					namespace: "test-namespace",
				},
				subscription: subscription1,
			},
			want: true,
		},
		{
			name: "match",
			args: args{
				items: []unstructured.Unstructured{subscription2},
				fnRef: functionReference{
					name:      "test-1",
					namespace: "test-namespace",
				},
				subscription: subscription1,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			predicate := buildMatchRemovedSubscriptionsPredicate(tt.args.fnRef, tt.args.items)
			got, err := predicate(tt.args.subscription.Object)
			if (err != nil) != tt.wantErr {
				t.Errorf("buildMatchRemovedSubscriptionsPredicate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("predicate() bool = %v, want %v", got, tt.want)
			}
		})
	}
}

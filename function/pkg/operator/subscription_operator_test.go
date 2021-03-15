package operator

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/kyma-incubator/hydroform/function/pkg/client"
	mockclient "github.com/kyma-incubator/hydroform/function/pkg/client/automock"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/watch"
)

func Test_subscriptionOperator_Apply(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type fields struct {
		fnRef  functionReference
		items  []unstructured.Unstructured
		Client client.Client
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
					testObject := fixUnstructured("test", "test")
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
		t.Run(tt.name, func(t *testing.T) {
			tr := subscriptionOperator{
				fnRef:  tt.fields.fnRef,
				items:  tt.fields.items,
				Client: tt.fields.Client,
			}
			if err := tr.Apply(tt.args.ctx, tt.args.opts); (err != nil) != tt.wantErr {
				t.Errorf("subscriptionOperator.Apply() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_subscriptionOperator_Delete(t *testing.T) {
	type fields struct {
		fnRef  functionReference
		items  []unstructured.Unstructured
		Client client.Client
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
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := subscriptionOperator{
				fnRef:  tt.fields.fnRef,
				items:  tt.fields.items,
				Client: tt.fields.Client,
			}
			if err := tr.Delete(tt.args.ctx, tt.args.opts); (err != nil) != tt.wantErr {
				t.Errorf("subscriptionOperator.Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

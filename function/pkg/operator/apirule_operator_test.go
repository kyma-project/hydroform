package operator

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/kyma-project/hydroform/function/pkg/client"
	mock_client "github.com/kyma-project/hydroform/function/pkg/client/automock"
	"github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func Test_buildMatchRemovedApiRulePredicate(t *testing.T) {
	g := gomega.NewWithT(t)
	apiRules := []unstructured.Unstructured{
		fixRawAPIRule("test-name-1", "fn-name"),
		fixRawAPIRuleWithOwnerRef("test-name-2", "fn-name", fixOwnerRef("fn-name")),
		fixRawAPIRuleWithOwnerRef("test-name-3", "fn-name", fixOwnerRef("fn-name")),
	}

	tests := []struct {
		name     string
		fnName   string
		givenObj unstructured.Unstructured
		wantErr  gomega.OmegaMatcher
		wantBool bool
	}{
		{
			name:     "should predicate to remove given apiRule",
			fnName:   "fn-name",
			givenObj: fixRawAPIRule("test-name-4", "fn-name"),
			wantErr:  gomega.BeNil(),
			wantBool: true,
		},
		{
			name:     "should return false because apiRule is one of items on the list",
			fnName:   "fn-name",
			givenObj: fixRawAPIRule("test-name-3", "fn-name"),
			wantErr:  gomega.BeNil(),
			wantBool: false,
		},
		{
			name:     "should return false because apiRule.service.name != fnName",
			fnName:   "fn-name",
			givenObj: fixRawAPIRule("test-name-4", "fn-name-2"),
			wantErr:  gomega.BeNil(),
			wantBool: false,
		},
		{
			name:     "should return false because owner reference",
			fnName:   "fn-name",
			givenObj: fixRawAPIRuleWithOwnerRef("test-name-4", "fn-name", fixOwnerRef("fn-name")),
			wantErr:  gomega.BeNil(),
			wantBool: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			predicateFn := buildMatchRemovedAPIRulePredicate(tt.fnName, apiRules)
			out, err := predicateFn(tt.givenObj.Object)
			g.Expect(err).To(tt.wantErr)
			g.Expect(out).To(gomega.Equal(tt.wantBool))
		})
	}
}

func Test_apiRuleOperator_Apply(t *testing.T) {
	g := gomega.NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name    string
		client  client.Client
		wantErr gomega.OmegaMatcher
	}{
		{
			name: "should return nil",
			client: func() client.Client {
				c := mock_client.NewMockClient(ctrl)

				c.EXPECT().List(gomock.Any(), gomock.Any()).
					Return(&unstructured.UnstructuredList{}, nil).
					Times(1)

				return c
			}(),
			wantErr: gomega.BeNil(),
		},
		{
			name: "should return error handled from wipeRemoved method",
			client: func() client.Client {
				c := mock_client.NewMockClient(ctrl)

				c.EXPECT().List(gomock.Any(), gomock.Any()).
					Return(&unstructured.UnstructuredList{}, errors.New("test error")).
					Times(1)

				return c
			}(),
			wantErr: gomega.HaveOccurred(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := NewAPIRuleOperator(tt.client, "fn-name", []unstructured.Unstructured{}...)

			err := o.Apply(context.Background(), ApplyOptions{})

			g.Expect(err).To(tt.wantErr)
		})
	}
}

func Test_apiRuleOperator_Delete(t *testing.T) {
	g := gomega.NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name    string
		client  client.Client
		wantErr gomega.OmegaMatcher
	}{
		{
			name: "should return nil",
			client: func() client.Client {
				c := mock_client.NewMockClient(ctrl)

				c.EXPECT().Delete(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil).
					Times(1)

				return c
			}(),
			wantErr: gomega.BeNil(),
		},
		{
			name: "should return error handled from Delete method",
			client: func() client.Client {
				c := mock_client.NewMockClient(ctrl)

				c.EXPECT().Delete(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("test error")).
					Times(1)

				return c
			}(),
			wantErr: gomega.HaveOccurred(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := NewAPIRuleOperator(tt.client, "fn-name", []unstructured.Unstructured{{}}...)

			err := o.Delete(context.Background(), DeleteOptions{})

			g.Expect(err).To(tt.wantErr)
		})
	}
}

func fixRawAPIRule(name, fnName string) unstructured.Unstructured {
	return unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "gateway.kyma-project.io/v1alpha1",
			"kind":       "APIRule",
			"metadata": map[string]interface{}{
				"name": name,
			},
			"spec": map[string]interface{}{
				"service": map[string]interface{}{
					"host": "test-host",
					"name": fnName,
					"port": int64(80),
				},
				"rules": []interface{}{
					map[string]interface{}{
						"methods": []interface{}{"PUT"},
						"path":    "/.*",
						"accessStrategies": []interface{}{
							map[string]interface{}{
								"handler": "allow",
							},
						},
					},
				},
				"gateway": "kyma-gateway.kyma-system.svc.cluster.local",
			},
		},
	}
}

func fixRawAPIRuleWithOwnerRef(name, fnName string, ref metav1.OwnerReference) unstructured.Unstructured {
	obj := fixRawAPIRule(name, fnName)
	obj.SetOwnerReferences([]metav1.OwnerReference{ref})

	return obj
}

func fixOwnerRef(name string) metav1.OwnerReference {
	return metav1.OwnerReference{
		APIVersion: "serverless.kyma-project.io/v1alpha1",
		Kind:       "Function",
		Name:       name,
		UID:        "1092378129381283128738189",
	}
}

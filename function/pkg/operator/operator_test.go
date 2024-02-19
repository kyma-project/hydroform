package operator

import (
	"context"
	errs "errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"

	"github.com/kyma-project/hydroform/function/pkg/client"
	mockclient "github.com/kyma-project/hydroform/function/pkg/client/automock"
)

var (
	testObj = unstructured.Unstructured{
		Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"name":      "test-obj",
				"namespace": "test-namespace",
			},
			"spec": map[string]interface{}{
				"test": "me",
				"subscriber": map[string]interface{}{
					"ref": map[string]interface{}{
						"kind":      "Service",
						"name":      "test-function-name",
						"namespace": "test-namespace",
					},
				},
			},
		},
	}
	testObj2 = unstructured.Unstructured{
		Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"name":      "test-obj2",
				"namespace": "test-namespace",
			},
			"spec": map[string]interface{}{
				"test": "me2",
				"subscriber": map[string]interface{}{
					"ref": map[string]interface{}{
						"kind":      "Service",
						"name":      "test-function-name",
						"namespace": "test-namespace",
					},
				},
			},
		},
	}
	testObjModifiedSpec = unstructured.Unstructured{
		Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"name":      "test-obj",
				"namespace": "test-namespace",
			},
			"spec": map[string]interface{}{
				"test": "me3",
				"subscriber": map[string]interface{}{
					"ref": map[string]interface{}{
						"kind":      "Service",
						"name":      "test-function-name",
						"namespace": "test-namespace",
					},
				},
			},
		},
	}
)

func Test_applyObject(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testObjWithLabelsAndAnnotations := *(testObj.DeepCopy())
	testObjWithLabelsAndAnnotations.Object["metadata"] = map[string]interface{}{
		"name":      "test-obj",
		"namespace": "test-namespace",
		"labels": map[string]interface{}{
			"aa": "bb",
			"cc": "dd",
		},
		"annotations": map[string]interface{}{
			"aa": "bb",
			"cc": "dd",
		},
	}

	type args struct {
		ctx    context.Context
		c      client.Client
		u      unstructured.Unstructured
		stages []string
	}
	tests := []struct {
		name    string
		args    args
		want    *unstructured.Unstructured
		want1   client.PostStatusEntry
		wantErr bool
	}{
		{
			name: "get returns error",
			args: args{
				c: func() client.Client {
					result := mockclient.NewMockClient(ctrl)

					result.EXPECT().
						Get(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(nil, fmt.Errorf("get failed")).
						Times(1)

					return result
				}(),
				u:      unstructured.Unstructured{},
				stages: []string{},
			},
			want:    &unstructured.Unstructured{},
			want1:   client.NewPostStatusEntryApplyFailed(unstructured.Unstructured{}),
			wantErr: true,
		},
		{
			name: "object is equal",
			args: args{
				c: func() client.Client {
					result := mockclient.NewMockClient(ctrl)
					result.EXPECT().
						Get(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(testObj.DeepCopy(), nil).
						Times(1)

					return result
				}(),
				u:      testObj,
				stages: []string{},
			},
			want:    &testObj,
			want1:   client.NewPostStatusEntrySkipped(testObj),
			wantErr: false,
		},
		{
			name: "update returns error",
			args: args{
				c: func() client.Client {
					result := mockclient.NewMockClient(ctrl)
					result.EXPECT().
						Get(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(testObj.DeepCopy(), nil).
						Times(1)

					result.EXPECT().
						Update(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(nil, fmt.Errorf("update failed")).
						AnyTimes()

					return result
				}(),
				u:      testObjModifiedSpec,
				stages: []string{},
			},
			want:    &testObjModifiedSpec,
			want1:   client.NewPostStatusEntryApplyFailed(testObjModifiedSpec),
			wantErr: true,
		},
		{
			name: "updated",
			args: args{
				c: func() client.Client {
					result := mockclient.NewMockClient(ctrl)

					result.EXPECT().
						Get(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(testObj2.DeepCopy(), nil).
						Times(1)

					result.EXPECT().
						Update(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(testObj.DeepCopy(), nil).
						AnyTimes()

					return result
				}(),
				u:      testObj,
				stages: []string{},
			},
			want:    &testObj,
			want1:   client.NewPostStatusEntryUpdated(testObj),
			wantErr: false,
		},
		{
			name: "updated with retries",
			args: args{
				c: func() client.Client {
					result := mockclient.NewMockClient(ctrl)

					result.EXPECT().
						Get(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(&testObj, fmt.Errorf("get failed, this should not stop the test")).
						Times(4).
						Return(testObj2.DeepCopy(), nil).
						Times(1)

					result.EXPECT().
						Update(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(testObj.DeepCopy(), nil).
						AnyTimes()

					return result
				}(),
				u:      testObj,
				stages: []string{},
			},
			want:    &testObj,
			want1:   client.NewPostStatusEntryUpdated(testObj),
			wantErr: false,
		},
		{
			name: "create fails",
			args: args{
				c: func() client.Client {
					result := mockclient.NewMockClient(ctrl)

					result.EXPECT().
						Get(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(nil, errors.NewNotFound(schema.GroupResource{}, "test error")).
						Times(1)

					result.EXPECT().
						Create(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(nil, fmt.Errorf("create error"))

					return result
				}(),
				u:      testObj,
				stages: []string{},
			},
			want:    &testObj,
			want1:   client.NewPostStatusEntryApplyFailed(testObj),
			wantErr: true,
		},
		{
			name: "create",
			args: args{
				c: func() client.Client {
					result := mockclient.NewMockClient(ctrl)

					result.EXPECT().
						Get(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(nil, errors.NewNotFound(schema.GroupResource{}, "test error")).
						Times(1)

					result.EXPECT().
						Create(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(testObj.DeepCopy(), nil)

					return result
				}(),
				u:      testObj,
				stages: []string{},
			},
			want:    &testObj,
			want1:   client.NewStatusEntryCreated(testObj),
			wantErr: false,
		},
		{
			name: "updated with labels and annotations",
			args: args{
				c: func() client.Client {
					result := mockclient.NewMockClient(ctrl)

					result.EXPECT().
						Get(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(testObj.DeepCopy(), nil).
						Times(1)

					result.EXPECT().
						Update(gomock.Any(), gomock.Eq(&testObjWithLabelsAndAnnotations), gomock.Any()).
						Return(testObjWithLabelsAndAnnotations.DeepCopy(), nil).
						Times(1)

					return result
				}(),
				u:      testObjWithLabelsAndAnnotations,
				stages: []string{},
			},
			want:    &testObjWithLabelsAndAnnotations,
			want1:   client.NewPostStatusEntryUpdated(testObjWithLabelsAndAnnotations),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := applyObject(tt.args.ctx, tt.args.c, tt.args.u, tt.args.stages)
			if (err != nil) != tt.wantErr {
				t.Errorf("applyObject() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			require.Equal(t, tt.want, got)
			require.Equal(t, tt.want1, got1)
		})
	}
}

func Test_deleteObject(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type args struct {
		i   client.Client
		u   unstructured.Unstructured
		ops DeleteOptions
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		want    client.PostStatusEntry
		wantErr bool
	}{
		{
			name: "delete failed",
			args: args{
				u: testObj,
				ops: DeleteOptions{
					DeletionPropagation: v1.DeletePropagationForeground,
				},
				i: func() client.Client {
					result := mockclient.NewMockClient(ctrl)

					result.EXPECT().
						Delete(gomock.Any(), testObj.GetName(), gomock.Any()).
						Return(fmt.Errorf("delete error")).
						Times(1)

					return result
				}(),
			},
			want:    client.NewPostStatusEntryDeleteFailed(testObj),
			wantErr: true,
		},
		{
			name: "delete",
			args: args{
				u: testObj,
				ops: DeleteOptions{
					DeletionPropagation: v1.DeletePropagationForeground,
				},
				i: func() client.Client {
					result := mockclient.NewMockClient(ctrl)

					result.EXPECT().
						Delete(gomock.Any(), testObj.GetName(), gomock.Any()).
						Return(nil).
						Times(1)

					return result
				}(),
			},
			want:    client.NewPostStatusEntryDeleted(testObj),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := deleteObject(tt.args.ctx, tt.args.i, tt.args.u, tt.args.ops)
			if (err != nil) != tt.wantErr {
				t.Errorf("deleteObject() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("deleteObject() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_fireCallbacks(t *testing.T) {
	type args struct {
		e   client.PostStatusEntry
		err error
		c   []Callback
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "panic recover",
			args: args{
				e:   client.NewStatusEntryCreated(testObj),
				err: nil,
				c: []Callback{
					func(_ interface{}, err error) error {
						panic("this is fine")
					},
				},
			},
			wantErr: true,
		},
		{
			name: "ioc error",
			args: args{
				e:   client.NewStatusEntryCreated(testObj),
				err: nil,
				c: []Callback{
					func(_ interface{}, err error) error {
						return err
					},
					func(v interface{}, err error) error {
						entry, ok := v.(client.PostStatusEntry)
						if !ok {
							return fmt.Errorf("invalid callback argument type")
						}
						if err != nil || entry.StatusType == client.StatusTypeCreated {
							return fmt.Errorf("this is not fine")
						}
						return nil
					},
				},
			},
			wantErr: true,
		},
		{
			name: "OK",
			args: args{
				e:   client.NewPostStatusEntryUpdated(testObj),
				err: nil,
				c: []Callback{
					func(_ interface{}, err error) error {
						return err
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := fireCallbacks(tt.args.e, tt.args.err, tt.args.c...); (err != nil) != tt.wantErr {
				t.Errorf("fireCallbacks() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_waitForObject(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	timeoutedContext, cencel := context.WithTimeout(context.Background(), time.Second)
	defer cencel()

	type args struct {
		ctx context.Context
		c   client.Client
		u   unstructured.Unstructured
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "should receive event type ADDED and return nil",
			args: args{
				ctx: context.Background(),
				c: func() client.Client {
					result := mockclient.NewMockClient(ctrl)
					fakeWatcher := watch.NewRaceFreeFake()
					testObject := fixUnstructured()
					fakeWatcher.Add(&testObject)

					result.EXPECT().
						Watch(gomock.Any(), gomock.Any()).
						Return(fakeWatcher, nil).
						Times(1)

					return result
				}(),
				u: fixUnstructured(),
			},
			wantErr: false,
		},
		{
			name: "should receive event type MODIFIED and return error",
			args: args{
				ctx: timeoutedContext,
				c: func() client.Client {
					result := mockclient.NewMockClient(ctrl)
					fakeWatcher := watch.NewRaceFreeFake()
					testObject := fixUnstructured()
					fakeWatcher.Modify(&testObject)

					result.EXPECT().
						Watch(gomock.Any(), gomock.Any()).
						Return(fakeWatcher, nil).
						Times(1)

					return result
				}(),
				u: fixUnstructured(),
			},
			wantErr: true,
		},
		{
			name: "should return error when watcher throws error",
			args: args{
				ctx: context.Background(),
				c: func() client.Client {
					result := mockclient.NewMockClient(ctrl)

					result.EXPECT().
						Watch(gomock.Any(), gomock.Any()).
						Return(nil, errs.New("sample error")).
						Times(1)

					return result
				}(),
				u: fixUnstructured(),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := waitForObject(tt.args.ctx, tt.args.c, tt.args.u); (err != nil) != tt.wantErr {
				t.Errorf("waitForObject() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_configurationObjectsAreEquivalent(t *testing.T) {
	defaultConfigurationObject := unstructured.Unstructured{
		Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"labels": map[string]interface{}{
					"aa": "bb",
					"cc": "dd",
				},
				"annotations": map[string]interface{}{
					"vv": "ww",
					"xx": "yy",
				},
				"omitted": "xx",
			},
			"spec": map[string]interface{}{
				"test": "me",
				"subscriber": map[string]interface{}{
					"ref": map[string]interface{}{
						"kind":      "Service",
						"name":      "test-function-name",
						"namespace": "test-namespace",
					},
				},
			},
			"omitted": "xx",
		},
	}

	type args struct {
		first  unstructured.Unstructured
		second unstructured.Unstructured
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "the same configuration objects",
			args: args{
				first:  defaultConfigurationObject,
				second: defaultConfigurationObject,
			},
			want: true,
		},
		{
			name: "similar configuration objects - changed metadata/^(labels|annotations)/*",
			args: args{
				first: defaultConfigurationObject,
				second: func() unstructured.Unstructured {
					u := *(defaultConfigurationObject.DeepCopy())
					u.Object["metadata"].(map[string]interface{})["omitted"] = "abcd"
					return u
				}(),
			},
			want: true,
		},
		{
			name: "similar configuration objects - changed ^(spec|metadata)",
			args: args{
				first: defaultConfigurationObject,
				second: func() unstructured.Unstructured {
					u := *(defaultConfigurationObject.DeepCopy())
					u.Object["omitted"] = "abcd"
					return u
				}(),
			},
			want: true,
		},
		{
			name: "configuration object with changed element metadata/labels/*",
			args: args{
				first: defaultConfigurationObject,
				second: func() unstructured.Unstructured {
					u := *(defaultConfigurationObject.DeepCopy())
					u.Object["metadata"].(map[string]interface{})["labels"].(map[string]interface{})["aa"] = "abcd"
					return u
				}(),
			},
			want: false,
		},
		{
			name: "configuration object with changed element metadata/annotations/*",
			args: args{
				first: defaultConfigurationObject,
				second: func() unstructured.Unstructured {
					u := *(defaultConfigurationObject.DeepCopy())
					u.Object["metadata"].(map[string]interface{})["annotations"].(map[string]interface{})["xx"] = "abcd"
					return u
				}(),
			},
			want: false,
		},
		{
			name: "configuration object with changed element spec/*",
			args: args{
				first: defaultConfigurationObject,
				second: func() unstructured.Unstructured {
					u := *(defaultConfigurationObject.DeepCopy())
					u.Object["spec"].(map[string]interface{})["test"] = "abcd"
					return u
				}(),
			},
			want: false,
		},
		{
			name: "configuration object with added element metadata/labels/*",
			args: args{
				first: defaultConfigurationObject,
				second: func() unstructured.Unstructured {
					u := *(defaultConfigurationObject.DeepCopy())
					u.Object["metadata"].(map[string]interface{})["labels"].(map[string]interface{})["new"] = "abcd"
					return u
				}(),
			},
			want: false,
		},
		{
			name: "configuration object with added element metadata/annotations/*",
			args: args{
				first: defaultConfigurationObject,
				second: func() unstructured.Unstructured {
					u := *(defaultConfigurationObject.DeepCopy())
					u.Object["metadata"].(map[string]interface{})["annotations"].(map[string]interface{})["new"] = "abcd"
					return u
				}(),
			},
			want: false,
		},
		{
			name: "configuration object with added element spec/*",
			args: args{
				first: defaultConfigurationObject,
				second: func() unstructured.Unstructured {
					u := *(defaultConfigurationObject.DeepCopy())
					u.Object["spec"].(map[string]interface{})["new"] = "abcd"
					return u
				}(),
			},
			want: false,
		},
		{
			name: "configuration object with removed element metadata/labels/*",
			args: args{
				first: defaultConfigurationObject,
				second: func() unstructured.Unstructured {
					u := *(defaultConfigurationObject.DeepCopy())
					delete(u.Object["metadata"].(map[string]interface{})["labels"].(map[string]interface{}), "cc")
					return u
				}(),
			},
			want: false,
		},
		{
			name: "configuration object with removed element metadata/annotations/*",
			args: args{
				first: defaultConfigurationObject,
				second: func() unstructured.Unstructured {
					u := *(defaultConfigurationObject.DeepCopy())
					delete(u.Object["metadata"].(map[string]interface{})["annotations"].(map[string]interface{}), "vv")
					return u
				}(),
			},
			want: false,
		},
		{
			name: "configuration object with removed element spec/*",
			args: args{
				first: defaultConfigurationObject,
				second: func() unstructured.Unstructured {
					u := *(defaultConfigurationObject.DeepCopy())
					delete(u.Object["spec"].(map[string]interface{})["subscriber"].(map[string]interface{})["ref"].(map[string]interface{}),
						"kind")
					return u
				}(),
			},
			want: false,
		},
		{
			name: "configuration object with nil element metadata/labels",
			args: args{
				first: defaultConfigurationObject,
				second: func() unstructured.Unstructured {
					u := *(defaultConfigurationObject.DeepCopy())
					u.Object["metadata"].(map[string]interface{})["labels"] = nil
					return u
				}(),
			},
			want: false,
		},
		{
			name: "configuration object with nil element metadata/annotations",
			args: args{
				first: defaultConfigurationObject,
				second: func() unstructured.Unstructured {
					u := *(defaultConfigurationObject.DeepCopy())
					u.Object["metadata"].(map[string]interface{})["annotations"] = nil
					return u
				}(),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := configurationObjectsAreEquivalent(tt.args.first, tt.args.second)
			require.Equal(t, tt.want, got)
			require.Equal(t, nil, err)
		})
	}
}

func Test_updateConfigurationObject(t *testing.T) {
	defaultDestinationConfigurationObject := unstructured.Unstructured{
		Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"labels": map[string]interface{}{
					"destLabel1": "destLabelValue1",
					"destLabel2": "destLabelValue2",
				},
				"annotations": map[string]interface{}{
					"destAnnotation1": "destAnnotationValue1",
					"destAnnotation2": "destAnnotationValue2",
				},
				"destOtherMetadata": "destOtherMetadataValue",
			},
			"spec": map[string]interface{}{
				"destSpec1": "destSpecValue1",
				"destSpec2": map[string]interface{}{
					"destSomeSpec21": map[string]interface{}{
						"destSpec211": "destSpecValue211",
						"destSpec212": "destSpecValue212",
						"destSpec213": "destSpecValue213",
					},
				},
				"destSpec3": "destSpecValue3",
			},
			"destOther": "destOtherValue",
		},
	}
	defaultSourceConfigurationObject := unstructured.Unstructured{
		Object: map[string]interface{}{
			"srcOther1": "srcOtherValue1",
			"metadata": map[string]interface{}{
				"labels": map[string]interface{}{
					"srcLabel1": "srcLabelValue1",
					"srcLabel2": "srcLabelValue2",
				},
				"annotations": map[string]interface{}{
					"srcAnnotation1": "srcAnnotationValue1",
					"srcAnnotation2": "srcAnnotationValue2",
				},
				"srcOtherMetadata1": "srcOtherMetadataValue1",
				"srcOtherMetadata2": "srcOtherMetadataValue2",
			},
			"srcOther2": "srcOtherValue2",
			"spec": map[string]interface{}{
				"srcSpec1": "srcSpecValue1",
				"srcSpec2": "srcSpecValue2",
				"srcSpec3": map[string]interface{}{
					"srcSpec31": map[string]interface{}{
						"srcSpec311": "srcSpecValue311",
					},
				},
			},
			"srcOther3": "srcOtherValue3",
		},
	}
	defaultUpdatedConfigurationObject := unstructured.Unstructured{
		Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"labels": map[string]interface{}{
					"srcLabel1": "srcLabelValue1",
					"srcLabel2": "srcLabelValue2",
				},
				"annotations": map[string]interface{}{
					"srcAnnotation1": "srcAnnotationValue1",
					"srcAnnotation2": "srcAnnotationValue2",
				},
				"destOtherMetadata": "destOtherMetadataValue",
			},
			"spec": map[string]interface{}{
				"srcSpec1": "srcSpecValue1",
				"srcSpec2": "srcSpecValue2",
				"srcSpec3": map[string]interface{}{
					"srcSpec31": map[string]interface{}{
						"srcSpec311": "srcSpecValue311",
					},
				},
			},
			"destOther": "destOtherValue",
		},
	}

	type args struct {
		destination unstructured.Unstructured
		source      unstructured.Unstructured
	}
	tests := []struct {
		name string
		args args
		want unstructured.Unstructured
	}{
		{
			name: "should copy `spec`, `metadata/labels`, `metadata/annotations` and not touch other content",
			args: args{
				destination: defaultDestinationConfigurationObject,
				source:      defaultSourceConfigurationObject,
			},
			want: defaultUpdatedConfigurationObject,
		},
		{
			name: "should copy `metadata/labels`, `metadata/annotations` when destination doesn't contain these elements",
			args: args{
				destination: func() unstructured.Unstructured {
					u := *(defaultDestinationConfigurationObject.DeepCopy())
					delete(u.Object["metadata"].(map[string]interface{}), "labels")
					delete(u.Object["metadata"].(map[string]interface{}), "annotations")
					return u
				}(),
				source: defaultSourceConfigurationObject,
			},
			want: defaultUpdatedConfigurationObject,
		},
		{
			name: "should remove `metadata/labels`, `metadata/annotations` when source doesn't contain these elements",
			args: args{
				destination: defaultDestinationConfigurationObject,
				source: func() unstructured.Unstructured {
					u := *(defaultSourceConfigurationObject.DeepCopy())
					delete(u.Object["metadata"].(map[string]interface{}), "labels")
					delete(u.Object["metadata"].(map[string]interface{}), "annotations")
					return u
				}(),
			},
			want: func() unstructured.Unstructured {
				u := *(defaultUpdatedConfigurationObject.DeepCopy())
				delete(u.Object["metadata"].(map[string]interface{}), "labels")
				delete(u.Object["metadata"].(map[string]interface{}), "annotations")
				return u
			}(),
		},
		{
			name: "should not copy `metadata/labels`, `metadata/annotations` when destination and source doesn't contain these elements",
			args: args{
				destination: func() unstructured.Unstructured {
					u := *(defaultDestinationConfigurationObject.DeepCopy())
					delete(u.Object["metadata"].(map[string]interface{}), "labels")
					delete(u.Object["metadata"].(map[string]interface{}), "annotations")
					return u
				}(),
				source: func() unstructured.Unstructured {
					u := *(defaultSourceConfigurationObject.DeepCopy())
					delete(u.Object["metadata"].(map[string]interface{}), "labels")
					delete(u.Object["metadata"].(map[string]interface{}), "annotations")
					return u
				}(),
			},
			want: func() unstructured.Unstructured {
				u := *(defaultUpdatedConfigurationObject.DeepCopy())
				delete(u.Object["metadata"].(map[string]interface{}), "labels")
				delete(u.Object["metadata"].(map[string]interface{}), "annotations")
				return u
			}(),
		},
		{
			name: "should copy `metadata/labels`, `metadata/annotations` when destination contain these elements with nil value",
			args: args{
				destination: func() unstructured.Unstructured {
					u := *(defaultDestinationConfigurationObject.DeepCopy())
					u.Object["metadata"].(map[string]interface{})["labels"] = nil
					u.Object["metadata"].(map[string]interface{})["annotations"] = nil
					return u
				}(),
				source: defaultSourceConfigurationObject,
			},
			want: defaultUpdatedConfigurationObject,
		},
		{
			name: "should copy `metadata/labels`, `metadata/annotations` when source contain these elements with nil value",
			args: args{
				destination: defaultDestinationConfigurationObject,
				source: func() unstructured.Unstructured {
					u := *(defaultSourceConfigurationObject.DeepCopy())
					u.Object["metadata"].(map[string]interface{})["labels"] = nil
					u.Object["metadata"].(map[string]interface{})["annotations"] = nil
					return u
				}(),
			},
			want: func() unstructured.Unstructured {
				u := *(defaultUpdatedConfigurationObject.DeepCopy())
				u.Object["metadata"].(map[string]interface{})["labels"] = nil
				u.Object["metadata"].(map[string]interface{})["annotations"] = nil
				return u
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			destination := *(tt.args.destination.DeepCopy())
			updateConfigurationObject(&destination, tt.args.source)
			require.Equal(t, tt.want, destination)
		})
	}
}

func fixUnstructured() unstructured.Unstructured {
	gitRepo := unstructured.Unstructured{}
	gitRepo.SetAPIVersion("testapiversion")
	gitRepo.SetKind("testkind")
	gitRepo.SetName("test")
	gitRepo.SetNamespace("test")
	gitRepo.SetResourceVersion("testResourceVersion")
	return gitRepo
}

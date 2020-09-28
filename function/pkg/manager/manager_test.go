package manager

import (
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/kyma-incubator/hydroform/function/pkg/client"
	"github.com/kyma-incubator/hydroform/function/pkg/operator"
	mock_operator "github.com/kyma-incubator/hydroform/function/pkg/operator/automock"
	"github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"reflect"
	"testing"
)

func TestNewManager(t *testing.T) {
	type args struct {
		operators map[operator.Operator][]operator.Operator
	}
	tests := []struct {
		name string
		args args
		want manager
	}{
		{
			name: "should be ok",
			args: args{},
			want: manager{args{}.operators},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewManager(tt.args.operators); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewManager() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_manager_Do(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type fields struct {
		operators map[operator.Operator][]operator.Operator
	}
	type args struct {
		options ManagerOptions
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "should be ok with no operators",
			fields:  fields{},
			args:    args{},
			wantErr: false,
		},
		{
			name: "should be ok with parents",
			fields: fields{
				operators: map[operator.Operator][]operator.Operator{
					fixOperatorMock(ctrl, 1, 0): {},
				},
			},
			args:    args{},
			wantErr: false,
		},
		{
			name: "should be ok with parents and children",
			fields: fields{
				operators: map[operator.Operator][]operator.Operator{
					fixOperatorMock(ctrl, 1, 0): {
						fixOperatorMock(ctrl, 1, 0),
						fixOperatorMock(ctrl, 1, 0),
						fixOperatorMock(ctrl, 1, 0),
					},
					fixOperatorMock(ctrl, 1, 0): {
						fixOperatorMock(ctrl, 1, 0),
						fixOperatorMock(ctrl, 1, 0),
					},
				},
			},
			args:    args{},
			wantErr: false,
		},
		{
			name: "should be ok with nil parent and children",
			fields: fields{
				operators: map[operator.Operator][]operator.Operator{
					nil: {
						fixOperatorMock(ctrl, 1, 0),
						fixOperatorMock(ctrl, 1, 0),
						fixOperatorMock(ctrl, 1, 0),
					},
				},
			},
			args:    args{},
			wantErr: false,
		},
		{
			name: "should be ok with parent and nil child",
			fields: fields{
				operators: map[operator.Operator][]operator.Operator{
					fixOperatorMock(ctrl, 1, 0): {
						nil,
						fixOperatorMock(ctrl, 1, 0),
					},
				},
			},
			args:    args{},
			wantErr: false,
		},
		{
			name: "should be error without purge",
			fields: fields{
				operators: map[operator.Operator][]operator.Operator{
					fixOperatorMock(ctrl, 1, 0): {
						nil,
						fixOperatorMockWithError(ctrl, 1, 0, errors.New("any error")),
						fixOperatorMock(ctrl, 0, 0),
					},
					fixOperatorMock(ctrl, 0, 0): {
						fixOperatorMock(ctrl, 0, 0),
						fixOperatorMock(ctrl, 0, 0),
					},
				},
			},
			args: args{
				options: ManagerOptions{
					OnError: NothingOnError,
				},
			},
			wantErr: true,
		},
		{
			name: "should be purge after error",
			fields: fields{
				operators: map[operator.Operator][]operator.Operator{
					fixOperatorMock(ctrl, 1, 1): {
						nil,
						fixOperatorMockWithError(ctrl, 1, 0, errors.New("any error")),
						fixOperatorMock(ctrl, 0, 0),
					},
					fixOperatorMock(ctrl, 0, 1): {
						fixOperatorMock(ctrl, 0, 0),
						fixOperatorMock(ctrl, 0, 0),
					},
				},
			},
			args: args{
				options: ManagerOptions{
					OnError: PurgeOnError,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := manager{
				operators: tt.fields.operators,
			}
			if err := m.Do(tt.args.options); (err != nil) != tt.wantErr {
				t.Errorf("Do() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_manager_getDryRunFlag(t *testing.T) {
	type args struct {
		dryRun bool
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "should be ok without returned elements",
			args: args{
				dryRun: false,
			},
			want: nil,
		},
		{
			name: "should be ok with one returned elements",
			args: args{
				dryRun: true,
			},
			want: []string{metav1.DryRunAll},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &manager{
				operators: nil,
			}
			if got := m.getDryRunFlag(tt.args.dryRun); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getDryRunFlag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_manager_manageOperators(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type fields struct {
		operators map[operator.Operator][]operator.Operator
	}
	type args struct {
		options ManagerOptions
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "should be ok without operators",
			fields:  fields{},
			args:    args{},
			wantErr: false,
		},
		{
			name: "should be ok with operators",
			fields: fields{
				operators: map[operator.Operator][]operator.Operator{
					fixOperatorMock(ctrl, 1, 0): {},
					nil:                         {},
					fixOperatorMock(ctrl, 1, 0): {
						nil,
						fixOperatorMock(ctrl, 1, 0),
						fixOperatorMock(ctrl, 1, 0),
					},
				},
			},
			args:    args{},
			wantErr: false,
		},
		{
			name: "should be error without purge",
			fields: fields{
				operators: map[operator.Operator][]operator.Operator{
					fixOperatorMockWithError(ctrl, 1, 0, errors.New("any error")): {},
					fixOperatorMock(ctrl, 0, 0): {
						fixOperatorMock(ctrl, 0, 0),
						fixOperatorMock(ctrl, 0, 0),
					},
				},
			},
			args:    args{},
			wantErr: true,
		},
		{
			name: "should be error after child error occurred",
			fields: fields{
				operators: map[operator.Operator][]operator.Operator{
					fixOperatorMock(ctrl, 1, 0): {},
					fixOperatorMock(ctrl, 1, 0): {
						fixOperatorMock(ctrl, 1, 0),
						fixOperatorMockWithError(ctrl, 1, 0, errors.New("any error")),
						fixOperatorMock(ctrl, 0, 0),
					},
				},
			},
			args:    args{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &manager{
				operators: tt.fields.operators,
			}
			if err := m.manageOperators(tt.args.options); (err != nil) != tt.wantErr {
				t.Errorf("manageOperators() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_manager_ownerReferenceCallback(t *testing.T) {
	type args struct {
		callbacks operator.Callbacks
		list      *OwnerReferenceList
	}
	tests := []struct {
		name     string
		args     args
		wantPre  gomega.OmegaMatcher
		wantPost gomega.OmegaMatcher
	}{
		{
			name: "should be ok without list",
			args: args{
				callbacks: operator.Callbacks{},
				list:      nil,
			},
			wantPre:  gomega.HaveLen(0),
			wantPost: gomega.HaveLen(0),
		},
		{
			name: "should be ok with list",
			args: args{
				callbacks: operator.Callbacks{},
				list:      &OwnerReferenceList{},
			},
			wantPre:  gomega.HaveLen(0),
			wantPost: gomega.HaveLen(1),
		},
		{
			name: "should be ok with Callbacks and list",
			args: args{
				callbacks: operator.Callbacks{
					Pre: []operator.Callback{
						func(i interface{}, err error) error { return nil },
						func(i interface{}, err error) error { return nil },
					},
					Post: []operator.Callback{
						func(i interface{}, err error) error { return nil },
					},
				},
				list: &OwnerReferenceList{},
			},
			wantPre:  gomega.HaveLen(2),
			wantPost: gomega.HaveLen(2),
		},
		{
			name: "should be ok with Callback but without list",
			args: args{
				callbacks: operator.Callbacks{
					Pre: []operator.Callback{
						func(i interface{}, err error) error { return nil },
						func(i interface{}, err error) error { return nil },
					},
					Post: []operator.Callback{
						func(i interface{}, err error) error { return nil },
					},
				},
				list: nil,
			},
			wantPre:  gomega.HaveLen(2),
			wantPost: gomega.HaveLen(1),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewWithT(t)
			m := &manager{
				operators: nil,
			}
			got := m.ownerReferenceCallback(tt.args.callbacks, tt.args.list)
			g.Expect(got.Pre).To(tt.wantPre)
			g.Expect(got.Post).To(tt.wantPost)
		})
	}
}

func Test_manager_run_ownerReferenceCallback(t *testing.T) {

	tests := []struct {
		name           string
		givenInterface interface{}
		givenError     error
		expectedErr    gomega.OmegaMatcher
		expectedList   gomega.OmegaMatcher
	}{
		{
			name: "should be ok",
			givenInterface: client.PostStatusEntry{
				StatusType: client.StatusTypeCreated,
				Unstructured: unstructured.Unstructured{
					Object: fixCommonUnstructured(),
				},
			},
			givenError:  nil,
			expectedErr: gomega.BeNil(),
			expectedList: gomega.Equal(
				OwnerReferenceList{
					List: []metav1.OwnerReference{
						{
							APIVersion: "test_apiVersion",
							Kind:       "test_kind",
							Name:       "test_name",
							UID:        "test_uid",
						},
					},
				},
			),
		},
		{
			name:           "should be error on wrong input type",
			givenInterface: "bad type",
			givenError:     nil,
			expectedErr:    gomega.HaveOccurred(),
			expectedList:   gomega.Equal(OwnerReferenceList{}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewWithT(t)
			m := &manager{
				operators: nil,
			}
			list := &OwnerReferenceList{}

			got := m.ownerReferenceCallback(operator.Callbacks{}, list)
			err := got.Post[0](tt.givenInterface, tt.givenError)
			g.Expect(err).To(tt.expectedErr)
			g.Expect(*list).To(tt.expectedList)
		})
	}
}

func Test_manager_purgeParents(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type fields struct {
		operators map[operator.Operator][]operator.Operator
	}
	type args struct {
		options ManagerOptions
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "should do nothing",
			fields: fields{
				operators: map[operator.Operator][]operator.Operator{},
			},
			args: args{},
		},
		{
			name: "should purge all parents",
			fields: fields{
				operators: map[operator.Operator][]operator.Operator{
					fixOperatorMock(ctrl, 0, 1): {
						fixOperatorMock(ctrl, 0, 0),
						fixOperatorMock(ctrl, 0, 0),
					},
					fixOperatorMock(ctrl, 0, 1): {},
				},
			},
			args: args{},
		},
		{
			name: "should purge all parents even parents with error",
			fields: fields{
				operators: map[operator.Operator][]operator.Operator{
					fixOperatorMock(ctrl, 0, 1): {
						fixOperatorMock(ctrl, 0, 0),
						fixOperatorMock(ctrl, 0, 0),
					},
					fixOperatorMockWithError(ctrl, 0, 1, errors.New("any error")): {
						fixOperatorMock(ctrl, 0, 0),
					},
					fixOperatorMock(ctrl, 0, 1): {
						fixOperatorMock(ctrl, 0, 0),
					},
				},
			},
			args: args{},
		},
		{
			name: "should be ok with callbacks and nil parent",
			fields: fields{
				operators: map[operator.Operator][]operator.Operator{
					fixOperatorMock(ctrl, 0, 1): {
						nil,
					},
					nil:                         {},
					fixOperatorMock(ctrl, 0, 1): {},
				},
			},
			args: args{
				options: ManagerOptions{
					Callbacks: operator.Callbacks{
						Pre: []operator.Callback{
							func(i interface{}, err error) error { return nil },
						},
						Post: []operator.Callback{
							func(i interface{}, err error) error { return nil },
						},
					},
					DryRun: true,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &manager{
				operators: tt.fields.operators,
			}
			m.purgeParents(tt.args.options)
		})
	}
}

func Test_manager_useOperator(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type args struct {
		opr        operator.Operator
		options    ManagerOptions
		references []metav1.OwnerReference
	}
	tests := []struct {
		name    string
		args    args
		want    []metav1.OwnerReference
		wantErr bool
	}{
		{
			name: "should be ok with nil operator",
			args: args{
				opr:        nil,
				options:    ManagerOptions{},
				references: nil,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "should be ok with operator",
			args: args{
				opr:        fixOperatorMock(ctrl, 1, 0),
				options:    ManagerOptions{},
				references: nil,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "should be ok with operator and options",
			args: args{
				opr: fixOperatorMock(ctrl, 1, 0),
				options: ManagerOptions{
					Callbacks: operator.Callbacks{
						Pre: []operator.Callback{
							func(i interface{}, err error) error { return nil },
						},
						Post: []operator.Callback{
							func(i interface{}, err error) error { return nil },
						},
					},
					DryRun:             true,
					SetOwnerReferences: true,
				},
				references: nil,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "should be error with operator",
			args: args{
				opr:        fixOperatorMockWithError(ctrl, 1, 0, errors.New("any error")),
				options:    ManagerOptions{},
				references: nil,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &manager{
				operators: nil,
			}
			got, err := m.useOperator(tt.args.opr, tt.args.options, tt.args.references)
			if (err != nil) != tt.wantErr {
				t.Errorf("useOperator() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("useOperator() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func fixOperatorMock(ctrl *gomock.Controller, applyTimes, deleteTimes int) operator.Operator {
	return fixOperatorMockWithError(ctrl, applyTimes, deleteTimes, nil)
}

func fixOperatorMockWithError(ctrl *gomock.Controller, applyTimes, deleteTimes int, err error) operator.Operator {
	opr := mock_operator.NewMockOperator(ctrl)
	opr.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(err).Times(applyTimes)
	opr.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(err).Times(deleteTimes)
	return opr
}

func fixCommonUnstructured() map[string]interface{} {
	return map[string]interface{}{
		"apiVersion": "test_apiVersion",
		"kind":       "test_kind",
		"metadata": map[string]interface{}{
			"name": "test_name",
			"uid":  "test_uid",
		},
	}
}

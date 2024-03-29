// Code generated by MockGen. DO NOT EDIT.
// Source: operator.go

// Package mock_operator is a generated GoMock package.
package mock_operator

import (
	context "context"
	gomock "github.com/golang/mock/gomock"
	operator "github.com/kyma-project/hydroform/function/pkg/operator"
	reflect "reflect"
)

// MockOperator is a mock of Operator interface
type MockOperator struct {
	ctrl     *gomock.Controller
	recorder *MockOperatorMockRecorder
}

// MockOperatorMockRecorder is the mock recorder for MockOperator
type MockOperatorMockRecorder struct {
	mock *MockOperator
}

// NewMockOperator creates a new mock instance
func NewMockOperator(ctrl *gomock.Controller) *MockOperator {
	mock := &MockOperator{ctrl: ctrl}
	mock.recorder = &MockOperatorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockOperator) EXPECT() *MockOperatorMockRecorder {
	return m.recorder
}

// Apply mocks base method
func (m *MockOperator) Apply(arg0 context.Context, arg1 operator.ApplyOptions) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Apply", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Apply indicates an expected call of Apply
func (mr *MockOperatorMockRecorder) Apply(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Apply", reflect.TypeOf((*MockOperator)(nil).Apply), arg0, arg1)
}

// Delete mocks base method
func (m *MockOperator) Delete(arg0 context.Context, arg1 operator.DeleteOptions) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete
func (mr *MockOperatorMockRecorder) Delete(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockOperator)(nil).Delete), arg0, arg1)
}

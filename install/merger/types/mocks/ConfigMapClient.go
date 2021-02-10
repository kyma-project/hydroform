// Code generated by mockery v2.5.1. DO NOT EDIT.

package mocks

import (
	context "context"

	corev1 "k8s.io/api/core/v1"

	mock "github.com/stretchr/testify/mock"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ConfigMapClient is an autogenerated mock type for the ConfigMapClient type
type ConfigMapClient struct {
	mock.Mock
}

// Get provides a mock function with given fields: ctx, name, opts
func (_m *ConfigMapClient) Get(ctx context.Context, name string, opts v1.GetOptions) (*corev1.ConfigMap, error) {
	ret := _m.Called(ctx, name, opts)

	var r0 *corev1.ConfigMap
	if rf, ok := ret.Get(0).(func(context.Context, string, v1.GetOptions) *corev1.ConfigMap); ok {
		r0 = rf(ctx, name, opts)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*corev1.ConfigMap)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, v1.GetOptions) error); ok {
		r1 = rf(ctx, name, opts)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Update provides a mock function with given fields: ctx, configMap, opts
func (_m *ConfigMapClient) Update(ctx context.Context, configMap *corev1.ConfigMap, opts v1.UpdateOptions) (*corev1.ConfigMap, error) {
	ret := _m.Called(ctx, configMap, opts)

	var r0 *corev1.ConfigMap
	if rf, ok := ret.Get(0).(func(context.Context, *corev1.ConfigMap, v1.UpdateOptions) *corev1.ConfigMap); ok {
		r0 = rf(ctx, configMap, opts)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*corev1.ConfigMap)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *corev1.ConfigMap, v1.UpdateOptions) error); ok {
		r1 = rf(ctx, configMap, opts)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

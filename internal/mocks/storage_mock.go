// Code generated by mockery v2.52.1. DO NOT EDIT.

package mocks

import (
	context "context"

	models "github.com/runtime-metrics-course/internal/models"
	mock "github.com/stretchr/testify/mock"
)

// StorageIface is an autogenerated mock type for the StorageIface type
type StorageIface struct {
	mock.Mock
}

// GetMetrics provides a mock function with given fields: ctx
func (_m *StorageIface) GetMetrics(ctx context.Context) (models.Metrics, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for GetMetrics")
	}

	var r0 models.Metrics
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (models.Metrics, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) models.Metrics); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Get(0).(models.Metrics)
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Ping provides a mock function with no fields
func (_m *StorageIface) Ping() error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Ping")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateCounter provides a mock function with given fields: ctx, name, value
func (_m *StorageIface) UpdateCounter(ctx context.Context, name string, value int64) error {
	ret := _m.Called(ctx, name, value)

	if len(ret) == 0 {
		panic("no return value specified for UpdateCounter")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, int64) error); ok {
		r0 = rf(ctx, name, value)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateGauge provides a mock function with given fields: ctx, name, value
func (_m *StorageIface) UpdateGauge(ctx context.Context, name string, value float64) error {
	ret := _m.Called(ctx, name, value)

	if len(ret) == 0 {
		panic("no return value specified for UpdateGauge")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, float64) error); ok {
		r0 = rf(ctx, name, value)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewStorageIface creates a new instance of StorageIface. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewStorageIface(t interface {
	mock.TestingT
	Cleanup(func())
}) *StorageIface {
	mock := &StorageIface{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

// Code generated by mockery v1.1.2. DO NOT EDIT.

package kubernetes_mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	v1beta1 "k8s.io/api/policy/v1beta1"
)

// EvictionInterface is an autogenerated mock type for the EvictionInterface type
type EvictionInterface struct {
	mock.Mock
}

// Evict provides a mock function with given fields: ctx, eviction
func (_m *EvictionInterface) Evict(ctx context.Context, eviction *v1beta1.Eviction) error {
	ret := _m.Called(ctx, eviction)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *v1beta1.Eviction) error); ok {
		r0 = rf(ctx, eviction)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

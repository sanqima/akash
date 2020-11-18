// Code generated by mockery v1.1.2. DO NOT EDIT.

package kubernetes_mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	v1 "k8s.io/api/core/v1"
)

// NodeExpansion is an autogenerated mock type for the NodeExpansion type
type NodeExpansion struct {
	mock.Mock
}

// PatchStatus provides a mock function with given fields: ctx, nodeName, data
func (_m *NodeExpansion) PatchStatus(ctx context.Context, nodeName string, data []byte) (*v1.Node, error) {
	ret := _m.Called(ctx, nodeName, data)

	var r0 *v1.Node
	if rf, ok := ret.Get(0).(func(context.Context, string, []byte) *v1.Node); ok {
		r0 = rf(ctx, nodeName, data)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1.Node)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, []byte) error); ok {
		r1 = rf(ctx, nodeName, data)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
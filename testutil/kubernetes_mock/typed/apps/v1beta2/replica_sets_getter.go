// Code generated by mockery v1.1.2. DO NOT EDIT.

package kubernetes_mocks

import (
	mock "github.com/stretchr/testify/mock"
	v1beta2 "k8s.io/client-go/kubernetes/typed/apps/v1beta2"
)

// ReplicaSetsGetter is an autogenerated mock type for the ReplicaSetsGetter type
type ReplicaSetsGetter struct {
	mock.Mock
}

// ReplicaSets provides a mock function with given fields: namespace
func (_m *ReplicaSetsGetter) ReplicaSets(namespace string) v1beta2.ReplicaSetInterface {
	ret := _m.Called(namespace)

	var r0 v1beta2.ReplicaSetInterface
	if rf, ok := ret.Get(0).(func(string) v1beta2.ReplicaSetInterface); ok {
		r0 = rf(namespace)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(v1beta2.ReplicaSetInterface)
		}
	}

	return r0
}

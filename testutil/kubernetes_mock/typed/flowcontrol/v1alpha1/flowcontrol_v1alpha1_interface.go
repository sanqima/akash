// Code generated by mockery v1.1.2. DO NOT EDIT.

package kubernetes_mocks

import (
	mock "github.com/stretchr/testify/mock"
	rest "k8s.io/client-go/rest"

	v1alpha1 "k8s.io/client-go/kubernetes/typed/flowcontrol/v1alpha1"
)

// FlowcontrolV1alpha1Interface is an autogenerated mock type for the FlowcontrolV1alpha1Interface type
type FlowcontrolV1alpha1Interface struct {
	mock.Mock
}

// FlowSchemas provides a mock function with given fields:
func (_m *FlowcontrolV1alpha1Interface) FlowSchemas() v1alpha1.FlowSchemaInterface {
	ret := _m.Called()

	var r0 v1alpha1.FlowSchemaInterface
	if rf, ok := ret.Get(0).(func() v1alpha1.FlowSchemaInterface); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(v1alpha1.FlowSchemaInterface)
		}
	}

	return r0
}

// PriorityLevelConfigurations provides a mock function with given fields:
func (_m *FlowcontrolV1alpha1Interface) PriorityLevelConfigurations() v1alpha1.PriorityLevelConfigurationInterface {
	ret := _m.Called()

	var r0 v1alpha1.PriorityLevelConfigurationInterface
	if rf, ok := ret.Get(0).(func() v1alpha1.PriorityLevelConfigurationInterface); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(v1alpha1.PriorityLevelConfigurationInterface)
		}
	}

	return r0
}

// RESTClient provides a mock function with given fields:
func (_m *FlowcontrolV1alpha1Interface) RESTClient() rest.Interface {
	ret := _m.Called()

	var r0 rest.Interface
	if rf, ok := ret.Get(0).(func() rest.Interface); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(rest.Interface)
		}
	}

	return r0
}

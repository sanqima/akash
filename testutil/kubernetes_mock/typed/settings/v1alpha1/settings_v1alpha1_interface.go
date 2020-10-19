// Code generated by mockery v1.1.2. DO NOT EDIT.

package kubernetes_mocks

import (
	mock "github.com/stretchr/testify/mock"
	rest "k8s.io/client-go/rest"

	v1alpha1 "k8s.io/client-go/kubernetes/typed/settings/v1alpha1"
)

// SettingsV1alpha1Interface is an autogenerated mock type for the SettingsV1alpha1Interface type
type SettingsV1alpha1Interface struct {
	mock.Mock
}

// PodPresets provides a mock function with given fields: namespace
func (_m *SettingsV1alpha1Interface) PodPresets(namespace string) v1alpha1.PodPresetInterface {
	ret := _m.Called(namespace)

	var r0 v1alpha1.PodPresetInterface
	if rf, ok := ret.Get(0).(func(string) v1alpha1.PodPresetInterface); ok {
		r0 = rf(namespace)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(v1alpha1.PodPresetInterface)
		}
	}

	return r0
}

// RESTClient provides a mock function with given fields:
func (_m *SettingsV1alpha1Interface) RESTClient() rest.Interface {
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

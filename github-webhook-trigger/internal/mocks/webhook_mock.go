// Code generated by mockery v2.32.0. DO NOT EDIT.

package mocks

import (
	sdk "github.com/konstellation-io/kai-sdk/go-sdk/sdk"
	mock "github.com/stretchr/testify/mock"
)

// WebhookMock is an autogenerated mock type for the Webhook type
type WebhookMock struct {
	mock.Mock
}

// InitWebhook provides a mock function with given fields: kaiSDK
func (_m *WebhookMock) InitWebhook(kaiSDK sdk.KaiSDK) error {
	ret := _m.Called(kaiSDK)

	var r0 error
	if rf, ok := ret.Get(0).(func(sdk.KaiSDK) error); ok {
		r0 = rf(kaiSDK)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewWebhookMock creates a new instance of WebhookMock. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewWebhookMock(t interface {
	mock.TestingT
	Cleanup(func())
}) *WebhookMock {
	mock := &WebhookMock{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

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

// InitWebhook provides a mock function with given fields: events, githubSecret, kaiSDK
func (_m *WebhookMock) InitWebhook(events []string, githubSecret string, kaiSDK sdk.KaiSDK) {
	_m.Called(events, githubSecret, kaiSDK)
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

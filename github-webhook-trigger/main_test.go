package main

import (
	"testing"

	"github.com/konstellation-io/kai-processes/github-webhook-trigger/mocks"
	sdkMocks "github.com/konstellation-io/kai-sdk/go-sdk/mocks"
	"github.com/konstellation-io/kai-sdk/go-sdk/sdk"
	"github.com/konstellation-io/kai-sdk/go-sdk/sdk/messaging"
	"github.com/nats-io/nats.go"
)

func TestInitializer(t *testing.T) {
	t.Run("should init webhook", func(t *testing.T) {
		rawEvents := "push,pull_request,release,workflow_run,workflow_dispatch"
		cleanedEvents := []string{"push", "pull_request", "release", "workflow_run", "workflow_dispatch"}
		githubSecret := "mockedSecret"

		centralizedConfigMock := sdkMocks.NewCentralizedConfigMock(t)
		sdkMock := sdk.KaiSDK{
			CentralizedConfig: centralizedConfigMock,
		}
		centralizedConfigMock.On("GetConfig", "webhook_events", messaging.ProcessScope).Return(rawEvents, nil)
		centralizedConfigMock.On("GetConfig", "github_secret", messaging.ProcessScope).Return(githubSecret, nil)

		webhookMock := mocks.NewWebhookMock(t)
		webhookMock.On("InitWebhook", cleanedEvents, githubSecret, sdkMock)

		initFunc := initializer(webhookMock)
		initFunc(sdkMock)

		webhookMock.AssertExpectations(t)
	})
}

// Preguntar mañana si usamos el monkey patch o no
func TestInitializerNoEventsConfiguredError(t *testing.T) {
	t.Run("should exit if no events configured", func(t *testing.T) {
		centralizedConfigMock := sdkMocks.NewCentralizedConfigMock(t)
		sdkMock := sdk.KaiSDK{
			CentralizedConfig: centralizedConfigMock,
		}
		centralizedConfigMock.On("GetConfig", "webhook_events").Return("", nats.ErrKeyNotFound)

		webhookMock := mocks.NewWebhookMock(t)

		initFunc := initializer(webhookMock)
		initFunc(sdkMock)

		webhookMock.AssertExpectations(t)
	})
}

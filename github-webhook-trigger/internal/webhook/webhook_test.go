package webhook_test

import (
	"testing"

	"github.com/konstellation-io/kai-processes/github-webhook-trigger/internal/webhook"
	sdkMocks "github.com/konstellation-io/kai-sdk/go-sdk/mocks"
	"github.com/konstellation-io/kai-sdk/go-sdk/sdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestNewGithubWebhook(t *testing.T) {
	t.Run("should create a new GithubWebhook", func(t *testing.T) {
		webhook := webhook.NewGithubWebhook()

		assert.NotNil(t, webhook)
	})
}

func TestInitWebhook(t *testing.T) {
	t.Run("should init webhook", func(t *testing.T) {
		ghWebhook := webhook.NewGithubWebhook()

		eventConfig := "push, pull_request, release,workflow_run,workflow_dispatch"
		githubSecret := "mockedSecret"
		eventURL := "mockedUrl"
		m, err := structpb.NewValue(map[string]interface{}{
			"eventUrl": eventURL,
			"event":    webhook.PushEvent,
		})
		require.NoError(t, err)

		messagingMock := sdkMocks.NewMessagingMock(t)
		sdkMock := sdk.KaiSDK{
			Messaging: messagingMock,
		}
		messagingMock.On("SendOutputWithRequestID", m, mock.AnythingOfType("int")).Return(nil)

		ghWebhook.InitWebhook(eventConfig, githubSecret, sdkMock)

		messagingMock.AssertExpectations(t)

		assert.NotNil(t, ghWebhook)
	})
}

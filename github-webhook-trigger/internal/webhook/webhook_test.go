package webhook_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-logr/logr/testr"
	"github.com/go-playground/webhooks/v6/github"
	sdkMocks "github.com/konstellation-io/kai-sdk/go-sdk/mocks"
	"github.com/konstellation-io/kai-sdk/go-sdk/sdk"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/konstellation-io/kai-processes/github-webhook-trigger/internal/webhook"
)

type GithubWebhookSuite struct {
	suite.Suite

	githubWebhook *webhook.GithubWebhook
	kaiSdk        sdk.KaiSDK
	messaging     *sdkMocks.MessagingMock
}

func TestGithubWebhookSuite(t *testing.T) {
	suite.Run(t, new(GithubWebhookSuite))
}

func (s *GithubWebhookSuite) SetupSuite() {
	s.githubWebhook = webhook.NewTestGithubWebhook()

	s.messaging = sdkMocks.NewMessagingMock(s.T())

	s.kaiSdk = sdk.KaiSDK{
		Logger:    testr.New(s.T()),
		Messaging: s.messaging,
	}
}

func (s *GithubWebhookSuite) TestHandlerEventRequest_ExpectOk() {
	// Given
	payload, err := os.Open("../../testdata/push_event.json")
	s.Require().NoError(err)
	defer func() {
		_ = payload.Close()
	}()
	request := httptest.NewRequest(http.MethodPost, "/webhooks", payload)
	request.Header.Set("X-GitHub-Event", "push")
	responseWriter := httptest.NewRecorder()
	parser, err := github.New()
	s.Require().NoError(err)
	expectedResponse, err := structpb.NewValue(map[string]interface{}{
		"eventUrl": "https://github.com/binkkatal/sample_app",
		"event":    "push",
	})
	s.Require().NoError(err)

	s.messaging.On("SendOutputWithRequestID",
		expectedResponse,
		mock.AnythingOfType("string")).
		Return(nil)

	// WHEN
	handlerFunction := s.githubWebhook.HandleEventRequest(parser, []github.Event{github.PushEvent}, s.kaiSdk)
	handlerFunction(responseWriter, request)
}

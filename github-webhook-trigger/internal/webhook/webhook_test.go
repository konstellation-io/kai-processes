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

type test struct {
	name             string
	payloadPath      string
	expecetdEventURL string
	expectedEvent    string
	githubEvents     []github.Event
}

func (s *GithubWebhookSuite) TestHandlerEventRequest_ExpectOk() {
	// Given
	allTests := []test{
		{
			name:             "push event",
			payloadPath:      "../../testdata/push_event.json",
			expecetdEventURL: "https://github.com/binkkatal/sample_app",
			expectedEvent:    "push",
			githubEvents:     []github.Event{github.PushEvent},
		},
	}

	parser, err := github.New()
	s.Require().NoError(err)

	for _, tc := range allTests {
		s.T().Run(tc.name, func(t *testing.T) {
			payload, err := os.Open(tc.payloadPath)
			s.Require().NoError(err)
			defer func() {
				_ = payload.Close()
			}()
			request := httptest.NewRequest(http.MethodPost, "/webhooks", payload)
			request.Header.Set("X-GitHub-Event", "push")
			responseWriter := httptest.NewRecorder()

			expectedResponse, err := structpb.NewValue(map[string]interface{}{
				"eventUrl": tc.expecetdEventURL,
				"event":    tc.expectedEvent,
			})
			s.Require().NoError(err)

			s.messaging.On("SendOutputWithRequestID",
				expectedResponse,
				mock.AnythingOfType("string")).
				Return(nil)

			// WHEN
			handlerFunction := s.githubWebhook.HandleEventRequest(parser, tc.githubEvents, s.kaiSdk)
			handlerFunction(responseWriter, request)
		})
	}
}

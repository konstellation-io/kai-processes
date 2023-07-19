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
	expectedEventURL string
	expectedEvent    string
	githubEvents     []github.Event
	isIgnored        bool
}

func (s *GithubWebhookSuite) TestHandlerEventRequest() {
	// Given
	okTests := []test{
		{
			name:             "push event",
			payloadPath:      "../../testdata/push_event.json",
			expectedEventURL: "https://github.com/binkkatal/sample_app",
			expectedEvent:    "push",
			githubEvents:     []github.Event{github.PushEvent},
			isIgnored:        false,
		},
		{
			name:             "pull request event",
			payloadPath:      "../../testdata/pull_request.json",
			expectedEventURL: "https://api.github.com/repos/baxterthehacker/public-repo/pulls/1",
			expectedEvent:    "pull_request",
			githubEvents:     []github.Event{github.PullRequestEvent},
			isIgnored:        false,
		},
		{
			name:             "release event",
			payloadPath:      "../../testdata/release.json",
			expectedEventURL: "https://api.github.com/repos/baxterthehacker/public-repo",
			expectedEvent:    "release",
			githubEvents:     []github.Event{github.ReleaseEvent},
			isIgnored:        false,
		},
		{
			name:             "workflow dispatch event",
			payloadPath:      "../../testdata/workflow_dispatch.json",
			expectedEventURL: "https://api.github.com/repos/baxterthehacker/public-repo",
			expectedEvent:    "workflow_dispatch",
			githubEvents:     []github.Event{github.WorkflowDispatchEvent},
			isIgnored:        false,
		},
		{
			name:             "workflow run event",
			payloadPath:      "../../testdata/workflow_run.json",
			expectedEventURL: "https://api.github.com/repos/baxterthehacker/public-repo",
			expectedEvent:    "workflow_run",
			githubEvents:     []github.Event{github.WorkflowRunEvent},
			isIgnored:        false,
		},
	}

	IgnoredTests := []test{
		{
			name:             "unsupported event",
			payloadPath:      "../../testdata/delete.json",
			expectedEventURL: "https://api.github.com/repos/baxterthehacker/public-repo",
			expectedEvent:    "delete",
			githubEvents:     []github.Event{github.DeleteEvent},
			isIgnored:        true,
		},
	}

	allTests := append(okTests, IgnoredTests...)

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
			request.Header.Set("X-GitHub-Event", tc.expectedEvent)
			responseWriter := httptest.NewRecorder()

			expectedResponse, err := structpb.NewValue(map[string]interface{}{
				"eventUrl": tc.expectedEventURL,
				"event":    tc.expectedEvent,
			})
			s.Require().NoError(err)

			if tc.isIgnored {
				s.messaging.ExpectedCalls = nil
			} else {
				s.messaging.On("SendOutputWithRequestID",
					expectedResponse,
					mock.AnythingOfType("string")).
					Return(nil)
			}

			// When
			handlerFunction := s.githubWebhook.HandleEventRequest(parser, tc.githubEvents, s.kaiSdk)
			handlerFunction(responseWriter, request)
		})
	}
}

func (s *GithubWebhookSuite) TestGetEventsFromConfig_OK() {
	// Given
	expectedEvents := []github.Event{github.PushEvent, github.PullRequestEvent, github.ReleaseEvent}
	eventConfig := "push,pull_request,release"

	// When
	events, err := s.githubWebhook.GetEventsFromConfig(eventConfig)

	// Then
	s.Require().NoError(err)
	s.Require().Equal(expectedEvents, events)
}

func (s *GithubWebhookSuite) TestGetEventsFromConfig_Error() {
	// Given
	eventConfig := "push,pull_request, delete"

	// When
	_, err := s.githubWebhook.GetEventsFromConfig(eventConfig)

	// Then
	s.Require().Error(err)
}
//go:build unit

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
	"github.com/konstellation-io/kai-sdk/go-sdk/sdk/messaging"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/konstellation-io/kai-processes/github-webhook-trigger/internal/webhook"
)

type GithubWebhookSuite struct {
	suite.Suite

	githubWebhook         webhook.Webhook
	githubWebhookTest     *webhook.GithubWebhookTestExporter
	centralizedConfigMock *sdkMocks.CentralizedConfigMock
	messagingMock         *sdkMocks.MessagingMock
	kaiSdk                sdk.KaiSDK
}

func TestGithubWebhookSuite(t *testing.T) {
	suite.Run(t, new(GithubWebhookSuite))
}

func (s *GithubWebhookSuite) SetupSuite() {
	s.githubWebhook = webhook.NewGithubWebhook()
	s.githubWebhookTest = webhook.NewGithubWebhookTestExporter()

	s.messagingMock = sdkMocks.NewMessagingMock(s.T())
	s.centralizedConfigMock = sdkMocks.NewCentralizedConfigMock(s.T())

	s.kaiSdk = sdk.KaiSDK{
		Logger:            testr.New(s.T()),
		Messaging:         s.messagingMock,
		CentralizedConfig: s.centralizedConfigMock,
	}
}

func (s *GithubWebhookSuite) TearDownTest() {
	s.messagingMock.AssertExpectations(s.T())
	s.centralizedConfigMock.AssertExpectations(s.T())
	s.messagingMock.ExpectedCalls = nil
	s.centralizedConfigMock.ExpectedCalls = nil
}

func (s *GithubWebhookSuite) TestInitializer() {
	rawEvents := "push,pull_request,release,workflow_run,workflow_dispatch"
	githubSecret := "mockedSecret"

	s.centralizedConfigMock.On("GetConfig", "webhook_events", messaging.ProcessScope).Return(rawEvents, nil)
	s.centralizedConfigMock.On("GetConfig", "github_secret", messaging.ProcessScope).Return(githubSecret, nil)

	err := s.githubWebhook.InitWebhook(s.kaiSdk)
	s.Assert().NoError(err)
	// pendiente hacer el monkey patch
}

func (s *GithubWebhookSuite) TestInitializerNoEventConfigError() {
	s.centralizedConfigMock.On("GetConfig", "webhook_events", messaging.ProcessScope).Return("", nats.ErrKeyNotFound)

	err := s.githubWebhook.InitWebhook(s.kaiSdk)
	s.Assert().Error(err)
}

func (s *GithubWebhookSuite) TestInitializerNoGithubSecretError() {
	rawEvents := "push,pull_request,release,workflow_run,workflow_dispatch"

	s.centralizedConfigMock.On("GetConfig", "webhook_events", messaging.ProcessScope).Return(rawEvents, nil)
	s.centralizedConfigMock.On("GetConfig", "github_secret", messaging.ProcessScope).Return("", nats.ErrKeyNotFound)

	err := s.githubWebhook.InitWebhook(s.kaiSdk)
	s.Assert().Error(err)
}

func (s *GithubWebhookSuite) TestInitializerEventNotSupportedError() {
	rawEvents := "push, pull_request, release, workflow_run, workflow_dispatch, unsupported"
	githubSecret := "mockedSecret"

	s.centralizedConfigMock.On("GetConfig", "webhook_events", messaging.ProcessScope).Return(rawEvents, nil)
	s.centralizedConfigMock.On("GetConfig", "github_secret", messaging.ProcessScope).Return(githubSecret, nil)

	err := s.githubWebhook.InitWebhook(s.kaiSdk)
	s.Require().Error(err)
	s.Assert().ErrorIs(err, webhook.ErrNotAValidEvent)
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
	const baxterPublicRepoExample = "https://api.github.com/repos/baxterthehacker/public-repo"

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
			expectedEventURL: baxterPublicRepoExample,
			expectedEvent:    "release",
			githubEvents:     []github.Event{github.ReleaseEvent},
			isIgnored:        false,
		},
		{
			name:             "workflow dispatch event",
			payloadPath:      "../../testdata/workflow_dispatch.json",
			expectedEventURL: baxterPublicRepoExample,
			expectedEvent:    "workflow_dispatch",
			githubEvents:     []github.Event{github.WorkflowDispatchEvent},
			isIgnored:        false,
		},
		{
			name:             "workflow run event",
			payloadPath:      "../../testdata/workflow_run.json",
			expectedEventURL: baxterPublicRepoExample,
			expectedEvent:    "workflow_run",
			githubEvents:     []github.Event{github.WorkflowRunEvent},
			isIgnored:        false,
		},
	}

	IgnoredTests := []test{
		{
			name:             "unsupported event",
			payloadPath:      "../../testdata/delete.json",
			expectedEventURL: baxterPublicRepoExample,
			expectedEvent:    "delete",
			githubEvents:     []github.Event{github.DeleteEvent},
			isIgnored:        true,
		},
	}

	allTests := make([]test, 0)
	allTests = append(allTests, okTests...)
	allTests = append(allTests, IgnoredTests...)

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
				s.messagingMock.ExpectedCalls = nil
			} else {
				s.messagingMock.On("SendOutputWithRequestID",
					expectedResponse,
					mock.AnythingOfType("string")).
					Return(nil)
			}

			// When
			handlerFunction := s.githubWebhookTest.HandleEventRequest(parser, tc.githubEvents, s.kaiSdk)
			handlerFunction(responseWriter, request)
		})
	}
}

func (s *GithubWebhookSuite) TestGetEventsFromConfigOK() {
	// Given
	expectedEvents := []github.Event{
		github.PushEvent, github.PullRequestEvent, github.ReleaseEvent, github.WorkflowDispatchEvent, github.WorkflowRunEvent,
	}
	eventConfig := "push, pull_request, release, workflow_dispatch, workflow_run"

	// When
	events, err := s.githubWebhookTest.GetEventsFromConfig(eventConfig)
	s.Require().NoError(err)

	// Then
	for _, event := range expectedEvents {
		s.Assert().Contains(events, event)
	}
}

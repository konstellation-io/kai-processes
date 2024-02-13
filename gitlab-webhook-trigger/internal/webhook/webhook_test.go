//go:build unit

package webhook_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"syscall"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/go-logr/logr/testr"
	"github.com/go-playground/webhooks/v6/gitlab"
	sdkMocks "github.com/konstellation-io/kai-sdk/go-sdk/v2/mocks"
	"github.com/konstellation-io/kai-sdk/go-sdk/v2/sdk"
	centralizedConfiguration "github.com/konstellation-io/kai-sdk/go-sdk/v2/sdk/centralized-configuration"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/konstellation-io/kai-processes/gitlab-webhook-trigger/internal/webhook"
)

const _rawTestEvents = "push, merge_request, comment, tag"

type GitlabWebhookSuite struct {
	suite.Suite

	gitlabWebhook         webhook.Webhook
	gitlabWebhookTest     *webhook.GitlabWebhookTestExporter
	centralizedConfigMock *sdkMocks.CentralizedConfigMock
	messagingMock         *sdkMocks.MessagingMock
	kaiSdk                sdk.KaiSDK
}

func TestGitlabWebhookSuite(t *testing.T) {
	suite.Run(t, new(GitlabWebhookSuite))
}

func (s *GitlabWebhookSuite) SetupSuite() {
	s.gitlabWebhook = webhook.NewGitlabWebhook()
	s.gitlabWebhookTest = webhook.NewGitlabWebhookTestExporter()

	s.messagingMock = sdkMocks.NewMessagingMock(s.T())
	s.centralizedConfigMock = sdkMocks.NewCentralizedConfigMock(s.T())

	s.kaiSdk = sdk.KaiSDK{
		Logger:            testr.New(s.T()),
		Messaging:         s.messagingMock,
		CentralizedConfig: s.centralizedConfigMock,
	}
}

func (s *GitlabWebhookSuite) TearDownTest() {
	s.messagingMock.AssertExpectations(s.T())
	s.centralizedConfigMock.AssertExpectations(s.T())
	s.messagingMock.ExpectedCalls = nil
	s.centralizedConfigMock.ExpectedCalls = nil
	s.messagingMock.Calls = nil
}

func (s *GitlabWebhookSuite) TestInitializer() {
	gitlabSecret := "mockedSecret"

	s.centralizedConfigMock.On("GetConfig", "webhook_events", centralizedConfiguration.ProcessScope).Return(_rawTestEvents, nil)
	s.centralizedConfigMock.On("GetConfig", "gitlab_secret", centralizedConfiguration.ProcessScope).Return(gitlabSecret, nil)

	go func() {
		err := s.gitlabWebhook.InitWebhook(s.kaiSdk)
		s.Assert().NoError(err)
	}()

	time.Sleep(time.Second)

	syscall.SIGTERM.Signal()
}

func (s *GitlabWebhookSuite) TestInitializerNoEventConfigError() {
	s.centralizedConfigMock.On("GetConfig", "webhook_events", centralizedConfiguration.ProcessScope).Return("", nats.ErrKeyNotFound)

	err := s.gitlabWebhook.InitWebhook(s.kaiSdk)
	s.Assert().Error(err)
}

func (s *GitlabWebhookSuite) TestInitializerNoGitlabSecretError() {
	s.centralizedConfigMock.On("GetConfig", "webhook_events", centralizedConfiguration.ProcessScope).Return(_rawTestEvents, nil)
	s.centralizedConfigMock.On("GetConfig", "gitlab_secret", centralizedConfiguration.ProcessScope).Return("", nats.ErrKeyNotFound)

	err := s.gitlabWebhook.InitWebhook(s.kaiSdk)
	s.Assert().Error(err)
}

func (s *GitlabWebhookSuite) TestInitializerEventNotSupportedError() {
	rawEventsWrong := "push,pull_request,release,workflow_run,workflow_dispatch"
	gitlabSecret := "mockedSecret"

	s.centralizedConfigMock.On("GetConfig", "webhook_events", centralizedConfiguration.ProcessScope).Return(rawEventsWrong, nil)
	s.centralizedConfigMock.On("GetConfig", "gitlab_secret", centralizedConfiguration.ProcessScope).Return(gitlabSecret, nil)

	err := s.gitlabWebhook.InitWebhook(s.kaiSdk)
	s.Require().Error(err)
	s.Assert().ErrorIs(err, webhook.ErrNotAValidEvent)
}

type test struct {
	name             string
	payloadPath      string
	expectedEventURL string
	expectedEvent    string
	gitlabEvents     []gitlab.Event
	isIgnored        bool
}

func (s *GitlabWebhookSuite) TestHandlerEventRequest() {
	const baxterPublicRepoExample = "https://api.github.com/repos/baxterthehacker/public-repo"

	testUseCases := []test{
		{
			name:             "push event",
			payloadPath:      "../../testdata/push_event.json",
			expectedEventURL: "git@example.com:mike/diaspora.git",
			expectedEvent:    string(gitlab.PushEvents),
			gitlabEvents:     []gitlab.Event{gitlab.PushEvents},
			isIgnored:        false,
		},
		{
			name:             "merge request event",
			payloadPath:      "../../testdata/merge_request_event.json",
			expectedEventURL: "http://example.com/gitlabhq/gitlab-test.git",
			expectedEvent:    string(gitlab.MergeRequestEvents),
			gitlabEvents:     []gitlab.Event{gitlab.MergeRequestEvents},
			isIgnored:        false,
		},
		{
			name:             "comment event",
			payloadPath:      "../../testdata/comment_commit_event.json",
			expectedEventURL: "http://example.com/gitlab-org/gitlab-test.git",
			expectedEvent:    string(gitlab.CommentEvents),
			gitlabEvents:     []gitlab.Event{gitlab.CommentEvents},
			isIgnored:        false,
		},
		{
			name:             "tag event",
			payloadPath:      "../../testdata/tag_event.json",
			expectedEventURL: "ssh://git@example.com/jsmith/example.git",
			expectedEvent:    string(gitlab.TagEvents),
			gitlabEvents:     []gitlab.Event{gitlab.TagEvents},
			isIgnored:        false,
		},
		{
			name:             "unsupported event",
			payloadPath:      "../../testdata/build_event.json",
			expectedEventURL: "git@192.168.64.1:gitlab-org/gitlab-test.git",
			expectedEvent:    string(gitlab.BuildEvents),
			gitlabEvents:     []gitlab.Event{gitlab.BuildEvents},
			isIgnored:        true,
		},
	}

	allTests := make([]test, 0)
	allTests = append(allTests, testUseCases...)

	parser, err := gitlab.New()
	s.Require().NoError(err)

	for _, tc := range allTests {
		s.T().Run(tc.name, func(t *testing.T) {
			payload, err := os.Open(tc.payloadPath)
			s.Require().NoError(err)
			defer func() {
				_ = payload.Close()
			}()
			request := httptest.NewRequest(http.MethodPost, "/gitlab-webhook", payload)
			request.Header.Set("X-GitLab-Event", tc.expectedEvent)
			responseWriter := httptest.NewRecorder()

			expectedResponse, err := structpb.NewValue(map[string]interface{}{
				"eventUrl": tc.expectedEventURL,
				"event":    tc.expectedEvent,
			})
			s.Require().NoError(err)

			if !tc.isIgnored {
				s.messagingMock.On("SendOutputWithRequestID",
					expectedResponse,
					mock.AnythingOfType("string")).
					Return(nil).
					Times(1)
			}

			handlerFunction := s.gitlabWebhookTest.HandleEventRequest(parser, tc.gitlabEvents, s.kaiSdk)
			handlerFunction(responseWriter, request)
		})
	}
}

func (s *GitlabWebhookSuite) TestGetEventsFromConfigOK() {
	expectedEvents := []gitlab.Event{
		gitlab.PushEvents, gitlab.MergeRequestEvents, gitlab.CommentEvents, gitlab.TagEvents,
	}

	events, err := s.gitlabWebhookTest.GetEventsFromConfig(_rawTestEvents)
	s.Require().NoError(err)

	s.ElementsMatch(expectedEvents, events)
}

func (s *GitlabWebhookSuite) TestHandlerEventRequestParseError() {
	fakeParse := func(*http.Request, ...gitlab.Event) (interface{}, error) {
		return nil, fmt.Errorf("fake error")
	}

	parser, err := gitlab.New()
	s.Require().NoError(err)

	patch := monkey.Patch(parser.Parse, fakeParse)
	defer patch.Unpatch()

	handlerFunction := s.gitlabWebhookTest.HandleEventRequest(parser, nil, s.kaiSdk)
	handlerFunction(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/gitlab-webhook", nil))

	s.messagingMock.AssertNotCalled(s.T(), "SendOutputWithRequestID", mock.Anything, mock.Anything)
}

func (s *GitlabWebhookSuite) TestInitWebhookCreatingWebhookError() {
	fakeNew := func(...gitlab.Option) (*gitlab.Webhook, error) {
		return nil, fmt.Errorf("fake error")
	}

	gitlabSecret := "mockedSecret"

	s.centralizedConfigMock.On("GetConfig", "webhook_events", centralizedConfiguration.ProcessScope).Return(_rawTestEvents, nil)
	s.centralizedConfigMock.On("GetConfig", "gitlab_secret", centralizedConfiguration.ProcessScope).Return(gitlabSecret, nil)

	patch := monkey.Patch(gitlab.New, fakeNew)
	defer patch.Unpatch()

	err := s.gitlabWebhook.InitWebhook(s.kaiSdk)
	s.Assert().Error(err)
}

func (s *GitlabWebhookSuite) TestInitWebhookGetConfigError() {
	gitlabSecret := "mockedSecret"

	s.centralizedConfigMock.On("GetConfig", "webhook_events", centralizedConfiguration.ProcessScope).Return(_rawTestEvents, nil)
	s.centralizedConfigMock.On("GetConfig", "gitlab_secret", centralizedConfiguration.ProcessScope).Return(gitlabSecret, nats.ErrKeyNotFound)

	err := s.gitlabWebhook.InitWebhook(s.kaiSdk)
	s.Assert().Error(err)
}

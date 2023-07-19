//go:build unit

package main

import (
	"os"
	"testing"

	"bou.ke/monkey"
	"github.com/konstellation-io/kai-processes/github-webhook-trigger/internal/mocks"
	"github.com/konstellation-io/kai-processes/github-webhook-trigger/internal/webhook"
	sdkMocks "github.com/konstellation-io/kai-sdk/go-sdk/mocks"
	"github.com/konstellation-io/kai-sdk/go-sdk/sdk"
	"github.com/konstellation-io/kai-sdk/go-sdk/sdk/messaging"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/suite"
)

type MainSuite struct {
	suite.Suite

	kaiSdkMock            sdk.KaiSDK
	githubWebhookMock     *mocks.WebhookMock
	centralizedConfigMock *sdkMocks.CentralizedConfigMock
}

func TestMainSuite(t *testing.T) {
	suite.Run(t, new(MainSuite))
}

func (s *MainSuite) SetupSuite() {
	s.githubWebhookMock = mocks.NewWebhookMock(s.T())
	s.centralizedConfigMock = sdkMocks.NewCentralizedConfigMock(s.T())
	s.kaiSdkMock = sdk.KaiSDK{
		CentralizedConfig: s.centralizedConfigMock,
	}
}

func (s *MainSuite) TearDownTest() {
	s.githubWebhookMock.AssertExpectations(s.T())
	s.centralizedConfigMock.AssertExpectations(s.T())
	s.githubWebhookMock.ExpectedCalls = nil
	s.centralizedConfigMock.ExpectedCalls = nil
	webhookEvents = ""
	githubSecret = ""
}

func (s *MainSuite) TestInitializer() {
	rawEvents := "push,pull_request,release,workflow_run,workflow_dispatch"
	githubSecret := "mockedSecret"

	s.centralizedConfigMock.On("GetConfig", "webhook_events", messaging.ProcessScope).Return(rawEvents, nil)
	s.centralizedConfigMock.On("GetConfig", "github_secret", messaging.ProcessScope).Return(githubSecret, nil)

	initializer(s.kaiSdkMock)
}

func (s *MainSuite) TestInitializerNoConfigError() {
	githubSecret := "mockedSecret"

	s.centralizedConfigMock.On("GetConfig", "webhook_events", messaging.ProcessScope).Return("", nats.ErrKeyNotFound)
	s.centralizedConfigMock.On("GetConfig", "github_secret", messaging.ProcessScope).Return(githubSecret, nil)

	fakeExitCalled := 0
	fakeExit := func(int) {
		fakeExitCalled++
	}

	patch := monkey.Patch(os.Exit, fakeExit)
	defer patch.Unpatch()

	initializer(s.kaiSdkMock)
	s.Equal(1, fakeExitCalled)
}

func (s *MainSuite) TestInitializerNoGithubSecretError() {
	rawEvents := "push,pull_request,release,workflow_run,workflow_dispatch"

	s.centralizedConfigMock.On("GetConfig", "webhook_events", messaging.ProcessScope).Return(rawEvents, nil)
	s.centralizedConfigMock.On("GetConfig", "github_secret", messaging.ProcessScope).Return("", nats.ErrKeyNotFound)

	fakeExitCalled := 0
	fakeExit := func(int) {
		fakeExitCalled++
	}

	patch := monkey.Patch(os.Exit, fakeExit)
	defer patch.Unpatch()

	initializer(s.kaiSdkMock)

	s.Equal(1, fakeExitCalled)
}

func (s *MainSuite) TestInitializerEventNotSupportedError() {
	fakeExitCalled := 0
	fakeExit := func(int) {
		fakeExitCalled++
	}

	patch := monkey.Patch(os.Exit, fakeExit)
	defer patch.Unpatch()

	s.githubWebhookMock.On("InitWebhook", "", "", s.kaiSdkMock).Return(webhook.ErrEventNotSupported)
	runnerFunc(s.githubWebhookMock)(nil, s.kaiSdkMock)

	s.Equal(1, fakeExitCalled)
}

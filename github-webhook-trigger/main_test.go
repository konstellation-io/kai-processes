package main

import (
	"os"
	"testing"

	"bou.ke/monkey"
	"github.com/konstellation-io/kai-processes/github-webhook-trigger/mocks"
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
}

func (s *MainSuite) TestInitializer() {
	rawEvents := "push,pull_request,release,workflow_run,workflow_dispatch"
	githubSecret := "mockedSecret"

	s.centralizedConfigMock.On("GetConfig", "webhook_events", messaging.ProcessScope).Return(rawEvents, nil)
	s.centralizedConfigMock.On("GetConfig", "github_secret", messaging.ProcessScope).Return(githubSecret, nil)

	initializer(s.kaiSdkMock)
}

// Esperar a que David me pase lo que captura el exit status 1
func (s *MainSuite) TestInitializerNoConfigError() {
	s.centralizedConfigMock.On("GetConfig", "webhook_events", messaging.ProcessScope).Return("", nats.ErrKeyNotFound)
	s.centralizedConfigMock.On("GetConfig", "github_secret", messaging.ProcessScope).Return("", nats.ErrKeyNotFound)

	fakeExitCalled := 0
	fakeExit := func(int) {
		fakeExitCalled++
	}

	patch := monkey.Patch(os.Exit, fakeExit)
	defer patch.Unpatch()

	initializer(s.kaiSdkMock)
	s.Equal(2, fakeExitCalled)
}

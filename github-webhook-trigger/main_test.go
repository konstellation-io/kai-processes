//go:build unit

package main

import (
	"fmt"
	"os"
	"testing"

	"bou.ke/monkey"
	"github.com/konstellation-io/kai-processes/github-webhook-trigger/internal/mocks"
	"github.com/konstellation-io/kai-sdk/go-sdk/sdk"
	"github.com/stretchr/testify/suite"
)

type MainSuite struct {
	suite.Suite

	kaiSdkMock        sdk.KaiSDK
	githubWebhookMock *mocks.WebhookMock
}

func TestMainSuite(t *testing.T) {
	suite.Run(t, new(MainSuite))
}

func (s *MainSuite) SetupSuite() {
	s.githubWebhookMock = mocks.NewWebhookMock(s.T())
	s.kaiSdkMock = sdk.KaiSDK{}
}

func (s *MainSuite) TearDownTest() {
	s.githubWebhookMock.AssertExpectations(s.T())
	s.githubWebhookMock.ExpectedCalls = nil
}

func (s *MainSuite) TestInitializer() {
	s.githubWebhookMock.On("InitWebhook", s.kaiSdkMock).Return(nil)

	initializer(s.githubWebhookMock)(s.kaiSdkMock)
}

func (s *MainSuite) TestInitializerError() {
	s.githubWebhookMock.On("InitWebhook", s.kaiSdkMock).Return(fmt.Errorf("mocked error"))

	fakeExitCalled := 0
	fakeExit := func(int) {
		fakeExitCalled++
	}

	patch := monkey.Patch(os.Exit, fakeExit)
	defer patch.Unpatch()

	initializer(s.githubWebhookMock)(s.kaiSdkMock)
}

func (s *MainSuite) TestRunnerFunc() {
	runnerFunc(nil, s.kaiSdkMock)
}

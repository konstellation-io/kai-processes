//go:build unit

package main

import (
	"fmt"
	"os"
	"testing"

	"bou.ke/monkey"
	"github.com/konstellation-io/kai-processes/gitlab-webhook-trigger/internal/mocks"
	"github.com/konstellation-io/kai-sdk/go-sdk/v2/sdk"
	"github.com/stretchr/testify/suite"
)

type MainSuite struct {
	suite.Suite

	kaiSdkMock        sdk.KaiSDK
	gitlabWebhookMock *mocks.WebhookMock
}

func TestMainSuite(t *testing.T) {
	suite.Run(t, new(MainSuite))
}

func (s *MainSuite) SetupSuite() {
	s.gitlabWebhookMock = mocks.NewWebhookMock(s.T())
	s.kaiSdkMock = sdk.KaiSDK{}
}

func (s *MainSuite) TearDownTest() {
	s.gitlabWebhookMock.AssertExpectations(s.T())
	s.gitlabWebhookMock.ExpectedCalls = nil
}

func (s *MainSuite) TestInitializer() {
	initializer(s.kaiSdkMock)
}

func (s *MainSuite) TestRunnerFunc() {
	s.gitlabWebhookMock.On("InitWebhook", s.kaiSdkMock).Return(nil)

	runnerFunc(s.gitlabWebhookMock)(nil, s.kaiSdkMock)
}

func (s *MainSuite) TestRunnerFuncError() {
	s.gitlabWebhookMock.On("InitWebhook", s.kaiSdkMock).Return(fmt.Errorf("mocked error"))

	fakeExitCalled := 0
	fakeExit := func(int) {
		fakeExitCalled++
	}

	patch := monkey.Patch(os.Exit, fakeExit)
	defer patch.Unpatch()

	runnerFunc(s.gitlabWebhookMock)(nil, s.kaiSdkMock)
}

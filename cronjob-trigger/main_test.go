//go:build unit

package main

import (
	"fmt"
	"os"
	"testing"

	"bou.ke/monkey"
	"github.com/go-logr/logr/testr"
	sdkMocks "github.com/konstellation-io/kai-sdk/go-sdk/mocks"
	"github.com/konstellation-io/kai-sdk/go-sdk/sdk"
	"github.com/konstellation-io/kai-sdk/go-sdk/sdk/messaging"
	"github.com/stretchr/testify/suite"
)

type MainSuite struct {
	suite.Suite

	centralizedConfigMock *sdkMocks.CentralizedConfigMock
	messagingMock         *sdkMocks.MessagingMock
	kaiSdk                sdk.KaiSDK
}

func TestMainSuite(t *testing.T) {
	suite.Run(t, new(MainSuite))
}

func (s *MainSuite) SetupSuite() {
	s.messagingMock = sdkMocks.NewMessagingMock(s.T())
	s.centralizedConfigMock = sdkMocks.NewCentralizedConfigMock(s.T())

	s.kaiSdk = sdk.KaiSDK{
		Logger:            testr.New(s.T()),
		Messaging:         s.messagingMock,
		CentralizedConfig: s.centralizedConfigMock,
	}
}

func (s *MainSuite) TearDownTest() {
	s.messagingMock.AssertExpectations(s.T())
	s.centralizedConfigMock.AssertExpectations(s.T())
	s.messagingMock.ExpectedCalls = nil
	s.centralizedConfigMock.ExpectedCalls = nil
	s.messagingMock.Calls = nil
}

func (s *MainSuite) TestInitializer() {
	initializer(s.kaiSdk)
}

func (s *MainSuite) TestCronjobRunnerFunc() {
	s.centralizedConfigMock.On("GetConfig", "cron").Return("30 * * * * *", nil)
	s.centralizedConfigMock.On("GetConfig", "message").Return("test message", nil)

	cronjobRunner(nil, s.kaiSdk)
}

func (s *MainSuite) TestCronjobRunnerFunc_Error() {
	s.centralizedConfigMock.On("GetConfig", "cron").Return("30 * * * * *", nil)
	s.centralizedConfigMock.On("GetConfig", "message").Return("", fmt.Errorf("mocked error"))

	fakeExitCalled := 0
	fakeExit := func(int) {
		fakeExitCalled++
	}

	patch := monkey.Patch(os.Exit, fakeExit)
	defer patch.Unpatch()

	cronjobRunner(nil, s.kaiSdk)
}

//go:build unit

package main

import (
	"testing"
	"time"

	"github.com/go-logr/logr/testr"
	sdkMocks "github.com/konstellation-io/kai-sdk/go-sdk/v2/mocks"
	"github.com/konstellation-io/kai-sdk/go-sdk/v2/sdk"
	centralizedConfiguration "github.com/konstellation-io/kai-sdk/go-sdk/v2/sdk/centralized-configuration"
	"github.com/stretchr/testify/mock"
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
	s.centralizedConfigMock.On("GetConfig", "cron", centralizedConfiguration.ProcessScope).Return("@every 1s", nil)
	s.centralizedConfigMock.On("GetConfig", "message", centralizedConfiguration.ProcessScope).Return("test message", nil)
	s.messagingMock.On("SendOutputWithRequestID", mock.Anything, mock.Anything).Return(nil)

	go func() {
		cronjobRunner(nil, s.kaiSdk)
	}()

	time.Sleep(1 * time.Second) // wait for cronjob to run
}

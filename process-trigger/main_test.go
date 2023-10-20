//go:build unit

package main

import (
	"testing"

	"github.com/go-logr/logr/testr"
	sdkMocks "github.com/konstellation-io/kai-sdk/go-sdk/mocks"
	"github.com/konstellation-io/kai-sdk/go-sdk/sdk"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type MainSuite struct {
	suite.Suite

	centralizedConfigMock *sdkMocks.CentralizedConfigMock
	metadataMock          *sdkMocks.MetadataMock
	messagingMock         *sdkMocks.MessagingMock
	kaiSdk                sdk.KaiSDK
}

func TestMainSuite(t *testing.T) {
	suite.Run(t, new(MainSuite))
}

func (s *MainSuite) SetupSuite() {
	s.messagingMock = sdkMocks.NewMessagingMock(s.T())
	s.metadataMock = sdkMocks.NewMetadataMock(s.T())
	s.centralizedConfigMock = sdkMocks.NewCentralizedConfigMock(s.T())

	s.kaiSdk = sdk.KaiSDK{
		Logger:            testr.New(s.T()),
		Messaging:         s.messagingMock,
		Metadata:          s.metadataMock,
		CentralizedConfig: s.centralizedConfigMock,
	}
}

func (s *MainSuite) TearDownTest() {
	s.messagingMock.AssertExpectations(s.T())
	s.metadataMock.AssertExpectations(s.T())
	s.centralizedConfigMock.AssertExpectations(s.T())
	s.messagingMock.ExpectedCalls = nil
	s.metadataMock.ExpectedCalls = nil
	s.centralizedConfigMock.ExpectedCalls = nil
	s.messagingMock.Calls = nil
}

func (s *MainSuite) TestInitializer() {
	initializer(s.kaiSdk)
}

func (s *MainSuite) TestProcessRunnerFunc() {
	s.centralizedConfigMock.On("GetConfig", "product").Return("product", nil)
	s.centralizedConfigMock.On("GetConfig", "version").Return("version", nil)
	s.centralizedConfigMock.On("GetConfig", "workflow").Return("workflow", nil)
	s.centralizedConfigMock.On("GetConfig", "process").Return("process", nil)
	s.metadataMock.On("GetProduct").Return("productID", nil)
	s.metadataMock.On("GetVersion").Return("versionID", nil)
	s.metadataMock.On("GetWorkflow").Return("workflowID", nil)
	s.metadataMock.On("GetProcess").Return("processID", nil)
	s.centralizedConfigMock.On("GetConfig", "retain-execution-id").Return("true", nil)
	s.messagingMock.On("SendOutputWithRequestID", mock.Anything, mock.Anything).Return(nil)

	go func() {
		processSubscriberRunner(nil, s.kaiSdk)
	}()
}

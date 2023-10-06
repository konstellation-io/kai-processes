//go:build unit

package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/konstellation-io/kai-sdk/go-sdk/sdk"
	"github.com/stretchr/testify/suite"
)

type MainSuite struct {
	suite.Suite

	kaiSdkMock        sdk.KaiSDK
	cronjobMock *mocks.CronjobMock
}

func TestMainSuite(t *testing.T) {
	suite.Run(t, new(MainSuite))
}

func (s *MainSuite) SetupSuite() {
	s.cronjobMock = mocks.NewCronjobMock(s.T())
	s.kaiSdkMock = sdk.KaiSDK{}
}

func (s *MainSuite) TearDownTest() {
	s.cronjobMock.AssertExpectations(s.T())
	s.cronjobMock.ExpectedCalls = nil
}

func (s *MainSuite) TestInitializer() {
	initializer(s.cronjobMock)
}

func (s *MainSuite) TestRunnerFunc() {
	s.cronjobMock.On("CronjobMock", s.cronjobMock).Return(nil)

	runnerFunc(s.cronjobMock)(nil, s.cronjobMock)
}

func (s *MainSuite) TestRunnerFuncError() {
	s.cronjobMock.On("CronjobMock", s.cronjobMock).Return(fmt.Errorf("mocked error"))

	fakeExitCalled := 0
	fakeExit := func(int) {
		fakeExitCalled++
	}

	patch := monkey.Patch(os.Exit, fakeExit)
	defer patch.Unpatch()

	runnerFunc(s.cronjobMock)(nil, s.cronjobMock)
}

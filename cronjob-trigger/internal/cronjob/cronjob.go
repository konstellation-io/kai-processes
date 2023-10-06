package cronjob

import (
	"github.com/konstellation-io/kai-sdk/go-sdk/sdk"
)

const (
	path = "/cronjob"
)

//go:generate mockery --name Cronjob --output ../mocks --filename cronjob_mock.go --structname CronjobMock
type Cronjob interface {
	InitCronjob(kaiSDK sdk.KaiSDK) error
}

type CronjobImpl struct {
}

func NewCronjob() Cronjob {
	return &CronjobImpl{}
}

func (cr *CronjobImpl) InitCronjob(kaiSDK sdk.KaiSDK) error {
	return nil
}

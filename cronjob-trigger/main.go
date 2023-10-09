package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/google/uuid"
	"github.com/konstellation-io/kai-sdk/go-sdk/runner"
	"github.com/konstellation-io/kai-sdk/go-sdk/runner/trigger"
	"github.com/konstellation-io/kai-sdk/go-sdk/sdk"
	"github.com/robfig/cron/v3"
)

func initializer(kaiSDK sdk.KaiSDK) {
	kaiSDK.Logger.Info("Initializing cronjob")
}

func cronjobRunner(tr *trigger.Runner, sdk sdk.KaiSDK) {
	sdk.Logger.Info("Starting cronjob runner")
	time, _ := sdk.CentralizedConfig.GetConfig("cron")
	message, _ := sdk.CentralizedConfig.GetConfig("message")

	c := cron.New(
		cron.WithLogger(sdk.Logger),
		cron.WithSeconds(),
	)

	_, err := c.AddFunc(time, func() {
		requestID := uuid.New().String()
		sdk.Logger.Info("Cronjob triggered, new message sent", "requestID", requestID)

		val := wrappers.StringValue{
			Value: message,
		}

		err := sdk.Messaging.SendOutputWithRequestID(&val, requestID)
		if err != nil {
			sdk.Logger.Error(err, "Error sending output")
			return
		}
	})
	if err != nil {
		sdk.Logger.Error(err, "Error adding cronjob")
		return
	}

	c.Start()

	// Handle sigterm and await termChan signal
	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)
	<-termChan

	c.Stop()
}

func main() {
	r := runner.NewRunner().
		TriggerRunner().
		WithInitializer(initializer).
		WithRunner(cronjobRunner)

	r.Run()
}

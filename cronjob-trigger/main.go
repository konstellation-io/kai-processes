package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/konstellation-io/kai-sdk/go-sdk/v2/runner"
	"github.com/konstellation-io/kai-sdk/go-sdk/v2/runner/trigger"
	"github.com/konstellation-io/kai-sdk/go-sdk/v2/sdk"
	centralizedConfiguration "github.com/konstellation-io/kai-sdk/go-sdk/v2/sdk/centralized-configuration"
	"github.com/robfig/cron/v3"
	"google.golang.org/protobuf/types/known/structpb"
)

func initializer(kaiSDK sdk.KaiSDK) {
	kaiSDK.Logger.Info("Initializing cronjob")
}

func cronjobRunner(tr *trigger.Runner, kaiSDK sdk.KaiSDK) {
	kaiSDK.Logger.Info("Starting cronjob runner")
	cronTime, _ := kaiSDK.CentralizedConfig.GetConfig("cron", centralizedConfiguration.ProcessScope)
	message, _ := kaiSDK.CentralizedConfig.GetConfig("message", centralizedConfiguration.ProcessScope)

	c := cron.New(
		cron.WithLogger(kaiSDK.Logger),
		cron.WithSeconds(),
	)

	_, err := c.AddFunc(cronTime, func() {
		requestID := uuid.New().String()
		kaiSDK.Logger.Info("Cronjob triggered, new message sent", "requestID", requestID)

		m, err := structpb.NewValue(map[string]interface{}{
			"message": message,
			"time":    time.Now().Format("Mon Jan 2 15:04:05 MST 2006"),
		})
		if err != nil {
			kaiSDK.Logger.Error(err, "error creating response")
			os.Exit(1)
		}

		err = kaiSDK.Messaging.SendOutputWithRequestID(m, requestID)
		if err != nil {
			kaiSDK.Logger.Error(err, "Error sending output")
			os.Exit(1)
		}
	})
	if err != nil {
		kaiSDK.Logger.Error(err, "Error adding cronjob")
		os.Exit(1)
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

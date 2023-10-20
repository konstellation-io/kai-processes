package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/google/uuid"
	"github.com/konstellation-io/kai-sdk/go-sdk/runner"
	"github.com/konstellation-io/kai-sdk/go-sdk/runner/trigger"
	"github.com/konstellation-io/kai-sdk/go-sdk/sdk"
	"github.com/nats-io/nats.go"
)

func initializer(kaiSDK sdk.KaiSDK) {
	kaiSDK.Logger.Info("Initializing process")
}

func processRunner(tr *trigger.Runner, kaiSDK sdk.KaiSDK) {
	kaiSDK.Logger.Info("Starting nats subscriber")
	process, _ := kaiSDK.CentralizedConfig.GetConfig("process")
	retainExecutionId, _ := kaiSDK.CentralizedConfig.GetConfig("retain-execution-id")

	nc, _ := nats.Connect("nats://localhost:4222")
	js, err := nc.JetStream()
	if err != nil {
		panic(err)
	}

	s, err := js.QueueSubscribe(
		process,
		"nats-trigger-queue",
		func(msg *nats.Msg) {
			val := &wrappers.StringValue{
				Value: "Hi there, I'm a NATS subscriber!",
			}

			requestID := uuid.New().String()
			if retainExecutionId == "true" {
				requestID = kaiSDK.GetRequestID()
			}
			responseChannel := tr.GetResponseChannel(requestID)

			err = kaiSDK.Messaging.SendOutputWithRequestID(val, requestID)
			if err != nil {
				kaiSDK.Logger.Error(err, "Error sending output")
				return
			}

			// Wait for the response before ACKing the message
			<-responseChannel

			kaiSDK.Logger.Info("Message received, acking message")

			err = msg.Ack()
			if err != nil {
				kaiSDK.Logger.Error(err, "Error acking message")
				return
			}
		},
		nats.DeliverNew(),
		nats.Durable("nats-trigger"),
		nats.ManualAck(),
		nats.AckWait(22*time.Hour),
	)

	// Handle sigterm and await termChan signal
	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)
	<-termChan

	err = s.Unsubscribe()
	if err != nil {
		kaiSDK.Logger.Error(err, "Error unsubscribing")
		return
	}
}

func main() {
	r := runner.NewRunner().
		TriggerRunner().
		WithInitializer(initializer).
		WithRunner(processRunner)

	r.Run()
}

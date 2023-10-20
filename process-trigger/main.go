package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
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
	kaiSDK.Logger.Info("Initializing process trigger")
}

func processSubscriberRunner(tr *trigger.Runner, kaiSDK sdk.KaiSDK) {
	kaiSDK.Logger.Info("Starting process subscriber runner")

	product, _ := kaiSDK.CentralizedConfig.GetConfig("product")
	version, _ := kaiSDK.CentralizedConfig.GetConfig("version")
	workflow, _ := kaiSDK.CentralizedConfig.GetConfig("workflow")
	process, _ := kaiSDK.CentralizedConfig.GetConfig("process")
	productID := kaiSDK.Metadata.GetProduct()
	productID = strings.ReplaceAll(strings.ToLower(productID), " ", "_")
	versionID := kaiSDK.Metadata.GetVersion()
	versionID = strings.ReplaceAll(strings.ToLower(versionID), " ", "_")
	workflowID := kaiSDK.Metadata.GetWorkflow()
	workflowID = strings.ReplaceAll(strings.ToLower(workflowID), " ", "_")
	processID := kaiSDK.Metadata.GetProcess()
	processID = strings.ReplaceAll(strings.ToLower(processID), " ", "_")
	streamName := fmt.Sprintf("%s_%s_%s.%s", product, version, workflow, process)
	consumerName := fmt.Sprintf("%s_%s_%s_%s", productID, versionID, workflowID, processID)
	retainExecutionId, _ := kaiSDK.CentralizedConfig.GetConfig("retain-execution-id")

	nc, _ := nats.Connect("nats://localhost:4222")
	js, err := nc.JetStream()
	if err != nil {
		panic(err)
	}

	kaiSDK.Logger.Info("Subscribing to stream", "stream", streamName, "consumer", consumerName)
	s, err := js.QueueSubscribe(
		streamName,
		consumerName,
		func(msg *nats.Msg) {
			requestID := uuid.New().String()

			if retainExecutionId == "true" {
				requestID = kaiSDK.GetRequestID()
			}

			responseChannel := tr.GetResponseChannel(requestID)

			val := &wrappers.StringValue{
				Value: string(msg.Data),
			}

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
		nats.Durable(consumerName),
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
	runner.
		NewRunner().
		TriggerRunner().
		WithInitializer(initializer).
		WithRunner(processSubscriberRunner).
		Run()
}

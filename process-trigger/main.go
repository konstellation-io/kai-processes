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
	"github.com/konstellation-io/kai-sdk/go-sdk/v2/runner"
	"github.com/konstellation-io/kai-sdk/go-sdk/v2/runner/trigger"
	"github.com/konstellation-io/kai-sdk/go-sdk/v2/sdk"
	"github.com/nats-io/nats.go"
	"github.com/spf13/viper"
)

const AckWait = 22 * time.Hour

func initializer(kaiSDK sdk.KaiSDK) {
	kaiSDK.Logger.Info("Initializing process trigger")
}

func processSubscriberRunner(tr *trigger.Runner, kaiSDK sdk.KaiSDK) {
	kaiSDK.Logger.Info("Starting process subscriber")

	targetProduct, _ := kaiSDK.CentralizedConfig.GetConfig("product")
	targetVersion, _ := kaiSDK.CentralizedConfig.GetConfig("version")
	targetWorkflow, _ := kaiSDK.CentralizedConfig.GetConfig("workflow")
	targetProcess, _ := kaiSDK.CentralizedConfig.GetConfig("process")

	subjectName := fmt.Sprintf("%s_%s_%s_%s", targetProduct, targetVersion, targetWorkflow, targetProcess)

	productID := kaiSDK.Metadata.GetProduct()
	productID = strings.ReplaceAll(strings.ToLower(productID), " ", "_")
	productID = strings.ReplaceAll(productID, ".", "_")
	versionID := kaiSDK.Metadata.GetVersion()
	versionID = strings.ReplaceAll(strings.ToLower(versionID), " ", "_")
	versionID = strings.ReplaceAll(versionID, ".", "_")
	workflowID := kaiSDK.Metadata.GetWorkflow()
	workflowID = strings.ReplaceAll(strings.ToLower(workflowID), " ", "_")
	workflowID = strings.ReplaceAll(workflowID, ".", "_")
	processID := kaiSDK.Metadata.GetProcess()
	processID = strings.ReplaceAll(strings.ToLower(processID), " ", "_")
	processID = strings.ReplaceAll(processID, ".", "_")

	queueName := fmt.Sprintf("%s_%s_%s_%s", productID, versionID, workflowID, processID)

	retainExecutionID, _ := kaiSDK.CentralizedConfig.GetConfig("retain-execution-id")

	nc, _ := nats.Connect(viper.GetString("nats.url"))

	js, err := nc.JetStream()
	if err != nil {
		panic(err)
	}

	s, err := js.QueueSubscribe(
		subjectName,
		queueName,
		func(msg *nats.Msg) {
			kaiSDK.Logger.Info("Message received", "subject", subjectName, "queue", queueName)
			requestID, err := kaiSDK.Messaging.GetRequestID(msg)
			if err != nil {
				kaiSDK.Logger.Error(err, "Error creating request message")
				return
			}

			if retainExecutionID == "true" {
				requestID = uuid.New().String()
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
		nats.Durable(queueName),
		nats.ManualAck(),
		nats.AckWait(AckWait),
	)

	if err != nil {
		kaiSDK.Logger.Error(err, "Error subscribing")
		return
	}

	kaiSDK.Logger.Info("Listening to subject", "subject", subjectName, "queue", queueName)

	// Handle sigterm and await termChan signal
	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)
	<-termChan

	err = s.Unsubscribe()
	if err != nil {
		kaiSDK.Logger.Error(err, "Error unsubscribing")
	}
}

func main() {
	runner.
		NewRunner().
		TriggerRunner().
		WithInitializer(initializer).
		WithRunner(processSubscriberRunner).
		WithFinalizer(func(kaiSDK sdk.KaiSDK) {
			kaiSDK.Logger.Info("Finishing process trigger")
		}).
		Run()
}

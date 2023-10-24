package main

import (
	"context"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/google/uuid"
	"github.com/konstellation-io/kai-sdk/go-sdk/runner"
	"github.com/konstellation-io/kai-sdk/go-sdk/runner/trigger"
	"github.com/konstellation-io/kai-sdk/go-sdk/sdk"
	"github.com/konstellation-io/kai-sdk/go-sdk/sdk/messaging"
	"github.com/segmentio/kafka-go"
	"google.golang.org/protobuf/types/known/structpb"
)

var config struct {
	Brokers   []string
	GroupID   string
	Topic     string
	Partition int
}

func main() {
	runner.
		NewRunner().
		TriggerRunner().
		WithInitializer(initializer).
		WithRunner(kafkaRunner).
		WithFinalizer(func(kaiSDK sdk.KaiSDK) {
			kaiSDK.Logger.Info("Finalizer")
		}).
		Run()
}

func initializer(kaiSDK sdk.KaiSDK) {
	kaiSDK.Logger.Info("Initializer")

	var errMsg = "error getting config"
	brokers, err := kaiSDK.CentralizedConfig.GetConfig("brokers", messaging.ProcessScope)
	if err != nil {
		kaiSDK.Logger.Error(err, errMsg)
		os.Exit(1)
	}
	config.Brokers = strings.Split(brokers, ",")

	groupID, err := kaiSDK.CentralizedConfig.GetConfig("groupID", messaging.ProcessScope)
	if err != nil {
		kaiSDK.Logger.Error(err, errMsg)
		os.Exit(1)
	}
	config.GroupID = groupID

	topic, err := kaiSDK.CentralizedConfig.GetConfig("topic", messaging.ProcessScope)
	if err != nil {
		kaiSDK.Logger.Error(err, errMsg)
		os.Exit(1)
	}
	config.Topic = topic

	partitionConfig, err := kaiSDK.CentralizedConfig.GetConfig("partition", messaging.ProcessScope)
	if err != nil {
		kaiSDK.Logger.Error(err, errMsg)
		os.Exit(1)
	}
	partition, err := strconv.Atoi(partitionConfig)
	if err != nil {
		kaiSDK.Logger.Error(err, errMsg)
		os.Exit(1)
	}
	config.Partition = partition

	kaiSDK.Logger.Info("Config loaded",
		"brokers", config.Brokers,
		"groupID", config.GroupID,
		"topic", config.Topic,
		"partition", config.Partition,
	)
}

func kafkaRunner(tr *trigger.Runner, kaiSDK sdk.KaiSDK) {
	kaiSDK.Logger.Info("Starting kafka runner")

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  config.Brokers,
		GroupID:  config.GroupID,
		Topic:    config.Topic,
		MaxBytes: 10e6, // 10MB
	})

	go func() {
		for {
			m, err := r.ReadMessage(context.Background())
			if err != nil {
				kaiSDK.Logger.Error(err, "error reading message")
				break
			}

			kaiSDK.Logger.Info("Incoming message",
				"topic", m.Topic,
				"partition", m.Partition,
				"offset", m.Offset,
				"key", string(m.Key),
				"value", string(m.Value),
			)

			err = sendMessage(kaiSDK, config.Topic, m.Value)
			if err != nil {
				kaiSDK.Logger.Error(err, "error sending message")
				break
			}
		}

		err := r.Close() // reader must be gracefully closed
		if err != nil {
			kaiSDK.Logger.Error(err, "error closing reader")
			os.Exit(1)
		}
	}()

	// Handle sigterm and await termChan signal
	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)
	<-termChan

	err := r.Close() // reader must be gracefully closed
	if err != nil {
		kaiSDK.Logger.Error(err, "error closing reader")
		os.Exit(1)
	}
}

func sendMessage(kaiSDK sdk.KaiSDK, topic string, message []byte) error {
	requestID := uuid.New().String()
	kaiSDK.Logger.Info("Triggering workflow", "requestID", requestID)

	m, err := structpb.NewValue(map[string]interface{}{
		"topic":   topic,
		"message": message,
	})
	if err != nil {
		return err
	}

	err = kaiSDK.Messaging.SendOutputWithRequestID(m, requestID)
	if err != nil {
		return err
	}

	return nil
}

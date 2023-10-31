package main

import (
	"context"
	"crypto/tls"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/konstellation-io/kai-sdk/go-sdk/runner"
	"github.com/konstellation-io/kai-sdk/go-sdk/runner/trigger"
	"github.com/konstellation-io/kai-sdk/go-sdk/sdk"
	"github.com/konstellation-io/kai-sdk/go-sdk/sdk/messaging"
	"github.com/segmentio/kafka-go"
	"google.golang.org/protobuf/types/known/structpb"
)

var config struct {
	Brokers            []string
	GroupID            string
	Topic              string
	TLSEnabled         bool
	InsecureSkipVerify bool
}

func main() {
	runner.
		NewRunner().
		TriggerRunner().
		WithInitializer(initializer).
		WithRunner(kafkaRunner).
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

	groupID, err := kaiSDK.CentralizedConfig.GetConfig("groupid", messaging.ProcessScope)
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

	tlsEnabledConfig, err := kaiSDK.CentralizedConfig.GetConfig("tls_enabled", messaging.ProcessScope)
	if err == nil { // optional config
		tlsEnabled, err := strconv.ParseBool(tlsEnabledConfig)
		if err != nil {
			kaiSDK.Logger.Error(err, errMsg)
			os.Exit(1)
		}
		config.TLSEnabled = tlsEnabled
	}

	insecureSkipVerifyConfig, err := kaiSDK.CentralizedConfig.GetConfig("insecure_skip_verify", messaging.ProcessScope)
	if err == nil { // optional config
		insecureSkipVerify, err := strconv.ParseBool(insecureSkipVerifyConfig)
		if err != nil {
			kaiSDK.Logger.Error(err, errMsg)
			os.Exit(1)
		}
		config.InsecureSkipVerify = insecureSkipVerify
	}

	kaiSDK.Logger.Info("Config loaded",
		"brokers", config.Brokers,
		"groupID", config.GroupID,
		"topic", config.Topic,
		"tlsEnabled", config.TLSEnabled,
		"insecureSkipVerify", config.InsecureSkipVerify,
	)
}

func kafkaRunner(tr *trigger.Runner, kaiSDK sdk.KaiSDK) {
	var err error

	var dialer *kafka.Dialer
	if config.TLSEnabled {
		dialer = &kafka.Dialer{
			Timeout:   10 * time.Second,
			DualStack: true,
			TLS: &tls.Config{
				InsecureSkipVerify: config.InsecureSkipVerify,
			},
		}
	}

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  config.Brokers,
		GroupID:  config.GroupID,
		Topic:    config.Topic,
		MaxBytes: 10e6, // 10MB
		Dialer:   dialer,
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
		}
		os.Exit(1)
	}()

	// Handle sigterm and await termChan signal
	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)
	<-termChan

	err = r.Close() // reader must be gracefully closed
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

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
	"github.com/konstellation-io/kai-sdk/go-sdk/v2/runner"
	"github.com/konstellation-io/kai-sdk/go-sdk/v2/runner/trigger"
	"github.com/konstellation-io/kai-sdk/go-sdk/v2/sdk"
	centralizedConfiguration "github.com/konstellation-io/kai-sdk/go-sdk/v2/sdk/centralized-configuration"
	"github.com/segmentio/kafka-go"
	"google.golang.org/protobuf/types/known/structpb"
)

const dialerTimeout = 10 * time.Second

type kafkaConfig struct {
	Brokers            []string
	GroupID            string
	Topic              string
	TLSEnabled         bool
	InsecureSkipVerify bool
}

var config kafkaConfig //nolint: gochecknoglobals

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

	brokers, err := kaiSDK.CentralizedConfig.GetConfig("brokers", centralizedConfiguration.ProcessScope)
	if err != nil {
		kaiSDK.Logger.Error(err, errMsg)
		os.Exit(1)
	}

	config.Brokers = strings.Split(brokers, ",")

	groupID, err := kaiSDK.CentralizedConfig.GetConfig("groupid", centralizedConfiguration.ProcessScope)
	if err != nil {
		kaiSDK.Logger.Error(err, errMsg)
		os.Exit(1)
	}

	config.GroupID = groupID

	topic, err := kaiSDK.CentralizedConfig.GetConfig("topic", centralizedConfiguration.ProcessScope)
	if err != nil {
		kaiSDK.Logger.Error(err, errMsg)
		os.Exit(1)
	}

	config.Topic = topic

	tlsEnabledConfig, err := kaiSDK.CentralizedConfig.GetConfig("tls_enabled", centralizedConfiguration.ProcessScope)
	if err == nil { // optional config
		tlsEnabled, err := strconv.ParseBool(tlsEnabledConfig)
		if err != nil {
			kaiSDK.Logger.Error(err, errMsg)
			os.Exit(1)
		}

		config.TLSEnabled = tlsEnabled
	}

	insecureSkipVerifyConfig, err := kaiSDK.CentralizedConfig.GetConfig("skip_tls_verify", centralizedConfiguration.ProcessScope)
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
	var dialer *kafka.Dialer
	if config.TLSEnabled {
		dialer = &kafka.Dialer{
			Timeout:   dialerTimeout,
			DualStack: true,
			TLS: &tls.Config{
				InsecureSkipVerify: config.InsecureSkipVerify, //nolint: gosec
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
	defer r.Close()

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

		os.Exit(1)
	}()

	// Handle sigterm and await termChan signal
	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)
	<-termChan
}

func sendMessage(kaiSDK sdk.KaiSDK, topic string, payload []byte) error {
	requestID := uuid.New().String()
	kaiSDK.Logger.Info("Triggering workflow", "requestID", requestID)

	m, err := structpb.NewValue(map[string]interface{}{
		"topic":   topic,
		"payload": payload,
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

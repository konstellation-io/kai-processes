//go:build integration

package main

import (
	"context"
	"fmt"
	"github.com/go-logr/logr/testr"
	"github.com/konstellation-io/kai-sdk/go-sdk/mocks"
	"github.com/konstellation-io/kai-sdk/go-sdk/sdk"
	kafkago "github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/kafka"
	"google.golang.org/protobuf/reflect/protoreflect"
	"log"
	"sync"
	"testing"
)

var (
	productID = "productID"
	ownerID   = "ownerID"
)

type MainSuite struct {
	suite.Suite
	kaiSdkMock     sdk.KaiSDK
	messagingMock  *mocks.MessagingMock
	kafkaContainer *kafka.KafkaContainer
	brokerAddress  string
	topic          string
	conn           *kafkago.Conn
}

func TestMainSuite(t *testing.T) {
	suite.Run(t, new(MainSuite))
}

func (s *MainSuite) SetupSuite() {
	var err error
	ctx := context.Background()

	s.kafkaContainer, err = kafka.RunContainer(ctx,
		kafka.WithClusterID("test-cluster"),
		testcontainers.WithImage("confluentinc/confluent-local:7.5.0"),
	)
	s.Require().NoError(err)

	brokers, err := s.kafkaContainer.Brokers(ctx)

	s.brokerAddress = brokers[0]

	s.topic = "test-topic"

	s.messagingMock = mocks.NewMessagingMock(s.T())
	s.kaiSdkMock = sdk.KaiSDK{
		Logger:            testr.NewWithOptions(s.T(), testr.Options{Verbosity: 1}),
		CentralizedConfig: mocks.NewCentralizedConfigMock(s.T()),
		Messaging:         s.messagingMock,
	}
}

func (s *MainSuite) TearDownSuite() {
	s.Require().NoError(s.kafkaContainer.Terminate(context.Background()))
}

func (s *MainSuite) TearDownTest() {
	config = kafkaConfig{}
}

func (s *MainSuite) TestRunnerFunc() {
	// GIVEN
	wg := sync.WaitGroup{}
	wg.Add(1)
	s.messagingMock.EXPECT().SendOutputWithRequestID(mock.Anything, mock.Anything).
		RunAndReturn(func(message protoreflect.ProtoMessage, s string, s2 ...string) error {
			wg.Done()

			return nil
		})

	s.createTestTopic()

	// WHEN
	go func() {
		config = kafkaConfig{
			Brokers:            []string{s.brokerAddress},
			GroupID:            "test-group",
			Topic:              s.topic,
			TLSEnabled:         false,
			InsecureSkipVerify: true,
		}

		kafkaRunner(nil, s.kaiSdkMock)
	}()

	s.produceKafkaMessages()

	//THEN
	wg.Wait()
}

func (s *MainSuite) produceKafkaMessages() {
	kafkaProducer := kafkago.Writer{
		Addr:     kafkago.TCP(s.brokerAddress),
		Topic:    s.topic,
		Balancer: &kafkago.LeastBytes{},
	}

	defer kafkaProducer.Close()

	fmt.Printf("Producing messages into kafka...\n")

	err := kafkaProducer.WriteMessages(context.Background(),
		kafkago.Message{
			Key:   []byte("Key-A"),
			Value: []byte("Hello World!"),
		},
	)
	if err != nil {
		log.Fatal("failed to write messages:", err)
	}
}

func (s *MainSuite) createTestTopic() {
	conn, err := kafkago.Dial("tcp", s.brokerAddress)
	if err != nil {
		panic(err.Error())
	}

	topicConfigs := []kafkago.TopicConfig{
		{
			Topic:             s.topic,
			NumPartitions:     1,
			ReplicationFactor: 1,
		},
	}

	err = conn.CreateTopics(topicConfigs...)
	if err != nil {
		panic(err.Error())
	}

	err = conn.Close()
	if err != nil {
		panic(err.Error())
	}
}

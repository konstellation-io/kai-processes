//go:build integration

package main

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/go-logr/logr/testr"
	"github.com/konstellation-io/kai-sdk/go-sdk/mocks"
	"github.com/konstellation-io/kai-sdk/go-sdk/sdk"
	"github.com/konstellation-io/kai-sdk/go-sdk/sdk/messaging"
	kafkago "github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/kafka"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type MainSuite struct {
	suite.Suite
	kaiSdkMock            sdk.KaiSDK
	centralizedConfigMock *mocks.CentralizedConfigMock
	messagingMock         *mocks.MessagingMock
	kafkaContainer        *kafka.KafkaContainer
	brokerAddress         string
	topic                 string
	conn                  *kafkago.Conn
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

	s.Require().Len(brokers, 1)
	s.brokerAddress = brokers[0]
	s.topic = "test-topic"

	s.centralizedConfigMock = mocks.NewCentralizedConfigMock(s.T())
	s.messagingMock = mocks.NewMessagingMock(s.T())

	s.kaiSdkMock = sdk.KaiSDK{
		Logger:            testr.NewWithOptions(s.T(), testr.Options{Verbosity: 1}),
		CentralizedConfig: s.centralizedConfigMock,
		Messaging:         s.messagingMock,
	}
}

func (s *MainSuite) TearDownSuite() {
	s.Require().NoError(s.kafkaContainer.Terminate(context.Background()))
}

func (s *MainSuite) TearDownTest() {
	config = kafkaConfig{}
}

func (s *MainSuite) waitOrTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()
	select {
	case <-c:
		return true
	case <-time.After(timeout):
		return false
	}
}

func (s *MainSuite) TestInitializer() {
	rawBrokers := "broker1,broker2"
	brokers := []string{"broker1", "broker2"}
	groupID := "test-group"
	topic := "test-topic"

	s.centralizedConfigMock.EXPECT().GetConfig("brokers", messaging.ProcessScope).Return(rawBrokers, nil)
	s.centralizedConfigMock.EXPECT().GetConfig("groupid", messaging.ProcessScope).Return(groupID, nil)
	s.centralizedConfigMock.EXPECT().GetConfig("topic", messaging.ProcessScope).Return(topic, nil)
	s.centralizedConfigMock.EXPECT().GetConfig("tls_enabled", messaging.ProcessScope).Return("", fmt.Errorf("not found"))
	s.centralizedConfigMock.EXPECT().GetConfig("skip_tls_verify", messaging.ProcessScope).Return("", fmt.Errorf("not found"))

	initializer(s.kaiSdkMock)

	s.Require().Equal(brokers, config.Brokers)
	s.Require().Equal(groupID, config.GroupID)
	s.Require().Equal(topic, config.Topic)
	s.Require().False(config.TLSEnabled)
	s.Require().False(config.InsecureSkipVerify)
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
	s.waitOrTimeout(&wg, 30*time.Second)
}

func (s *MainSuite) produceKafkaMessages() {
	kafkaProducer := kafkago.Writer{
		Addr:     kafkago.TCP(s.brokerAddress),
		Topic:    s.topic,
		Balancer: &kafkago.LeastBytes{},
	}
	defer kafkaProducer.Close()

	err := kafkaProducer.WriteMessages(context.Background(),
		kafkago.Message{
			Key:   []byte("Key-A"),
			Value: []byte("Hello World!"),
		},
	)
	s.Require().NoError(err)
}

func (s *MainSuite) createTestTopic() {
	conn, err := kafkago.Dial("tcp", s.brokerAddress)
	s.Require().NoError(err)

	defer conn.Close()

	topicConfigs := []kafkago.TopicConfig{
		{
			Topic:             s.topic,
			NumPartitions:     1,
			ReplicationFactor: 1,
		},
	}

	err = conn.CreateTopics(topicConfigs...)
	s.Require().NoError(err)
}

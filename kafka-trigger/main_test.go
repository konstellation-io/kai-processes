//go:build integration

package main

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/go-logr/logr/testr"
	"github.com/konstellation-io/kai-sdk/go-sdk/mocks"
	"github.com/konstellation-io/kai-sdk/go-sdk/sdk"
	kafkago "github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/kafka"
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
	hostAddress    string
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
		// testcontainers.WithConfigModifier(
		// 	func(config *container.Config) {
		// 		config.Env = append(config.Env, "KAFKA_AUTO_CREATE_TOPICS_ENABLE=true")
		// 		config.ExposedPorts = nat.PortSet{
		// 			"9092": {},
		// 			"9093": {},
		// 			"9094": {},
		// 		}
		// 	},
		// ),
	)
	s.Require().NoError(err)

	// req := testcontainers.ContainerRequest{
	// 	Image:        "confluentinc/confluent-local:7.5.0",
	// 	ExposedPorts: []string{"9092", "9093", "9094"},
	// 	Env:          map[string]string{
	// 		//"KAFKA_AUTO_CREATE_TOPICS_ENABLE": "true",
	// 	},
	// 	WaitingFor: wait.ForLog("Server started"),
	// }

	// s.kafkaContainer, err = testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
	// 	ContainerRequest: req,
	// 	Started:          true,
	// })
	// s.Require().NoError(err)

	// host, err := s.kafkaContainer.Host(ctx)
	// if err != nil {
	// 	panic(err)
	// }
	// port, err := s.kafkaContainer.MappedPort(ctx, KAFKA_CLIENT_PORT)
	// if err != nil {
	// 	panic(err)
	// }
	// brokerPort, err := s.kafkaContainer.MappedPort(ctx, KAFKA_BROKER_PORT)
	// if err != nil {
	// 	panic(err)
	// }

	brokers, err := s.kafkaContainer.Brokers(ctx)
	fmt.Println("++++++++++++++++++++++++++")
	fmt.Println(brokers)

	// s.hostAddress = host + ":" + port.Port()
	// s.brokerAddress = host + ":" + brokerPort.Port()

	s.hostAddress = brokers[0]
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
	s.messagingMock.EXPECT().SendOutputWithRequestID(mock.Anything, mock.Anything).Return(nil)

	s.createTestTopic()
	s.listTopics()

	go func() {
		config = kafkaConfig{
			Brokers:            []string{s.hostAddress},
			GroupID:            "test-group",
			Topic:              s.topic,
			TLSEnabled:         false,
			InsecureSkipVerify: true,
		}

		kafkaRunner(nil, s.kaiSdkMock)
	}()

	s.produceKafkaMessages()

	time.Sleep(10 * time.Second)

}

func (s *MainSuite) produceKafkaMessages() {
	kafkaProducer := kafkago.NewWriter(
		kafkago.WriterConfig{
			Brokers:  []string{s.hostAddress},
			Topic:    s.topic,
			Balancer: &kafkago.LeastBytes{},
		},
	)

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

func (s *MainSuite) listTopics() {
	conn, err := kafkago.Dial("tcp", s.brokerAddress)
	if err != nil {
		panic(err.Error())
	}

	partitions, err := conn.ReadPartitions()
	if err != nil {
		panic(err.Error())
	}

	m := map[string]struct{}{}

	for _, p := range partitions {
		m[p.Topic] = struct{}{}
	}
	for k := range m {
		fmt.Println(k)
	}

	err = conn.Close()
	if err != nil {
		panic(err.Error())
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

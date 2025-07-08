package infrastructure

import (
	"context"
	"strconv"
	"testing"
	"time"

	"thanhnt208/container-adm-service/config"

	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	tcKafka "github.com/testcontainers/testcontainers-go/modules/kafka"
)

func createTopic(brokers []string, topic string) error {
	conn, err := kafka.Dial("tcp", brokers[0])
	if err != nil {
		return err
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		return err
	}

	controllerAddr := controller.Host + ":" + strconv.Itoa(controller.Port)
	conn, err = kafka.Dial("tcp", controllerAddr)
	if err != nil {
		return err
	}
	defer conn.Close()

	return conn.CreateTopics(kafka.TopicConfig{
		Topic:             topic,
		NumPartitions:     1,
		ReplicationFactor: 1,
	})
}

func TestKafka_ProduceConsume_WithRealKafka(t *testing.T) {
	ctx := context.Background()

	kafkaContainer, err := tcKafka.RunContainer(ctx,
		testcontainers.WithImage("confluentinc/cp-kafka:7.3.0"),
		tcKafka.WithClusterID("test-cluster"),
	)
	assert.NoError(t, err)
	defer kafkaContainer.Terminate(ctx)

	brokers, err := kafkaContainer.Brokers(ctx)
	assert.NoError(t, err)
	assert.NotEmpty(t, brokers)

	testTopic := "test-topic"
	testGroupID := "test-group"

	// Explicitly create the topic
	err = createTopic(brokers, testTopic)
	assert.NoError(t, err)

	cfg := &config.Config{
		KafkaBrokers: brokers,
		KafkaTopic:   testTopic,
		KafkaGroupID: testGroupID,
	}

	k := NewKafka(cfg)

	// Connect producer
	producer, err := k.ConnectProducer()
	assert.NoError(t, err)
	assert.NotNil(t, producer)

	// Send message
	err = producer.WriteMessages(ctx, kafka.Message{
		Key:   []byte("key"),
		Value: []byte("hello world"),
	})
	assert.NoError(t, err)

	// Connect consumer
	consumer, err := k.ConnectConsumer([]string{testTopic})
	assert.NoError(t, err)
	assert.NotNil(t, consumer)

	// Wait for message up to 10s
	var msg kafka.Message
	found := false
	readCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	for {
		msg, err = consumer.ReadMessage(readCtx)
		if err == nil && string(msg.Value) == "hello world" {
			found = true
			break
		}
		if readCtx.Err() != nil {
			break
		}
	}
	assert.True(t, found, "Did not receive expected Kafka message")

	// Clean up
	err = k.Close()
	assert.NoError(t, err)
}

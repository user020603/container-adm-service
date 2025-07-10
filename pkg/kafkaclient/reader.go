package kafkaclient

import (
	"context"

	"github.com/segmentio/kafka-go"
)

type IKafkaReader interface {
	ReadMessage(ctx context.Context) (kafka.Message, error)
	Close() error
}

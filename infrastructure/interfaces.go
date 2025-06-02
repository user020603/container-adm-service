package infrastructure

import (
	"context"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"gorm.io/gorm"
)

type IDatabase interface {
	ConnectDB() (*gorm.DB, error)
	Close() error
}

type IRedis interface {
	ConnectClient() (*redis.Client, error)
	Ping(ctx context.Context) error
	Close() error
}

type IElasticsearch interface {
	ConnectElasticsearch() (*elasticsearch.Client, error)
	Close() error
}

type IKafka interface {
	ConnectProducer() (*kafka.Writer, error)
	ConnectConsumer(topics []string) (*kafka.Reader, error)
	Close() error
}

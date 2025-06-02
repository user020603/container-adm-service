package infrastructure

import (
	"fmt"
	"thanhnt208/container-adm-service/config"
	"time"

	"github.com/segmentio/kafka-go"
)

type Kafka struct {
	brokers []string
	writer  *kafka.Writer
	reader  *kafka.Reader
	cfg     *config.Config
}

func NewKafka(cfg *config.Config) IKafka {
	return &Kafka{
		brokers: cfg.KafkaBrokers,
		cfg:     cfg,
	}
}

func (k *Kafka) ConnectProducer() (*kafka.Writer, error) {
	k.writer = &kafka.Writer{
		Addr:         kafka.TCP(k.brokers...),
		Topic:        k.cfg.KafkaTopic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireAll,
		Async:        false,
	}

	return k.writer, nil
}

func (k *Kafka) ConnectConsumer(topics []string) (*kafka.Reader, error) {
	k.reader = kafka.NewReader(kafka.ReaderConfig{
		Brokers:     k.brokers,
		GroupID:     k.cfg.KafkaGroupID,
		Topic:       topics[0],
		MinBytes:    10e3,
		MaxBytes:    10e6,
		StartOffset: kafka.FirstOffset,
		MaxWait:     1 * time.Second,
	})
	return k.reader, nil
}

func (k *Kafka) Close() error {
	var writerErr, readerErr error

	if k.writer != nil {
		writerErr = k.writer.Close()
		k.writer = nil
	}

	if k.reader != nil {
		readerErr = k.reader.Close()
		k.reader = nil
	}

	if writerErr != nil && readerErr != nil {
		return fmt.Errorf("failed to close Kafka writer (%w) and reader (%w)", writerErr, readerErr)
	}
	if writerErr != nil {
		return fmt.Errorf("failed to close Kafka writer: %w", writerErr)
	}
	if readerErr != nil {
		return fmt.Errorf("failed to close Kafka reader: %w", readerErr)
	}
	return nil
}

package main

import (
	"context"
	"fmt"
	"syscall"
	"testing"
	"thanhnt208/container-adm-service/config"
	"thanhnt208/container-adm-service/external/client"
	"thanhnt208/container-adm-service/infrastructure"
	kafkaHandler "thanhnt208/container-adm-service/internal/delivery/kafka"
	"thanhnt208/container-adm-service/internal/mocks"
	"thanhnt208/container-adm-service/internal/service"
	"thanhnt208/container-adm-service/pkg/logger"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/golang/mock/gomock"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type mockDatabase struct {
	connectErr error
}

func (m *mockDatabase) ConnectDB() (*gorm.DB, error) {
	return nil, m.connectErr
}
func (m *mockDatabase) Close() error {
	return nil
}

type mockElasticsearch struct {
	connectErr error
}

func (m *mockElasticsearch) ConnectElasticsearch() (*elasticsearch.Client, error) {
	return nil, m.connectErr
}
func (m *mockElasticsearch) Close() error {
	return nil
}

type mockDockerClient struct{}

func (m *mockDockerClient) StartContainer(ctx context.Context, containerName, imageName string) (string, error) {
	return "mock-container-id", nil
}
func (m *mockDockerClient) StopContainer(ctx context.Context, containerID string) error {
	return nil
}
func (m *mockDockerClient) RemoveContainer(ctx context.Context, containerID string) error {
	return nil
}
func (m *mockDockerClient) StartExistingContainer(ctx context.Context, containerID string) error {
	return nil
}

type mockKafkaConsumerHandler struct {
	mock.Mock
}

func (m *mockKafkaConsumerHandler) StartConsume(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *mockKafkaConsumerHandler) Close() error {
	return nil
}

type mockKafka struct {
	reader client.IKafkaReader
}

func (m *mockKafka) ConnectProducer() (*kafka.Writer, error) {
	return nil, nil
}

func (m *mockKafka) ConnectConsumer(topics []string) (*kafka.Reader, error) {
	// bạn không thể trả mockKafkaReader ở đây nếu hàm yêu cầu *kafka.Reader
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"localhost:9092"}, // giả lập thôi, không cần thực kết nối
		Topic:   "test-topic",
		GroupID: "test-group",
	}), nil
}

func (m *mockKafka) Close() error {
	return nil
}

type mockKafkaReader struct{}

func (m *mockKafkaReader) ReadMessage(ctx context.Context) (kafka.Message, error) {
	<-ctx.Done()
	return kafka.Message{}, ctx.Err()
}

func (m *mockKafkaReader) Close() error {
	return nil
}

func TestRun_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLog := mocks.NewMockILogger(ctrl)
	mockLog.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLog.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

	originalLogger := logger.NewLogger
	logger.NewLogger = func(level, file string) (logger.ILogger, error) {
		return mockLog, nil
	}
	defer func() { logger.NewLogger = originalLogger }()

	originalDB := infrastructure.NewDatabase
	infrastructure.NewDatabase = func(cfg *config.Config) infrastructure.IDatabase {
		return &mockDatabase{connectErr: nil}
	}
	defer func() { infrastructure.NewDatabase = originalDB }()

	originalES := infrastructure.NewElasticsearch
	infrastructure.NewElasticsearch = func(cfg *config.Config) infrastructure.IElasticsearch {
		return &mockElasticsearch{connectErr: nil}
	}
	defer func() { infrastructure.NewElasticsearch = originalES }()

	originalDockerClient := client.NewDockerClient
	client.NewDockerClient = func() (client.IDockerClient, error) {
		return &mockDockerClient{}, nil
	}
	defer func() { client.NewDockerClient = originalDockerClient }()

	originalKafka := infrastructure.NewKafka
	infrastructure.NewKafka = func(cfg *config.Config) infrastructure.IKafka {
		return &mockKafka{reader: &mockKafkaReader{}}
	}
	defer func() { infrastructure.NewKafka = originalKafka }()

	originalHandler := newKafkaConsumerHandler
	mockHandler := new(mockKafkaConsumerHandler)
	mockHandler.On("StartConsume", mock.Anything).Return(context.Canceled)

	newKafkaConsumerHandler = func(
		svc service.IContainerService,
		log logger.ILogger,
		reader client.IKafkaReader,
	) kafkaHandler.IKafkaConsumerHandler {
		return mockHandler
	}
	defer func() { newKafkaConsumerHandler = originalHandler }()

	go func() {
		time.Sleep(100 * time.Millisecond) 
		_ = syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	}()

	err := Run()
	assert.NoError(t, err, "Expected no error, got %v", err)
}

func TestRun_FailToInitLogger(t *testing.T) {
	originalLogger := logger.NewLogger
	logger.NewLogger = func(level, file string) (logger.ILogger, error) {
		return nil, fmt.Errorf("logger error")
	}
	defer func() { logger.NewLogger = originalLogger }()

	assert.PanicsWithValue(t, "Failed to initialize logger: logger error", func() {
		_ = Run()
	})
}

func TestRun_DBConnectionFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLog := mocks.NewMockILogger(ctrl)
	mockLog.EXPECT().
		Error(gomock.Any(), gomock.Any(), gomock.Any()).
		Times(1)

	originalLogger := logger.NewLogger
	logger.NewLogger = func(level, file string) (logger.ILogger, error) {
		return mockLog, nil
	}
	defer func() { logger.NewLogger = originalLogger }()

	originalDB := infrastructure.NewDatabase
	infrastructure.NewDatabase = func(cfg *config.Config) infrastructure.IDatabase {
		return &mockDatabase{connectErr: fmt.Errorf("db error")}
	}
	defer func() { infrastructure.NewDatabase = originalDB }()

	assert.PanicsWithValue(t, "Failed to connect to the database: db error", func() {
		_ = Run()
	})
}

func TestRun_ElasticsearchConnectionFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLog := mocks.NewMockILogger(ctrl)
	mockLog.EXPECT().Error(
		gomock.Eq("Failed to connect to Elasticsearch"),
		gomock.Eq("error"),
		gomock.Any(),
	).Times(1)

	originalLogger := logger.NewLogger
	logger.NewLogger = func(level, file string) (logger.ILogger, error) {
		return mockLog, nil
	}
	defer func() { logger.NewLogger = originalLogger }()

	originalDB := infrastructure.NewDatabase
	infrastructure.NewDatabase = func(cfg *config.Config) infrastructure.IDatabase {
		return &mockDatabase{connectErr: nil}
	}
	defer func() { infrastructure.NewDatabase = originalDB }()

	originalES := infrastructure.NewElasticsearch
	infrastructure.NewElasticsearch = func(cfg *config.Config) infrastructure.IElasticsearch {
		return &mockElasticsearch{connectErr: fmt.Errorf("es connection failed")}
	}
	defer func() { infrastructure.NewElasticsearch = originalES }()

	assert.PanicsWithValue(t,
		"Failed to connect to Elasticsearch: es connection failed",
		func() {
			_ = Run()
		})
}

func TestRun_DockerClientInitFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLog := mocks.NewMockILogger(ctrl)
	mockLog.EXPECT().Error(
		gomock.Eq("Failed to create Docker client"),
		gomock.Eq("error"),
		gomock.Any(),
	).Times(1)

	originalLogger := logger.NewLogger
	logger.NewLogger = func(level, file string) (logger.ILogger, error) {
		return mockLog, nil
	}
	defer func() { logger.NewLogger = originalLogger }()

	originalDB := infrastructure.NewDatabase
	infrastructure.NewDatabase = func(cfg *config.Config) infrastructure.IDatabase {
		return &mockDatabase{connectErr: nil}
	}
	defer func() { infrastructure.NewDatabase = originalDB }()

	originalES := infrastructure.NewElasticsearch
	infrastructure.NewElasticsearch = func(cfg *config.Config) infrastructure.IElasticsearch {
		return &mockElasticsearch{connectErr: nil}
	}
	defer func() { infrastructure.NewElasticsearch = originalES }()

	// Mock Kafka success
	originalKafka := infrastructure.NewKafka
	infrastructure.NewKafka = func(cfg *config.Config) infrastructure.IKafka {
		return &mockKafka{
			reader: &mockKafkaReader{},
		}
	}
	defer func() { infrastructure.NewKafka = originalKafka }()

	originalDockerClient := client.NewDockerClient
	client.NewDockerClient = func() (client.IDockerClient, error) {
		return nil, fmt.Errorf("docker init failed")
	}
	defer func() { client.NewDockerClient = originalDockerClient }()

	assert.PanicsWithValue(t,
		"Failed to create Docker client: docker init failed",
		func() {
			_ = Run()
		})
}

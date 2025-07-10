package main

import (
	"errors"
	"fmt"
	"sync"
	"syscall"
	"testing"
	"thanhnt208/container-adm-service/config"
	"thanhnt208/container-adm-service/external/client"
	"thanhnt208/container-adm-service/infrastructure"
	"thanhnt208/container-adm-service/internal/mocks"
	"thanhnt208/container-adm-service/internal/service"
	"thanhnt208/container-adm-service/pkg/logger"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/golang/mock/gomock"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/net/context"
	"gorm.io/gorm"

	kafkaHandler "thanhnt208/container-adm-service/internal/delivery/kafka"
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
	logger.NewLogger = func(level string, file string) (logger.ILogger, error) {
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
		return &mockKafka{
			reader: &mockKafkaReader{},
		}
	}
	defer func() { infrastructure.NewKafka = originalKafka }()

	originalHandler := newKafkaConsumerHandler
	mockHandler := new(mockKafkaConsumerHandler)
	mockHandler.On("StartConsume", mock.Anything).Return(context.Canceled)

	newKafkaConsumerHandler = func(service.IContainerService, logger.ILogger, client.IKafkaReader) kafkaHandler.IKafkaConsumerHandler {
		return mockHandler
	}
	defer func() { newKafkaConsumerHandler = originalHandler }()

	go func() {
		syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	}()

	err := Run()
	if err != nil {
		assert.NoError(t, err, "Expected no error, got %v", err)
	}
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
	// Expect the error log to be called once with any arguments
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

	// Mock DB success
	originalDB := infrastructure.NewDatabase
	infrastructure.NewDatabase = func(cfg *config.Config) infrastructure.IDatabase {
		return &mockDatabase{connectErr: nil}
	}
	defer func() { infrastructure.NewDatabase = originalDB }()

	// Mock Elasticsearch success
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

	// Mock Docker client fail
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

func TestRun_KafkaHandlerConsume_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var wg sync.WaitGroup
	wg.Add(1)

	mockLog := mocks.NewMockILogger(ctrl)
	mockLog.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLog.EXPECT().
		Error(gomock.Any(), gomock.Any(), gomock.Any()).
		Do(func(msg string, key string, val interface{}) {
			t.Logf("MockLog.Error called with: %s - %s: %v", msg, key, val)
		}).
		AnyTimes()

	mockHandler := new(mockKafkaConsumerHandler)
	mockHandler.On("StartConsume", mock.Anything).
		Run(func(args mock.Arguments) {
			// Giả lập xử lý rồi mới signal
			time.Sleep(100 * time.Millisecond)
			wg.Done()
		}).
		Return(errors.New("consume failed"))

	mockHandler.On("Close").Return(nil)

	// Mock DB success
	originalDB := infrastructure.NewDatabase
	infrastructure.NewDatabase = func(cfg *config.Config) infrastructure.IDatabase {
		return &mockDatabase{connectErr: nil}
	}
	defer func() { infrastructure.NewDatabase = originalDB }()

	// Inject mock handler
	originalHandler := newKafkaConsumerHandler
	newKafkaConsumerHandler = func(service.IContainerService, logger.ILogger, client.IKafkaReader) kafkaHandler.IKafkaConsumerHandler {
		return mockHandler
	}
	defer func() { newKafkaConsumerHandler = originalHandler }()

	// Gửi SIGINT sau khi tiêu thụ xong
	go func() {
		wg.Wait()
		time.Sleep(100 * time.Millisecond) // đảm bảo log.Error được gọi
		syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	}()

	err := Run()
	assert.NoError(t, err)
}

func TestRun_KafkaHandlerConsume_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var wg sync.WaitGroup
	wg.Add(1)

	mockLog := mocks.NewMockILogger(ctrl)

	// Log các Info() để dễ debug
	mockLog.EXPECT().
		Info(gomock.Any(), gomock.Any()).
		Do(func(msg string, key interface{}) {
			t.Logf("MockLog.Info called with: %s - %v", msg, key)
		}).
		AnyTimes()

	mockLog.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	mockHandler := new(mockKafkaConsumerHandler)
	mockHandler.On("StartConsume", mock.Anything).
		Run(func(args mock.Arguments) {
			time.Sleep(100 * time.Millisecond)
			wg.Done()
		}).
		Return(nil)

	mockHandler.On("Close").Return(nil)

	// Mock DB
	originalDB := infrastructure.NewDatabase
	infrastructure.NewDatabase = func(cfg *config.Config) infrastructure.IDatabase {
		return &mockDatabase{connectErr: nil}
	}
	defer func() { infrastructure.NewDatabase = originalDB }()

	// Inject mock handler
	originalHandler := newKafkaConsumerHandler
	newKafkaConsumerHandler = func(service.IContainerService, logger.ILogger, client.IKafkaReader) kafkaHandler.IKafkaConsumerHandler {
		return mockHandler
	}
	defer func() { newKafkaConsumerHandler = originalHandler }()

	// Đợi goroutine xong
	go func() {
		wg.Wait()
	}()

	err := Run()
	assert.NoError(t, err)
}

func TestRun_KafkaConsumer_FinishWithNonCanceledError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var wg sync.WaitGroup
	wg.Add(1)

	mockLog := mocks.NewMockILogger(ctrl)

	// KHÔNG dùng .AnyTimes() cho Info nữa – gomock sẽ nuốt mất call cụ thể
	mockLog.EXPECT().
		Info(gomock.Eq("Starting Kafka consumer"), gomock.Any()).
		Times(1)

	mockLog.EXPECT().
		Error(gomock.Eq("Kafka consumer did not finish gracefully"), gomock.Eq("error"), gomock.Any()).
		Times(1)

	mockLog.EXPECT().
		Info(gomock.Eq("Kafka consumer closed successfully"), gomock.Any()).
		Times(1)

	mockLog.EXPECT().
		Info(gomock.Eq("Service shutdown complete"), gomock.Any()).
		Times(1)

	// mock handler
	mockHandler := new(mockKafkaConsumerHandler)
	mockHandler.On("StartConsume", mock.Anything).
		Run(func(args mock.Arguments) {
			wg.Done()
		}).
		Return(errors.New("unexpected error"))

	mockHandler.On("Close").Return(nil)

	// mock DB
	originalDB := infrastructure.NewDatabase
	infrastructure.NewDatabase = func(cfg *config.Config) infrastructure.IDatabase {
		return &mockDatabase{connectErr: nil}
	}
	defer func() { infrastructure.NewDatabase = originalDB }()

	// inject handler
	originalHandler := newKafkaConsumerHandler
	newKafkaConsumerHandler = func(service.IContainerService, logger.ILogger, client.IKafkaReader) kafkaHandler.IKafkaConsumerHandler {
		return mockHandler
	}
	defer func() { newKafkaConsumerHandler = originalHandler }()

	// gửi SIGINT
	go func() {
		wg.Wait()
		time.Sleep(50 * time.Millisecond)
		syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	}()

	err := Run()
	assert.NoError(t, err)
}

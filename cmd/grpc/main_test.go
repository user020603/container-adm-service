package main

import (
	"context"
	"errors"
	"syscall"
	"testing"
	"thanhnt208/container-adm-service/config"
	"thanhnt208/container-adm-service/external/client"
	"thanhnt208/container-adm-service/infrastructure"
	"thanhnt208/container-adm-service/internal/mocks"
	"thanhnt208/container-adm-service/pkg/logger"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
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

func TestRun_FailInitLogger(t *testing.T) {
	original := logger.NewLogger
	defer func() { logger.NewLogger = original }()

	logger.NewLogger = func(level, file string) (logger.ILogger, error) {
		return nil, errors.New("logger init fail")
	}

	assert.PanicsWithValue(t, "Failed to initialize logger: logger init fail", func() {
		Run()
	})
}

func TestRun_FailConnectDB(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLog := mocks.NewMockILogger(ctrl)
	mockLog.EXPECT().Error("Failed to connect to the database", "error", gomock.Any()).Times(1)

	originalLogger := logger.NewLogger
	defer func() { logger.NewLogger = originalLogger }()
	logger.NewLogger = func(level, file string) (logger.ILogger, error) {
		return mockLog, nil
	}

	originalNewDatabase := infrastructure.NewDatabase
	defer func() { infrastructure.NewDatabase = originalNewDatabase }()
	infrastructure.NewDatabase = func(cfg *config.Config) infrastructure.IDatabase {
		return &mockDatabase{connectErr: errors.New("db error")}
	}

	assert.PanicsWithValue(t, "Failed to connect to the database: db error", func() {
		Run()
	})
}

func TestRun_FailConnectElasticsearch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLog := mocks.NewMockILogger(ctrl)
	mockLog.EXPECT().Error("Failed to connect to Elasticsearch", "error", gomock.Any()).Times(1)

	logger.NewLogger = func(level, file string) (logger.ILogger, error) {
		return mockLog, nil
	}

	infrastructure.NewDatabase = func(cfg *config.Config) infrastructure.IDatabase {
		return &mockDatabase{connectErr: nil}
	}

	infrastructure.NewElasticsearch = func(cfg *config.Config) infrastructure.IElasticsearch {
		return &mockElasticsearch{connectErr: errors.New("es error")}
	}

	assert.PanicsWithValue(t, "Failed to connect to Elasticsearch: es error", func() {
		Run()
	})
}

func TestRun_FailDockerClient(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLog := mocks.NewMockILogger(ctrl)
	mockLog.EXPECT().Error("Failed to create Docker client", "error", gomock.Any()).Times(1)

	logger.NewLogger = func(level, file string) (logger.ILogger, error) {
		return mockLog, nil
	}

	infrastructure.NewDatabase = func(cfg *config.Config) infrastructure.IDatabase {
		return &mockDatabase{}
	}

	infrastructure.NewElasticsearch = func(cfg *config.Config) infrastructure.IElasticsearch {
		return &mockElasticsearch{}
	}

	client.NewDockerClient = func() (client.IDockerClient, error) {
		return nil, errors.New("docker error")
	}

	assert.PanicsWithValue(t, "Failed to create Docker client: docker error", func() {
		Run()
	})
}

func TestRun_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLog := mocks.NewMockILogger(ctrl)
	mockLog.EXPECT().Info("Starting gRPC server", "port", gomock.Any()).AnyTimes()
	mockLog.EXPECT().Info("Shutting down gRPC server...").AnyTimes()
	mockLog.EXPECT().Info("gRPC server exiting").AnyTimes()

	originalLogger := logger.NewLogger
	defer func() { logger.NewLogger = originalLogger }()
	logger.NewLogger = func(level, file string) (logger.ILogger, error) {
		return mockLog, nil
	}

	// --- Mock Database ---
	originalDB := infrastructure.NewDatabase
	defer func() { infrastructure.NewDatabase = originalDB }()
	infrastructure.NewDatabase = func(cfg *config.Config) infrastructure.IDatabase {
		return &mockDatabase{}
	}

	// --- Mock Elasticsearch ---
	originalES := infrastructure.NewElasticsearch
	defer func() { infrastructure.NewElasticsearch = originalES }()
	infrastructure.NewElasticsearch = func(cfg *config.Config) infrastructure.IElasticsearch {
		return &mockElasticsearch{}
	}

	// --- Mock Docker ---
	originalDocker := client.NewDockerClient
	defer func() { client.NewDockerClient = originalDocker }()
	client.NewDockerClient = func() (client.IDockerClient, error) {
		return &mockDockerClient{}, nil
	}

	// --- Simulate SIGINT ---
	go func() {
		time.Sleep(100 * time.Millisecond)
		_ = syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	}()

	// --- Run test ---
	err := Run()
	assert.NoError(t, err)
}

func TestMain_PanicOnRunError(t *testing.T) {
	originalRun := Run
	defer func() { Run = originalRun }()

	Run = func() error {
		return errors.New("mock run error")
	}

	assert.PanicsWithValue(t, "Application failed: mock run error", func() {
		main()
	})
}

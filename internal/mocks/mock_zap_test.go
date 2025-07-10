package mocks

import (
	"errors"
	"testing"
	"thanhnt208/container-adm-service/pkg/logger"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func ProcessSomething(logger logger.ILogger) {
	logger.Info("Processing started", "module", "test")
	logger.Info("Processing done", "module", "test")
}

func TestProcessSomething_LogsInfo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockILogger(ctrl)

	mockLogger.EXPECT().Info("Processing started", "module", "test")
	mockLogger.EXPECT().Info("Processing done", "module", "test")

	ProcessSomething(mockLogger)
}

func TestMockILogger_Methods(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockILogger(ctrl)

	mockLogger.EXPECT().Debug("debug message", "key", "value")
	mockLogger.EXPECT().Info("info message", "key", 123)
	mockLogger.EXPECT().Warn("warn message")
	mockLogger.EXPECT().Error("error message", "err", errors.New("fail"))
	mockLogger.EXPECT().Fatal("fatal message")
	mockLogger.EXPECT().Sync().Return(nil)

	mockLogger.Debug("debug message", "key", "value")
	mockLogger.Info("info message", "key", 123)
	mockLogger.Warn("warn message")
	mockLogger.Error("error message", "err", errors.New("fail"))
	mockLogger.Fatal("fatal message")
	err := mockLogger.Sync()
	require.NoError(t, err)
}

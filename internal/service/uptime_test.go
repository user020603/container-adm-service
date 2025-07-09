package service

import (
	"context"
	"errors"
	"testing"
	"thanhnt208/container-adm-service/internal/dto"
	"thanhnt208/container-adm-service/internal/mocks"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestGetContainerUptimeRatio_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockIContainerRepository(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	service := NewContainerService(mockRepo, mockLogger, nil)

	start := time.Now().Add(-24 * time.Hour)
	end := time.Now()
	expectedRatio := 0.95

	mockRepo.EXPECT().
		GetContainerUptimeRatio(gomock.Any(), start, end).
		Return(expectedRatio, nil)

	mockLogger.EXPECT().
		Info("Container uptime ratio retrieved successfully", "uptimeRatio", expectedRatio)

	ratio, err := service.GetContainerUptimeRatio(context.TODO(), start, end)

	assert.NoError(t, err)
	assert.Equal(t, expectedRatio, ratio)
}

func TestGetContainerUptimeRatio_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockIContainerRepository(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	service := NewContainerService(mockRepo, mockLogger, nil)

	start := time.Now().Add(-24 * time.Hour)
	end := time.Now()
	repoErr := errors.New("database error")

	mockRepo.EXPECT().
		GetContainerUptimeRatio(gomock.Any(), start, end).
		Return(0.0, repoErr)

	mockLogger.EXPECT().
		Error("Failed to get container uptime ratio", "startTime", start, "endTime", end, "error", repoErr)

	ratio, err := service.GetContainerUptimeRatio(context.TODO(), start, end)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get container uptime ratio")
	assert.Equal(t, 0.0, ratio)
}

func TestGetContainerUptimeDuration_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockIContainerRepository(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	service := NewContainerService(mockRepo, mockLogger, nil)

	start := time.Now().Add(-2 * time.Hour)
	end := time.Now()

	expectedDetails := &dto.UptimeDetails{
		TotalUptime: 90 * time.Minute,
		PerContainerUptime: map[string]time.Duration{
			"container1": 60 * time.Minute,
			"container2": 30 * time.Minute,
		},
	}

	mockRepo.EXPECT().
		GetContainerUptimeDuration(gomock.Any(), start, end).
		Return(expectedDetails, nil)

	mockLogger.EXPECT().
		Info("Container uptime duration retrieved successfully", "totalUptime", expectedDetails.TotalUptime)

	result, err := service.GetContainerUptimeDuration(context.TODO(), start, end)

	assert.NoError(t, err)
	assert.Equal(t, expectedDetails, result)
}

func TestGetContainerUptimeDuration_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockIContainerRepository(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	service := NewContainerService(mockRepo, mockLogger, nil)

	start := time.Now().Add(-2 * time.Hour)
	end := time.Now()

	repoErr := errors.New("database error")

	mockRepo.EXPECT().
		GetContainerUptimeDuration(gomock.Any(), start, end).
		Return(nil, repoErr)

	mockLogger.EXPECT().
		Error("Failed to get container uptime duration", "startTime", start, "endTime", end, "error", repoErr)

	result, err := service.GetContainerUptimeDuration(context.TODO(), start, end)

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get container uptime duration")
}

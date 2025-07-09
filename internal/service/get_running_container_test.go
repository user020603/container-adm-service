package service

import (
	"context"
	"errors"
	"testing"
	"thanhnt208/container-adm-service/internal/mocks"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestGetNumRunningContainers_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockIContainerRepository(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)

	service := NewContainerService(mockRepo, mockLogger, nil)

	expectedCount := int64(5)

	mockRepo.EXPECT().
		GetNumRunningContainers(gomock.Any()).
		Return(expectedCount, nil)

	mockLogger.EXPECT().
		Info("Number of running containers retrieved successfully", "count", expectedCount)

	count, err := service.GetNumRunningContainers(context.TODO())

	assert.NoError(t, err)
	assert.Equal(t, expectedCount, count)
}

func TestGetNumRunningContainers_RepoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockIContainerRepository(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)

	service := NewContainerService(mockRepo, mockLogger, nil)

	repoErr := errors.New("database failure")

	mockRepo.EXPECT().
		GetNumRunningContainers(gomock.Any()).
		Return(int64(0), repoErr)

	mockLogger.EXPECT().
		Error("Failed to get number of running containers", "error", repoErr)

	count, err := service.GetNumRunningContainers(context.TODO())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get number of running containers")
	assert.Equal(t, int64(0), count)
}

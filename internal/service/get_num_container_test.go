package service

import (
	"context"
	"errors"
	"testing"
	"thanhnt208/container-adm-service/internal/mocks"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestGetNumContainers_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockIContainerRepository(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)

	service := NewContainerService(mockRepo, mockLogger, nil)

	expectedCount := int64(5)

	mockRepo.EXPECT().
		GetNumContainers(gomock.Any()).
		Return(expectedCount, nil)

	mockLogger.EXPECT().
		Info("Number of containers retrieved successfully", "count", expectedCount)

	count, err := service.GetNumContainers(context.TODO())

	assert.NoError(t, err)
	assert.Equal(t, expectedCount, count)
}

func TestGetNumContainers_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockIContainerRepository(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)

	service := NewContainerService(mockRepo, mockLogger, nil)

	expectedErr := errors.New("database error")

	mockRepo.EXPECT().
		GetNumContainers(gomock.Any()).
		Return(int64(0), expectedErr)

	mockLogger.EXPECT().
		Error("Failed to get number of containers", "error", expectedErr)

	count, err := service.GetNumContainers(context.TODO())

	assert.Error(t, err)
	assert.Equal(t, int64(0), count)
	assert.Contains(t, err.Error(), "failed to get number of containers")
}

package service

import (
	"context"
	"errors"
	"testing"
	"thanhnt208/container-adm-service/internal/mocks"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestAddContainerStatus_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockIContainerRepository(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)

	mockService := NewContainerService(mockRepo, mockLogger, nil)

	containerID := uint(1)
	status := "running"

	mockRepo.EXPECT().
		AddContainerStatus(gomock.Any(), containerID, status).
		Return(nil)

	mockLogger.EXPECT().
		Info("Container status added successfully", "id", containerID, "status", status)

	err := mockService.AddContainerStatus(context.TODO(), containerID, status)

	assert.NoError(t, err)
}

func TestAddContainerStatus_RepoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockIContainerRepository(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)

	mockService := NewContainerService(mockRepo, mockLogger, nil)

	containerID := uint(1)
	status := "stopped"
	repoErr := errors.New("db error")

	mockRepo.EXPECT().
		AddContainerStatus(gomock.Any(), containerID, status).
		Return(repoErr)

	mockLogger.EXPECT().
		Error("Failed to add container status", "id", containerID, "status", status, "error", repoErr)

	err := mockService.AddContainerStatus(context.TODO(), containerID, status)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to add container status")
}

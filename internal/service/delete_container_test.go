package service

import (
	"context"
	"fmt"
	"testing"
	"thanhnt208/container-adm-service/internal/mocks"
	"thanhnt208/container-adm-service/internal/model"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestDeleteContainer_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockIContainerRepository(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	mockDockerClient := mocks.NewMockIDockerClient(ctrl)
	service := NewContainerService(mockRepo, mockLogger, mockDockerClient)

	ctx := context.TODO()
	id := uint(1)
	containerID := "abc123"
	container := &model.Container{
		ID:          id,
		ContainerID: containerID,
	}

	mockRepo.EXPECT().
		GetContainerByID(ctx, id).
		Return(container, nil)

	mockDockerClient.EXPECT().
		StopContainer(ctx, containerID).
		Return(nil)

	mockDockerClient.EXPECT().
		RemoveContainer(ctx, containerID).
		Return(nil)

	mockRepo.EXPECT().
		DeleteContainer(ctx, id).
		Return(nil)

	mockLogger.EXPECT().
		Info("Container deleted successfully", "id", id)

	err := service.DeleteContainer(ctx, id)

	assert.NoError(t, err)
}

func TestDeleteContainer_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockIContainerRepository(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	mockDockerClient := mocks.NewMockIDockerClient(ctrl)
	service := NewContainerService(mockRepo, mockLogger, mockDockerClient)

	ctx := context.TODO()
	id := uint(2)

	mockRepo.EXPECT().
		GetContainerByID(ctx, id).
		Return(nil, nil)

	mockLogger.EXPECT().
		Warn("Container not found for deletion", "id", id)

	err := service.DeleteContainer(ctx, id)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestDeleteContainer_GetContainerError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockIContainerRepository(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	mockDockerClient := mocks.NewMockIDockerClient(ctrl)
	service := NewContainerService(mockRepo, mockLogger, mockDockerClient)

	ctx := context.TODO()
	id := uint(3)
	expectedErr := fmt.Errorf("db error")

	mockRepo.EXPECT().
		GetContainerByID(ctx, id).
		Return(nil, expectedErr)

	mockLogger.EXPECT().
		Error("Failed to retrieve container for deletion", "id", id, "error", expectedErr)

	err := service.DeleteContainer(ctx, id)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to retrieve container for deletion")
}

func TestDeleteContainer_StopContainerError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockIContainerRepository(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	mockDockerClient := mocks.NewMockIDockerClient(ctrl)
	service := NewContainerService(mockRepo, mockLogger, mockDockerClient)

	ctx := context.TODO()
	id := uint(4)
	containerID := "abc123"
	stopErr := fmt.Errorf("stop error")

	container := &model.Container{
		ID:          id,
		ContainerID: containerID,
	}

	mockRepo.EXPECT().
		GetContainerByID(ctx, id).
		Return(container, nil)

	mockDockerClient.EXPECT().
		StopContainer(ctx, containerID).
		Return(stopErr)

	mockLogger.EXPECT().
		Error("Failed to stop Docker container before deletion", "containerID", containerID, "error", stopErr)

	err := service.DeleteContainer(ctx, id)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to stop Docker container")
}

func TestDeleteContainer_RemoveContainerError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockIContainerRepository(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	mockDockerClient := mocks.NewMockIDockerClient(ctrl)
	service := NewContainerService(mockRepo, mockLogger, mockDockerClient)

	ctx := context.TODO()
	id := uint(5)
	containerID := "abc123"
	removeErr := fmt.Errorf("remove error")

	container := &model.Container{
		ID:          id,
		ContainerID: containerID,
	}

	mockRepo.EXPECT().
		GetContainerByID(ctx, id).
		Return(container, nil)

	mockDockerClient.EXPECT().
		StopContainer(ctx, containerID).
		Return(nil)

	mockDockerClient.EXPECT().
		RemoveContainer(ctx, containerID).
		Return(removeErr)

	mockLogger.EXPECT().
		Error("Failed to remove Docker container", "containerID", containerID, "error", removeErr)

	err := service.DeleteContainer(ctx, id)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to remove Docker container")
}

func TestDeleteContainer_DeleteRepoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockIContainerRepository(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	mockDockerClient := mocks.NewMockIDockerClient(ctrl)
	service := NewContainerService(mockRepo, mockLogger, mockDockerClient)

	ctx := context.TODO()
	id := uint(6)
	containerID := "abc123"
	deleteErr := fmt.Errorf("repo delete error")

	container := &model.Container{
		ID:          id,
		ContainerID: containerID,
	}

	mockRepo.EXPECT().
		GetContainerByID(ctx, id).
		Return(container, nil)

	mockDockerClient.EXPECT().
		StopContainer(ctx, containerID).
		Return(nil)

	mockDockerClient.EXPECT().
		RemoveContainer(ctx, containerID).
		Return(nil)

	mockRepo.EXPECT().
		DeleteContainer(ctx, id).
		Return(deleteErr)

	mockLogger.EXPECT().
		Error("Failed to delete container from repository", "id", id, "error", deleteErr)

	err := service.DeleteContainer(ctx, id)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete container from repository")
}
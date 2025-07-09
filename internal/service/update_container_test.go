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

func TestUpdateContainer_DisallowNameUpdate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockIContainerRepository(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	mockDockerClient := mocks.NewMockIDockerClient(ctrl)
	mockService := NewContainerService(mockRepo, mockLogger, mockDockerClient)

	ctx := context.Background()
	id := uint(1)
	updateData := map[string]interface{}{
		"container_name": "not-allowed-name",
	}

	mockLogger.
		EXPECT().
		Warn("Container name update is not allowed", "id", id)

	container, err := mockService.UpdateContainer(ctx, id, updateData)

	assert.Error(t, err)
	assert.Nil(t, container)
	assert.Contains(t, err.Error(), "updating container name is not allowed")
}

func TestUpdateContainer_GetContainerError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockIContainerRepository(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	mockDockerClient := mocks.NewMockIDockerClient(ctrl)
	mockService := NewContainerService(mockRepo, mockLogger, mockDockerClient)

	ctx := context.TODO()
	id := uint(1)
	updateData := map[string]interface{}{}

	repoErr := fmt.Errorf("db error")

	mockRepo.
		EXPECT().
		GetContainerByID(ctx, id).
		Return(nil, repoErr)

	mockLogger.
		EXPECT().
		Error("Failed to retrieve container for update", "id", id, "error", repoErr)

	container, err := mockService.UpdateContainer(ctx, id, updateData)

	assert.Error(t, err)
	assert.Nil(t, container)
	assert.Contains(t, err.Error(), "failed to retrieve container for update")
}

func TestUpdateContainer_ContainerNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockIContainerRepository(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	mockDockerClient := mocks.NewMockIDockerClient(ctrl)
	service := NewContainerService(mockRepo, mockLogger, mockDockerClient)

	ctx := context.TODO()
	id := uint(3)
	updateData := map[string]interface{}{}

	mockRepo.
		EXPECT().
		GetContainerByID(ctx, id).
		Return(nil, nil)

	mockLogger.
		EXPECT().
		Warn("Container not found for update", "id", id)

	container, err := service.UpdateContainer(ctx, id, updateData)

	assert.Error(t, err)
	assert.Nil(t, container)
	assert.Contains(t, err.Error(), "container with ID")
}

func TestUpdateContainer_UpdateImageSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockIContainerRepository(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	mockDockerClient := mocks.NewMockIDockerClient(ctrl)
	mockService := NewContainerService(mockRepo, mockLogger, mockDockerClient)

	ctx := context.TODO()
	id := uint(4)
	oldImage := "nginx:latest"
	newImage := "nginx:1.25"
	oldContainerID := "abc123"
	newContainerID := "def456"
	containerName := "my-nginx"

	updateData := map[string]interface{}{
		"image_name": newImage,
	}

	existing := &model.Container{
		ID:            id,
		ContainerID:   oldContainerID,
		ContainerName: containerName,
		ImageName:     oldImage,
	}

	mockRepo.
		EXPECT().
		GetContainerByID(ctx, id).
		Return(existing, nil)

	mockDockerClient.
		EXPECT().
		StopContainer(ctx, oldContainerID).
		Return(nil)

	mockDockerClient.
		EXPECT().
		RemoveContainer(ctx, oldContainerID).
		Return(nil)

	mockDockerClient.
		EXPECT().
		StartContainer(ctx, containerName, newImage).
		Return(newContainerID, nil)

	mockRepo.
		EXPECT().
		UpdateContainer(ctx, id, gomock.Any()).
		DoAndReturn(func(_ context.Context, _ uint, data map[string]interface{}) (*model.Container, error) {
			assert.Equal(t, newContainerID, data["ContainerID"])
			assert.Equal(t, newImage, data["image_name"])
			return &model.Container{
				ID:            id,
				ContainerID:   newContainerID,
				ContainerName: containerName,
				ImageName:     newImage,
			}, nil
		})

	mockLogger.
		EXPECT().
		Info("Container updated successfully", "id", id, "containerID", newContainerID)

	container, err := mockService.UpdateContainer(ctx, id, updateData)

	assert.NoError(t, err)
	assert.NotNil(t, container)
	assert.Equal(t, newContainerID, container.ContainerID)
	assert.Equal(t, newImage, container.ImageName)
}

func TestUpdateContainer_StopContainerError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockIContainerRepository(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	mockDockerClient := mocks.NewMockIDockerClient(ctrl)
	mockService := NewContainerService(mockRepo, mockLogger, mockDockerClient)

	ctx := context.TODO()
	id := uint(5)
	oldContainerID := "xyz789"
	containerName := "my-container"
	oldImage := "alpine:latest"
	newImage := "alpine:3.15"

	updateData := map[string]interface{}{
		"image_name": newImage,
	}

	container := &model.Container{
		ID:            id,
		ContainerID:   oldContainerID,
		ContainerName: containerName,
		ImageName:     oldImage,
	}

	stopErr := fmt.Errorf("stop failure")

	mockRepo.
		EXPECT().
		GetContainerByID(ctx, id).
		Return(container, nil)

	mockDockerClient.
		EXPECT().
		StopContainer(ctx, oldContainerID).
		Return(stopErr)

	mockLogger.
		EXPECT().
		Error("Failed to update Docker container image", "containerID", oldContainerID, "error", stopErr)

	result, err := mockService.UpdateContainer(ctx, id, updateData)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to update Docker container image")
}

func TestUpdateContainer_RemoveContainerError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockIContainerRepository(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	mockDockerClient := mocks.NewMockIDockerClient(ctrl)
	mockService := NewContainerService(mockRepo, mockLogger, mockDockerClient)

	ctx := context.TODO()
	id := uint(5)
	oldContainerID := "xyz789"
	containerName := "my-container"
	oldImage := "alpine:latest"
	newImage := "alpine:3.15"

	updateData := map[string]interface{}{
		"image_name": newImage,
	}

	container := &model.Container{
		ID:            id,
		ContainerID:   oldContainerID,
		ContainerName: containerName,
		ImageName:     oldImage,
	}

	removeErr := fmt.Errorf("remove failure")

	mockRepo.
		EXPECT().
		GetContainerByID(ctx, id).
		Return(container, nil)

	mockDockerClient.
		EXPECT().
		StopContainer(ctx, oldContainerID).
		Return(nil)

	mockDockerClient.
		EXPECT().
		RemoveContainer(ctx, oldContainerID).
		Return(removeErr)

	mockLogger.
		EXPECT().
		Error("Failed to remove Docker container before updating image", "containerID", oldContainerID, "error", removeErr)

	result, err := mockService.UpdateContainer(ctx, id, updateData)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to remove Docker container before updating image")
}

func TestUpdateContainer_StartContainerError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockIContainerRepository(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	mockDockerClient := mocks.NewMockIDockerClient(ctrl)
	mockService := NewContainerService(mockRepo, mockLogger, mockDockerClient)

	ctx := context.TODO()
	id := uint(6)
	oldContainerID := "xyz789"
	containerName := "my-container"
	oldImage := "alpine:latest"
	newImage := "alpine:3.15"

	updateData := map[string]interface{}{
		"image_name": newImage,
	}

	container := &model.Container{
		ID:            id,
		ContainerID:   oldContainerID,
		ContainerName: containerName,
		ImageName:     oldImage,
	}

	startErr := fmt.Errorf("start failure")

	mockRepo.
		EXPECT().
		GetContainerByID(ctx, id).
		Return(container, nil)

	mockDockerClient.
		EXPECT().
		StopContainer(ctx, oldContainerID).
		Return(nil)

	mockDockerClient.
		EXPECT().
		RemoveContainer(ctx, oldContainerID).
		Return(nil)

	mockDockerClient.
		EXPECT().
		StartContainer(ctx, containerName, newImage).
		Return("", startErr)

	mockLogger.
		EXPECT().
		// s.logger.Error("Failed to start Docker container with new image", "containerName", container.ContainerName, "image", image, "error", err)
		Error("Failed to start Docker container with new image", "containerName", containerName, "image", newImage, "error", startErr)

	result, err := mockService.UpdateContainer(ctx, id, updateData)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to start Docker container with new image")
}

func TestUpdateContainer_UpdateStatusToRunning(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockIContainerRepository(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	mockDockerClient := mocks.NewMockIDockerClient(ctrl)
	mockService := NewContainerService(mockRepo, mockLogger, mockDockerClient)

	ctx := context.TODO()
	id := uint(10)
	containerID := "cid123"
	containerName := "web-app"
	currentStatus := "stopped"
	newStatus := "running"

	updateData := map[string]interface{}{
		"status": newStatus,
	}

	existing := &model.Container{
		ID:            id,
		ContainerID:   containerID,
		ContainerName: containerName,
		ImageName:     "nginx",
		Status:        currentStatus,
	}

	mockRepo.EXPECT().
		GetContainerByID(ctx, id).
		Return(existing, nil)

	mockDockerClient.EXPECT().
		StartExistingContainer(ctx, containerID).
		Return(nil)

	mockRepo.EXPECT().
		UpdateContainer(ctx, id, gomock.Any()).
		DoAndReturn(func(_ context.Context, _ uint, data map[string]interface{}) (*model.Container, error) {
			assert.Equal(t, newStatus, data["status"])
			return &model.Container{
				ID:            id,
				ContainerID:   containerID,
				ContainerName: containerName,
				ImageName:     "nginx",
				Status:        newStatus,
			}, nil
		})

	mockLogger.
		EXPECT().
		Info("Container updated successfully", "id", id, "containerID", containerID)

	updated, err := mockService.UpdateContainer(ctx, id, updateData)

	assert.NoError(t, err)
	assert.Equal(t, newStatus, updated.Status)
}

func TestUpdateContainer_UpdateStatusToStopped(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockIContainerRepository(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	mockDockerClient := mocks.NewMockIDockerClient(ctrl)
	mockService := NewContainerService(mockRepo, mockLogger, mockDockerClient)

	ctx := context.TODO()
	id := uint(11)
	containerID := "cid456"
	containerName := "api-service"
	currentStatus := "running"
	newStatus := "stopped"

	updateData := map[string]interface{}{
		"status": newStatus,
	}

	existing := &model.Container{
		ID:            id,
		ContainerID:   containerID,
		ContainerName: containerName,
		ImageName:     "api:latest",
		Status:        currentStatus,
	}

	mockRepo.EXPECT().
		GetContainerByID(ctx, id).
		Return(existing, nil)

	mockDockerClient.EXPECT().
		StopContainer(ctx, containerID).
		Return(nil)

	mockRepo.EXPECT().
		UpdateContainer(ctx, id, gomock.Any()).
		DoAndReturn(func(_ context.Context, _ uint, data map[string]interface{}) (*model.Container, error) {
			assert.Equal(t, newStatus, data["status"])
			return &model.Container{
				ID:            id,
				ContainerID:   containerID,
				ContainerName: containerName,
				ImageName:     "api:latest",
				Status:        newStatus,
			}, nil
		})

	mockLogger.
		EXPECT().
		Info("Container updated successfully", "id", id, "containerID", containerID)

	updated, err := mockService.UpdateContainer(ctx, id, updateData)

	assert.NoError(t, err)
	assert.Equal(t, newStatus, updated.Status)
}

func TestUpdateContainer_StartExistingContainerError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockIContainerRepository(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	mockDockerClient := mocks.NewMockIDockerClient(ctrl)
	mockService := NewContainerService(mockRepo, mockLogger, mockDockerClient)

	ctx := context.TODO()
	id := uint(12)
	containerID := "cid789"
	containerName := "db-service"
	currentStatus := "stopped"
	newStatus := "running"

	updateData := map[string]interface{}{
		"status": newStatus,
	}

	existing := &model.Container{
		ID:            id,
		ContainerID:   containerID,
		ContainerName: containerName,
		ImageName:     "db:latest",
		Status:        currentStatus,
	}

	startErr := fmt.Errorf("failed to start existing container")

	mockRepo.EXPECT().
		GetContainerByID(ctx, id).
		Return(existing, nil)

	mockDockerClient.EXPECT().
		StartExistingContainer(ctx, containerID).
		Return(startErr)

	mockLogger.
		EXPECT().
		Error("Failed to start Docker container", "containerID", containerID, "error", startErr)

	result, err := mockService.UpdateContainer(ctx, id, updateData)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to start Docker container")
}

func TestUpdateContainer_Status_StopContainerError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockIContainerRepository(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	mockDockerClient := mocks.NewMockIDockerClient(ctrl)
	mockService := NewContainerService(mockRepo, mockLogger, mockDockerClient)

	ctx := context.TODO()
	id := uint(13)
	containerID := "cid123"
	containerName := "test-container"
	currentStatus := "running"
	newStatus := "stopped"

	updateData := map[string]interface{}{
		"status": newStatus,
	}

	existing := &model.Container{
		ID:            id,
		ContainerID:   containerID,
		ContainerName: containerName,
		ImageName:     "test-image",
		Status:        currentStatus,
	}

	stopErr := fmt.Errorf("failed to stop container")

	mockRepo.EXPECT().
		GetContainerByID(ctx, id).
		Return(existing, nil)

	mockDockerClient.EXPECT().
		StopContainer(ctx, containerID).
		Return(stopErr)

	mockLogger.
		EXPECT().
		Error("Failed to stop Docker container", "containerID", containerID, "error", stopErr)

	result, err := mockService.UpdateContainer(ctx, id, updateData)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to stop Docker container")
}

func TestUpdateContainer_UpdateContainerRepoErr(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockIContainerRepository(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	mockDockerClient := mocks.NewMockIDockerClient(ctrl)
	mockService := NewContainerService(mockRepo, mockLogger, mockDockerClient)

	ctx := context.TODO()
	id := uint(14)
	containerID := "cid123"
	containerName := "test-container"
	currentStatus := "running"
	newStatus := "stopped"

	updateData := map[string]interface{}{
		"status": newStatus,
	}

	existing := &model.Container{
		ID:            id,
		ContainerID:   containerID,
		ContainerName: containerName,
		ImageName:     "test-image",
		Status:        currentStatus,
	}

	repoErr := fmt.Errorf("repository error")

	mockRepo.EXPECT().
		GetContainerByID(ctx, id).
		Return(existing, nil)

	mockDockerClient.EXPECT().
		StopContainer(ctx, containerID).
		Return(nil)

	mockRepo.EXPECT().
		UpdateContainer(ctx, id, gomock.Any()).
		Return(nil, repoErr)

	mockLogger.
		EXPECT().
		Error("Failed to update container in repository", "id", id, "error", repoErr)

	result, err := mockService.UpdateContainer(ctx, id, updateData)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to update container in repository")
}

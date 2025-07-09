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

func TestCreateContainer_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockIContainerRepository(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	mockDockerClient := mocks.NewMockIDockerClient(ctrl)

	mockService := NewContainerService(mockRepo, mockLogger, mockDockerClient)
	ctx := context.TODO()
	containerName := "test-container"
	imageName := "test-image"
	containerID := "abc123"
	expectedID := 1

	mockDockerClient.
		EXPECT().
		StartContainer(ctx, containerName, imageName).
		Return(containerID, nil)

	mockRepo.
		EXPECT().
		CreateContainer(ctx, gomock.Any()).
		DoAndReturn(func(_ context.Context, c *model.Container) (int, error) {
			assert.Equal(t, containerName, c.ContainerName)
			assert.Equal(t, imageName, c.ImageName)
			assert.Equal(t, containerID, c.ContainerID)
			return expectedID, nil
		})

	mockLogger.
		EXPECT().
		Info("Container created successfully", "databaseID", expectedID, "containerID", containerID)

	id, err := mockService.CreateContainer(ctx, containerName, imageName)

	assert.NoError(t, err)
	assert.Equal(t, expectedID, id)
}

func TestCreateContainer_DockerStartError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockIContainerRepository(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	mockDockerClient := mocks.NewMockIDockerClient(ctrl)

	mockService := NewContainerService(mockRepo, mockLogger, mockDockerClient)

	ctx := context.TODO()
	containerName := "fail-container"
	imageName := "fail-image"
	startErr := fmt.Errorf("docker error")

	mockDockerClient.
		EXPECT().
		StartContainer(ctx, containerName, imageName).
		Return("", startErr)

	mockLogger.
		EXPECT().
		Error("Failed to start Docker container", "error", startErr)

	id, err := mockService.CreateContainer(ctx, containerName, imageName)

	assert.Error(t, err)
	assert.Equal(t, 0, id)
	assert.Contains(t, err.Error(), "failed to start Docker container")
}

func TestCreateContainer_RepoCreateError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockIContainerRepository(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	mockDockerClient := mocks.NewMockIDockerClient(ctrl)

	mockService := NewContainerService(mockRepo, mockLogger, mockDockerClient)

	ctx := context.TODO()
	containerName := "test-container"
	imageName := "test-image"
	containerID := "xyz789"
	repoErr := fmt.Errorf("db error")

	mockDockerClient.
		EXPECT().
		StartContainer(ctx, containerName, imageName).
		Return(containerID, nil)

	mockRepo.
		EXPECT().
		CreateContainer(ctx, gomock.Any()).
		Return(0, repoErr)

	mockDockerClient.
		EXPECT().
		StopContainer(ctx, containerID).
		Return(nil)

	mockLogger.
		EXPECT().
		Error("Failed to create container in repository", "error", repoErr, "container", gomock.Any())

	mockLogger.
		EXPECT().
		Info("Stopped Docker container after repository creation failure", "containerID", containerID)

	id, err := mockService.CreateContainer(ctx, containerName, imageName)

	assert.Error(t, err)
	assert.Equal(t, 0, id)
	assert.Contains(t, err.Error(), "failed to create container in repository")
}

func TestCreateContainer_RepoCreateError_StopContainerError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockIContainerRepository(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	mockDockerClient := mocks.NewMockIDockerClient(ctrl)

	mockService := NewContainerService(mockRepo, mockLogger, mockDockerClient)

	ctx := context.TODO()
	containerName := "test_container"
	imageName := "test_image"
	containerID := "xyz789"
	repoErr := fmt.Errorf("db error")
	stopErr := fmt.Errorf("stop error")

	mockDockerClient.
		EXPECT().
		StartContainer(ctx, containerName, imageName).
		Return(containerID, nil)

	mockRepo.
		EXPECT().
		CreateContainer(ctx, gomock.Any()).
		Return(0, repoErr)

	mockDockerClient.
		EXPECT().
		StopContainer(ctx, containerID).
		Return(stopErr)

	mockLogger.
		EXPECT().
		Error("Failed to create container in repository", "error", repoErr, "container", gomock.Any())

	mockLogger.
		EXPECT().
		Error("Failed to stop Docker container after repository creation failure", "containerID", containerID, "error", stopErr)

	id, err := mockService.CreateContainer(ctx, containerName, imageName)

	assert.Error(t, err)
	assert.Equal(t, 0, id)
	assert.Contains(t, err.Error(), "failed to create container in repository")
}

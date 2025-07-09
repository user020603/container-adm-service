package service

import (
	"context"
	"fmt"
	"testing"
	"thanhnt208/container-adm-service/internal/dto"
	"thanhnt208/container-adm-service/internal/mocks"
	"thanhnt208/container-adm-service/internal/model"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestViewAllContainers_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockIContainerRepository(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	mockDockerClient := mocks.NewMockIDockerClient(ctrl)

	mockService := NewContainerService(mockRepo, mockLogger, mockDockerClient)

	ctx := context.TODO()
	filter := &dto.ContainerFilter{Status: "running"}
	from := 0
	to := 10
	sortBy := "created_at"
	sortOrder := "desc"

	expectedCount := int64(2)
	expectedContainers := []model.Container{
		{ID: 1, ContainerName: "container-1", ImageName: "nginx"},
		{ID: 2, ContainerName: "container-2", ImageName: "redis"},
	}

	mockRepo.
		EXPECT().
		ViewAllContainers(ctx, filter, from, to, sortBy, sortOrder).
		Return(expectedCount, expectedContainers, nil)

	mockLogger.
		EXPECT().
		Info("Retrieved containers successfully", "count", expectedCount)

	count, containers, err := mockService.ViewAllContainers(ctx, filter, from, to, sortBy, sortOrder)

	assert.NoError(t, err)
	assert.Equal(t, expectedCount, count)
	assert.Equal(t, expectedContainers, containers)
}

func TestViewAllContainers_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockIContainerRepository(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	mockDockerClient := mocks.NewMockIDockerClient(ctrl)

	mockService := NewContainerService(mockRepo, mockLogger, mockDockerClient)

	ctx := context.TODO()
	filter := &dto.ContainerFilter{}
	from := 0
	to := 10
	sortBy := "name"
	sortOrder := "asc"

	repoErr := fmt.Errorf("db failure")

	mockRepo.
		EXPECT().
		ViewAllContainers(ctx, filter, from, to, sortBy, sortOrder).
		Return(int64(0), nil, repoErr)

	mockLogger.
		EXPECT().
		Error("Failed to retrieve containers", "error", repoErr)

	count, containers, err := mockService.ViewAllContainers(ctx, filter, from, to, sortBy, sortOrder)

	assert.Error(t, err)
	assert.Equal(t, int64(0), count)
	assert.Nil(t, containers)
	assert.Contains(t, err.Error(), "failed to retrieve containers")
}

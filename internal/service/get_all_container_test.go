package service

import (
	"context"
	"errors"
	"testing"
	"thanhnt208/container-adm-service/internal/dto"
	"thanhnt208/container-adm-service/internal/mocks"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestGetAllContainers_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockIContainerRepository(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)

	mockService := NewContainerService(mockRepo, mockLogger, nil)

	expected := []dto.ContainerName{
		{Id: 1, ContainerName: "nginx"},
		{Id: 2, ContainerName: "redis"},
	}

	mockRepo.EXPECT().
		GetContainerInfo(gomock.Any()).
		Return(expected, nil)

	result, err := mockService.GetAllContainers(context.TODO())

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestGetAllContainers_RepoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockIContainerRepository(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)

	mockService := NewContainerService(mockRepo, mockLogger, nil)

	mockRepo.EXPECT().
		GetContainerInfo(gomock.Any()).
		Return(nil, errors.New("database error"))

	mockLogger.EXPECT().
		Error("Failed to retrieve container names", "error", gomock.Any())

	result, err := mockService.GetAllContainers(context.TODO())

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to retrieve container names")
}

package service

import (
	"context"
	"errors"
	"testing"
	"thanhnt208/container-adm-service/internal/mocks"
	"thanhnt208/container-adm-service/internal/model"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestExportContainers_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockIContainerRepository(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)

	service := NewContainerService(mockRepo, mockLogger, nil)

	containers := []model.Container{
		{ID: 1, ContainerID: "abc123", ContainerName: "test1", ImageName: "nginx", Status: "running"},
		{ID: 2, ContainerID: "def456", ContainerName: "test2", ImageName: "redis", Status: "stopped"},
	}

	mockRepo.EXPECT().
		ViewAllContainers(gomock.Any(), gomock.Nil(), 0, 10, "", "").
		Return(int64(len(containers)), containers, nil)

	mockLogger.EXPECT().
		Info("Containers exported successfully", "count", len(containers))

	result, err := service.ExportContainers(context.TODO(), nil, 0, 10, "", "")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Contains(t, result.FileName, "containers_")
	assert.Equal(t, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", result.ContentType)
	assert.True(t, len(result.Data) > 0)
}

func TestExportContainers_ViewAllContainersError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockIContainerRepository(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)

	service := NewContainerService(mockRepo, mockLogger, nil)

	mockRepo.EXPECT().
		ViewAllContainers(gomock.Any(), gomock.Nil(), 0, 10, "", "").
		Return(int64(0), nil, errors.New("db error"))

	mockLogger.EXPECT().
		Error("Failed to retrieve containers for export", "error", gomock.Any())

	result, err := service.ExportContainers(context.TODO(), nil, 0, 10, "", "")
	assert.Error(t, err)
	assert.Nil(t, result)
}

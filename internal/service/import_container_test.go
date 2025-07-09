package service

import (
	"context"
	"fmt"
	"testing"
	"thanhnt208/container-adm-service/internal/mocks"
	"thanhnt208/container-adm-service/internal/model"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/xuri/excelize/v2"
)

func TestImportContainers_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockIContainerRepository(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	mockDockerClient := mocks.NewMockIDockerClient(ctrl)
	service := NewContainerService(mockRepo, mockLogger, mockDockerClient)

	f := excelize.NewFile()
	sheet := f.GetSheetName(0)
	_ = f.SetSheetRow(sheet, "A1", &[]interface{}{"container_name", "image_name"})
	_ = f.SetSheetRow(sheet, "A2", &[]interface{}{"test-container", "nginx:latest"})
	buf, _ := f.WriteToBuffer()

	mockDockerClient.EXPECT().
		StartContainer(gomock.Any(), "test-container", "nginx:latest").
		Return("docker-id-1", nil)

	mockRepo.EXPECT().
		CreateManyContainers(gomock.Any(), gomock.Any()).
		Return([]model.Container{{ID: 123, ContainerName: "test-container"}}, nil, nil)

	mockLogger.EXPECT().
		Info("Containers imported successfully", "successfulCount", 1, "failedCount", 0)

	result, err := service.ImportContainers(context.TODO(), buf.Bytes())

	assert.NoError(t, err)
	assert.Equal(t, 1, result.SuccessfulCount)
	assert.Equal(t, 0, result.FailedCount)
	assert.Contains(t, result.SuccessfulItems[0], "test-container")
}

func TestImportContainers_InvalidExcel(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockIContainerRepository(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	mockDockerClient := mocks.NewMockIDockerClient(ctrl)
	service := NewContainerService(mockRepo, mockLogger, mockDockerClient)

	invalidData := []byte("not excel")

	mockLogger.EXPECT().
		Error("Failed to open Excel file", "error", gomock.Any())

	result, err := service.ImportContainers(context.TODO(), invalidData)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestImportContainers_EmptyData(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockIContainerRepository(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	mockDockerClient := mocks.NewMockIDockerClient(ctrl)
	service := NewContainerService(mockRepo, mockLogger, mockDockerClient)

	f := excelize.NewFile()
	sheet := f.GetSheetName(0)
	_ = f.SetSheetRow(sheet, "A1", &[]interface{}{"container_name", "image_name"})
	buf, _ := f.WriteToBuffer()

	mockLogger.EXPECT().
		Warn("No data found in Excel file")

	result, err := service.ImportContainers(context.TODO(), buf.Bytes())

	assert.NoError(t, err)
	assert.Equal(t, 0, result.SuccessfulCount)
	assert.Equal(t, 0, result.FailedCount)
}

func TestImportContainers_MissingFields(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockIContainerRepository(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	mockDockerClient := mocks.NewMockIDockerClient(ctrl)
	service := NewContainerService(mockRepo, mockLogger, mockDockerClient)

	f := excelize.NewFile()
	sheet := f.GetSheetName(0)
	_ = f.SetSheetRow(sheet, "A1", &[]interface{}{"container_name", "image_name"})
	_ = f.SetSheetRow(sheet, "A2", &[]interface{}{"", "nginx:latest"})
	buf, _ := f.WriteToBuffer()

	mockLogger.EXPECT().
		Error("No valid containers to import", "errors", gomock.Any())

	result, err := service.ImportContainers(context.TODO(), buf.Bytes())

	assert.NoError(t, err)
	assert.Equal(t, 0, result.SuccessfulCount)
	assert.Equal(t, 1, result.FailedCount)
	assert.Contains(t, result.FailedItems[0], "Missing required fields")
}

func TestImportContainers_DockerStartError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockIContainerRepository(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	mockDockerClient := mocks.NewMockIDockerClient(ctrl)
	service := NewContainerService(mockRepo, mockLogger, mockDockerClient)

	f := excelize.NewFile()
	sheet := f.GetSheetName(0)
	_ = f.SetSheetRow(sheet, "A1", &[]interface{}{"container_name", "image_name"})
	_ = f.SetSheetRow(sheet, "A2", &[]interface{}{"test-container", "nginx:latest"})
	buf, _ := f.WriteToBuffer()

	mockDockerClient.EXPECT().
		StartContainer(gomock.Any(), "test-container", "nginx:latest").
		Return("", fmt.Errorf("docker error"))

	mockLogger.EXPECT().
		Error("Failed to start Docker container", "error", gomock.Any())

	mockLogger.EXPECT().
		Error("No valid containers to import", "errors", gomock.Any())

	result, err := service.ImportContainers(context.TODO(), buf.Bytes())

	assert.NoError(t, err)
	assert.Equal(t, 0, result.SuccessfulCount)
	assert.Equal(t, 1, result.FailedCount)
	assert.Contains(t, result.FailedItems[0], "Failed to start container")
}

func TestImportContainers_RepoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockIContainerRepository(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	mockDockerClient := mocks.NewMockIDockerClient(ctrl)
	service := NewContainerService(mockRepo, mockLogger, mockDockerClient)

	f := excelize.NewFile()
	sheet := f.GetSheetName(0)
	_ = f.SetSheetRow(sheet, "A1", &[]interface{}{"container_name", "image_name"})
	_ = f.SetSheetRow(sheet, "A2", &[]interface{}{"test-container", "nginx:latest"})
	buf, _ := f.WriteToBuffer()

	mockDockerClient.EXPECT().
		StartContainer(gomock.Any(), "test-container", "nginx:latest").
		Return("docker-id", nil)

	mockRepo.EXPECT().
		CreateManyContainers(gomock.Any(), gomock.Any()).
		Return(nil, nil, fmt.Errorf("repo fail"))

	mockLogger.EXPECT().
		Error("Repository failed during CreateManyContainers", "error", gomock.Any())

	result, err := service.ImportContainers(context.TODO(), buf.Bytes())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "repository level")
	assert.Equal(t, 0, result.SuccessfulCount)
	assert.Equal(t, 1, result.FailedCount)
	assert.Contains(t, result.FailedItems[0], "repository error")
}

func TestImportContainers_RowTooShort(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockIContainerRepository(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	mockDockerClient := mocks.NewMockIDockerClient(ctrl)
	service := NewContainerService(mockRepo, mockLogger, mockDockerClient)

	f := excelize.NewFile()
	sheet := f.GetSheetName(0)
	_ = f.SetSheetRow(sheet, "A1", &[]interface{}{"container_name", "image_name"})
	_ = f.SetSheetRow(sheet, "A2", &[]interface{}{"only-one-column"})
	buf, _ := f.WriteToBuffer()

	mockLogger.EXPECT().
		Error("No valid containers to import", "errors", gomock.Any())

	result, err := service.ImportContainers(context.TODO(), buf.Bytes())

	assert.NoError(t, err)
	assert.Equal(t, 0, result.SuccessfulCount)
	assert.Equal(t, 1, result.FailedCount)
	assert.Contains(t, result.FailedItems[0], "Not enough columns")
}

func TestImportContainers_OpenReaderError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockIContainerRepository(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	mockDockerClient := mocks.NewMockIDockerClient(ctrl)
	service := NewContainerService(mockRepo, mockLogger, mockDockerClient)

	invalidExcel := []byte("not an excel file")

	mockLogger.EXPECT().
		Error("Failed to open Excel file", "error", gomock.Any())

	result, err := service.ImportContainers(context.TODO(), invalidExcel)

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open Excel file")
}

func TestImportContainers_FailedRepoImportsHandled(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockIContainerRepository(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	mockDockerClient := mocks.NewMockIDockerClient(ctrl)
	service := NewContainerService(mockRepo, mockLogger, mockDockerClient)

	f := excelize.NewFile()
	sheet := f.GetSheetName(0)
	_ = f.SetSheetRow(sheet, "A1", &[]interface{}{"container_name", "image_name"})
	_ = f.SetSheetRow(sheet, "A2", &[]interface{}{"ctn-fail", "nginx:1.21"})
	buf, _ := f.WriteToBuffer()

	mockDockerClient.
		EXPECT().
		StartContainer(gomock.Any(), "ctn-fail", "nginx:1.21").
		Return("def999", nil)

	mockRepo.
		EXPECT().
		CreateManyContainers(gomock.Any(), gomock.Any()).
		Return([]model.Container{}, []model.Container{
			{
				ContainerName: "ctn-fail",
				ContainerID:   "def999",
				ImageName:     "nginx:1.21",
			},
		}, nil)

	mockLogger.EXPECT().
		Info("Containers imported successfully", "successfulCount", 0, "failedCount", 1)

	result, err := service.ImportContainers(context.TODO(), buf.Bytes())

	assert.NoError(t, err)
	assert.Equal(t, 0, result.SuccessfulCount)
	assert.Equal(t, 1, result.FailedCount)
	assert.Contains(t, result.FailedItems[0], "ctn-fail (repository error)")
}

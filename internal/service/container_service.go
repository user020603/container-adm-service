package service

import (
	"bytes"
	"context"
	"fmt"
	"thanhnt208/container-adm-service/external/client"
	"thanhnt208/container-adm-service/internal/dto"
	"thanhnt208/container-adm-service/internal/model"
	"thanhnt208/container-adm-service/internal/repository"
	"thanhnt208/container-adm-service/pkg/logger"
	"time"

	"github.com/xuri/excelize/v2"
)

type IContainerService interface {
	CreateContainer(ctx context.Context, containerName, imageName string) (int, error)
	ViewAllContainers(ctx context.Context, containerFilter *dto.ContainerFilter, from, to int, sortBy string, sortOrder string) (int64, []model.Container, error)
	UpdateContainer(ctx context.Context, id uint, updateData map[string]interface{}) (*model.Container, error)
	DeleteContainer(ctx context.Context, id uint) error
	ImportContainers(ctx context.Context, buf []byte) (*dto.ImportResult, error)
	ExportContainers(ctx context.Context, containerFilter *dto.ContainerFilter, from, to int, sortBy, sortOrder string) (*dto.ExportData, error)
}

type containerService struct {
	repo         repository.IContainerRepository
	logger       logger.ILogger
	dockerClient client.IDockerClient
}

func NewContainerService(repo repository.IContainerRepository, logger logger.ILogger, dockerClient client.IDockerClient) IContainerService {
	return &containerService{
		repo:         repo,
		logger:       logger,
		dockerClient: dockerClient,
	}
}

func (s *containerService) CreateContainer(ctx context.Context, containerName, imageName string) (int, error) {
	containerID, err := s.dockerClient.StartContainer(ctx, containerName, imageName)
	if err != nil {
		s.logger.Error("Failed to start Docker container", "error", err)
		return 0, fmt.Errorf("failed to start Docker container: %w", err)
	}

	container := &model.Container{
		ContainerID:   containerID,
		ContainerName: containerName,
		ImageName:     imageName,
		Status:        "running",
	}

	id, err := s.repo.CreateContainer(ctx, container)
	if err != nil {
		s.logger.Error("Failed to create container in repository", "error", err, "container", container)
		if stopErr := s.dockerClient.StopContainer(ctx, containerID); stopErr != nil {
			s.logger.Error("Failed to stop Docker container after repository creation failure", "containerID", containerID, "error", stopErr)
		} else {
			s.logger.Info("Stopped Docker container after repository creation failure", "containerID", containerID)
		}
		return 0, fmt.Errorf("failed to create container in repository: %w", err)
	}

	s.logger.Info("Container created successfully", "databaseID", id, "containerID", containerID)
	return id, nil
}

func (s *containerService) ViewAllContainers(ctx context.Context, containerFilter *dto.ContainerFilter, from, to int, sortBy string, sortOrder string) (int64, []model.Container, error) {
	count, containers, err := s.repo.ViewAllContainers(ctx, containerFilter, from, to, sortBy, sortOrder)
	if err != nil {
		s.logger.Error("Failed to retrieve containers", "error", err)
		return 0, nil, fmt.Errorf("failed to retrieve containers: %w", err)
	}

	s.logger.Info("Retrieved containers successfully", "count", count)
	return count, containers, nil
}

func (s *containerService) UpdateContainer(ctx context.Context, id uint, updateData map[string]interface{}) (*model.Container, error) {
	container, err := s.repo.UpdateContainer(ctx, id, updateData)
	if err != nil {
		s.logger.Error("Failed to update container", "id", id, "error", err)
		return nil, fmt.Errorf("failed to update container: %w", err)
	}

	s.logger.Info("Container updated successfully", "id", id)
	return container, nil
}

func (s *containerService) DeleteContainer(ctx context.Context, id uint) error {
	container, err := s.repo.GetContainerByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to retrieve container for deletion", "id", id, "error", err)
		return fmt.Errorf("failed to retrieve container for deletion: %w", err)
	}

	if container == nil {
		s.logger.Warn("Container not found for deletion", "id", id)
		return fmt.Errorf("container with ID %d not found", id)
	}

	if err := s.dockerClient.StopContainer(ctx, container.ContainerID); err != nil {
		s.logger.Error("Failed to stop Docker container before deletion", "containerID", container.ContainerID, "error", err)
		return fmt.Errorf("failed to stop Docker container before deletion: %w", err)
	}

	if err := s.dockerClient.RemoveContainer(ctx, container.ContainerID); err != nil {
		s.logger.Error("Failed to remove Docker container", "containerID", container.ContainerID, "error", err)
		return fmt.Errorf("failed to remove Docker container: %w", err)
	}

	if err := s.repo.DeleteContainer(ctx, id); err != nil {
		s.logger.Error("Failed to delete container from repository", "id", id, "error", err)
		return fmt.Errorf("failed to delete container from repository: %w", err)
	}

	s.logger.Info("Container deleted successfully", "id", id)
	return nil
}

func (s *containerService) ImportContainers(ctx context.Context, buf []byte) (*dto.ImportResult, error) {
	f, err := excelize.OpenReader(bytes.NewReader(buf))
	if err != nil {
		s.logger.Error("Failed to open Excel file", "error", err)
		return nil, fmt.Errorf("failed to open Excel file: %w", err)
	}
	defer f.Close()

	rows, err := f.GetRows("Sheet1")
	if err != nil {
		s.logger.Error("Failed to get rows from Excel file", "error", err)
		return nil, fmt.Errorf("failed to get rows from Excel file: %w", err)
	}

	if len(rows) < 2 {
		s.logger.Warn("No data found in Excel file")
		return &dto.ImportResult{SuccessfulCount: 0, FailedCount: 0}, nil
	}

	var containersToCreate []model.Container
	var parsingErrors []string

	for i, row := range rows[1:] {
		rowNum := i + 2
		if len(row) < 2 {
			parsingErrors = append(parsingErrors, fmt.Sprintf("Row %d: Not enough columns", rowNum))
			continue
		}

		containerName := row[0]
		imageName := row[1]

		if containerName == "" || imageName == "" {
			parsingErrors = append(parsingErrors, fmt.Sprintf("Row %d: Missing required fields", rowNum))
			continue
		}

		containerID, err := s.dockerClient.StartContainer(ctx, containerName, imageName)
		if err != nil {
			s.logger.Error("Failed to start Docker container", "error", err)
			parsingErrors = append(parsingErrors, fmt.Sprintf("Row %d: Failed to start container - %s", rowNum, err.Error()))
			continue
		}

		container := model.Container{
			ContainerID:   containerID,
			ContainerName: containerName,
			ImageName:     imageName,
			Status:        "running",
		}
		containersToCreate = append(containersToCreate, container)
	}

	if len(containersToCreate) == 0 && len(parsingErrors) > 0 {
		s.logger.Error("No valid containers to import", "errors", parsingErrors)
		return &dto.ImportResult{
			SuccessfulCount: 0,
			FailedCount:     len(parsingErrors),
			FailedItems:     parsingErrors,
		}, nil
	}

	if len(containersToCreate) == 0 && len(parsingErrors) == 0 {
		s.logger.Warn("No valid data found in Excel file")
		return &dto.ImportResult{SuccessfulCount: 0, FailedCount: 0}, nil
	}

	successfulRepoImports, failedRepoImports, err := s.repo.CreateManyContainers(ctx, containersToCreate)
	if err != nil {
		s.logger.Error("Repository failed during CreateManyContainers", "error", err)
		failedItemsFromRepo := make([]string, len(containersToCreate))
		for i, ctn := range containersToCreate {
			failedItemsFromRepo[i] = fmt.Sprintf("%s (repository error)", ctn.ContainerName)
		}
		return &dto.ImportResult{
			FailedCount: len(containersToCreate),
			FailedItems: append(parsingErrors, failedItemsFromRepo...),
		}, fmt.Errorf("failed to import containers at repository level: %w", err)
	}

	result := &dto.ImportResult{
		SuccessfulCount: len(successfulRepoImports),
		SuccessfulItems: make([]string, len(successfulRepoImports)),
		FailedCount:     len(failedRepoImports) + len(parsingErrors),
		FailedItems:     make([]string, 0, len(failedRepoImports)+len(parsingErrors)),
	}

	for i, ctn := range successfulRepoImports {
		result.SuccessfulItems[i] = fmt.Sprintf("%s (ID: %d)", ctn.ContainerName, ctn.ID)
	}
	result.FailedItems = append(result.FailedItems, parsingErrors...)
	for _, ctn := range failedRepoImports {
		result.FailedItems = append(result.FailedItems, fmt.Sprintf("%s (repository error)", ctn.ContainerName))
	}

	s.logger.Info("Containers imported successfully", "successfulCount", result.SuccessfulCount, "failedCount", result.FailedCount)
	return result, nil
}

func (s *containerService) ExportContainers(ctx context.Context, containerFilter *dto.ContainerFilter, from, to int, sortBy, sortOrder string) (*dto.ExportData, error) {
	_, containers, err := s.repo.ViewAllContainers(ctx, containerFilter, from, to, sortBy, sortOrder)
	if err != nil {
		s.logger.Error("Failed to retrieve containers for export", "error", err)
		return nil, fmt.Errorf("failed to retrieve containers for export: %w", err)
	}

	f := excelize.NewFile()
	sheetName := "containers_export"

	if err := f.SetSheetName("Sheet1", sheetName); err != nil {
		s.logger.Error("Failed to set sheet name", "error", err)
		return nil, fmt.Errorf("failed to set sheet name: %w", err)
	}

	cols := []string{"ID", "Container ID", "Container Name", "Image Name", "Status"}
	for i, col := range cols {
		cell, err := excelize.CoordinatesToCellName(i+1, 1)
		if err != nil {
			s.logger.Error("Failed to convert coordinates to cell name", "error", err, "column", i+1)
			return nil, fmt.Errorf("failed to convert coordinates to cell name: %w", err)
		}
		if err := f.SetCellValue(sheetName, cell, col); err != nil {
			s.logger.Error("Failed to set cell value", "error", err, "cell", cell)
			return nil, fmt.Errorf("failed to set cell value: %w", err)
		}
	}

	for i, container := range containers {
		rowNum := i + 2
		values := []interface{}{
			container.ID,
			container.ContainerID,
			container.ContainerName,
			container.ImageName,
			container.Status,
		}

		for colIdx, value := range values {
			cell := fmt.Sprintf("%c%d", 'A'+colIdx, rowNum)
			if err := f.SetCellValue(sheetName, cell, value); err != nil {
				s.logger.Error("Failed to set cell value", "error", err, "cell", cell)
				return nil, fmt.Errorf("failed to set cell value: %w", err)
			}
		}
	}

	var buf bytes.Buffer
	err = f.Write(&buf)
	if err != nil {
		s.logger.Error("Failed to write Excel file", "error", err)
		return nil, fmt.Errorf("failed to write Excel file: %w", err)
	}

	s.logger.Info("Containers exported successfully", "count", len(containers))

	exportResult := &dto.ExportData{
		FileName:    fmt.Sprintf("containers_%s.xlsx", time.Now().Format("2006-01-02_15-04-05")),
		ContentType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		Data:        buf.Bytes(),
	}

	return exportResult, nil
}

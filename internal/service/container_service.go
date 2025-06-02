package service

import (
	"context"
	"fmt"
	"thanhnt208/container-adm-service/external/client"
	"thanhnt208/container-adm-service/internal/model"
	"thanhnt208/container-adm-service/internal/repository"
	"thanhnt208/container-adm-service/pkg/logger"
)

type IContainerService interface {
	CreateContainer(ctx context.Context, containerName, imageName string) (int, error)
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
		s.logger.Error("Failed to create container metadata", "error", err)

		if stopErr := s.dockerClient.StopContainer(ctx, containerID); stopErr != nil {
			s.logger.Error("Failed to cleanup container after database error", "containerID", containerID, "error", stopErr)
		}

		return 0, fmt.Errorf("failed to create container metadata: %w", err)
	}

	s.logger.Info("Container created successfully", "databaseID", id, "containerID", containerID)
	return id, nil
}

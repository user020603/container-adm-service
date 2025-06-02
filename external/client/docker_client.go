package client

import (
	"context"
	"fmt"
	"io"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
)

type IDockerClient interface {
	StartContainer(ctx context.Context, containerName, imageName string) (string, error)
	StopContainer(ctx context.Context, containerID string) error
}

type dockerClient struct {
	client *client.Client
}

func NewDockerClient() (IDockerClient, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}
	return &dockerClient{client: cli}, nil
}

func (d *dockerClient) StartContainer(ctx context.Context, containerName, imageName string) (string, error) {
	out, err := d.client.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to pull image %s: %w", imageName, err)
	}
	defer out.Close()
	io.Copy(io.Discard, out)

	resp, err := d.client.ContainerCreate(
		ctx,
		&container.Config{
			Image: imageName,
		},
		nil,
		nil,
		nil,
		containerName,
	)

	if err != nil {
		return "", fmt.Errorf("failed to create container %s: %w", containerName, err)
	}

	if err := d.client.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		// Clean up the container if it fails to start
		if removeErr := d.client.ContainerRemove(ctx, resp.ID, container.RemoveOptions{Force: true}); removeErr != nil {
			return "", fmt.Errorf("failed to remove container %s after start failure: %w", resp.ID, removeErr)
		}
		return "", fmt.Errorf("failed to start container %s: %w", containerName, err)
	}

	return resp.ID, nil
}

func (d *dockerClient) StopContainer(ctx context.Context, containerID string) error {
	if err := d.client.ContainerStop(ctx, containerID, container.StopOptions{}); err != nil {
		return fmt.Errorf("failed to stop container %s: %w", containerID, err)
	}
	return nil
}

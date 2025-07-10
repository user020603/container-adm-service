package client

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDockerClient_StartContainer_Success(t *testing.T) {
	dockerCli, err := NewDockerClient()
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	containerID, err := dockerCli.StartContainer(ctx, "test-container", "alpine")
	require.NoError(t, err)
	require.NotEmpty(t, containerID)

	err = dockerCli.StopContainer(ctx, containerID)
	require.NoError(t, err)

	err = dockerCli.RemoveContainer(ctx, containerID)
	require.NoError(t, err)
}

func TestDockerClient_StartContainer_InvalidImage(t *testing.T) {
	dockerCli, err := NewDockerClient()
	require.NoError(t, err)

	ctx := context.Background()

	containerID, err := dockerCli.StartContainer(ctx, "bad-container", "nonexistent:image123")
	require.Error(t, err)
	require.Empty(t, containerID)
}

func TestDockerClient_StopContainer_NotExist(t *testing.T) {
	dockerCli, err := NewDockerClient()
	require.NoError(t, err)

	ctx := context.Background()
	err = dockerCli.StopContainer(ctx, "non-existent-id")
	require.Error(t, err)
}

func TestDockerClient_StartExistingContainer_Invalid(t *testing.T) {
	dockerCli, err := NewDockerClient()
	require.NoError(t, err)

	ctx := context.Background()
	err = dockerCli.StartExistingContainer(ctx, "invalid-id")
	require.Error(t, err)
}

func TestDockerClient_RemoveContainer_NotExist(t *testing.T) {
	dockerCli, err := NewDockerClient()
	require.NoError(t, err)

	ctx := context.Background()
	err = dockerCli.RemoveContainer(ctx, "non-existent-id")
	require.Error(t, err)
}

func TestDockerClient_StartContainer_StartFailure(t *testing.T) {
	dockerCli, err := NewDockerClient()
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	containerID, err := dockerCli.StartContainer(ctx, "quick-exit", "busybox")
	require.NoError(t, err)
	require.NotEmpty(t, containerID)

	time.Sleep(1 * time.Second)

	err = dockerCli.StartExistingContainer(ctx, containerID)
	require.NoError(t, err)

	_ = dockerCli.StopContainer(ctx, containerID)
	_ = dockerCli.RemoveContainer(ctx, containerID)
}

func TestDockerClient_StartContainer_StartFailure_CleanupSuccess(t *testing.T) {
	dockerCli, err := NewDockerClient()
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	containerID, err := dockerCli.StartContainer(ctx, "failing-container", "alpine")
	require.NoError(t, err)
	require.NotEmpty(t, containerID)

	_ = dockerCli.StopContainer(ctx, containerID)

	_ = dockerCli.RemoveContainer(ctx, containerID)

	containerID, _ = dockerCli.StartContainer(ctx, "failing-start", "busybox:latest")
	require.Empty(t, containerID)
}

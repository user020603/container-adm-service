package grpc

import (
	"context"
	"errors"
	"testing"
	"thanhnt208/container-adm-service/internal/dto"
	"thanhnt208/container-adm-service/internal/mocks"
	"thanhnt208/container-adm-service/proto/pb"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestGetContainerUptimeDuration_RequestNil(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockIContainerService(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	handler := NewGrpcServerHandler(mockService, mockLogger)
	ctx := context.Background()

	mockLogger.EXPECT().Error("GetContainerUptimeDuration: request cannot be nil")

	resp, err := handler.GetContainerUptimeDuration(ctx, nil)
	assert.Nil(t, resp)
	assert.EqualError(t, err, "request cannot be nil")
}

func TestGetContainerUptimeDuration_InvalidTimestamps(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockIContainerService(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	handler := NewGrpcServerHandler(mockService, mockLogger)
	ctx := context.Background()

	req := &pb.GetContainerInfomationRequest{StartTime: 0, EndTime: 100}
	mockLogger.EXPECT().Error("GetContainerUptimeDuration: startTime and endTime must be greater than 0", "startTime", int64(0), "endTime", int64(100))

	resp, err := handler.GetContainerUptimeDuration(ctx, req)
	assert.Nil(t, resp)
	assert.EqualError(t, err, "startTime and endTime must be greater than 0")
}

func TestGetContainerUptimeDuration_StartAfterEnd(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockIContainerService(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	handler := NewGrpcServerHandler(mockService, mockLogger)
	ctx := context.Background()

	req := &pb.GetContainerInfomationRequest{StartTime: 200, EndTime: 100}
	mockLogger.EXPECT().Error("GetContainerUptimeDuration: startTime must be less than endTime", "startTime", int64(200), "endTime", int64(100))

	resp, err := handler.GetContainerUptimeDuration(ctx, req)
	assert.Nil(t, resp)
	assert.EqualError(t, err, "startTime must be less than endTime")
}

func TestGetContainerUptimeDuration_NumContainersError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockIContainerService(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	handler := NewGrpcServerHandler(mockService, mockLogger)
	ctx := context.Background()

	req := &pb.GetContainerInfomationRequest{StartTime: 100, EndTime: 200}
	mockService.EXPECT().GetNumContainers(ctx).Return(int64(0), errors.New("db error"))
	mockLogger.EXPECT().Error("GetContainerUptimeDuration: failed to get number of containers", "error", gomock.Any())

	resp, err := handler.GetContainerUptimeDuration(ctx, req)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "failed to get number of containers")
}

func TestGetContainerUptimeDuration_NumRunningError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockIContainerService(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	handler := NewGrpcServerHandler(mockService, mockLogger)
	ctx := context.Background()

	req := &pb.GetContainerInfomationRequest{StartTime: 100, EndTime: 200}
	mockService.EXPECT().GetNumContainers(ctx).Return(int64(5), nil)
	mockService.EXPECT().GetNumRunningContainers(ctx).Return(int64(0), errors.New("run err"))
	mockLogger.EXPECT().Error("GetContainerUptimeDuration: failed to get number of running containers", "error", gomock.Any())

	resp, err := handler.GetContainerUptimeDuration(ctx, req)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "failed to get number of running containers")
}

func TestGetContainerUptimeDuration_GetUptimeError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockIContainerService(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	handler := NewGrpcServerHandler(mockService, mockLogger)
	ctx := context.Background()

	req := &pb.GetContainerInfomationRequest{StartTime: 100, EndTime: 200}
	mockService.EXPECT().GetNumContainers(ctx).Return(int64(3), nil)
	mockService.EXPECT().GetNumRunningContainers(ctx).Return(int64(2), nil)
	mockService.EXPECT().GetContainerUptimeDuration(ctx, time.Unix(100, 0).UTC(), time.Unix(200, 0).UTC()).Return(nil, errors.New("uptime fail"))
	mockLogger.EXPECT().Error("GetContainerUptimeDuration: failed to get uptime duration", "error", gomock.Any())

	resp, err := handler.GetContainerUptimeDuration(ctx, req)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "failed to get uptime duration")
}

func TestGetContainerUptimeDuration_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockIContainerService(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	handler := NewGrpcServerHandler(mockService, mockLogger)
	ctx := context.Background()

	req := &pb.GetContainerInfomationRequest{StartTime: 100, EndTime: 200}
	startTime := time.Unix(100, 0).UTC()
	endTime := time.Unix(200, 0).UTC()

	mockService.EXPECT().GetNumContainers(ctx).Return(int64(4), nil)
	mockService.EXPECT().GetNumRunningContainers(ctx).Return(int64(2), nil)

	uptimeDetails := &dto.UptimeDetails{
		TotalUptime: 10 * time.Second,
		PerContainerUptime: map[string]time.Duration{
			"c1": 3 * time.Second,
			"c2": 7 * time.Second,
		},
	}

	mockService.EXPECT().GetContainerUptimeDuration(ctx, startTime, endTime).Return(uptimeDetails, nil)

	mockLogger.EXPECT().Info(
		"GetContainerUptimeDuration: successfully retrieved uptime duration",
		"numContainers", int64(4),
		"numRunningContainers", int64(2),
		"numStoppedContainers", int64(2),
		"totalUptime", int64(10_000),
	)

	resp, err := handler.GetContainerUptimeDuration(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, int64(4), resp.NumContainers)
	assert.Equal(t, int64(2), resp.NumRunningContainers)
	assert.Equal(t, int64(2), resp.NumStoppedContainers)
	assert.Equal(t, int64(10_000), resp.UptimeDetails.TotalUptime)
	assert.Equal(t, int64(3000), resp.UptimeDetails.PerContainerUptime["c1"])
	assert.Equal(t, int64(7000), resp.UptimeDetails.PerContainerUptime["c2"])
}

func TestGetContainerUptimeDuration_StoppedNegative(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockIContainerService(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	handler := NewGrpcServerHandler(mockService, mockLogger)
	ctx := context.Background()

	req := &pb.GetContainerInfomationRequest{StartTime: 100, EndTime: 200}
	mockService.EXPECT().GetNumContainers(ctx).Return(int64(2), nil)
	mockService.EXPECT().GetNumRunningContainers(ctx).Return(int64(5), nil)

	mockService.EXPECT().GetContainerUptimeDuration(ctx, time.Unix(100, 0).UTC(), time.Unix(200, 0).UTC()).Return(
		&dto.UptimeDetails{
			TotalUptime: 1 * time.Second,
			PerContainerUptime: map[string]time.Duration{
				"c1": 1 * time.Second,
			},
		},
		nil,
	)

	mockLogger.EXPECT().Info(
		"GetContainerUptimeDuration: successfully retrieved uptime duration",
		"numContainers", int64(2),
		"numRunningContainers", int64(5),
		"numStoppedContainers", int64(0), // clamped
		"totalUptime", int64(1000),
	)

	resp, err := handler.GetContainerUptimeDuration(ctx, req)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), resp.NumStoppedContainers)
}

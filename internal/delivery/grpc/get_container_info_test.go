package grpc

import (
	"context"
	"errors"
	"testing"
	"thanhnt208/container-adm-service/internal/mocks"
	"thanhnt208/container-adm-service/proto/pb"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestGetContainerInformation_RequestNil(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockIContainerService(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)

	handler := NewGrpcServerHandler(mockService, mockLogger)
	ctx := context.Background()

	mockLogger.EXPECT().Error("GetContainerInformation: request cannot be nil")

	resp, err := handler.GetContainerInformation(ctx, nil)

	assert.Nil(t, resp)
	assert.EqualError(t, err, "request cannot be nil")
}

func TestGetContainerInformation_InvalidTimestamps(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockIContainerService(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)

	handler := NewGrpcServerHandler(mockService, mockLogger)
	ctx := context.Background()

	req := &pb.GetContainerInfomationRequest{
		StartTime: 0,
		EndTime:   100,
	}

	mockLogger.EXPECT().Error(
		"GetContainerInformation: startTime and endTime must be greater than 0",
		"startTime", int64(0),
		"endTime", int64(100),
	)

	resp, err := handler.GetContainerInformation(ctx, req)

	assert.Nil(t, resp)
	assert.EqualError(t, err, "startTime and endTime must be greater than 0")
}

func TestGetContainerInformation_StartAfterEnd(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockIContainerService(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)

	handler := NewGrpcServerHandler(mockService, mockLogger)
	ctx := context.Background()

	req := &pb.GetContainerInfomationRequest{
		StartTime: 200,
		EndTime:   100,
	}

	mockLogger.EXPECT().Error(
		"GetContainerInformation: startTime must be less than endTime",
		"startTime", int64(200),
		"endTime", int64(100),
	)

	resp, err := handler.GetContainerInformation(ctx, req)

	assert.Nil(t, resp)
	assert.EqualError(t, err, "startTime must be less than endTime")
}

func TestGetContainerInformation_NumContainersError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockIContainerService(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)

	handler := NewGrpcServerHandler(mockService, mockLogger)
	ctx := context.Background()

	req := &pb.GetContainerInfomationRequest{
		StartTime: 100,
		EndTime:   200,
	}

	errExpected := errors.New("db error")

	mockService.EXPECT().GetNumContainers(ctx).Return(int64(0), errExpected)
	mockLogger.EXPECT().Error("GetContainerInformation: failed to get number of containers", "error", errExpected)

	resp, err := handler.GetContainerInformation(ctx, req)

	assert.Nil(t, resp)
	assert.EqualError(t, err, "failed to get number of containers: db error")
}

func TestGetContainerInformation_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockIContainerService(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)

	handler := NewGrpcServerHandler(mockService, mockLogger)
	ctx := context.Background()

	req := &pb.GetContainerInfomationRequest{
		StartTime: 100,
		EndTime:   200,
	}

	mockService.EXPECT().GetNumContainers(ctx).Return(int64(10), nil)
	mockService.EXPECT().GetNumRunningContainers(ctx).Return(int64(7), nil)
	mockService.EXPECT().GetContainerUptimeRatio(ctx, time.Unix(100, 0), time.Unix(200, 0)).Return(0.85, nil)

	mockLogger.EXPECT().Info(
		gomock.Eq("GetContainerInformation: successfully retrieved container information"),
		gomock.Any(),
	)

	resp, err := handler.GetContainerInformation(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, int64(10), resp.NumContainers)
	assert.Equal(t, int64(7), resp.NumRunningContainers)
	assert.Equal(t, int64(3), resp.NumStoppedContainers)
	assert.Equal(t, float32(0.85), resp.MeanUptimeRatio)
}

func TestGetContainerInformation_GetNumRunningContainersError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockIContainerService(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	handler := NewGrpcServerHandler(mockService, mockLogger)

	ctx := context.Background()
	req := &pb.GetContainerInfomationRequest{StartTime: 100, EndTime: 200}

	mockService.EXPECT().GetNumContainers(ctx).Return(int64(10), nil)
	mockService.EXPECT().GetNumRunningContainers(ctx).Return(int64(0), errors.New("db error"))

	mockLogger.EXPECT().Error(
		"GetContainerInformation: failed to get number of running containers",
		"error", gomock.Any(),
	)

	resp, err := handler.GetContainerInformation(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "failed to get number of running containers")
}

func TestGetContainerInformation_GetUptimeRatioError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockIContainerService(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	handler := NewGrpcServerHandler(mockService, mockLogger)

	ctx := context.Background()
	req := &pb.GetContainerInfomationRequest{StartTime: 100, EndTime: 200}

	mockService.EXPECT().GetNumContainers(ctx).Return(int64(5), nil)
	mockService.EXPECT().GetNumRunningContainers(ctx).Return(int64(2), nil)
	mockService.EXPECT().GetContainerUptimeRatio(ctx, time.Unix(100, 0), time.Unix(200, 0)).Return(0.0, errors.New("ratio error"))

	mockLogger.EXPECT().Error(
		"GetContainerInformation: failed to get uptime ratio",
		"error", gomock.Any(),
	)

	resp, err := handler.GetContainerInformation(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "failed to get uptime ratio")
}

func TestGetContainerInformation_InvalidUptimeRatio(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockIContainerService(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	handler := NewGrpcServerHandler(mockService, mockLogger)

	ctx := context.Background()
	req := &pb.GetContainerInfomationRequest{StartTime: 100, EndTime: 200}

	mockService.EXPECT().GetNumContainers(ctx).Return(int64(5), nil)
	mockService.EXPECT().GetNumRunningContainers(ctx).Return(int64(3), nil)
	mockService.EXPECT().GetContainerUptimeRatio(ctx, time.Unix(100, 0), time.Unix(200, 0)).Return(1.5, nil)

	mockLogger.EXPECT().Error(
		"GetContainerInformation: uptime ratio must be between 0 and 1",
		"uptimeRatio", 1.5,
	)

	resp, err := handler.GetContainerInformation(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "uptime ratio must be between 0 and 1")
}

func TestGetContainerInformation_StoppedContainersNegative(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockIContainerService(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	handler := NewGrpcServerHandler(mockService, mockLogger)

	ctx := context.Background()
	req := &pb.GetContainerInfomationRequest{StartTime: 100, EndTime: 200}

	mockService.EXPECT().GetNumContainers(ctx).Return(int64(3), nil)
	mockService.EXPECT().GetNumRunningContainers(ctx).Return(int64(5), nil) // > total
	mockService.EXPECT().GetContainerUptimeRatio(ctx, time.Unix(100, 0), time.Unix(200, 0)).Return(0.5, nil)

	mockLogger.EXPECT().Info(
		gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
		gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
		gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
		gomock.Any(),
	)

	resp, err := handler.GetContainerInformation(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, int64(3), resp.NumContainers)
	assert.Equal(t, int64(5), resp.NumRunningContainers)
	assert.Equal(t, int64(0), resp.NumStoppedContainers) // expected to be clamped
}

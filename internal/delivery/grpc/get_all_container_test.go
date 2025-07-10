package grpc

import (
	"context"
	"errors"
	"testing"
	"thanhnt208/container-adm-service/internal/dto"
	"thanhnt208/container-adm-service/internal/mocks"
	"thanhnt208/container-adm-service/proto/pb"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestGetAllContainers_RequestNil(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockIContainerService(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)

	handler := NewGrpcServerHandler(mockService, mockLogger)

	ctx := context.Background()

	mockLogger.EXPECT().Error("GetAllContainers: request cannot be nil")

	resp, err := handler.GetAllContainers(ctx, nil)

	assert.Nil(t, resp)
	assert.EqualError(t, err, "request cannot be nil")
}

func TestGetAllContainers_ServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockIContainerService(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)

	handler := NewGrpcServerHandler(mockService, mockLogger)

	ctx := context.Background()
	expectedErr := errors.New("db error")

	mockService.EXPECT().GetAllContainers(ctx).Return(nil, expectedErr)
	mockLogger.EXPECT().Error("GetAllContainers: failed to retrieve containers", "error", expectedErr)

	resp, err := handler.GetAllContainers(ctx, &pb.EmptyRequest{})

	assert.Nil(t, resp)
	assert.EqualError(t, err, "failed to retrieve containers: db error")
}

func TestGetAllContainers_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockIContainerService(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)

	handler := NewGrpcServerHandler(mockService, mockLogger)

	ctx := context.Background()

	mockContainers := []dto.ContainerName{
		{Id: 1, ContainerName: "container-1"},
		{Id: 2, ContainerName: "container-2"},
	}

	mockService.EXPECT().GetAllContainers(ctx).Return(mockContainers, nil)
	mockLogger.EXPECT().Info("GetAllContainers: successfully retrieved containers", "count", 2)

	resp, err := handler.GetAllContainers(ctx, &pb.EmptyRequest{})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Containers, 2)
	assert.Equal(t, uint64(1), resp.Containers[0].Id)
	assert.Equal(t, "container-1", resp.Containers[0].ContainerName)
	assert.Equal(t, uint64(2), resp.Containers[1].Id)
	assert.Equal(t, "container-2", resp.Containers[1].ContainerName)
}

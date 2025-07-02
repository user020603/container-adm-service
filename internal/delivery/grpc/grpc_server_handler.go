package grpc

import (
	"context"
	"fmt"
	"thanhnt208/container-adm-service/internal/service"
	"thanhnt208/container-adm-service/pkg/logger"
	"thanhnt208/container-adm-service/proto/pb"
	"time"
)

type GrpcServerHandler struct {
	service service.IContainerService
	pb.UnimplementedContainerAdmServiceServer
	logger logger.ILogger
}

func NewGrpcServerHandler(service service.IContainerService, logger logger.ILogger) *GrpcServerHandler {
	return &GrpcServerHandler{
		service: service,
		logger:  logger,
	}
}

func (h *GrpcServerHandler) GetAllContainers(ctx context.Context, req *pb.EmptyRequest) (*pb.ContainerResponse, error) {
	if req == nil {
		h.logger.Error("GetAllContainers: request cannot be nil")
		return nil, fmt.Errorf("request cannot be nil")
	}

	containers, err := h.service.GetAllContainers(ctx)
	if err != nil {
		h.logger.Error("GetAllContainers: failed to retrieve containers", "error", err)
		return nil, fmt.Errorf("failed to retrieve containers: %w", err)
	}

	containerName := make([]*pb.ContainerName, len(containers))
	for i, container := range containers {
		containerName[i] = &pb.ContainerName{
			Id:            uint64(container.Id),
			ContainerName: container.ContainerName,
		}
	}

	h.logger.Info("GetAllContainers: successfully retrieved containers", "count", len(containerName))

	return &pb.ContainerResponse{
		Containers: containerName,
	}, nil
}

func (h *GrpcServerHandler) GetContainerInformation(ctx context.Context, req *pb.GetContainerInfomationRequest) (*pb.GetContainerInfomationResponse, error) {
	if req == nil {
		h.logger.Error("GetContainerInformation: request cannot be nil")
		return nil, fmt.Errorf("request cannot be nil")
	}

	startTime := req.GetStartTime()
	endTime := req.GetEndTime()

	if startTime <= 0 || endTime <= 0 {
		h.logger.Error("GetContainerInformation: startTime and endTime must be greater than 0", "startTime", startTime, "endTime", endTime)
		return nil, fmt.Errorf("startTime and endTime must be greater than 0")
	}

	if startTime >= endTime {
		h.logger.Error("GetContainerInformation: startTime must be less than endTime", "startTime", startTime, "endTime", endTime)
		return nil, fmt.Errorf("startTime must be less than endTime")
	}

	numContainers, err := h.service.GetNumContainers(ctx)
	if err != nil {
		h.logger.Error("GetContainerInformation: failed to get number of containers", "error", err)
		return nil, fmt.Errorf("failed to get number of containers: %w", err)
	}

	numRunningContainers, err := h.service.GetNumRunningContainers(ctx)
	if err != nil {
		h.logger.Error("GetContainerInformation: failed to get number of running containers", "error", err)
		return nil, fmt.Errorf("failed to get number of running containers: %w", err)
	}

	numStoppedContainers := numContainers - numRunningContainers
	if numStoppedContainers < 0 {
		numStoppedContainers = 0
	}

	startTimeObj := time.Unix(startTime, 0)
	endTimeObj := time.Unix(endTime, 0)

	uptimeRatio, err := h.service.GetContainerUptimeRatio(ctx, startTimeObj, endTimeObj)
	if err != nil {
		h.logger.Error("GetContainerInformation: failed to get uptime ratio", "error", err)
		return nil, fmt.Errorf("failed to get uptime ratio: %w", err)
	}

	if uptimeRatio < 0 || uptimeRatio > 1 {
		h.logger.Error("GetContainerInformation: uptime ratio must be between 0 and 1", "uptimeRatio", uptimeRatio)
		return nil, fmt.Errorf("uptime ratio must be between 0 and 1")
	}

	h.logger.Info("GetContainerInformation: successfully retrieved container information",
		"numContainers", numContainers,
		"numRunningContainers", numRunningContainers,
		"numStoppedContainers", numStoppedContainers,
		"meanUptimeRatio", uptimeRatio,
		"startTime", startTimeObj,
		"endTime", endTimeObj,
	)

	return &pb.GetContainerInfomationResponse{
		NumContainers:        int64(numContainers),
		NumRunningContainers: int64(numRunningContainers),
		NumStoppedContainers: int64(numStoppedContainers),
		MeanUptimeRatio:      float32(uptimeRatio),
	}, nil
}

func (h *GrpcServerHandler) GetContainerUptimeDuration(ctx context.Context, req *pb.GetContainerInfomationRequest) (*pb.GetContainerUptimeDurationResponse, error) {
    if req == nil {
        h.logger.Error("GetContainerUptimeDuration: request cannot be nil")
        return nil, fmt.Errorf("request cannot be nil")
    }

    startTime := req.GetStartTime()
    endTime := req.GetEndTime()

    if startTime <= 0 || endTime <= 0 {
        h.logger.Error("GetContainerUptimeDuration: startTime and endTime must be greater than 0", "startTime", startTime, "endTime", endTime)
        return nil, fmt.Errorf("startTime and endTime must be greater than 0")
    }

    if startTime >= endTime {
        h.logger.Error("GetContainerUptimeDuration: startTime must be less than endTime", "startTime", startTime, "endTime", endTime)
        return nil, fmt.Errorf("startTime must be less than endTime")
    }

    numContainers, err := h.service.GetNumContainers(ctx)
    if err != nil {
        h.logger.Error("GetContainerUptimeDuration: failed to get number of containers", "error", err)
        return nil, fmt.Errorf("failed to get number of containers: %w", err)
    }

    numRunningContainers, err := h.service.GetNumRunningContainers(ctx)
    if err != nil {
        h.logger.Error("GetContainerUptimeDuration: failed to get number of running containers", "error", err)
        return nil, fmt.Errorf("failed to get number of running containers: %w", err)
    }

    numStoppedContainers := numContainers - numRunningContainers
    if numStoppedContainers < 0 {
        numStoppedContainers = 0
    }

    startTimeObj := time.Unix(startTime, 0).UTC()
    endTimeObj := time.Unix(endTime, 0).UTC()

    uptimeDetails, err := h.service.GetContainerUptimeDuration(ctx, startTimeObj, endTimeObj)
    if err != nil {
        h.logger.Error("GetContainerUptimeDuration: failed to get uptime duration", "error", err)
        return nil, fmt.Errorf("failed to get uptime duration: %w", err)
    }

    perContainerUptime := make(map[string]int64)
    for k, v := range uptimeDetails.PerContainerUptime {
        perContainerUptime[k] = int64(v.Milliseconds())
    }

    h.logger.Info("GetContainerUptimeDuration: successfully retrieved uptime duration",
        "numContainers", numContainers,
        "numRunningContainers", numRunningContainers,
        "numStoppedContainers", numStoppedContainers,
        "totalUptime", int64(uptimeDetails.TotalUptime.Milliseconds()),
    )

    return &pb.GetContainerUptimeDurationResponse{
        NumContainers:        int64(numContainers),
        NumRunningContainers: int64(numRunningContainers),
        NumStoppedContainers: int64(numStoppedContainers),
        UptimeDetails: &pb.ContainerUptimeDetails{
            TotalUptime:       int64(uptimeDetails.TotalUptime.Milliseconds()),
            PerContainerUptime: perContainerUptime,
        },
    }, nil
}


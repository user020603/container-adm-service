package rest

import (
	"net/http"
	"strconv"
	"strings"
	"thanhnt208/container-adm-service/internal/dto"
	"thanhnt208/container-adm-service/internal/service"
	"thanhnt208/container-adm-service/pkg/logger"

	"github.com/gin-gonic/gin"
)

const (
	MaxRangeLimit = 1000 // Maximum range limit for pagination
)

type RestServerHandler struct {
	service service.IContainerService
	logger  logger.ILogger
}

func NewRestServerHandler(service service.IContainerService, logger logger.ILogger) *RestServerHandler {
	return &RestServerHandler{
		service: service,
		logger:  logger,
	}
}

func (h *RestServerHandler) respondWithError(c *gin.Context, statusCode int, message string, err error) {
	if err != nil {
		h.logger.Error(message, "error", err)
	} else {
		h.logger.Error(message)
	}
	c.JSON(statusCode, gin.H{
		"error": message,
	})
}

func (h *RestServerHandler) respondWithSuccess(c *gin.Context, statusCode int, data gin.H) {
	c.JSON(statusCode, data)
}

func (h *RestServerHandler) CreateContainer(c *gin.Context) {
	var req dto.CreateContainerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondWithError(c, http.StatusBadRequest, "Invalid request data", err)
		return
	}

	if strings.TrimSpace(req.ContainerName) == "" || strings.TrimSpace(req.ImageName) == "" {
		h.respondWithError(c, http.StatusBadRequest, "Container name and image name are required", nil)
		return
	}

	id, err := h.service.CreateContainer(c, req.ContainerName, req.ImageName)
	if err != nil {
		h.respondWithError(c, http.StatusInternalServerError, "Failed to create container", err)
		return
	}

	h.respondWithSuccess(c, http.StatusCreated, gin.H{
		"message": "Container created successfully",
		"id":      id,
	})
}

func (h *RestServerHandler) ViewContainers(c *gin.Context) {
	var filter *dto.ContainerFilter
	if err := c.ShouldBindJSON(&filter); err != nil {
		h.respondWithError(c, http.StatusBadRequest, "Invalid filter data", err)
		return
	}

	from, err := strconv.Atoi(c.DefaultQuery("from", "0"))
	if err != nil || from < 0 {
		h.respondWithError(c, http.StatusBadRequest, "Invalid 'from' parameter", err)
		return
	}

	to, err := strconv.Atoi(c.DefaultQuery("to", "100"))
	if err != nil || to < from {
		h.respondWithError(c, http.StatusBadRequest, "Invalid 'to' parameter", err)
		return
	}

	if to-from > MaxRangeLimit {
		h.respondWithError(c, http.StatusBadRequest, "Range limit exceeded", nil)
		return
	}

	sortBy := c.DefaultQuery("sortBy", "id")
	sortOrder := strings.ToUpper(c.DefaultQuery("sortOrder", "ASC"))

	if sortOrder != "ASC" && sortOrder != "DESC" {
		h.respondWithError(c, http.StatusBadRequest, "Invalid sort order", nil)
		return
	}

	count, containers, err := h.service.ViewAllContainers(c, filter, from, to, sortBy, sortOrder)
	if err != nil {
		h.respondWithError(c, http.StatusInternalServerError, "Failed to retrieve containers", err)
		return
	}

	h.respondWithSuccess(c, http.StatusOK, gin.H{
		"count":      count,
		"containers": containers,
	})
}

func (h *RestServerHandler) UpdateContainer(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		h.respondWithError(c, http.StatusBadRequest, "ID is required", nil)
		return
	}

	idUint, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		h.respondWithError(c, http.StatusBadRequest, "Invalid ID format", err)
		return
	}

	var updateReq map[string]interface{}
	if err := c.ShouldBindJSON(&updateReq); err != nil {
		h.respondWithError(c, http.StatusBadRequest, "Invalid update data", err)
		return
	}

	if len(updateReq) == 0 {
		h.respondWithError(c, http.StatusBadRequest, "No fields to update", nil)
		return
	}

	updatedData, err := h.service.UpdateContainer(c, uint(idUint), updateReq)
	if err != nil {
		h.respondWithError(c, http.StatusInternalServerError, "Failed to update container", err)
		return
	}

	h.respondWithSuccess(c, http.StatusOK, gin.H{
		"message": "Container updated successfully",
		"id":      idUint,
		"data":    updatedData,
	})
}

func (h *RestServerHandler) DeleteContainer(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		h.respondWithError(c, http.StatusBadRequest, "ID is required", nil)
		return
	}

	idUint, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		h.respondWithError(c, http.StatusBadRequest, "Invalid ID format", err)
		return
	}

	err = h.service.DeleteContainer(c, uint(idUint))
	if err != nil {
		h.respondWithError(c, http.StatusInternalServerError, "Failed to delete container", err)
		return
	}

	h.respondWithSuccess(c, http.StatusOK, gin.H{
		"message": "Container deleted successfully",
	})
}

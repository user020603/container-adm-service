package rest

import (
	"io"
	"net/http"
	"strconv"
	"strings"
	"thanhnt208/container-adm-service/internal/dto"
	"thanhnt208/container-adm-service/internal/service"
	"thanhnt208/container-adm-service/pkg/logger"

	"github.com/gin-gonic/gin"
)

const (
	MaxRangeLimit     = 100             // Maximum range limit for pagination
	MaxFileSize       = 10 * 1024 * 1024 // Maximum file size for import (10 MB)
	SupportedFileType = ".xlsx"          // Supported file type for import
)

type RestContainerHandler struct {
	service service.IContainerService
	logger  logger.ILogger
}

func NewRestServerHandler(service service.IContainerService, logger logger.ILogger) *RestContainerHandler {
	return &RestContainerHandler{
		service: service,
		logger:  logger,
	}
}

func (h *RestContainerHandler) respondWithError(c *gin.Context, statusCode int, message string, err error) {
	if err != nil {
		h.logger.Error(message, "error", err)
	} else {
		h.logger.Error(message)
	}
	c.JSON(statusCode, gin.H{
		"error": message,
	})
}

func (h *RestContainerHandler) respondWithSuccess(c *gin.Context, statusCode int, data gin.H) {
	c.JSON(statusCode, data)
}

func (h *RestContainerHandler) CreateContainer(c *gin.Context) {
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

func (h *RestContainerHandler) ViewContainers(c *gin.Context) {
	var filter dto.ContainerFilter
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

	if to-from >= MaxRangeLimit {
		h.respondWithError(c, http.StatusBadRequest, "Range limit exceeded", nil)
		return
	}

	sortBy := c.DefaultQuery("sortBy", "id")
	sortOrder := strings.ToUpper(c.DefaultQuery("sortOrder", "ASC"))

	if sortOrder != "ASC" && sortOrder != "DESC" {
		h.respondWithError(c, http.StatusBadRequest, "Invalid sort order", nil)
		return
	}

	count, containers, err := h.service.ViewAllContainers(c, &filter, from, to, sortBy, sortOrder)
	if err != nil {
		h.respondWithError(c, http.StatusInternalServerError, "Failed to retrieve containers", err)
		return
	}

	h.respondWithSuccess(c, http.StatusOK, gin.H{
		"count":      count,
		"containers": containers,
	})
}

func (h *RestContainerHandler) UpdateContainer(c *gin.Context) {
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

func (h *RestContainerHandler) DeleteContainer(c *gin.Context) {
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

func (h *RestContainerHandler) ImportContainers(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		h.respondWithError(c, http.StatusBadRequest, "File is required", err)
		return
	}

	if file.Size > MaxFileSize {
		h.respondWithError(c, http.StatusBadRequest, "File size exceeds the maximum limit of 10 MB", nil)
		return
	}

	if file.Size == 0 {
		h.respondWithError(c, http.StatusBadRequest, "File is empty", nil)
		return
	}

	if !strings.HasSuffix(strings.ToLower(file.Filename), SupportedFileType) {
		h.respondWithError(c, http.StatusBadRequest, "Unsupported file type, only .xlsx is allowed", nil)
		return
	}

	fileContent, err := file.Open()
	if err != nil {
		h.respondWithError(c, http.StatusInternalServerError, "Failed to open file", err)
		return
	}
	defer fileContent.Close()

	buf, err := io.ReadAll(fileContent)
	if err != nil {
		h.respondWithError(c, http.StatusInternalServerError, "Failed to read file content", err)
		return
	}

	var importResult *dto.ImportResult
	importResult, err = h.service.ImportContainers(c, buf)
	if err != nil {
		h.respondWithError(c, http.StatusInternalServerError, "Failed to import containers", err)
		return
	}
	if importResult == nil {
		h.respondWithError(c, http.StatusInternalServerError, "Import result is nil", nil)
		return
	}

	h.respondWithSuccess(c, http.StatusOK, gin.H{
		"message":          "Containers imported successfully",
		"file_name":        file.Filename,
		"successful_count": importResult.SuccessfulCount,
		"successful_items": importResult.SuccessfulItems,
		"failed_count":     importResult.FailedCount,
		"failed_items":     importResult.FailedItems,
	})
}

func (h *RestContainerHandler) ExportContainers(c *gin.Context) {
	var filter dto.ContainerFilter
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

	if to-from >= MaxRangeLimit {
		h.respondWithError(c, http.StatusBadRequest, "Range limit exceeded", nil)
		return
	}

	sortBy := c.DefaultQuery("sortBy", "id")
	sortOrder := strings.ToUpper(c.DefaultQuery("sortOrder", "ASC"))

	if sortOrder != "ASC" && sortOrder != "DESC" {
		h.respondWithError(c, http.StatusBadRequest, "Invalid sort order", nil)
		return
	}

	var exportData *dto.ExportData
	exportData, err = h.service.ExportContainers(c, &filter, from, to, sortBy, sortOrder)
	if err != nil {
		h.respondWithError(c, http.StatusInternalServerError, "Failed to export containers", err)
		return
	}
	if exportData == nil {
		h.respondWithError(c, http.StatusInternalServerError, "Export data is nil", nil)
		return
	}

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Access-Control-Expose-Headers", "Content-Disposition")
	c.Header("Content-Disposition", "attachment; filename="+exportData.FileName)
	c.Header("File-Name", exportData.FileName)
	c.Status(http.StatusOK)
	c.Writer.Write(exportData.Data)
}

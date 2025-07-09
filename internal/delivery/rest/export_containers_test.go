package rest

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"thanhnt208/container-adm-service/internal/dto"
	"thanhnt208/container-adm-service/internal/mocks"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestExportContainers_InvalidFilter(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockILogger(ctrl)
	mockService := mocks.NewMockIContainerService(ctrl)
	handler := NewRestServerHandler(mockService, mockLogger)

	body := strings.NewReader("invalid-json")
	req := httptest.NewRequest(http.MethodPost, "/export?from=0&to=10", body)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	mockLogger.EXPECT().Error("Invalid filter data", "error", gomock.Any())

	handler.ExportContainers(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid filter data")
}

func TestExportContainers_InvalidFromParam(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockILogger(ctrl)
	mockService := mocks.NewMockIContainerService(ctrl)
	handler := NewRestServerHandler(mockService, mockLogger)

	body := strings.NewReader(`{}`)
	req := httptest.NewRequest(http.MethodPost, "/export?from=abc", body)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	mockLogger.EXPECT().Error("Invalid 'from' parameter", "error", gomock.Any())

	handler.ExportContainers(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid 'from' parameter")
}

func TestExportContainers_InvalidToParam(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockILogger(ctrl)
	mockService := mocks.NewMockIContainerService(ctrl)
	handler := NewRestServerHandler(mockService, mockLogger)

	body := strings.NewReader(`{}`)
	req := httptest.NewRequest(http.MethodPost, "/export?from=10&to=5", body)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	mockLogger.EXPECT().Error("Invalid 'to' parameter")

	handler.ExportContainers(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid 'to' parameter")
}

func TestExportContainers_RangeLimitExceeded(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockILogger(ctrl)
	mockService := mocks.NewMockIContainerService(ctrl)
	handler := NewRestServerHandler(mockService, mockLogger)

	body := strings.NewReader(`{}`)
	req := httptest.NewRequest(http.MethodPost, "/export?from=0&to=1000", body)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	mockLogger.EXPECT().Error("Range limit exceeded")

	handler.ExportContainers(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Range limit exceeded")
}

func TestExportContainers_InvalidSortOrder(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockILogger(ctrl)
	mockService := mocks.NewMockIContainerService(ctrl)
	handler := NewRestServerHandler(mockService, mockLogger)

	body := strings.NewReader(`{}`)
	req := httptest.NewRequest(http.MethodPost, "/export?from=0&to=10&sortOrder=INVALID", body)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	mockLogger.EXPECT().Error("Invalid sort order")

	handler.ExportContainers(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid sort order")
}

func TestExportContainers_ServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockILogger(ctrl)
	mockService := mocks.NewMockIContainerService(ctrl)
	handler := NewRestServerHandler(mockService, mockLogger)

	body := strings.NewReader(`{}`)
	req := httptest.NewRequest(http.MethodPost, "/export?from=0&to=10", body)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	mockService.EXPECT().ExportContainers(gomock.Any(), gomock.Any(), 0, 10, "id", "ASC").
		Return(nil, errors.New("service failed"))
	mockLogger.EXPECT().Error("Failed to export containers", "error", gomock.Any())

	handler.ExportContainers(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Failed to export containers")
}

func TestExportContainers_NilExportData(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockILogger(ctrl)
	mockService := mocks.NewMockIContainerService(ctrl)
	handler := NewRestServerHandler(mockService, mockLogger)

	body := strings.NewReader(`{}`)
	req := httptest.NewRequest(http.MethodPost, "/export?from=0&to=10", body)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	mockService.EXPECT().ExportContainers(gomock.Any(), gomock.Any(), 0, 10, "id", "ASC").
		Return(nil, nil)
	mockLogger.EXPECT().Error("Export data is nil")

	handler.ExportContainers(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Export data is nil")
}

func TestExportContainers_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockILogger(ctrl)
	mockService := mocks.NewMockIContainerService(ctrl)
	handler := NewRestServerHandler(mockService, mockLogger)

	exportData := &dto.ExportData{
		FileName: "containers.xlsx",
		Data:     []byte("excel-content"),
	}

	body := strings.NewReader(`{}`)
	req := httptest.NewRequest(http.MethodPost, "/export?from=0&to=10", body)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	mockService.EXPECT().
		ExportContainers(gomock.Any(), gomock.Any(), 0, 10, "id", "ASC").
		Return(exportData, nil)

	handler.ExportContainers(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "excel-content", w.Body.String())
	assert.Equal(t, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", w.Header().Get("Content-Type"))
	assert.Equal(t, "containers.xlsx", w.Header().Get("File-Name"))
}

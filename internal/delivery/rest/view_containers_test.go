package rest

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"thanhnt208/container-adm-service/internal/dto"
	"thanhnt208/container-adm-service/internal/mocks"
	"thanhnt208/container-adm-service/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestViewContainers_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockIContainerService(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	handler := NewRestServerHandler(mockService, mockLogger)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	filterJSON := `{"container_name":"test"}`
	req := httptest.NewRequest(http.MethodPost, "/containers/view?from=0&to=2&sortBy=id&sortOrder=asc", bytes.NewBufferString(filterJSON))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	expectedFilter := dto.ContainerFilter{ContainerName: "test"}
	exptectedResult := []model.Container{
		{ID: 1, ContainerName: "test", ImageName: "nginx"},
	}
	mockService.EXPECT().
		ViewAllContainers(c, &expectedFilter, 0, 2, "id", "ASC").
		Return(int64(1), exptectedResult, nil)

	handler.ViewContainers(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"count":1`)
	assert.Contains(t, w.Body.String(), `"containers"`)
}

func TestViewContainers_InvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockIContainerService(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	handler := NewRestServerHandler(mockService, mockLogger)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodPost, "/containers/view", bytes.NewBufferString("{invalid_json"))
	c.Request.Header.Set("Content-Type", "application/json")

	mockLogger.EXPECT().Error("Invalid filter data", "error", gomock.Any())

	handler.ViewContainers(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), `"error":"Invalid filter data"`)
}

func TestViewContainers_InvalidFromParam(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockIContainerService(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	handler := NewRestServerHandler(mockService, mockLogger)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := httptest.NewRequest(http.MethodPost, "/containers/view?from=-1", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	mockLogger.EXPECT().Error("Invalid 'from' parameter")

	handler.ViewContainers(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), `"error":"Invalid 'from' parameter"`)
}

func TestViewContainers_InvalidToParam(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockIContainerService(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	handler := NewRestServerHandler(mockService, mockLogger)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := httptest.NewRequest(http.MethodPost, "/containers/view?from=5&to=3", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	mockLogger.EXPECT().Error("Invalid 'to' parameter")

	handler.ViewContainers(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), `"error":"Invalid 'to' parameter"`)
}

func TestViewContainers_ExceedRangeLimit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockIContainerService(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	handler := NewRestServerHandler(mockService, mockLogger)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := `{"name":"test"}`
	req := httptest.NewRequest(http.MethodPost, "/containers/view?from=0&to=101", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	mockLogger.EXPECT().Error("Range limit exceeded")

	handler.ViewContainers(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), `"error":"Range limit exceeded"`)
}

func TestViewContainers_InvalidSortOrder(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockIContainerService(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	handler := NewRestServerHandler(mockService, mockLogger)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := httptest.NewRequest(http.MethodPost, "/containers/view?from=0&to=99&sortOrder=INVALID", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	mockLogger.EXPECT().Error("Invalid sort order")

	handler.ViewContainers(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), `"error":"Invalid sort order"`)
}

func TestViewContainers_ServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockIContainerService(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	handler := NewRestServerHandler(mockService, mockLogger)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := httptest.NewRequest(http.MethodPost, "/containers/view?from=0&to=1", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	mockService.EXPECT().
		ViewAllContainers(gomock.Any(), gomock.Any(), 0, 1, "id", "ASC").
		Return(int64(0), nil, errors.New("db failed"))

	mockLogger.EXPECT().Error("Failed to retrieve containers", "error", gomock.Any())

	handler.ViewContainers(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), `"error":"Failed to retrieve containers"`)
}

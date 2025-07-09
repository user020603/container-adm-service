package rest

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"thanhnt208/container-adm-service/internal/mocks"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestCreateContainer_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockIContainerService(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	handler := NewRestServerHandler(mockService, mockLogger)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := `
		{
			"container_name": "test-container",
			"image_name": "test-image"
		}
	`
	c.Request = httptest.NewRequest(http.MethodPost, "/containers", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")

	mockService.EXPECT().
		CreateContainer(gomock.Any(), "test-container", "test-image").
		Return(int(123), nil)

	handler.CreateContainer(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), `"id":123`)
	assert.Contains(t, w.Body.String(), `"message":"Container created successfully"`)
}

func TestCreateContainer_InvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockIContainerService(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	handler := NewRestServerHandler(mockService, mockLogger)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := `{invalid json`

	c.Request = httptest.NewRequest(http.MethodPost, "/containers", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")

	mockLogger.EXPECT().Error("Invalid request data", "error", gomock.Any())

	handler.CreateContainer(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), `"error":"Invalid request data"`)
}

func TestCreateContainer_MissingFields(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockIContainerService(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	handler := NewRestServerHandler(mockService, mockLogger)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := `{"container_name": " ", "image_name": " "}`
	c.Request = httptest.NewRequest(http.MethodPost, "/containers", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")

	mockLogger.EXPECT().Error("Container name and image name are required")

	handler.CreateContainer(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), `"error":"Container name and image name are required"`)
}

func TestCreateContainer_ServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockIContainerService(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	handler := NewRestServerHandler(mockService, mockLogger)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := `{"container_name": "test-container", "image_name": "test-image"}`
	c.Request = httptest.NewRequest(http.MethodPost, "/containers", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")

	mockService.EXPECT().
		CreateContainer(gomock.Any(), "test-container", "test-image").
		Return(0, errors.New("service error"))
	
	mockLogger.EXPECT(). 
		Error("Failed to create container", "error", gomock.Any())

	handler.CreateContainer(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), `"error":"Failed to create container"`)
}

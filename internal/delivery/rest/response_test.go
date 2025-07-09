package rest

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"thanhnt208/container-adm-service/internal/mocks"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestRespondWithError_WithError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockILogger(ctrl)
	mockService := mocks.NewMockIContainerService(ctrl)

	handler := NewRestServerHandler(mockService, mockLogger)

	mockLogger.EXPECT().Error("Test error message", "error", gomock.Any()).Times(1)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	handler.respondWithError(c, http.StatusInternalServerError, "Test error message", errors.New("some error"))

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), `"error":"Test error message"`)
}

func TestRespondWithError_WithoutError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockILogger(ctrl)
	mockService := mocks.NewMockIContainerService(ctrl)

	handler := NewRestServerHandler(mockService, mockLogger)

	mockLogger.EXPECT().Error("Test error message").Times(1)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	handler.respondWithError(c, http.StatusBadRequest, "Test error message", nil)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.True(t, strings.Contains(w.Body.String(), `"error":"Test error message"`))
}

func TestRespondWithSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockILogger(ctrl)
	mockService := mocks.NewMockIContainerService(ctrl)

	handler := NewRestServerHandler(mockService, mockLogger)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	data := gin.H{
		"message": "Success",
		"count":   5,
	}

	handler.respondWithSuccess(c, http.StatusOK, data)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "Success", resp["message"])
	assert.Equal(t, float64(5), resp["count"]) // json decodes numbers to float64
}

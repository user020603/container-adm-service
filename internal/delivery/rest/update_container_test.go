package rest

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"thanhnt208/container-adm-service/internal/mocks"
	"thanhnt208/container-adm-service/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestUpdateContainer_MissingID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockIContainerService(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	handler := NewRestServerHandler(mockService, mockLogger)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := httptest.NewRequest(http.MethodPut, "/containers/update", strings.NewReader(`{"name":"new"}`))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	c.Params = gin.Params{}

	mockLogger.EXPECT().Error("ID is required")

	handler.UpdateContainer(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), `"error":"ID is required"`)
}

func TestUpdateContainer_InvalidIDFormat(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockIContainerService(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	handler := NewRestServerHandler(mockService, mockLogger)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := httptest.NewRequest(http.MethodPut, "/containers/update/abc", strings.NewReader(`{"name":"new"}`))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	c.Params = gin.Params{gin.Param{Key: "id", Value: "abc"}}

	mockLogger.EXPECT().Error("Invalid ID format", "error", gomock.Any())

	handler.UpdateContainer(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), `"error":"Invalid ID format"`)
}

func TestUpdateContainer_InvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockIContainerService(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	handler := NewRestServerHandler(mockService, mockLogger)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := httptest.NewRequest(http.MethodPut, "/containers/update/1", strings.NewReader(`invalid-json`))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}

	mockLogger.EXPECT().Error("Invalid update data", "error", gomock.Any())

	handler.UpdateContainer(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), `"error":"Invalid update data"`)
}

func TestUpdateContainer_EmptyUpdateFields(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockIContainerService(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	handler := NewRestServerHandler(mockService, mockLogger)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := httptest.NewRequest(http.MethodPut, "/containers/update/1", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}

	mockLogger.EXPECT().Error("No fields to update")

	handler.UpdateContainer(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), `"error":"No fields to update"`)
}

func TestUpdateContainer_ServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockIContainerService(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	handler := NewRestServerHandler(mockService, mockLogger)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	updateData := `{"status":"running"}`
	req := httptest.NewRequest(http.MethodPut, "/containers/update/1", strings.NewReader(updateData))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}

	mockService.EXPECT().
		UpdateContainer(gomock.Any(), uint(1), map[string]interface{}{"status": "running"}).
		Return(nil, errors.New("db error"))

	mockLogger.EXPECT().Error("Failed to update container", "error", gomock.Any())

	handler.UpdateContainer(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), `"error":"Failed to update container"`)
}

func TestUpdateContainer_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockIContainerService(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	handler := NewRestServerHandler(mockService, mockLogger)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	updateData := `{"status":"running"}`
	req := httptest.NewRequest(http.MethodPut, "/containers/update/1", strings.NewReader(updateData))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}

	expectedResult := &model.Container{
		ID:     1,
		Status: "running",
	}

	mockService.EXPECT().
		UpdateContainer(gomock.Any(), uint(1), map[string]interface{}{"status": "running"}).
		Return(expectedResult, nil)

	handler.UpdateContainer(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"message":"Container updated successfully"`)
	assert.Contains(t, w.Body.String(), `"id":1`)
}

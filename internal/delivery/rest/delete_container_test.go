package rest

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"thanhnt208/container-adm-service/internal/mocks"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestDeleteContainer_MissingID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockIContainerService(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	handler := NewRestServerHandler(mockService, mockLogger)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := httptest.NewRequest(http.MethodDelete, "/containers/delete", nil)
	c.Request = req
	c.Params = gin.Params{} // Không có ID

	mockLogger.EXPECT().Error("ID is required")

	handler.DeleteContainer(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), `"error":"ID is required"`)
}

func TestDeleteContainer_InvalidIDFormat(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockIContainerService(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	handler := NewRestServerHandler(mockService, mockLogger)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := httptest.NewRequest(http.MethodDelete, "/containers/delete/abc", nil)
	c.Request = req
	c.Params = gin.Params{gin.Param{Key: "id", Value: "abc"}}

	mockLogger.EXPECT().Error("Invalid ID format", "error", gomock.Any())

	handler.DeleteContainer(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), `"error":"Invalid ID format"`)
}

func TestDeleteContainer_ServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockIContainerService(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	handler := NewRestServerHandler(mockService, mockLogger)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := httptest.NewRequest(http.MethodDelete, "/containers/delete/1", nil)
	c.Request = req
	c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}

	mockService.EXPECT().
		DeleteContainer(gomock.Any(), uint(1)).
		Return(errors.New("delete failed"))

	mockLogger.EXPECT().Error("Failed to delete container", "error", gomock.Any())

	handler.DeleteContainer(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), `"error":"Failed to delete container"`)
}

func TestDeleteContainer_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockIContainerService(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	handler := NewRestServerHandler(mockService, mockLogger)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := httptest.NewRequest(http.MethodDelete, "/containers/delete/1", nil)
	c.Request = req
	c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}

	mockService.EXPECT().
		DeleteContainer(gomock.Any(), uint(1)).
		Return(nil)

	handler.DeleteContainer(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"message":"Container deleted successfully"`)
}


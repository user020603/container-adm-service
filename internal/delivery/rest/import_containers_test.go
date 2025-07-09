package rest

import (
	"bytes"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"thanhnt208/container-adm-service/internal/dto"
	"thanhnt208/container-adm-service/internal/mocks"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestImportContainers_MissingFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockIContainerService(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	handler := NewRestServerHandler(mockService, mockLogger)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	req := httptest.NewRequest(http.MethodPost, "/containers/import", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	c.Request = req

	mockLogger.EXPECT().Error("File is required", "error", gomock.Any())

	handler.ImportContainers(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), `"error":"File is required"`)
}

func TestImportContainers_ExceedMaxSize(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockIContainerService(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	handler := NewRestServerHandler(mockService, mockLogger)

	fileContent := bytes.Repeat([]byte("a"), MaxFileSize+1)
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, _ := writer.CreateFormFile("file", "too_large.xlsx")
	part.Write(fileContent)
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/containers/import", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	mockLogger.EXPECT().Error("File size exceeds the maximum limit of 10 MB", gomock.Any())

	handler.ImportContainers(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), `"error":"File size exceeds the maximum limit of 10 MB"`)
}

func TestImportContainers_EmptyFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockIContainerService(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	handler := NewRestServerHandler(mockService, mockLogger)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, _ := writer.CreateFormFile("file", "empty.xlsx")
	part.Write([]byte{})
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/containers/import", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	mockLogger.EXPECT().Error("File is empty")

	handler.ImportContainers(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), `"error":"File is empty"`)
}

func TestImportContainers_UnsupportedFileType(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockIContainerService(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	handler := NewRestServerHandler(mockService, mockLogger)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "invalid.txt")
	part.Write([]byte("dummy data"))
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/containers/import", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	mockLogger.EXPECT().Error("Unsupported file type, only .xlsx is allowed")

	handler.ImportContainers(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), `"error":"Unsupported file type, only .xlsx is allowed"`)
}

func TestImportContainers_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockIContainerService(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	handler := NewRestServerHandler(mockService, mockLogger)

	fileContent := []byte("mock content")
	file := &bytes.Buffer{}
	writer := multipart.NewWriter(file)
	part, _ := writer.CreateFormFile("file", "import.xlsx")
	part.Write(fileContent)
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/containers/import", file)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	importResult := &dto.ImportResult{
		SuccessfulCount: 1,
		SuccessfulItems: []string{"A"},
		FailedCount:     0,
		FailedItems:     []string{},
	}

	mockService.EXPECT().
		ImportContainers(gomock.Any(), gomock.Any()).
		Return(importResult, nil)

	handler.ImportContainers(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"message":"Containers imported successfully"`)
}

func TestImportContainers_ServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockIContainerService(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	handler := NewRestServerHandler(mockService, mockLogger)

	content := []byte("test xlsx content")
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "valid.xlsx")
	part.Write(content)
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/containers/import", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	mockService.EXPECT().ImportContainers(gomock.Any(), content).Return(nil, errors.New("import failed"))
	mockLogger.EXPECT().Error("Failed to import containers", "error", gomock.Any())

	handler.ImportContainers(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), `"error":"Failed to import containers"`)
}

func TestImportContainers_NilResult(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockIContainerService(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	handler := NewRestServerHandler(mockService, mockLogger)

	content := []byte("test xlsx content")
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "valid.xlsx")
	part.Write(content)
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/containers/import", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	mockService.EXPECT().ImportContainers(gomock.Any(), content).Return(nil, nil)
	mockLogger.EXPECT().Error("Import result is nil")

	handler.ImportContainers(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), `"error":"Import result is nil"`)
}

package rest

import (
	"testing"
	"thanhnt208/container-adm-service/internal/mocks"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestNewRestServerHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockIContainerService(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)

	handler := NewRestServerHandler(mockService, mockLogger)

	assert.NotNil(t, handler)
}

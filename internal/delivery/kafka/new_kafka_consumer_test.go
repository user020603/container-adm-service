package kafka

import (
	"testing"
	"thanhnt208/container-adm-service/internal/mocks"

	"github.com/golang/mock/gomock"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
)

func TestNewKafkaConsumerHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockIContainerService(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	mockReader := &kafka.Reader{}

	handler := NewKafkaConsumerHandler(mockService, mockLogger, mockReader)

	assert.NotNil(t, handler)
}

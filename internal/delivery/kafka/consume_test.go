package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"thanhnt208/container-adm-service/internal/mocks"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
)

func TestKafkaConsumerHandler_StartConsume_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockIContainerService(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	mockReader := mocks.NewMockIKafkaReader(ctrl)

	handler := NewKafkaConsumerHandler(mockService, mockLogger, mockReader)

	messagePayload := map[string]interface{}{
		"id":             123,
		"container_name": "nginx",
		"status":         true,
	}
	msgBytes, _ := json.Marshal(messagePayload)
	testMsg := kafka.Message{
		Key:   []byte("test-key"),
		Value: msgBytes,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	readCalled := make(chan struct{})
	callCount := 0

	mockReader.EXPECT().
		ReadMessage(gomock.Any()).
		DoAndReturn(func(ctx context.Context) (kafka.Message, error) {
			callCount++
			if callCount == 1 {
				close(readCalled)
				return testMsg, nil
			}
			<-ctx.Done()
			return kafka.Message{}, ctx.Err()
		}).
		AnyTimes()

	var wg sync.WaitGroup
	wg.Add(1)

	mockService.EXPECT().
		UpdateContainer(gomock.Any(), uint(123), map[string]interface{}{"status": "running"}).
		Return(nil, nil)

	mockService.EXPECT().
		AddContainerStatus(gomock.Any(), uint(123), "running").
		DoAndReturn(func(context.Context, uint, string) error {
			wg.Done()
			return nil
		})

	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info("Stopping Kafka consumer due to context cancellation").AnyTimes()

	go func() {
		_ = handler.StartConsume(ctx)
	}()

	<-readCalled
	wg.Wait()
	cancel()
}

func TestKafkaConsumerHandler_Close_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockIContainerService(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	mockReader := mocks.NewMockIKafkaReader(ctrl)

	handler := NewKafkaConsumerHandler(mockService, mockLogger, mockReader)

	mockReader.EXPECT().Close().Return(nil)
	mockLogger.EXPECT().Info("Kafka reader closed successfully")

	err := handler.Close()
	assert.NoError(t, err)
}

func TestKafkaConsumerHandler_Close_Failure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockIContainerService(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	mockReader := mocks.NewMockIKafkaReader(ctrl)

	handler := NewKafkaConsumerHandler(mockService, mockLogger, mockReader)

	expectedErr := errors.New("close error")

	mockReader.EXPECT().Close().Return(expectedErr)
	mockLogger.EXPECT().Error("Failed to close Kafka reader", "error", expectedErr)

	err := handler.Close()
	assert.EqualError(t, err, "close error")
}

func TestKafkaConsumerHandler_StartConsume_UnmarshalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockIContainerService(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	mockReader := mocks.NewMockIKafkaReader(ctrl)

	handler := NewKafkaConsumerHandler(mockService, mockLogger, mockReader)

	// Invalid JSON
	testMsg := kafka.Message{
		Key:   []byte("key"),
		Value: []byte("invalid-json"),
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mockReader.EXPECT().
		ReadMessage(gomock.Any()).
		DoAndReturn(func(ctx context.Context) (kafka.Message, error) {
			cancel()
			return testMsg, nil
		})

	mockLogger.EXPECT().
		Info("Received message from Kafka", "key", "key", "value", "invalid-json").
		AnyTimes()

	mockLogger.EXPECT().
		Error(
			gomock.Eq("Failed to unmarshal message"),
			gomock.Eq("error"),
			gomock.Any(),
			gomock.Eq("message"),
			gomock.Eq("invalid-json"),
		).AnyTimes()

	mockLogger.EXPECT().
		Info("Stopping Kafka consumer due to context cancellation")

	err := handler.StartConsume(ctx)
	assert.EqualError(t, err, context.Canceled.Error())
}

func TestKafkaConsumerHandler_StartConsume_UpdateContainerError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockIContainerService(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	mockReader := mocks.NewMockIKafkaReader(ctrl)

	handler := NewKafkaConsumerHandler(mockService, mockLogger, mockReader)

	payload := map[string]interface{}{
		"id":             42,
		"container_name": "test-container",
		"status":         true,
	}
	msgBytes, _ := json.Marshal(payload)
	testMsg := kafka.Message{
		Key:   []byte("some-key"),
		Value: msgBytes,
	}

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})

	mockReader.EXPECT().
		ReadMessage(gomock.Any()).
		Return(testMsg, nil).
		Times(1)

	mockReader.EXPECT().
		ReadMessage(gomock.Any()).
		DoAndReturn(func(ctx context.Context) (kafka.Message, error) {
			cancel()
			return kafka.Message{}, context.Canceled
		}).
		AnyTimes()

	mockLogger.EXPECT().
		Info("Received message from Kafka", "key", "some-key", "value", string(msgBytes))

	mockLogger.EXPECT().
		Info("Processing message", "ID", uint(42), "containerName", "test-container")

	mockService.EXPECT().
		UpdateContainer(gomock.Any(), uint(42), map[string]interface{}{"status": "running"}).
		DoAndReturn(func(ctx context.Context, id uint, data map[string]interface{}) (interface{}, error) {
			defer close(done)
			return nil, errors.New("update failed")
		})

	mockLogger.EXPECT().
		Error("Failed to read message from Kafka", "error", context.Canceled).
		AnyTimes()

	mockLogger.EXPECT().
		Error("Failed to update container status", "ID", uint(42), "error", gomock.Any())

	mockLogger.EXPECT().
		Info("Stopping Kafka consumer due to context cancellation")

	go func() {
		_ = handler.StartConsume(ctx)
	}()

	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("goroutine did not finish in time")
	}
}

func TestKafkaConsumerHandler_StartConsume_AddContainerStatusError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockIContainerService(ctrl)
	mockLogger := mocks.NewMockILogger(ctrl)
	mockReader := mocks.NewMockIKafkaReader(ctrl)

	handler := NewKafkaConsumerHandler(mockService, mockLogger, mockReader)

	messagePayload := map[string]interface{}{
		"id":             101,
		"container_name": "redis",
		"status":         true,
	}
	msgBytes, _ := json.Marshal(messagePayload)
	testMsg := kafka.Message{
		Key:   []byte("test-key"),
		Value: msgBytes,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan struct{})

	// 1st message: success
	mockReader.EXPECT().
		ReadMessage(gomock.Any()).
		Return(testMsg, nil).
		Times(1)

	// 2nd message: ctx canceled
	mockReader.EXPECT().
		ReadMessage(gomock.Any()).
		DoAndReturn(func(ctx context.Context) (kafka.Message, error) {
			<-time.After(50 * time.Millisecond)
			cancel()
			return kafka.Message{}, context.Canceled
		}).
		Times(1)

	mockLogger.EXPECT().
		Info("Received message from Kafka", "key", "test-key", "value", string(msgBytes))

	mockLogger.EXPECT().
		Info("Processing message", "ID", uint(101), "containerName", "redis")

	mockService.EXPECT().
		UpdateContainer(gomock.Any(), uint(101), map[string]interface{}{"status": "running"}).
		Return(nil, nil)

	mockLogger.EXPECT().
		Info("Successfully updated container status", "ID", uint(101), "status", "running")

	mockLogger.EXPECT().
		Info("Write to ES", "ID", uint(101), "status", "running")

	mockService.EXPECT().
		AddContainerStatus(gomock.Any(), uint(101), "running").
		DoAndReturn(func(ctx context.Context, id uint, status string) error {
			defer close(done)
			return errors.New("mock add error")
		})

	mockLogger.EXPECT().
		Error("Failed to add container status", "ID", uint(101), "error", gomock.Any())

	mockLogger.EXPECT().
		Error("Failed to read message from Kafka", "error", gomock.Any()).
		AnyTimes()

	mockLogger.EXPECT().
		Info("Stopping Kafka consumer due to context cancellation").
		AnyTimes()

	go func() {
		_ = handler.StartConsume(ctx)
	}()

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for goroutine")
	}
}

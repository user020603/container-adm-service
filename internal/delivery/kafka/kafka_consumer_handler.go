package kafka

import (
	"context"
	"encoding/json"
	"thanhnt208/container-adm-service/internal/service"
	"thanhnt208/container-adm-service/pkg/kafkaClient"
	"thanhnt208/container-adm-service/pkg/logger"

	"github.com/segmentio/kafka-go"
)

type KafkaConsumerHandler struct {
	service service.IContainerService
	logger  logger.ILogger
	reader  kafkaClient.IKafkaReader
}

func NewKafkaConsumerHandler(service service.IContainerService, logger logger.ILogger, reader kafkaClient.IKafkaReader) *KafkaConsumerHandler {
	return &KafkaConsumerHandler{
		service: service,
		logger:  logger,
		reader:  reader,
	}
}

func (h *KafkaConsumerHandler) StartConsume(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			h.logger.Info("Stopping Kafka consumer due to context cancellation")
			return ctx.Err()
		default:
			msg, err := h.reader.ReadMessage(ctx)
			if err != nil {
				h.logger.Error("Failed to read message from Kafka", "error", err)
				continue
			}

			go func(message kafka.Message) {
				h.logger.Info("Received message from Kafka", "key", string(msg.Key), "value", string(msg.Value))

				var containerMessage struct {
					ID            uint   `json:"id"`
					ContainerName string `json:"container_name"`
					Status        bool   `json:"status"`
				}

				if err := json.Unmarshal(message.Value, &containerMessage); err != nil {
					h.logger.Error("Failed to unmarshal message", "error", err, "message", string(message.Value))
					return
				}

				h.logger.Info("Processing message", "ID", containerMessage.ID, "containerName", containerMessage.ContainerName)

				status := "stopped"
				if containerMessage.Status {
					status = "running"
				}

				updateData := map[string]interface{}{
					"status": status,
				}

				_, err = h.service.UpdateContainer(ctx, containerMessage.ID, updateData)
				if err != nil {
					h.logger.Error("Failed to update container status", "ID", containerMessage.ID, "error", err)
					return
				}
				h.logger.Info("Successfully updated container status", "ID", containerMessage.ID, "status", status)

				h.logger.Info("Write to ES", "ID", containerMessage.ID, "status", status)
				err = h.service.AddContainerStatus(ctx, containerMessage.ID, status)
				if err != nil {
					h.logger.Error("Failed to add container status", "ID", containerMessage.ID, "error", err)
					return
				}

				h.logger.Info("Successfully added container status", "ID", containerMessage.ID, "status", status)
			}(msg)
		}
	}
}

func (h *KafkaConsumerHandler) Close() error {
	if err := h.reader.Close(); err != nil {
		h.logger.Error("Failed to close Kafka reader", "error", err)
		return err
	}
	h.logger.Info("Kafka reader closed successfully")
	return nil
}

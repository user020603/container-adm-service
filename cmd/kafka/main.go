package main

import (
	"os"
	"os/signal"
	"syscall"
	"thanhnt208/container-adm-service/config"
	"thanhnt208/container-adm-service/external/client"
	"thanhnt208/container-adm-service/infrastructure"
	kafkaHandler "thanhnt208/container-adm-service/internal/delivery/kafka"
	"thanhnt208/container-adm-service/internal/repository"
	"thanhnt208/container-adm-service/internal/service"
	"thanhnt208/container-adm-service/pkg/logger"

	"golang.org/x/net/context"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.LoadConfig()
	log, err := logger.NewLogger(cfg.LogLevel, cfg.LogFile)
	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}

	postgresDB := infrastructure.NewDatabase(cfg)
	db, err := postgresDB.ConnectDB()
	if err != nil {
		log.Error("Failed to connect to the database", "error", err)
		panic("Failed to connect to the database: " + err.Error())
	}
	defer postgresDB.Close()

	elasticsearchClient := infrastructure.NewElasticsearch(cfg)
	esClient, err := elasticsearchClient.ConnectElasticsearch()
	if err != nil {
		log.Error("Failed to connect to Elasticsearch", "error", err)
		panic("Failed to connect to Elasticsearch: " + err.Error())
	}

	kafkaInfra := infrastructure.NewKafka(cfg)
	kafkaConsumer, err := kafkaInfra.ConnectConsumer([]string{cfg.KafkaTopic})
	if err != nil {
		log.Error("Failed to connect to Kafka consumer", "error", err)
		panic("Failed to connect to Kafka consumer: " + err.Error())
	}

	dockerClient, err := client.NewDockerClient()
	if err != nil {
		log.Error("Failed to create Docker client", "error", err)
		panic("Failed to create Docker client: " + err.Error())
	}

	containerRepository := repository.NewContainerRepository(db, esClient, log)
	containerService := service.NewContainerService(containerRepository, log, dockerClient)
	kafkaConsumerHandler := kafkaHandler.NewKafkaConsumerHandler(containerService, log, kafkaConsumer)

	consumerDone := make(chan error, 1)

	go func() {
		log.Info("Starting Kafka consumer")
		if err := kafkaConsumerHandler.StartConsume(ctx); err != nil {
			log.Error("Error consuming messages from Kafka", "error", err)
			consumerDone <- err
		} else {
			log.Info("Kafka consumer stopped gracefully")
			consumerDone <- nil
		}
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-consumerDone:
		if err != nil {
			log.Error("Kafka consumer encountered an error", "error", err)
		} else {
			log.Info("Kafka consumer finished successfully")
		}
	case sig := <-sigs:
		log.Info("Received signal, shutting down", "signal", sig)
		cancel()

		log.Info("Waiting for Kafka consumer to finish")
		err := <-consumerDone
		if err != nil && err != context.Canceled {
			log.Error("Kafka consumer did not finish gracefully", "error", err)
		} else {
			log.Info("Kafka consumer finished gracefully")
		}
	}

	if err := kafkaConsumer.Close(); err != nil {
		log.Error("Failed to close Kafka consumer", "error", err)
	} else {
		log.Info("Kafka consumer closed successfully")
	}

	log.Info("Service shutdown complete")
}

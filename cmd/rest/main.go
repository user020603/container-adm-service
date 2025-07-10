package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"thanhnt208/container-adm-service/api/routes"
	"thanhnt208/container-adm-service/config"
	"thanhnt208/container-adm-service/external/client"
	"thanhnt208/container-adm-service/infrastructure"
	"thanhnt208/container-adm-service/internal/delivery/rest"
	"thanhnt208/container-adm-service/internal/repository"
	"thanhnt208/container-adm-service/internal/service"
	"thanhnt208/container-adm-service/pkg/logger"
	"time"
)

func main() {
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
	esRawClient, err := elasticsearchClient.ConnectElasticsearch()
	if err != nil {
		log.Error("Failed to connect to Elasticsearch", "error", err)
		panic("Failed to connect to Elasticsearch: " + err.Error())
	}

	esClient := &client.RealESClient{Client: esRawClient}

	dockerClient, err := client.NewDockerClient()
	if err != nil {
		log.Error("Failed to create Docker client", "error", err)
		panic("Failed to create Docker client: " + err.Error())
	}

	containerRepository := repository.NewContainerRepository(db, esClient, log)
	containerService := service.NewContainerService(containerRepository, log, dockerClient)
	containerRestHandler := rest.NewRestServerHandler(containerService, log)

	r := routes.SetupContainerRoutes(containerRestHandler)

	port := cfg.ServerPort
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		log.Info("Starting REST server", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("Failed to start server", "error", err)
			panic("Failed to start server: " + err.Error())
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("Shutting down server gracefully...")

	ctxShutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctxShutdown); err != nil {
		log.Fatal("REST server forced to shutdown:", "error", err)
	}

	log.Info("REST server exiting")
}

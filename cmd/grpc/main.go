package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"thanhnt208/container-adm-service/config"
	"thanhnt208/container-adm-service/external/client"
	"thanhnt208/container-adm-service/infrastructure"
	"thanhnt208/container-adm-service/internal/delivery/grpc"
	"thanhnt208/container-adm-service/internal/repository"
	"thanhnt208/container-adm-service/internal/service"
	"thanhnt208/container-adm-service/pkg/logger"
	"thanhnt208/container-adm-service/proto/pb"

	grpcServer "google.golang.org/grpc"
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
	containerHandler := grpc.NewGrpcServerHandler(containerService, log)

	grpcPort := cfg.GrpcPort
	log.Info("Starting gRPC server", "port", grpcPort)

	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Error("Failed to listen on port", "port", grpcPort, "error", err)
		panic("Failed to listen on port: " + err.Error())
	}

	grpcServer := grpcServer.NewServer()
	pb.RegisterContainerAdmServiceServer(grpcServer, containerHandler)

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Error("Failed to serve gRPC server", "error", err)
			panic("Failed to serve gRPC server: " + err.Error())
		}
	}()

	fmt.Printf("gRPC server is running on port %s\n", grpcPort)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("Shutting down gRPC server...")
	grpcServer.GracefulStop()
	log.Info("gRPC server exiting")
}

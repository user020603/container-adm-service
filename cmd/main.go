package main

import (
	"context"
	"fmt"
	"os"
	"thanhnt208/container-adm-service/config"
	"thanhnt208/container-adm-service/external/client"
	"thanhnt208/container-adm-service/infrastructure"
	"thanhnt208/container-adm-service/internal/repository"
	"thanhnt208/container-adm-service/internal/service"
	"thanhnt208/container-adm-service/pkg/logger"
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
	esClient, err := elasticsearchClient.ConnectElasticsearch()
	if err != nil {
		log.Error("Failed to connect to Elasticsearch", "error", err)
		panic("Failed to connect to Elasticsearch: " + err.Error())
	}

	dockerClient, err := client.NewDockerClient()
	if err != nil {
		log.Error("Failed to create Docker client", "error", err)
		panic("Failed to create Docker client: " + err.Error())
	}

	containerRepository := repository.NewContainerRepository(db, esClient, log)
	containerService := service.NewContainerService(containerRepository, log, dockerClient)

	fmt.Println("Exporting containers to Excel...")

	exportData, err := containerService.ExportContainers(context.Background(), nil, 0, 100, "", "")
	if err != nil {
		log.Error("Failed to export containers", "error", err)
		fmt.Printf("❌ Export failed: %v\n", err)
		return
	}

	if err := os.WriteFile(exportData.FileName, exportData.Data, 0644); err != nil {
		log.Error("Failed to save exported Excel file", "error", err)
		fmt.Printf("❌ Failed to save exported Excel file: %v\n", err)
		return
	}

	fmt.Printf("✅ Exported containers to file: %s\n", exportData.FileName)
	fmt.Println("\n🎉 Export test completed!")
}

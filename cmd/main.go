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

	"github.com/xuri/excelize/v2"
)

func createSampleExcelFile() error {
	f := excelize.NewFile()

	headers := []string{"Container Name", "Image Name"}
	for i, header := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i)
		f.SetCellValue("Sheet1", cell, header)
	}

	sampleData := [][]string{
		{"nginx-container-1", "nginx:latest"},
		{"redis-container-1", "redis:alpine"},
		{"postgres-container-1", "postgres:13"},
		{"ubuntu-container-1", "ubuntu:20.04"},
		{"mysql-container-1", "mysql:8.0"},
	}

	for i, row := range sampleData {
		rowNum := i + 2 
		for j, value := range row {
			cell := fmt.Sprintf("%c%d", 'A'+j, rowNum)
			f.SetCellValue("Sheet1", cell, value)
		}
	}

	if err := f.SaveAs("containers.xlsx"); err != nil {
		return fmt.Errorf("failed to save Excel file: %w", err)
	}

	fmt.Println("✅ Created sample Excel file: containers.xlsx")
	return nil
}

func main() {
	if err := createSampleExcelFile(); err != nil {
		fmt.Printf("❌ Failed to create sample Excel file: %v\n", err)
		return
	}

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

	fileData, err := os.ReadFile("containers.xlsx")
	if err != nil {
		log.Error("Failed to read Excel file", "error", err)
		fmt.Printf("❌ Failed to read Excel file: %v\n", err)
		return
	}

	fmt.Println("🚀 Starting container import...")

	importResult, err := containerService.ImportContainers(context.Background(), fileData)
	if err != nil {
		log.Error("Failed to import containers", "error", err)
		fmt.Printf("❌ Import failed: %v\n", err)
		return
	}

	fmt.Println("\n📊 Import Results:")
	fmt.Printf("✅ Successfully imported: %d containers\n", importResult.SuccessfulCount)
	fmt.Printf("❌ Failed to import: %d containers\n", importResult.FailedCount)

	if len(importResult.SuccessfulItems) > 0 {
		fmt.Println("\n✅ Successfully imported containers:")
		for _, item := range importResult.SuccessfulItems {
			fmt.Printf("  - %s\n", item)
		}
	}

	if len(importResult.FailedItems) > 0 {
		fmt.Println("\n❌ Failed to import containers:")
		for _, item := range importResult.FailedItems {
			fmt.Printf("  - %s\n", item)
		}
	}

	fmt.Println("\n🎉 Import test completed!")
}

package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"thanhnt208/container-adm-service/internal/dto"
	"thanhnt208/container-adm-service/internal/model"
	"thanhnt208/container-adm-service/pkg/esclient"
	"thanhnt208/container-adm-service/pkg/logger"
	"time"

	"github.com/elastic/go-elasticsearch/esapi"
	"gorm.io/gorm"
)

const (
	StatusRunning = "running"
	StatusStopped = "stopped"
)

type IContainerRepository interface {
	CreateContainer(ctx context.Context, container *model.Container) (int, error)
	CreateManyContainers(ctx context.Context, containers []model.Container) ([]model.Container, []model.Container, error)
	ViewAllContainers(ctx context.Context, containerFilter *dto.ContainerFilter, from, to int, sortBy string, sortOrder string) (int64, []model.Container, error)
	UpdateContainer(ctx context.Context, id uint, updateData map[string]interface{}) (*model.Container, error)
	DeleteContainer(ctx context.Context, id uint) error
	GetContainerByID(ctx context.Context, id uint) (*model.Container, error)

	GetContainerInfo(ctx context.Context) ([]dto.ContainerName, error)

	AddContainerStatus(ctx context.Context, id uint, status string) error
	GetNumContainers(ctx context.Context) (int64, error)
	GetNumRunningContainers(ctx context.Context) (int64, error)
	GetContainerUptimeRatio(ctx context.Context, startTime, endTime time.Time) (float64, error)
	GetContainerUptimeDuration(ctx context.Context, startTime, endTime time.Time) (*dto.UptimeDetails, error)
}

type containerRepository struct {
	db     *gorm.DB
	es     esclient.ElasticsearchClient
	logger logger.ILogger
}

func NewContainerRepository(db *gorm.DB, es esclient.ElasticsearchClient, logger logger.ILogger) IContainerRepository {
	return &containerRepository{
		db:     db,
		es:     es,
		logger: logger,
	}
}

func (r *containerRepository) CreateContainer(ctx context.Context, container *model.Container) (int, error) {
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		r.logger.Error("Failed to begin transaction", "error", tx.Error)
		return 0, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	defer func() {
		if rec := recover(); rec != nil {
			tx.Rollback()
			r.logger.Error("Recovered from panic in CreateContainer", "error", rec)
			panic(rec)
		}
	}()

	if err := tx.Create(container).Error; err != nil {
		tx.Rollback()
		r.logger.Error("Failed to create container", "error", err)
		return 0, fmt.Errorf("failed to create container: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		r.logger.Error("Failed to commit transaction", "error", err)
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	r.logger.Info("Container created successfully", "container_id", container.ID)
	return int(container.ID), nil
}

func (r *containerRepository) CreateManyContainers(ctx context.Context, containers []model.Container) ([]model.Container, []model.Container, error) {
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		r.logger.Error("Failed to begin transaction", "error", tx.Error)
		return nil, nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	defer func() {
		if rec := recover(); rec != nil {
			tx.Rollback()
			r.logger.Error("Recovered from panic in CreateManyContainers", "error", rec)
			panic(rec)
		}
	}()

	var createdContainers []model.Container
	var failedContainers []model.Container

	for _, container := range containers {
		if err := tx.Create(&container).Error; err != nil {
			failedContainers = append(failedContainers, container)
			r.logger.Error("Failed to create container", "error", err, "container_id", container.ContainerID)
			continue
		}
		createdContainers = append(createdContainers, container)
	}

	if len(createdContainers) == 0 && len(failedContainers) > 0 {
		tx.Rollback()
		r.logger.Warn("No containers created, all failed", "failed_containers", failedContainers)
		return nil, failedContainers, fmt.Errorf("no containers created, all failed")
	}

	if err := tx.Commit().Error; err != nil {
		r.logger.Error("Failed to commit transaction", "error", err)
		return nil, failedContainers, fmt.Errorf("failed to commit transaction: %w", err)
	}

	r.logger.Info("Containers inserted successfully", "createdCount", len(createdContainers), "failedCount", len(failedContainers))
	return createdContainers, failedContainers, nil
}

func (r *containerRepository) ViewAllContainers(ctx context.Context, containerFilter *dto.ContainerFilter, from, to int, sortBy string, sortOrder string) (int64, []model.Container, error) {
	query := r.db.WithContext(ctx).Model(&model.Container{})

	if containerFilter != nil {
		if containerFilter.ContainerID != "" {
			query = query.Where("container_id = ?", containerFilter.ContainerID)
		}

		if containerFilter.ContainerName != "" {
			query = query.Where("container_name LIKE ?", "%"+containerFilter.ContainerName+"%")
		}

		if containerFilter.ImageName != "" {
			query = query.Where("image_name LIKE ?", "%"+containerFilter.ImageName+"%")
		}

		if containerFilter.Status != "" {
			query = query.Where("status = ?", containerFilter.Status)
		}
	}

	var totalContainers int64
	if err := query.Count(&totalContainers).Error; err != nil {
		r.logger.Error("Failed to count containers", "error", err)
		return 0, nil, fmt.Errorf("failed to count containers: %w", err)
	}

	if sortBy != "" {
		allowedSortFields := map[string]bool{
			"id":             true,
			"container_id":   true,
			"container_name": true,
			"image_name":     true,
			"status":         true,
			"created_at":     true,
			"updated_at":     true,
		}

		if !allowedSortFields[sortBy] {
			r.logger.Warn("Invalid sort field", "sortBy", sortBy)
			return 0, nil, fmt.Errorf("invalid sort field: %s", sortBy)
		}

		orderDirection := "ASC"
		if strings.ToLower(sortOrder) == "desc" {
			orderDirection = "DESC"
		}
		query = query.Order(sortBy + " " + orderDirection)
	} else {
		query = query.Order("created_at DESC")
	}

	var containers []model.Container
	limit := to - from
	if limit <= 0 {
		limit = 10
	}

	if err := query.Offset(from).Limit(limit).Find(&containers).Error; err != nil {
		r.logger.Error("Failed to retrieve containers", "error", err)
		return 0, nil, fmt.Errorf("failed to retrieve containers: %w", err)
	}

	r.logger.Info("Containers retrieved successfully", "total", totalContainers, "count", len(containers))
	return totalContainers, containers, nil
}

func (r *containerRepository) UpdateContainer(ctx context.Context, id uint, updateData map[string]interface{}) (*model.Container, error) {
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		r.logger.Error("Failed to begin transaction", "error", tx.Error)
		return nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	defer func() {
		if rec := recover(); rec != nil {
			tx.Rollback()
			r.logger.Error("Recovered from panic in UpdateContainer", "error", rec)
			panic(rec)
		}
	}()

	var container model.Container
	if err := tx.First(&container, id).Error; err != nil {
		tx.Rollback()
		r.logger.Error("Container not found", "id", id, "error", err)
		return nil, fmt.Errorf("container not found: %w", err)
	}

	if err := tx.Model(&container).Updates(updateData).Error; err != nil {
		tx.Rollback()
		r.logger.Error("Failed to update container", "id", id, "error", err)
		return nil, fmt.Errorf("failed to update container: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		r.logger.Error("Failed to commit transaction", "error", err)
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	r.logger.Info("Container updated successfully", "id", id)
	return &container, nil
}

func (r *containerRepository) DeleteContainer(ctx context.Context, id uint) error {
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		r.logger.Error("Failed to begin transaction", "error", tx.Error)
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	defer func() {
		if rec := recover(); rec != nil {
			tx.Rollback()
			r.logger.Error("Recovered from panic in DeleteContainer", "error", rec)
			panic(rec)
		}
	}()

	var container model.Container
	if err := tx.First(&container, id).Error; err != nil {
		tx.Rollback()
		r.logger.Error("Container not found", "id", id, "error", err)
		return fmt.Errorf("container not found: %w", err)
	}

	if err := tx.Delete(&container).Error; err != nil {
		tx.Rollback()
		r.logger.Error("Failed to delete container", "id", id, "error", err)
		return fmt.Errorf("failed to delete container: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		r.logger.Error("Failed to commit transaction", "error", err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	r.logger.Info("Container deleted successfully", "id", id)
	return nil
}

func (r *containerRepository) GetContainerByID(ctx context.Context, id uint) (*model.Container, error) {
	var container model.Container
	if err := r.db.WithContext(ctx).First(&container, id).Error; err != nil {
		r.logger.Error("Container not found", "id", id, "error", err)
		return nil, fmt.Errorf("container not found: %w", err)
	}

	r.logger.Info("Container retrieved successfully", "id", id)
	return &container, nil
}

func (r *containerRepository) GetContainerInfo(ctx context.Context) ([]dto.ContainerName, error) {
	query := r.db.WithContext(ctx).Model(&model.Container{}).Select("id, container_name")

	var containerNames []dto.ContainerName
	if err := query.Find(&containerNames).Error; err != nil {
		r.logger.Error("Failed to retrieve container names", "error", err)
		return nil, fmt.Errorf("failed to retrieve container names: %w", err)
	}

	r.logger.Info("Container names retrieved successfully", "count", len(containerNames))
	return containerNames, nil
}

func (r *containerRepository) AddContainerStatus(ctx context.Context, id uint, status string) error {
	if status != StatusRunning && status != StatusStopped {
		r.logger.Error("Invalid status provided", "status", status)
		return fmt.Errorf("invalid status: %s", status)
	}

	doc := map[string]interface{}{
		"id":        id,
		"status":    status,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(doc); err != nil {
		r.logger.Error("Failed to encode document for Elasticsearch", "error", err)
		return fmt.Errorf("encode error: %w", err)
	}

	req := esapi.IndexRequest{
		Index:   "container_status",
		Body:    &buf,
		Refresh: "true",
	}
	res, err := r.es.Do(ctx, req)
	if err != nil {
		r.logger.Error("ES index request failed", "error", err)
		return fmt.Errorf("index request failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		r.logger.Error("Elasticsearch response error", "status", res.Status(), "body", string(body))
		return fmt.Errorf("elasticsearch error: %s", res.Status())
	}

	r.logger.Info("Container status indexed", "id", id, "status", status)
	return nil
}

func (r *containerRepository) GetNumContainers(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.Container{}).Count(&count).Error; err != nil {
		r.logger.Error("Failed to count containers", "error", err)
		return 0, fmt.Errorf("failed to count containers: %w", err)
	}

	r.logger.Info("Number of containers retrieved successfully", "count", count)
	return count, nil
}

func (r *containerRepository) GetNumRunningContainers(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.Container{}).Where("status = ?", StatusRunning).Count(&count).Error; err != nil {
		r.logger.Error("Failed to count running containers", "error", err)
		return 0, fmt.Errorf("failed to count running containers: %w", err)
	}

	r.logger.Info("Number of running containers retrieved successfully", "count", count)
	return count, nil
}

func (r *containerRepository) GetContainerUptimeRatio(ctx context.Context, startTime, endTime time.Time) (float64, error) {
	if startTime.After(endTime) {
		r.logger.Error("Start time cannot be after end time", "startTime", startTime, "endTime", endTime)
		return 0, fmt.Errorf("start time cannot be after end time")
	}

	query := map[string]interface{}{
		"size": 0,
		"query": map[string]interface{}{
			"range": map[string]interface{}{
				"timestamp": map[string]interface{}{
					"gte": startTime.UTC().Format(time.RFC3339),
					"lte": endTime.UTC().Format(time.RFC3339),
				},
			},
		},
		"aggs": map[string]interface{}{
			"per_container": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": "id",
					"size":  10000,
				},
				"aggs": map[string]interface{}{
					"total_docs": map[string]interface{}{
						"value_count": map[string]interface{}{
							"field": "id",
						},
					},
					"on_count": map[string]interface{}{
						"filter": map[string]interface{}{
							"term": map[string]interface{}{
								"status.keyword": StatusRunning,
							},
						},
					},
					"on_ratio": map[string]interface{}{
						"bucket_script": map[string]interface{}{
							"buckets_path": map[string]interface{}{
								"on":  "on_count._count",
								"all": "total_docs.value",
							},
							"script": "params.all > 0 ? params.on / params.all : 0",
						},
					},
				},
			},
			"avg_ratio": map[string]interface{}{
				"avg_bucket": map[string]interface{}{
					"buckets_path": "per_container>on_ratio.value",
				},
			},
		},
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		r.logger.Error("Failed to encode search query", "error", err)
		return 0, fmt.Errorf("encode error: %w", err)
	}

	req := esapi.SearchRequest{
		Index:          []string{"container_status"},
		Body:           &buf,
		TrackTotalHits: esapi.BoolPtr(true),
	}
	res, err := r.es.Do(ctx, req)
	if err != nil {
		r.logger.Error("ES search request failed", "error", err)
		return 0, err
	}
	defer res.Body.Close()

	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		r.logger.Error("Elasticsearch search error", "status", res.Status(), "body", string(body))
		return 0, fmt.Errorf("search error: %s", res.Status())
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		r.logger.Error("Failed to decode ES search response", "error", err)
		return 0, fmt.Errorf("decode error: %w", err)
	}

	aggs := resp["aggregations"].(map[string]interface{})
	avg := aggs["avg_ratio"].(map[string]interface{})["value"].(float64)
	return avg, nil
}

func (r *containerRepository) GetContainerUptimeDuration(ctx context.Context, startTime, endTime time.Time) (*dto.UptimeDetails, error) {
	if startTime.After(endTime) {
		r.logger.Error("Start time cannot be after end time", "startTime", startTime, "endTime", endTime)
		return nil, fmt.Errorf("start time cannot be after end time")
	}

	query := map[string]interface{}{
		"size": 0,
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{
					{
						"range": map[string]interface{}{
							"timestamp": map[string]interface{}{
								"gte": startTime.UTC().Format(time.RFC3339),
								"lte": endTime.UTC().Format(time.RFC3339),
							},
						},
					},
					{
						"term": map[string]interface{}{
							"status.keyword": StatusRunning,
						},
					},
				},
			},
		},
		"aggs": map[string]interface{}{
			"containers": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": "id",
					"size":  10000,
				},
				"aggs": map[string]interface{}{
					"running_count": map[string]interface{}{
						"value_count": map[string]interface{}{
							"field": "id",
						},
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	_ = json.NewEncoder(&buf).Encode(query)

	req := esapi.SearchRequest{
		Index:          []string{"container_status"},
		Body:           &buf,
		TrackTotalHits: esapi.BoolPtr(true),
	}
	res, err := r.es.Do(ctx, req)
	if err != nil {
		r.logger.Error("ES search request failed", "error", err)
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		r.logger.Error("Elasticsearch search error", "status", res.Status(), "body", string(body))
		return nil, fmt.Errorf("search error: %s", res.Status())
	}

	var resp map[string]interface{}
	_ = json.NewDecoder(res.Body).Decode(&resp)

	buckets := resp["aggregations"].(map[string]interface{})["containers"].(map[string]interface{})["buckets"].([]interface{})
	per := make(map[string]time.Duration)
	var total time.Duration
	for _, b := range buckets {
		m := b.(map[string]interface{})
		key := fmt.Sprintf("%v", m["key"])
		cnt := int(m["running_count"].(map[string]interface{})["value"].(float64))
		dur := time.Duration(cnt) * time.Minute
		per[key] = dur
		total += dur
	}

	return &dto.UptimeDetails{TotalUptime: total, PerContainerUptime: per}, nil
}

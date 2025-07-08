package repository

import (
	"context"
	"errors"
	"testing"
	"thanhnt208/container-adm-service/internal/dto"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

// Test ViewAllContainers
func TestViewAllContainers_Success_NoFilter(t *testing.T) {
	db, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	logger := new(MockLogger)
	repo := &containerRepository{db: db, logger: logger}

	rows := sqlmock.NewRows([]string{"id", "container_id", "container_name", "image_name", "status"}).
		AddRow(1, "id1", "name1", "img1", "running").
		AddRow(2, "id2", "name2", "img2", "stopped")

	mock.ExpectQuery("SELECT count\\(\\*\\) FROM .*containers.*").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	mock.ExpectQuery("SELECT .* FROM .*containers.*").
		WillReturnRows(rows)

	total, result, err := repo.ViewAllContainers(context.Background(), nil, 0, 10, "", "")

	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, result, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestViewAllContainers_WithFilter(t *testing.T) {
	db, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	logger := new(MockLogger)
	repo := &containerRepository{db: db, logger: logger}

	filter := &dto.ContainerFilter{
		ContainerID:   "id1",
		ContainerName: "app",
		ImageName:     "nginx",
		Status:        "running",
	}

	mock.ExpectQuery("SELECT count\\(\\*\\) FROM .*containers.*WHERE.*container_id = .*container_name LIKE.*image_name LIKE.*status = .*").
		WithArgs(filter.ContainerID, "%"+filter.ContainerName+"%", "%"+filter.ImageName+"%", filter.Status).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery("SELECT .* FROM .*containers.*WHERE.*").
		WithArgs(
			filter.ContainerID,
			"%"+filter.ContainerName+"%",
			"%"+filter.ImageName+"%",
			filter.Status,
			sqlmock.AnyArg(), // LIMIT
		).
		WillReturnRows(sqlmock.NewRows([]string{"id", "container_id", "container_name", "image_name", "status"}).
			AddRow(1, "id1", "app-server", "nginx", "running"))

	total, result, err := repo.ViewAllContainers(context.Background(), filter, 0, 10, "", "")

	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, result, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestViewAllContainers_CountError(t *testing.T) {
	db, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	logger := new(MockLogger)
	repo := &containerRepository{db: db, logger: logger}

	mock.ExpectQuery("SELECT count\\(\\*\\) FROM .*containers.*").
		WillReturnError(errors.New("count failed"))

	total, result, err := repo.ViewAllContainers(context.Background(), nil, 0, 10, "", "")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to count containers")
	assert.Equal(t, int64(0), total)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestViewAllContainers_FindError(t *testing.T) {
	db, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	logger := new(MockLogger)
	repo := &containerRepository{db: db, logger: logger}

	mock.ExpectQuery("SELECT count\\(\\*\\) FROM .*containers.*").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery("SELECT .* FROM .*containers.*").
		WillReturnError(errors.New("query failed"))

	total, result, err := repo.ViewAllContainers(context.Background(), nil, 0, 10, "", "")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to retrieve containers")
	assert.Equal(t, int64(0), total)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestViewAllContainers_InvalidSortField(t *testing.T) {
	db, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	logger := new(MockLogger)
	repo := &containerRepository{db: db, logger: logger}

	mock.ExpectQuery("SELECT count\\(\\*\\) FROM .*containers").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0)) // cần dòng này

	total, result, err := repo.ViewAllContainers(context.Background(), nil, 0, 10, "invalid_field", "asc")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid sort field")
	assert.Equal(t, int64(0), total)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestViewAllContainers_LimitLessThanZero(t *testing.T) {
	db, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	logger := new(MockLogger)
	repo := &containerRepository{db: db, logger: logger}

	mock.ExpectQuery("SELECT count\\(\\*\\) FROM .*containers.*").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery("SELECT .* FROM .*containers.*").
		WillReturnRows(sqlmock.NewRows([]string{"id", "container_id", "container_name", "image_name", "status"}).
			AddRow(1, "id1", "name1", "img1", "running"))

	total, result, err := repo.ViewAllContainers(context.Background(), nil, 5, 5, "", "")

	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, result, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

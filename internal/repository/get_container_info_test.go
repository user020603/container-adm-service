package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestGetContainerInfo_Success(t *testing.T) {
	db, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	logger := new(MockLogger)
	repo := &containerRepository{db: db, logger: logger}

	rows := sqlmock.NewRows([]string{"id", "container_name"}).
		AddRow(1, "web").
		AddRow(2, "api")

	mock.ExpectQuery("SELECT id, container_name FROM \"containers\"").
		WillReturnRows(rows)

	result, err := repo.GetContainerInfo(context.Background())

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, uint(1), result[0].Id)
	assert.Equal(t, "web", result[0].ContainerName)
	assert.Equal(t, uint(2), result[1].Id)
	assert.Equal(t, "api", result[1].ContainerName)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetContainerInfo_DBError(t *testing.T) {
	db, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	logger := new(MockLogger)
	repo := &containerRepository{db: db, logger: logger}

	mock.ExpectQuery("SELECT id, container_name FROM \"containers\"").
		WillReturnError(errors.New("query failed"))

	result, err := repo.GetContainerInfo(context.Background())

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to retrieve container names")
	assert.NoError(t, mock.ExpectationsWereMet())
}

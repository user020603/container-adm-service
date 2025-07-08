package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestGetNumContainers_Success(t *testing.T) {
	db, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	logger := new(MockLogger)
	repo := &containerRepository{db: db, logger: logger}

	// Setup mock query
	mock.ExpectQuery(`SELECT count\(\*\) FROM "containers"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	count, err := repo.GetNumContainers(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, int64(5), count)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetNumContainers_Error(t *testing.T) {
	db, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	logger := new(MockLogger)
	repo := &containerRepository{db: db, logger: logger}

	mock.ExpectQuery(`SELECT count\(\*\) FROM "containers"`).
		WillReturnError(errors.New("db count failed"))

	count, err := repo.GetNumContainers(context.Background())
	assert.Error(t, err)
	assert.Equal(t, int64(0), count)
	assert.Contains(t, err.Error(), "failed to count containers")

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetNumRunningContainers_Success(t *testing.T) {
	db, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	logger := new(MockLogger)
	repo := &containerRepository{db: db, logger: logger}

	mock.ExpectQuery(`SELECT count\(\*\) FROM "containers" WHERE status = \$1`).
		WithArgs("running").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

	count, err := repo.GetNumRunningContainers(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, int64(3), count)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetNumRunningContainers_Error(t *testing.T) {
	db, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	logger := new(MockLogger)
	repo := &containerRepository{db: db, logger: logger}

	mock.ExpectQuery(`SELECT count\(\*\) FROM "containers" WHERE status = \$1`).
		WithArgs("running").
		WillReturnError(errors.New("db count error"))

	count, err := repo.GetNumRunningContainers(context.Background())
	assert.Error(t, err)
	assert.Equal(t, int64(0), count)
	assert.Contains(t, err.Error(), "failed to count running containers")
	assert.NoError(t, mock.ExpectationsWereMet())
}

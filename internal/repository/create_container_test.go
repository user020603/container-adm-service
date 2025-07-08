package repository

import (
	"context"
	"errors"
	"testing"
	"thanhnt208/container-adm-service/internal/model"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestCreateContainer_Success(t *testing.T) {
	db, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	logger := new(MockLogger)
	repo := &containerRepository{db: db, logger: logger}

	container := &model.Container{
		ContainerID:   "abc123",
		ContainerName: "test-container",
		ImageName:     "nginx",
		Status:        "running",
	}

	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO .*containers.*").
		WithArgs(
			container.ContainerID,
			container.ContainerName,
			container.ImageName,
			container.Status,
			sqlmock.AnyArg(), // created_at
			sqlmock.AnyArg(), // updated_at
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	id, err := repo.CreateContainer(context.Background(), container)

	assert.NoError(t, err)
	assert.Equal(t, 1, id)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateContainer_BeginError(t *testing.T) {
	db, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	logger := new(MockLogger)
	repo := &containerRepository{db: db, logger: logger}

	container := &model.Container{}

	mock.ExpectBegin().WillReturnError(errors.New("begin error"))

	id, err := repo.CreateContainer(context.Background(), container)

	assert.Error(t, err)
	assert.Equal(t, 0, id)
	assert.Contains(t, err.Error(), "begin error")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateContainer_CreateError(t *testing.T) {
	db, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	logger := new(MockLogger)
	repo := &containerRepository{db: db, logger: logger}

	container := &model.Container{
		ContainerID:   "abc123",
		ContainerName: "test-container",
		ImageName:     "nginx",
		Status:        "running",
	}

	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO .*containers.*").
		WithArgs(container.ContainerID, container.ContainerName, container.ImageName, container.Status, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(errors.New("insert error"))
	mock.ExpectRollback()

	id, err := repo.CreateContainer(context.Background(), container)

	assert.Error(t, err)
	assert.Equal(t, 0, id)
	assert.Contains(t, err.Error(), "insert error")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateContainer_CommitError(t *testing.T) {
	db, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	logger := new(MockLogger)
	repo := &containerRepository{db: db, logger: logger}

	container := &model.Container{
		ContainerID:   "abc123",
		ContainerName: "test-container",
		ImageName:     "nginx",
		Status:        "running",
	}

	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO .*containers.*").
		WithArgs(container.ContainerID, container.ContainerName, container.ImageName, container.Status, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit().WillReturnError(errors.New("commit error"))

	id, err := repo.CreateContainer(context.Background(), container)

	assert.Error(t, err)
	assert.Equal(t, 0, id)
	assert.Contains(t, err.Error(), "commit error")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateContainer_PanicRecoveredAndLogged(t *testing.T) {
	db, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	logger := new(MockLogger)

	repo := &containerRepository{
		db:     db,
		logger: logger,
	}

	container := &model.Container{
		ContainerID:   "panic123",
		ContainerName: "panic-container",
		ImageName:     "nginx",
		Status:        "running",
	}

	mock.ExpectBegin()
	mock.ExpectRollback()

	db.Callback().Create().Replace("gorm:create", func(db *gorm.DB) {
		panic("simulated panic")
	})

	defer func() {
		if r := recover(); r != nil {
			assert.Equal(t, "simulated panic", r)
			assert.NoError(t, mock.ExpectationsWereMet())
			logger.AssertExpectations(t)
		} else {
			t.Errorf("expected panic")
		}
	}()

	_, _ = repo.CreateContainer(context.Background(), container)
}

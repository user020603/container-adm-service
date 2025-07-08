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

// Test CreateManyContainers
func TestCreateManyContainers_Success(t *testing.T) {
	db, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	logger := new(MockLogger)
	repo := &containerRepository{db: db, logger: logger}

	containers := []model.Container{
		{
			ContainerID:   "id1",
			ContainerName: "name1",
			ImageName:     "img1",
			Status:        "running",
		},
		{
			ContainerID:   "id2",
			ContainerName: "name2",
			ImageName:     "img2",
			Status:        "stopped",
		},
	}

	mock.ExpectBegin()

	for _, c := range containers {
		mock.ExpectQuery("INSERT INTO .*containers.*").
			WithArgs(c.ContainerID, c.ContainerName, c.ImageName, c.Status, sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	}

	mock.ExpectCommit()

	created, failed, err := repo.CreateManyContainers(context.Background(), containers)

	assert.NoError(t, err)
	assert.Len(t, created, 2)
	assert.Len(t, failed, 0)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateManyContainers_BeginError(t *testing.T) {
	db, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	logger := new(MockLogger)
	repo := &containerRepository{db: db, logger: logger}

	mock.ExpectBegin().WillReturnError(errors.New("begin error"))

	created, failed, err := repo.CreateManyContainers(context.Background(), []model.Container{})

	assert.Error(t, err)
	assert.Nil(t, created)
	assert.Nil(t, failed)
	assert.Contains(t, err.Error(), "begin error")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateManyContainers_PartialSuccess(t *testing.T) {
	db, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	logger := new(MockLogger)
	repo := &containerRepository{db: db, logger: logger}

	containers := []model.Container{
		{
			ContainerID:   "id1",
			ContainerName: "name1",
			ImageName:     "img1",
			Status:        "running",
		},
		{
			ContainerID:   "id2",
			ContainerName: "name2",
			ImageName:     "img2",
			Status:        "running",
		},
	}

	mock.ExpectBegin()

	// First: success
	mock.ExpectQuery("INSERT INTO .*containers.*").
		WithArgs(containers[0].ContainerID, containers[0].ContainerName, containers[0].ImageName, containers[0].Status, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	// Second: error
	mock.ExpectQuery("INSERT INTO .*containers.*").
		WithArgs(containers[1].ContainerID, containers[1].ContainerName, containers[1].ImageName, containers[1].Status, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(errors.New("insert error"))

	mock.ExpectCommit()

	created, failed, err := repo.CreateManyContainers(context.Background(), containers)

	assert.NoError(t, err)
	assert.Len(t, created, 1)
	assert.Len(t, failed, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateManyContainers_AllFailed(t *testing.T) {
	db, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	logger := new(MockLogger)
	repo := &containerRepository{db: db, logger: logger}

	containers := []model.Container{
		{
			ContainerID:   "id1",
			ContainerName: "name1",
			ImageName:     "img1",
			Status:        "running",
		},
		{
			ContainerID:   "id2",
			ContainerName: "name2",
			ImageName:     "img2",
			Status:        "running",
		},
	}

	mock.ExpectBegin()

	for _, c := range containers {
		mock.ExpectQuery("INSERT INTO .*containers.*").
			WithArgs(c.ContainerID, c.ContainerName, c.ImageName, c.Status, sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnError(errors.New("insert error"))
	}

	mock.ExpectRollback()

	created, failed, err := repo.CreateManyContainers(context.Background(), containers)

	assert.Error(t, err)
	assert.Nil(t, created)
	assert.Len(t, failed, 2)
	assert.Contains(t, err.Error(), "no containers created, all failed")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateManyContainers_CommitError(t *testing.T) {
	db, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	logger := new(MockLogger)
	repo := &containerRepository{db: db, logger: logger}

	containers := []model.Container{
		{
			ContainerID:   "id1",
			ContainerName: "name1",
			ImageName:     "img1",
			Status:        "running",
		},
	}

	mock.ExpectBegin()

	mock.ExpectQuery("INSERT INTO .*containers.*").
		WithArgs(containers[0].ContainerID, containers[0].ContainerName, containers[0].ImageName, containers[0].Status, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	mock.ExpectCommit().WillReturnError(errors.New("commit error"))

	created, failed, err := repo.CreateManyContainers(context.Background(), containers)

	assert.Error(t, err)
	assert.Nil(t, created)
	assert.Len(t, failed, 0)
	assert.Contains(t, err.Error(), "commit error")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateManyContainers_PanicRecovered(t *testing.T) {
	db, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	logger := new(MockLogger)
	repo := &containerRepository{db: db, logger: logger}

	containers := []model.Container{
		{
			ContainerID:   "panic-id",
			ContainerName: "panic-name",
			ImageName:     "nginx",
			Status:        "running",
		},
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
		} else {
			t.Errorf("expected panic")
		}
	}()

	_, _, _ = repo.CreateManyContainers(context.Background(), containers)
}

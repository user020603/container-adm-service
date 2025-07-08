package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"thanhnt208/container-adm-service/internal/dto"
	"thanhnt208/container-adm-service/internal/model"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type MockLogger struct {
	mock.Mock
}

func (l *MockLogger) Error(msg string, keyvals ...interface{}) {}
func (l *MockLogger) Info(msg string, keyvals ...interface{})  {}
func (l *MockLogger) Debug(msg string, keyvals ...interface{}) {}
func (l *MockLogger) Warn(msg string, keyvals ...interface{})  {}
func (l *MockLogger) Fatal(msg string, keyvals ...interface{}) {}
func (l *MockLogger) Sync() error                              { return nil }

func TestNewContainerRepository(t *testing.T) {
	db := &gorm.DB{}
	es := &elasticsearch.Client{}
	loggerMock := new(MockLogger)

	repo := NewContainerRepository(db, es, loggerMock)

	assert.NotNil(t, repo)
}

// Test CreateContainer
func setupMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, *sql.DB) {
	sqlDB, mock, err := sqlmock.New()
	assert.NoError(t, err)

	dialector := postgres.New(postgres.Config{
		Conn:       sqlDB,
		DriverName: "postgres",
	})
	db, err := gorm.Open(dialector, &gorm.Config{})
	assert.NoError(t, err)

	return db, mock, sqlDB
}

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

// Test UpdateContainer
func TestUpdateContainer_Success(t *testing.T) {
	db, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	repo := &containerRepository{db: db, logger: new(MockLogger)}
	id := uint(1)
	updateData := map[string]interface{}{"status": "stopped"}

	mock.ExpectBegin()

	mock.ExpectQuery(`SELECT \* FROM "containers" WHERE "containers"\."id" = \$1 .* LIMIT \$2`).
		WithArgs(id, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "container_id", "container_name", "image_name", "status"}).
			AddRow(1, "abc123", "web", "nginx", "running"))

	mock.ExpectExec(`UPDATE "containers" SET "status"=\$1,"updated_at"=\$2 WHERE "id" = \$3`).
		WithArgs("stopped", sqlmock.AnyArg(), id).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	container, err := repo.UpdateContainer(context.Background(), id, updateData)

	assert.NoError(t, err)
	assert.Equal(t, uint(1), container.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateContainer_BeginError(t *testing.T) {
	db, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	repo := &containerRepository{db: db, logger: new(MockLogger)}

	mock.ExpectBegin().WillReturnError(errors.New("begin error"))

	container, err := repo.UpdateContainer(context.Background(), 1, map[string]interface{}{"status": "stopped"})

	assert.Error(t, err)
	assert.Nil(t, container)
	assert.Contains(t, err.Error(), "begin error")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateContainer_NotFound(t *testing.T) {
	db, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	repo := &containerRepository{db: db, logger: new(MockLogger)}

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT \* FROM "containers" WHERE "containers"\."id" = \$1 .* LIMIT \$2`).
		WithArgs(1, 1).
		WillReturnError(gorm.ErrRecordNotFound)
	mock.ExpectRollback()

	container, err := repo.UpdateContainer(context.Background(), 1, map[string]interface{}{"status": "stopped"})

	assert.Error(t, err)
	assert.Nil(t, container)
	assert.Contains(t, err.Error(), "container not found")
	assert.NoError(t, mock.ExpectationsWereMet())
}


func TestUpdateContainer_UpdateError(t *testing.T) {
	db, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	repo := &containerRepository{db: db, logger: new(MockLogger)}

	mock.ExpectBegin()

	mock.ExpectQuery(`SELECT \* FROM "containers" WHERE "containers"\."id" = \$1 .* LIMIT \$2`).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "container_id", "container_name", "image_name", "status"}).
			AddRow(1, "abc123", "web", "nginx", "running"))

	mock.ExpectExec(`UPDATE "containers" SET "status"=\$1,"updated_at"=\$2 WHERE "id" = \$3`).
		WithArgs("stopped", sqlmock.AnyArg(), 1).
		WillReturnError(errors.New("update failed"))

	mock.ExpectRollback()

	container, err := repo.UpdateContainer(context.Background(), 1, map[string]interface{}{"status": "stopped"})

	assert.Error(t, err)
	assert.Nil(t, container)
	assert.Contains(t, err.Error(), "failed to update container")
	assert.NoError(t, mock.ExpectationsWereMet())
}


func TestUpdateContainer_CommitError(t *testing.T) {
	db, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	repo := &containerRepository{db: db, logger: new(MockLogger)}

	mock.ExpectBegin()

	mock.ExpectQuery(`SELECT \* FROM "containers" WHERE "containers"\."id" = \$1 .* LIMIT \$2`).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "container_id", "container_name", "image_name", "status"}).
			AddRow(1, "abc123", "web", "nginx", "running"))

	mock.ExpectExec(`UPDATE "containers" SET "status"=\$1,"updated_at"=\$2 WHERE "id" = \$3`).
		WithArgs("stopped", sqlmock.AnyArg(), 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit().WillReturnError(errors.New("commit failed"))

	container, err := repo.UpdateContainer(context.Background(), 1, map[string]interface{}{"status": "stopped"})

	assert.Error(t, err)
	assert.Nil(t, container)
	assert.Contains(t, err.Error(), "commit failed")
	assert.NoError(t, mock.ExpectationsWereMet())
}


func TestUpdateContainer_PanicRecovery(t *testing.T) {
	db, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	repo := &containerRepository{db: db, logger: new(MockLogger)}

	mock.ExpectBegin()

	mock.ExpectQuery(`SELECT \* FROM "containers" WHERE "containers"\."id" = \$1 .* LIMIT \$2`).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "container_id", "container_name", "image_name", "status"}).
			AddRow(1, "abc123", "web", "nginx", "running"))

	mock.ExpectRollback()

	db.Callback().Update().Replace("gorm:update", func(db *gorm.DB) {
		panic("simulated panic")
	})

	defer func() {
		if r := recover(); r != nil {
			assert.Equal(t, "simulated panic", r)
			assert.NoError(t, mock.ExpectationsWereMet())
		} else {
			t.Errorf("expected panic but did not occur")
		}
	}()

	_, _ = repo.UpdateContainer(context.Background(), 1, map[string]interface{}{"status": "stopped"})
}

// Test DeleteContainer
func TestDeleteContainer_Success(t *testing.T) {
	db, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()
	repo := &containerRepository{db: db, logger: new(MockLogger)}

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT \* FROM "containers" WHERE "containers"\."id" = \$1 .* LIMIT \$2`).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "container_id", "container_name", "image_name", "status"}).
			AddRow(1, "abc123", "web", "nginx", "running"))

	mock.ExpectExec(`DELETE FROM "containers" WHERE "containers"\."id" = \$1`).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	err := repo.DeleteContainer(context.Background(), 1)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteContainer_NotFound(t *testing.T) {
	db, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()
	repo := &containerRepository{db: db, logger: new(MockLogger)}

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT \* FROM "containers" WHERE "containers"\."id" = \$1 .* LIMIT \$2`).
		WithArgs(1, 1).
		WillReturnError(gorm.ErrRecordNotFound)
	mock.ExpectRollback()

	err := repo.DeleteContainer(context.Background(), 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "container not found")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteContainer_DeleteError(t *testing.T) {
	db, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()
	repo := &containerRepository{db: db, logger: new(MockLogger)}

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT \* FROM "containers" WHERE "containers"\."id" = \$1 .* LIMIT \$2`).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "container_id", "container_name", "image_name", "status"}).
			AddRow(1, "abc123", "web", "nginx", "running"))

	mock.ExpectExec(`DELETE FROM "containers" WHERE "containers"\."id" = \$1`).
		WithArgs(1).
		WillReturnError(errors.New("delete failed"))

	mock.ExpectRollback()

	err := repo.DeleteContainer(context.Background(), 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete container")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteContainer_CommitError(t *testing.T) {
	db, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()
	repo := &containerRepository{db: db, logger: new(MockLogger)}

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT \* FROM "containers" WHERE "containers"\."id" = \$1 .* LIMIT \$2`).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "container_id", "container_name", "image_name", "status"}).
			AddRow(1, "abc123", "web", "nginx", "running"))

	mock.ExpectExec(`DELETE FROM "containers" WHERE "containers"\."id" = \$1`).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit().WillReturnError(errors.New("commit failed"))

	err := repo.DeleteContainer(context.Background(), 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "commit failed")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteContainer_PanicRecovery(t *testing.T) {
	db, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()
	repo := &containerRepository{db: db, logger: new(MockLogger)}

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT \* FROM "containers" WHERE "containers"\."id" = \$1 .* LIMIT \$2`).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "container_id", "container_name", "image_name", "status"}).
			AddRow(1, "abc123", "web", "nginx", "running"))

	mock.ExpectRollback()

	db.Callback().Delete().Replace("gorm:delete", func(db *gorm.DB) {
		panic("simulated panic")
	})

	defer func() {
		if r := recover(); r != nil {
			assert.Equal(t, "simulated panic", r)
			assert.NoError(t, mock.ExpectationsWereMet())
		} else {
			t.Errorf("expected panic but none occurred")
		}
	}()

	_ = repo.DeleteContainer(context.Background(), 1)
}


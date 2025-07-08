package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

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

func TestDeleteContainer_BeginError(t *testing.T) {
	db, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	repo := &containerRepository{db: db, logger: new(MockLogger)}

	mock.ExpectBegin().WillReturnError(errors.New("begin error"))

	err := repo.DeleteContainer(context.Background(), 1)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "begin error")
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

package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

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

package repository

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestGetContainerByID_Success(t *testing.T) {
	db, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	repo := &containerRepository{db: db, logger: new(MockLogger)}
	id := uint(1)

	mock.ExpectQuery(`SELECT \* FROM "containers" WHERE "containers"\."id" = \$1 .* LIMIT \$2`).
		WithArgs(id, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "container_id", "container_name", "image_name", "status"}).
			AddRow(1, "abc123", "web", "nginx", "running"))

	container, err := repo.GetContainerByID(context.Background(), id)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if container.ID != id {
		t.Errorf("expected container ID %d, got %d", id, container.ID)
	}

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetContainerByID_NotFound(t *testing.T) {
	db, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()
	repo := &containerRepository{db: db, logger: new(MockLogger)}

	mock.ExpectQuery(`SELECT \* FROM "containers" WHERE "containers"\."id" = \$1 .* LIMIT \$2`).
		WithArgs(999, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	container, err := repo.GetContainerByID(context.Background(), 999)

	assert.Error(t, err)
	assert.Nil(t, container)
	assert.Contains(t, err.Error(), "container not found")
	assert.NoError(t, mock.ExpectationsWereMet())
}

package repository

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/elastic/go-elasticsearch/esapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type MockLogger struct {
	mock.Mock
}

type MockESClient struct {
	mock.Mock
}

func (m *MockESClient) Do(ctx context.Context, req esapi.Request) (*esapi.Response, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*esapi.Response), args.Error(1)
}

func (l *MockLogger) Error(msg string, keyvals ...interface{}) {}
func (l *MockLogger) Info(msg string, keyvals ...interface{})  {}
func (l *MockLogger) Debug(msg string, keyvals ...interface{}) {}
func (l *MockLogger) Warn(msg string, keyvals ...interface{})  {}
func (l *MockLogger) Fatal(msg string, keyvals ...interface{}) {}
func (l *MockLogger) Sync() error                              { return nil }

func TestNewContainerRepository(t *testing.T) {
	db := &gorm.DB{}
	es := new(MockESClient)
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

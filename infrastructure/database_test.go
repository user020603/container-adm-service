package infrastructure

import (
	"context"
	"fmt"
	"testing"
	"thanhnt208/container-adm-service/config"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/docker/go-connections/nat"
	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestNewDatabase(t *testing.T) {
	cfg := &config.Config{DBName: "testdb"}
	got := NewDatabase(cfg)

	assert.NotNil(t, got)
	assert.Equal(t, cfg, got.(*Database).cfg)
}

func TestDatabase_ConnectDB(t *testing.T) {
	testDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	cfg := &config.Config{
		DBHost:     "localhost",
		DBPort:     "5432",
		DBUser:     "user",
		DBPassword: "pass",
		DBName:     "testdb",
	}

	db := &Database{
		db:  testDB,
		cfg: cfg,
	}

	got, err := db.ConnectDB()
	assert.NoError(t, err)
	assert.Same(t, testDB, got)
}

func TestDatabase_ConnectDB_Failure(t *testing.T) {
	cfg := &config.Config{
		DBHost:     "invalid_host",
		DBPort:     "5432",
		DBUser:     "invalid_user",
		DBPassword: "failure",
		DBName:     "testdb",
	}
	db := &Database{cfg: cfg}
	_, err := db.ConnectDB()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to connect to database")
}

func TestDatabase_Close(t *testing.T) {
	t.Run("close when db is nil", func(t *testing.T) {
		db := &Database{}
		err := db.Close()
		assert.NoError(t, err)
	})

	t.Run("close when db is valid", func(t *testing.T) {
		dbConn, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		assert.NoError(t, err)

		db := &Database{db: dbConn}
		err = db.Close()
		assert.NoError(t, err)
	})
}

func TestDatabase_ConnectDB_WithRealPostgres(t *testing.T) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:latest",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpass",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForSQL("5432/tcp", "postgres", func(host string, port nat.Port) string {
			return fmt.Sprintf("host=%s port=%s user=testuser password=testpass dbname=testdb sslmode=disable", host, port.Port())
		}).WithStartupTimeout(30 * time.Second),
	}
	postgresC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	assert.NoError(t, err)
	defer postgresC.Terminate(ctx)

	host, _ := postgresC.Host(ctx)
	port, _ := postgresC.MappedPort(ctx, "5432")

	cfg := &config.Config{
		DBHost:     host,
		DBPort:     port.Port(),
		DBUser:     "testuser",
		DBPassword: "testpass",
		DBName:     "testdb",
	}

	db := &Database{cfg: cfg}
	conn, err := db.ConnectDB()
	assert.NoError(t, err)
	assert.NotNil(t, conn)
}

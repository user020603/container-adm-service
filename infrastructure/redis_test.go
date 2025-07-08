package infrastructure

import (
	"context"
	"testing"
	"thanhnt208/container-adm-service/config"

	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestRedis_Ping_Mocked(t *testing.T) {
	db, mock := redismock.NewClientMock()
	defer db.Close()

	mock.ExpectPing().SetVal("PONG")

	redis := &Redis{
		client: db,
	}

	err := redis.Ping(context.Background())
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRedis_Ping_ConnectFallback_WithContainer(t *testing.T) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "redis:latest",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForListeningPort("6379/tcp"),
	}
	redisC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	assert.NoError(t, err)
	defer func() { _ = redisC.Terminate(ctx) }()

	host, _ := redisC.Host(ctx)
	port, _ := redisC.MappedPort(ctx, "6379")

	cfg := &config.Config{
		RedisAddr: host + ":" + port.Port(),
	}

	r := NewRedis(cfg)

	err = r.Ping(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, r.(*Redis).client)
}

func TestRedis_Ping_FailedConnect(t *testing.T) {
	r := &Redis{
		cfg: &config.Config{RedisAddr: "invalid:9999"},
	}

	err := r.Ping(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to connect to Redis")
}

func TestRedis_Close_WithClient(t *testing.T) {
	db, _ := redismock.NewClientMock()

	r := &Redis{
		client: db,
	}

	err := r.Close()
	assert.NoError(t, err, "Close should not return an error when client is initialized")
}

func TestRedis_Close_WithoutClient(t *testing.T) {
	r := &Redis{
		client: nil,
	}

	err := r.Close()
	assert.NoError(t, err, "Close should not return an error when client is nil")
}

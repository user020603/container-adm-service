package infrastructure

import (
	"context"
	"fmt"
	"thanhnt208/container-adm-service/config"
	"time"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	client *redis.Client
	cfg    *config.Config
}

func NewRedis(cfg *config.Config) IRedis {
	return &Redis{
		cfg: cfg,
	}
}

func (r *Redis) ConnectClient() (*redis.Client, error) {
	if r.client != nil {
		return r.client, nil
	}

	r.client = redis.NewClient(&redis.Options{
		Addr:     r.cfg.RedisAddr,
		Password: r.cfg.RedisPassword,
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := r.client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return r.client, nil
}

func (r *Redis) Ping(ctx context.Context) error {
	if r.client == nil {
		_, err := r.ConnectClient()
		if err != nil {
			return fmt.Errorf("failed to connect to Redis: %w", err)
		}
	}
	return r.client.Ping(ctx).Err()
}

func (r *Redis) Close() error {
	if r.client != nil {
		return r.client.Close()
	}
	return nil
}

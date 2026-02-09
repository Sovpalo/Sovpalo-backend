package repository

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type RedisHealthRepository struct {
	client *redis.Client
}

func NewRedisHealthRepository(client *redis.Client) *RedisHealthRepository {
	return &RedisHealthRepository{client: client}
}

func (r *RedisHealthRepository) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

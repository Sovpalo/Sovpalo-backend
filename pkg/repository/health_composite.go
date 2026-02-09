package repository

import "context"

type CompositeHealthRepository struct {
	postgres HealthRepository
	redis    HealthRepository
}

func NewCompositeHealthRepository(postgres HealthRepository, redis HealthRepository) *CompositeHealthRepository {
	return &CompositeHealthRepository{
		postgres: postgres,
		redis:    redis,
	}
}

func (r *CompositeHealthRepository) Ping(ctx context.Context) error {
	if err := r.postgres.Ping(ctx); err != nil {
		return err
	}
	if err := r.redis.Ping(ctx); err != nil {
		return err
	}
	return nil
}

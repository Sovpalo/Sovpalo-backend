package repository

import "github.com/jackc/pgx/v5/pgxpool"

type AvailabilityPostgres struct {
	pool *pgxpool.Pool
}

func NewAvailabilityRepository(pool *pgxpool.Pool) *AvailabilityPostgres {
	return &AvailabilityPostgres{pool: pool}
}

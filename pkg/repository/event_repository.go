package repository

import "github.com/jackc/pgx/v5/pgxpool"

type EventPostgres struct {
	pool *pgxpool.Pool
}

func NewEventRepository(pool *pgxpool.Pool) *EventPostgres {
	return &EventPostgres{pool: pool}
}

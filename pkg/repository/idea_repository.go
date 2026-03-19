package repository

import "github.com/jackc/pgx/v5/pgxpool"

type IdeaPostgres struct {
	pool *pgxpool.Pool
}

func NewIdeaRepository(pool *pgxpool.Pool) *IdeaPostgres {
	return &IdeaPostgres{pool: pool}
}

package repository

import "github.com/jackc/pgx/v5/pgxpool"

type ChatPostgres struct {
	pool *pgxpool.Pool
}

func NewChatRepository(pool *pgxpool.Pool) *ChatPostgres {
	return &ChatPostgres{pool: pool}
}

package repository

import "github.com/jackc/pgx/v5/pgxpool"

type NotificationPostgres struct {
	pool *pgxpool.Pool
}

func NewNotificationRepository(pool *pgxpool.Pool) *NotificationPostgres {
	return &NotificationPostgres{pool: pool}
}

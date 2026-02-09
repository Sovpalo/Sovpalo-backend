package repository

import (
	"context"
	"database/sql"

	"github.com/Sovpalo/sovpalo-backend/pkg/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthPostgres struct {
	pool *pgxpool.Pool
}

func NewAuthPostgres(pool *pgxpool.Pool) *AuthPostgres {
	return &AuthPostgres{pool: pool}
}

func (r *AuthPostgres) UserExists(email string) (bool, error) {
	query := "SELECT COUNT(*) FROM users WHERE email = $1"
	var count int
	err := r.pool.QueryRow(context.Background(), query, email).Scan(&count)
	if err != nil {
		if err == sql.ErrNoRows || err == pgx.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return count > 0, nil
}

func (r *AuthPostgres) CreateUser(user model.User) (int, error) {
	var id int
	query := "INSERT INTO users (username, email, password) VALUES ($1, $2, $3) RETURNING id"
	row := r.pool.QueryRow(context.Background(), query, user.Username, user.Email, user.Password)
	if err := row.Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func (r *AuthPostgres) GetUser(email, password string) (model.User, error) {
	var user model.User
	query := "SELECT id FROM users WHERE email = $1 AND password_hash = $2"
	err := r.pool.QueryRow(context.Background(), query, email, password).Scan(&user.ID)
	return user, err
}

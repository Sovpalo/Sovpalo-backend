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

func (r *AuthPostgres) UsernameExists(username string) (bool, error) {
	query := "SELECT COUNT(*) FROM users WHERE username = $1"
	var count int
	err := r.pool.QueryRow(context.Background(), query, username).Scan(&count)
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
	query := "SELECT id FROM users WHERE email = $1 AND password = $2"
	err := r.pool.QueryRow(context.Background(), query, email, password).Scan(&user.ID)
	return user, err
}

func (r *AuthPostgres) GetUserByEmail(email string) (model.User, error) {
	var user model.User
	query := "SELECT id, email, username, password FROM users WHERE email = $1"
	err := r.pool.QueryRow(context.Background(), query, email).Scan(&user.ID, &user.Email, &user.Username, &user.Password)
	return user, err
}

func (r *AuthPostgres) GetUserByID(userID int64) (model.User, error) {
	var user model.User
	query := "SELECT id, email, username, avatar_url FROM users WHERE id = $1"
	err := r.pool.QueryRow(context.Background(), query, userID).Scan(&user.ID, &user.Email, &user.Username, &user.AvatarURL)
	return user, err
}

func (r *AuthPostgres) UpdateUserAvatar(userID int64, avatarURL *string) error {
	query := "UPDATE users SET avatar_url = $1, updated_at = NOW() WHERE id = $2"
	tag, err := r.pool.Exec(context.Background(), query, avatarURL, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *AuthPostgres) DeleteUser(userID int64) error {
	ctx := context.Background()
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "DELETE FROM companies WHERE created_by = $1", userID); err != nil {
		return err
	}

	tag, err := tx.Exec(ctx, "DELETE FROM users WHERE id = $1", userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return tx.Commit(ctx)
}

func (r *AuthPostgres) UpdateUserPassword(email string, passwordHash string) error {
	query := "UPDATE users SET password = $1, updated_at = NOW() WHERE email = $2"
	_, err := r.pool.Exec(context.Background(), query, passwordHash, email)
	return err
}

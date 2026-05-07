package repository

import (
	"context"
	"database/sql"
	"fmt"

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
	query := "SELECT COUNT(*) FROM auth_identities WHERE provider = $1 AND email = $2"
	var count int
	err := r.pool.QueryRow(context.Background(), query, model.AuthProviderPassword, email).Scan(&count)
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
	ctx := context.Background()
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	var id int
	query := "INSERT INTO users (username, avatar_url) VALUES ($1, $2) RETURNING id"
	row := tx.QueryRow(ctx, query, user.Username, user.AvatarURL)
	if err := row.Scan(&id); err != nil {
		return 0, err
	}

	hasPrimary := false
	if user.Email != nil && user.Password != "" {
		if err := r.insertIdentity(ctx, tx, int64(id), model.AuthProviderPassword, nil, user.Email, &user.Password, true); err != nil {
			return 0, err
		}
		hasPrimary = true
	}

	if user.TelegramID != nil {
		providerUserID := int64ToString(*user.TelegramID)
		if err := r.insertIdentity(ctx, tx, int64(id), model.AuthProviderTelegram, &providerUserID, nil, nil, !hasPrimary); err != nil {
			return 0, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}

	return id, nil
}

func (r *AuthPostgres) GetUser(email, password string) (model.User, error) {
	var user model.User
	query := `
		SELECT u.id
		FROM users u
		JOIN auth_identities ai ON ai.user_id = u.id
		WHERE ai.provider = $1 AND ai.email = $2 AND ai.password_hash = $3
		LIMIT 1
	`
	err := r.pool.QueryRow(context.Background(), query, model.AuthProviderPassword, email, password).Scan(&user.ID)
	return user, err
}

func (r *AuthPostgres) GetUserByEmail(email string) (model.User, error) {
	var user model.User
	query := `
		SELECT u.id, ai.email, u.username, ai.password_hash
		FROM users u
		JOIN auth_identities ai ON ai.user_id = u.id
		WHERE ai.provider = $1 AND ai.email = $2
		LIMIT 1
	`
	err := r.pool.QueryRow(context.Background(), query, model.AuthProviderPassword, email).Scan(&user.ID, &user.Email, &user.Username, &user.Password)
	return user, err
}

func (r *AuthPostgres) GetUserByTelegramID(telegramID int64) (model.User, error) {
	var user model.User
	query := `
		SELECT u.id, u.username
		FROM users u
		JOIN auth_identities ai ON ai.user_id = u.id
		WHERE ai.provider = $1 AND ai.provider_user_id = $2
		LIMIT 1
	`
	err := r.pool.QueryRow(context.Background(), query, model.AuthProviderTelegram, int64ToString(telegramID)).Scan(&user.ID, &user.Username)
	if err != nil {
		return user, err
	}
	user.TelegramID = &telegramID
	return user, err
}

func (r *AuthPostgres) GetUserByID(userID int64) (model.User, error) {
	var user model.User
	query := `
		SELECT
			u.id,
			(
				SELECT ai.email
				FROM auth_identities ai
				WHERE ai.user_id = u.id AND ai.provider = $2
				ORDER BY ai.is_primary DESC, ai.id ASC
				LIMIT 1
			) AS email,
			u.username,
			u.avatar_url
		FROM users u
		WHERE u.id = $1
	`
	err := r.pool.QueryRow(context.Background(), query, userID, model.AuthProviderPassword).Scan(&user.ID, &user.Email, &user.Username, &user.AvatarURL)
	if err != nil {
		return user, err
	}

	providers, err := r.listUserProviders(context.Background(), userID)
	if err != nil {
		return model.User{}, err
	}
	user.Providers = providers

	return user, nil
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
	query := "UPDATE auth_identities SET password_hash = $1, updated_at = NOW() WHERE provider = $2 AND email = $3"
	tag, err := r.pool.Exec(context.Background(), query, passwordHash, model.AuthProviderPassword, email)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return err
}

func (r *AuthPostgres) insertIdentity(
	ctx context.Context,
	tx pgx.Tx,
	userID int64,
	provider model.AuthProvider,
	providerUserID *string,
	email *string,
	passwordHash *string,
	isPrimary bool,
) error {
	query := `
		INSERT INTO auth_identities (user_id, provider, provider_user_id, email, password_hash, is_primary)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := tx.Exec(ctx, query, userID, provider, providerUserID, email, passwordHash, isPrimary)
	return err
}

func (r *AuthPostgres) listUserProviders(ctx context.Context, userID int64) ([]model.AuthProvider, error) {
	rows, err := r.pool.Query(ctx, "SELECT provider FROM auth_identities WHERE user_id = $1 ORDER BY id", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	providers := make([]model.AuthProvider, 0, 2)
	for rows.Next() {
		var provider string
		if err := rows.Scan(&provider); err != nil {
			return nil, err
		}
		providers = append(providers, model.AuthProvider(provider))
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	if len(providers) == 0 {
		return nil, pgx.ErrNoRows
	}

	return providers, nil
}

func int64ToString(value int64) string {
	return fmt.Sprintf("%d", value)
}

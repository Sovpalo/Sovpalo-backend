package repository

import (
	"context"

	"github.com/Sovpalo/sovpalo-backend/pkg/model"
)

func (r *NotificationPostgres) UpsertPushDeviceToken(userID int64, token string, platform string) error {
	ctx := context.Background()
	_, err := r.pool.Exec(ctx, `
		INSERT INTO push_device_tokens (user_id, token, platform)
		VALUES ($1, $2, $3)
		ON CONFLICT (token) DO UPDATE
		SET user_id = EXCLUDED.user_id,
		    platform = EXCLUDED.platform,
		    updated_at = NOW()
	`, userID, token, platform)
	return err
}

func (r *NotificationPostgres) DeletePushDeviceToken(userID int64, token string) error {
	ctx := context.Background()
	_, err := r.pool.Exec(ctx, "DELETE FROM push_device_tokens WHERE user_id = $1 AND token = $2", userID, token)
	return err
}

func (r *NotificationPostgres) CreateNotification(notification model.PushNotification) error {
	ctx := context.Background()
	_, err := r.pool.Exec(ctx, `
		INSERT INTO notifications (user_id, type, title, message, related_entity_type, related_entity_id)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, notification.UserID, notification.Type, notification.Title, notification.Message, notification.RelatedEntityType, notification.RelatedEntityID)
	return err
}

func (r *NotificationPostgres) ListPushTokens(userID int64) ([]model.PushDeviceToken, error) {
	ctx := context.Background()
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, token, platform, created_at, updated_at
		FROM push_device_tokens
		WHERE user_id = $1
		ORDER BY updated_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tokens := []model.PushDeviceToken{}
	for rows.Next() {
		var token model.PushDeviceToken
		if err := rows.Scan(&token.ID, &token.UserID, &token.Token, &token.Platform, &token.CreatedAt, &token.UpdatedAt); err != nil {
			return nil, err
		}
		tokens = append(tokens, token)
	}
	return tokens, rows.Err()
}

func (r *NotificationPostgres) RemovePushDeviceToken(token string) error {
	ctx := context.Background()
	_, err := r.pool.Exec(ctx, "DELETE FROM push_device_tokens WHERE token = $1", token)
	return err
}

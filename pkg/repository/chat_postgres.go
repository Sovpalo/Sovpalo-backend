package repository

import (
	"context"
	"errors"
	"time"

	"github.com/Sovpalo/sovpalo-backend/pkg/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func (r *ChatPostgres) ListCompanyChatMessages(companyID int64, userID int64, beforeMessageID int64, limit int) ([]model.ChatMessageView, error) {
	ctx := context.Background()
	if err := r.ensureCompanyMember(ctx, companyID, userID); err != nil {
		return nil, err
	}

	query := `
		SELECT m.id, m.company_id, m.sender_id, u.username, u.avatar_url, m.text, m.created_at, rr.read_at
		FROM company_chat_messages m
		JOIN users u ON u.id = m.sender_id
		LEFT JOIN company_chat_message_reads rr
			ON rr.message_id = m.id AND rr.user_id = $2
		WHERE m.company_id = $1
		  AND ($3 = 0 OR m.id < $3)
		ORDER BY m.id DESC
		LIMIT $4
	`

	rows, err := r.pool.Query(ctx, query, companyID, userID, beforeMessageID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages := make([]model.ChatMessageView, 0, limit)
	messageIDs := make([]int64, 0, limit)
	for rows.Next() {
		var item model.ChatMessageView
		if err := rows.Scan(
			&item.ID,
			&item.CompanyID,
			&item.SenderID,
			&item.SenderUsername,
			&item.SenderAvatarURL,
			&item.Text,
			&item.CreatedAt,
			&item.ReadAt,
		); err != nil {
			return nil, err
		}
		item.IsReadByCurrentUser = item.ReadAt != nil
		item.Attachments = []model.ChatAttachment{}
		messages = append(messages, item)
		messageIDs = append(messageIDs, item.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	attachmentsByMessage, err := r.listAttachmentsByMessageIDs(ctx, messageIDs)
	if err != nil {
		return nil, err
	}
	for i := range messages {
		messages[i].Attachments = attachmentsByMessage[messages[i].ID]
	}

	// Reverse to chronological order for chat rendering.
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

func (r *ChatPostgres) CreateCompanyChatMessage(companyID int64, userID int64, text *string, attachments []model.ChatAttachmentCreate) (model.ChatMessageView, error) {
	ctx := context.Background()
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return model.ChatMessageView{}, err
	}
	defer tx.Rollback(ctx)

	if err := r.ensureCompanyMemberTx(ctx, tx, companyID, userID); err != nil {
		return model.ChatMessageView{}, err
	}

	var senderUsername string
	var senderAvatarURL *string
	if err := tx.QueryRow(ctx, "SELECT username, avatar_url FROM users WHERE id = $1", userID).Scan(&senderUsername, &senderAvatarURL); err != nil {
		return model.ChatMessageView{}, err
	}

	var message model.ChatMessageView
	if err := tx.QueryRow(
		ctx,
		`INSERT INTO company_chat_messages (company_id, sender_id, text)
		 VALUES ($1, $2, $3)
		 RETURNING id, company_id, sender_id, text, created_at`,
		companyID, userID, text,
	).Scan(&message.ID, &message.CompanyID, &message.SenderID, &message.Text, &message.CreatedAt); err != nil {
		return model.ChatMessageView{}, err
	}

	message.SenderUsername = senderUsername
	message.SenderAvatarURL = senderAvatarURL
	message.Attachments = make([]model.ChatAttachment, 0, len(attachments))

	for _, attachment := range attachments {
		var saved model.ChatAttachment
		if err := tx.QueryRow(
			ctx,
			`INSERT INTO company_chat_attachments (message_id, file_name, file_url, file_type, file_size, media_type)
			 VALUES ($1, $2, $3, $4, $5, $6)
			 RETURNING id, message_id, file_name, file_url, file_type, file_size, media_type, created_at`,
			message.ID,
			attachment.FileName,
			attachment.FileURL,
			attachment.FileType,
			attachment.FileSize,
			attachment.MediaType,
		).Scan(
			&saved.ID,
			&saved.MessageID,
			&saved.FileName,
			&saved.FileURL,
			&saved.FileType,
			&saved.FileSize,
			&saved.MediaType,
			&saved.CreatedAt,
		); err != nil {
			return model.ChatMessageView{}, err
		}
		message.Attachments = append(message.Attachments, saved)
	}

	if _, err := tx.Exec(
		ctx,
		`INSERT INTO company_chat_message_reads (message_id, user_id, read_at)
		 VALUES ($1, $2, NOW())`,
		message.ID,
		userID,
	); err != nil {
		return model.ChatMessageView{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return model.ChatMessageView{}, err
	}

	now := time.Now().UTC()
	message.IsReadByCurrentUser = true
	message.ReadAt = &now
	return message, nil
}

func (r *ChatPostgres) ListCompanyChatRecipientIDs(companyID int64, senderID int64) ([]int64, error) {
	ctx := context.Background()
	if err := r.ensureCompanyMember(ctx, companyID, senderID); err != nil {
		return nil, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT user_id
		FROM company_members
		WHERE company_id = $1 AND user_id <> $2
	`, companyID, senderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	recipients := []int64{}
	for rows.Next() {
		var userID int64
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		recipients = append(recipients, userID)
	}
	return recipients, rows.Err()
}

func (r *ChatPostgres) DeleteCompanyChatMessage(companyID int64, messageID int64, userID int64) ([]model.ChatAttachment, error) {
	ctx := context.Background()
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	if err := r.ensureCompanyMemberTx(ctx, tx, companyID, userID); err != nil {
		return nil, err
	}

	var senderID int64
	err = tx.QueryRow(ctx, "SELECT sender_id FROM company_chat_messages WHERE id = $1 AND company_id = $2", messageID, companyID).Scan(&senderID)
	if err != nil {
		return nil, err
	}
	if senderID != userID {
		return nil, errors.New("only author can delete message")
	}

	rows, err := tx.Query(
		ctx,
		`SELECT id, message_id, file_name, file_url, file_type, file_size, media_type, created_at
		 FROM company_chat_attachments
		 WHERE message_id = $1
		 ORDER BY id ASC`,
		messageID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	attachments := make([]model.ChatAttachment, 0)
	for rows.Next() {
		var item model.ChatAttachment
		if err := rows.Scan(
			&item.ID,
			&item.MessageID,
			&item.FileName,
			&item.FileURL,
			&item.FileType,
			&item.FileSize,
			&item.MediaType,
			&item.CreatedAt,
		); err != nil {
			return nil, err
		}
		attachments = append(attachments, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	tag, err := tx.Exec(ctx, "DELETE FROM company_chat_messages WHERE id = $1 AND company_id = $2", messageID, companyID)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, pgx.ErrNoRows
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return attachments, nil
}

func (r *ChatPostgres) MarkCompanyChatMessagesRead(companyID int64, userID int64, messageIDs []int64) ([]int64, time.Time, error) {
	ctx := context.Background()
	if err := r.ensureCompanyMember(ctx, companyID, userID); err != nil {
		return nil, time.Time{}, err
	}

	readAt := time.Now().UTC()
	rows, err := r.pool.Query(
		ctx,
		`INSERT INTO company_chat_message_reads (message_id, user_id, read_at)
		 SELECT m.id, $2, $4
		 FROM company_chat_messages m
		 WHERE m.company_id = $1
		   AND m.id = ANY($3)
		   AND m.sender_id <> $2
		 ON CONFLICT (message_id, user_id) DO NOTHING
		 RETURNING message_id`,
		companyID,
		userID,
		messageIDs,
		readAt,
	)
	if err != nil {
		return nil, time.Time{}, err
	}
	defer rows.Close()

	inserted := make([]int64, 0, len(messageIDs))
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, time.Time{}, err
		}
		inserted = append(inserted, id)
	}
	if err := rows.Err(); err != nil {
		return nil, time.Time{}, err
	}

	return inserted, readAt, nil
}

func (r *ChatPostgres) GetCompanyChatUnreadCount(companyID int64, userID int64) (int64, error) {
	ctx := context.Background()
	if err := r.ensureCompanyMember(ctx, companyID, userID); err != nil {
		return 0, err
	}

	var count int64
	err := r.pool.QueryRow(
		ctx,
		`SELECT COUNT(*)
		 FROM company_chat_messages m
		 LEFT JOIN company_chat_message_reads rr
		   ON rr.message_id = m.id AND rr.user_id = $2
		 WHERE m.company_id = $1
		   AND m.sender_id <> $2
		   AND rr.message_id IS NULL`,
		companyID,
		userID,
	).Scan(&count)
	return count, err
}

func (r *ChatPostgres) ensureCompanyMember(ctx context.Context, companyID int64, userID int64) error {
	var exists bool
	err := r.pool.QueryRow(
		ctx,
		"SELECT EXISTS (SELECT 1 FROM company_members WHERE company_id = $1 AND user_id = $2)",
		companyID,
		userID,
	).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("user is not a member of the company")
	}
	return nil
}

func (r *ChatPostgres) ensureCompanyMemberTx(ctx context.Context, tx pgx.Tx, companyID int64, userID int64) error {
	var exists bool
	err := tx.QueryRow(
		ctx,
		"SELECT EXISTS (SELECT 1 FROM company_members WHERE company_id = $1 AND user_id = $2)",
		companyID,
		userID,
	).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("user is not a member of the company")
	}
	return nil
}

func (r *ChatPostgres) listAttachmentsByMessageIDs(ctx context.Context, messageIDs []int64) (map[int64][]model.ChatAttachment, error) {
	result := make(map[int64][]model.ChatAttachment, len(messageIDs))
	if len(messageIDs) == 0 {
		return result, nil
	}

	rows, err := r.pool.Query(
		ctx,
		`SELECT id, message_id, file_name, file_url, file_type, file_size, media_type, created_at
		 FROM company_chat_attachments
		 WHERE message_id = ANY($1)
		 ORDER BY id ASC`,
		messageIDs,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item model.ChatAttachment
		if err := rows.Scan(
			&item.ID,
			&item.MessageID,
			&item.FileName,
			&item.FileURL,
			&item.FileType,
			&item.FileSize,
			&item.MediaType,
			&item.CreatedAt,
		); err != nil {
			return nil, err
		}
		result[item.MessageID] = append(result[item.MessageID], item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for _, messageID := range messageIDs {
		if _, ok := result[messageID]; !ok {
			result[messageID] = []model.ChatAttachment{}
		}
	}

	return result, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

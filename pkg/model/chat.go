package model

import "time"

type ChatAttachment struct {
	ID        int64     `db:"id" json:"id"`
	MessageID int64     `db:"message_id" json:"message_id"`
	FileName  string    `db:"file_name" json:"file_name"`
	FileURL   string    `db:"file_url" json:"file_url"`
	FileType  string    `db:"file_type" json:"file_type"`
	FileSize  int64     `db:"file_size" json:"file_size"`
	MediaType string    `db:"media_type" json:"media_type"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type ChatAttachmentCreate struct {
	FileName  string
	FileURL   string
	FileType  string
	FileSize  int64
	MediaType string
}

type ChatMessage struct {
	ID        int64      `db:"id" json:"id"`
	CompanyID int64      `db:"company_id" json:"company_id"`
	SenderID  int64      `db:"sender_id" json:"sender_id"`
	Text      *string    `db:"text" json:"text,omitempty"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	DeletedAt *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`
}

type ChatMessageView struct {
	ID                  int64            `json:"id"`
	CompanyID           int64            `json:"company_id"`
	SenderID            int64            `json:"sender_id"`
	SenderUsername      string           `json:"sender_username"`
	SenderAvatarURL     *string          `json:"sender_avatar_url,omitempty"`
	Text                *string          `json:"text,omitempty"`
	Attachments         []ChatAttachment `json:"attachments"`
	CreatedAt           time.Time        `json:"created_at"`
	IsReadByCurrentUser bool             `json:"is_read_by_current_user"`
	ReadAt              *time.Time       `json:"read_at,omitempty"`
}

type ChatReadResult struct {
	MessageIDs  []int64   `json:"message_ids"`
	ReadAt      time.Time `json:"read_at"`
	UnreadCount int64     `json:"unread_count"`
}

type ChatRealtimeEvent struct {
	Type        string           `json:"type"`
	Message     *ChatMessageView `json:"message,omitempty"`
	MessageID   *int64           `json:"message_id,omitempty"`
	MessageIDs  []int64          `json:"message_ids,omitempty"`
	CompanyID   int64            `json:"company_id"`
	UserID      *int64           `json:"user_id,omitempty"`
	UnreadCount *int64           `json:"unread_count,omitempty"`
	ReadAt      *time.Time       `json:"read_at,omitempty"`
}

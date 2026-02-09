package model

import (
	"encoding/json"
	"net"
	"time"
)

type User struct {
	ID         int64     `db:"id" json:"id"`
	Email      string    `db:"email" json:"email"`
	TelegramID *int64    `db:"telegram_id" json:"telegram_id,omitempty"`
	Username   string    `db:"username" json:"username"`
	AvatarURL  *string   `db:"avatar_url" json:"avatar_url,omitempty"`
	Password   string    `db:"password" json:"password"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time `db:"updated_at" json:"updated_at"`
}

type PasswordResetToken struct {
	ID        int64     `db:"id" json:"id"`
	UserID    int64     `db:"user_id" json:"user_id"`
	TokenHash string    `db:"token_hash" json:"token_hash"`
	ExpiresAt time.Time `db:"expires_at" json:"expires_at"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type Group struct {
	ID          int64     `db:"id" json:"id"`
	Name        string    `db:"name" json:"name"`
	Description *string   `db:"description" json:"description,omitempty"`
	AvatarURL   *string   `db:"avatar_url" json:"avatar_url,omitempty"`
	InviteCode  *string   `db:"invite_code" json:"invite_code,omitempty"`
	CreatedBy   int64     `db:"created_by" json:"created_by"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

type GroupMember struct {
	ID       int64     `db:"id" json:"id"`
	GroupID  int64     `db:"group_id" json:"group_id"`
	UserID   int64     `db:"user_id" json:"user_id"`
	Role     string    `db:"role" json:"role"`
	JoinedAt time.Time `db:"joined_at" json:"joined_at"`
}

type Company struct {
	ID          int64     `db:"id" json:"id"`
	Name        string    `db:"name" json:"name"`
	Description *string   `db:"description" json:"description,omitempty"`
	CreatedBy   int64     `db:"created_by" json:"created_by"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

type CompanyMember struct {
	ID        int64     `db:"id" json:"id"`
	CompanyID int64     `db:"company_id" json:"company_id"`
	UserID    int64     `db:"user_id" json:"user_id"`
	Role      string    `db:"role" json:"role"`
	JoinedAt  time.Time `db:"joined_at" json:"joined_at"`
}

type CompanyInvitation struct {
	ID            int64      `db:"id" json:"id"`
	CompanyID     int64      `db:"company_id" json:"company_id"`
	InvitedUserID int64      `db:"invited_user_id" json:"invited_user_id"`
	InvitedBy     int64      `db:"invited_by" json:"invited_by"`
	Status        string     `db:"status" json:"status"`
	CreatedAt     time.Time  `db:"created_at" json:"created_at"`
	RespondedAt   *time.Time `db:"responded_at" json:"responded_at,omitempty"`
}

type CompanyInvitationView struct {
	ID                int64     `db:"id" json:"id"`
	CompanyID         int64     `db:"company_id" json:"company_id"`
	CompanyName       string    `db:"company_name" json:"company_name"`
	InvitedBy         int64     `db:"invited_by" json:"invited_by"`
	InvitedByUsername string    `db:"invited_by_username" json:"invited_by_username"`
	Status            string    `db:"status" json:"status"`
	CreatedAt         time.Time `db:"created_at" json:"created_at"`
}

type CompanyUpdateInput struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

type Event struct {
	ID          int64      `db:"id" json:"id"`
	GroupID     int64      `db:"group_id" json:"group_id"`
	CreatedBy   int64      `db:"created_by" json:"created_by"`
	Title       string     `db:"title" json:"title"`
	Description *string    `db:"description" json:"description,omitempty"`
	StartTime   *time.Time `db:"start_time" json:"start_time,omitempty"`
	EndTime     *time.Time `db:"end_time" json:"end_time,omitempty"`
	PlaceName   *string    `db:"place_name" json:"place_name,omitempty"`
	PlaceLink   *string    `db:"place_link" json:"place_link,omitempty"`
	Status      string     `db:"status" json:"status"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at" json:"updated_at"`
}

type EventParticipant struct {
	ID        int64     `db:"id" json:"id"`
	EventID   int64     `db:"event_id" json:"event_id"`
	UserID    int64     `db:"user_id" json:"user_id"`
	Status    string    `db:"status" json:"status"`
	Notified  bool      `db:"notified" json:"notified"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type Idea struct {
	ID          int64     `db:"id" json:"id"`
	GroupID     int64     `db:"group_id" json:"group_id"`
	CreatedBy   int64     `db:"created_by" json:"created_by"`
	Title       string    `db:"title" json:"title"`
	Description *string   `db:"description" json:"description,omitempty"`
	Source      string    `db:"source" json:"source"`
	LLMPrompt   *string   `db:"llm_prompt" json:"llm_prompt,omitempty"`
	IsSaved     bool      `db:"is_saved" json:"is_saved"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

type UserAvailability struct {
	ID        int64     `db:"id" json:"id"`
	UserID    int64     `db:"user_id" json:"user_id"`
	GroupID   int64     `db:"group_id" json:"group_id"`
	StartTime time.Time `db:"start_time" json:"start_time"`
	EndTime   time.Time `db:"end_time" json:"end_time"`
	Note      *string   `db:"note" json:"note,omitempty"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type MediaArchive struct {
	ID           int64           `db:"id" json:"id"`
	GroupID      int64           `db:"group_id" json:"group_id"`
	UploadedBy   int64           `db:"uploaded_by" json:"uploaded_by"`
	EventID      *int64          `db:"event_id" json:"event_id,omitempty"`
	FileName     string          `db:"file_name" json:"file_name"`
	FileURL      string          `db:"file_url" json:"file_url"`
	FileType     string          `db:"file_type" json:"file_type"`
	FileSize     int64           `db:"file_size" json:"file_size"`
	ThumbnailURL *string         `db:"thumbnail_url" json:"thumbnail_url,omitempty"`
	Description  *string         `db:"description" json:"description,omitempty"`
	Metadata     json.RawMessage `db:"metadata" json:"metadata,omitempty"`
	CreatedAt    time.Time       `db:"created_at" json:"created_at"`
}

type Notification struct {
	ID                int64     `db:"id" json:"id"`
	UserID            int64     `db:"user_id" json:"user_id"`
	Type              string    `db:"type" json:"type"`
	Title             string    `db:"title" json:"title"`
	Message           string    `db:"message" json:"message"`
	RelatedEntityType *string   `db:"related_entity_type" json:"related_entity_type,omitempty"`
	RelatedEntityID   *int64    `db:"related_entity_id" json:"related_entity_id,omitempty"`
	IsRead            bool      `db:"is_read" json:"is_read"`
	CreatedAt         time.Time `db:"created_at" json:"created_at"`
}

type UserSession struct {
	ID               int64     `db:"id" json:"id"`
	UserID           int64     `db:"user_id" json:"user_id"`
	RefreshTokenHash string    `db:"refresh_token_hash" json:"refresh_token_hash"`
	UserAgent        *string   `db:"user_agent" json:"user_agent,omitempty"`
	IPAddress        net.IP    `db:"ip_address" json:"ip_address,omitempty"`
	ExpiresAt        time.Time `db:"expires_at" json:"expires_at"`
	CreatedAt        time.Time `db:"created_at" json:"created_at"`
}

type UpdateUserInput struct {
	FirstName  *string `json:"first_name,omitempty"`
	SecondName *string `json:"second_name,omitempty"`
	Email      *string `json:"email,omitempty"`
}

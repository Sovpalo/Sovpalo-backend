package model

import "time"

type EventCreateInput struct {
	Title       string     `json:"title"`
	Description *string    `json:"description,omitempty"`
	PhotoURL    *string    `json:"photo_url,omitempty"`
	StartTime   *time.Time `json:"start_time,omitempty"`
	EndTime     *time.Time `json:"end_time,omitempty"`
	CompanyID   *int64     `json:"company_id,omitempty"`
}

type EventUpdateInput struct {
	Title       *string    `json:"title,omitempty"`
	Description *string    `json:"description,omitempty"`
	PhotoURL    *string    `json:"photo_url,omitempty"`
	StartTime   *time.Time `json:"start_time,omitempty"`
	EndTime     *time.Time `json:"end_time,omitempty"`
	CompanyID   *int64     `json:"company_id,omitempty"`
}

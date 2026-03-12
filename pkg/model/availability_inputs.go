package model

import "time"

type AvailabilityCreateInput struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Note      *string   `json:"note,omitempty"`
}

type AvailabilityRangeInput struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

package models

import "time"

type EventStatus string

const (
	EventStatusDraft     EventStatus = "draft"
	EventStatusPublished EventStatus = "published"
	EventStatusCompleted EventStatus = "completed"
	EventStatusCancelled EventStatus = "cancelled"
	EventStatusArchived  EventStatus = "archived"
)

type Event struct {
	ID           int64       `db:"id" json:"id"`
	Title        string      `db:"title" json:"title"`
	Description  *string     `db:"description" json:"description,omitempty"`
	StartTime    time.Time   `db:"start_time" json:"start_time"`
	EndTime      *time.Time  `db:"end_time" json:"end_time,omitempty"`
	Timezone     string      `db:"timezone" json:"timezone"`
	Location     *string     `db:"location" json:"location,omitempty"`
	Status       EventStatus `db:"status" json:"status"`
	CreatedBy    int64       `db:"created_by" json:"created_by"`
	Version      int         `db:"version" json:"version"`
	ICSSequence  int         `db:"ics_sequence" json:"ics_sequence"`
	MaxPlusOnes  int         `db:"max_plus_ones" json:"max_plus_ones"`
	RSVPDeadline *time.Time  `db:"rsvp_deadline" json:"rsvp_deadline,omitempty"`
	CreatedAt    time.Time   `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time   `db:"updated_at" json:"updated_at"`
}

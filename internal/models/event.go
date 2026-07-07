package models

import "time"

type EventStatus string

const (
	EventStatusDraft     EventStatus = "draft"
	EventStatusPublished EventStatus = "published"
	EventStatusCancelled EventStatus = "cancelled"
	EventStatusArchived  EventStatus = "archived"
)

type Event struct {
	ID                     int64       `db:"id" json:"id"`
	PublicID               *string     `db:"public_id" json:"public_id,omitempty"`
	FriendlyName           *string     `db:"friendly_name" json:"friendly_name,omitempty"`
	Title                  string      `db:"title" json:"title"`
	Description            *string     `db:"description" json:"description,omitempty"`
	StartTime              time.Time   `db:"start_time" json:"start_time"`
	EndTime                *time.Time  `db:"end_time" json:"end_time,omitempty"`
	Timezone               string      `db:"timezone" json:"timezone"`
	Location               *string     `db:"location" json:"location,omitempty"`
	Status                 EventStatus `db:"status" json:"status"`
	CreatedBy              int64       `db:"created_by" json:"created_by"`
	Version                int         `db:"version" json:"version"`
	ICSSequence            int         `db:"ics_sequence" json:"ics_sequence"`
	MaxPlusOnes            int         `db:"max_plus_ones" json:"max_plus_ones"`
	RSVPDeadline           *time.Time  `db:"rsvp_deadline" json:"rsvp_deadline,omitempty"`
	AllowRSVPAfterDeadline bool        `db:"allow_rsvp_after_deadline" json:"allow_rsvp_after_deadline"`
	AllowMaybeRSVP         bool        `db:"allow_maybe_rsvp" json:"allow_maybe_rsvp"`
	PrivateGuestList       bool        `db:"private_guest_list" json:"private_guest_list"`
	FamilyHeadcount        bool        `db:"family_headcount" json:"family_headcount"`
	EventCapacity          *int        `db:"event_capacity" json:"event_capacity,omitempty"`
	TemplateID             *int64      `db:"template_id" json:"template_id,omitempty"`
	CustomThemeImageURL    *string     `db:"custom_theme_image_url" json:"custom_theme_image_url,omitempty"`
	CustomThemeColor       *string     `db:"custom_theme_color" json:"custom_theme_color,omitempty"`
	ComponentOverrides     *string     `db:"component_overrides" json:"component_overrides,omitempty"`
	CreatedAt              time.Time   `db:"created_at" json:"created_at"`
	UpdatedAt              time.Time   `db:"updated_at" json:"updated_at"`
}

type EventWithStats struct {
	Event
	InviteCount int `db:"invite_count" json:"invite_count"`
	RSVPCount   int `db:"rsvp_count" json:"rsvp_count"`
	AcceptCount int `db:"accept_count" json:"accept_count"`
}

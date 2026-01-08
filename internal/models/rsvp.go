package models

import (
	"fmt"
	"time"
)

type RSVPResponse string

const (
	RSVPResponseYes   RSVPResponse = "yes"
	RSVPResponseNo    RSVPResponse = "no"
	RSVPResponseMaybe RSVPResponse = "maybe"
)

func (r RSVPResponse) Valid() bool {
	switch r {
	case RSVPResponseYes, RSVPResponseNo, RSVPResponseMaybe:
		return true
	default:
		return false
	}
}

type RSVP struct {
	ID        int64        `db:"id" json:"id"`
	InviteID  int64        `db:"invite_id" json:"invite_id"`
	Response  RSVPResponse `db:"response" json:"response"`
	PlusOnes  int          `db:"plus_ones" json:"plus_ones"`
	CreatedAt time.Time    `db:"created_at" json:"created_at"`
	UpdatedAt time.Time    `db:"updated_at" json:"updated_at"`
}

func (r *RSVP) Validate() error {
	if r.InviteID == 0 {
		return fmt.Errorf("invite_id is required")
	}

	if !r.Response.Valid() {
		return fmt.Errorf("response must be yes, no, or maybe")
	}

	if r.PlusOnes < 0 {
		return fmt.Errorf("plus_ones cannot be negative")
	}

	return nil
}

type RSVPAnswer struct {
	ID            int64     `db:"id" json:"id"`
	RSVPID        int64     `db:"rsvp_id" json:"rsvp_id"`
	QuestionID    int64     `db:"question_id" json:"question_id"`
	AnswerText    *string   `db:"answer_text" json:"answer_text,omitempty"`
	AnswerOption  *string   `db:"answer_option" json:"answer_option,omitempty"`
	AnswerBoolean *bool     `db:"answer_boolean" json:"answer_boolean,omitempty"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time `db:"updated_at" json:"updated_at"`
}

func (a *RSVPAnswer) Validate() error {
	if a.RSVPID == 0 {
		return fmt.Errorf("rsvp_id is required")
	}

	if a.QuestionID == 0 {
		return fmt.Errorf("question_id is required")
	}

	populated := 0
	if a.AnswerText != nil {
		populated++
		if *a.AnswerText == "" {
			return fmt.Errorf("answer_text cannot be empty")
		}
		if len(*a.AnswerText) > 500 {
			return fmt.Errorf("answer_text cannot exceed 500 characters")
		}
	}
	if a.AnswerOption != nil {
		populated++
		if *a.AnswerOption == "" {
			return fmt.Errorf("answer_option cannot be empty")
		}
	}
	if a.AnswerBoolean != nil {
		populated++
	}

	if populated != 1 {
		return fmt.Errorf("exactly one answer field must be populated")
	}

	return nil
}

package models

import (
	"encoding/json"
	"fmt"
	"time"
)

type EmailStatus string

const (
	EmailStatusPending   EmailStatus = "pending"
	EmailStatusSending   EmailStatus = "sending"
	EmailStatusSent      EmailStatus = "sent"
	EmailStatusFailed    EmailStatus = "failed"
	EmailStatusCancelled EmailStatus = "cancelled"
)

func (s EmailStatus) Valid() bool {
	switch s {
	case EmailStatusPending, EmailStatusSending, EmailStatusSent, EmailStatusFailed, EmailStatusCancelled:
		return true
	default:
		return false
	}
}

type EmailQueue struct {
	ID            int64       `json:"id"`
	ToEmail       string      `json:"to_email"`
	ToName        *string     `json:"to_name,omitempty"`
	Subject       string      `json:"subject"`
	BodyText      string      `json:"body_text"`
	BodyHTML      *string     `json:"body_html,omitempty"`
	Attachments   []byte      `json:"attachments,omitempty"`
	Status        EmailStatus `json:"status"`
	Attempts      int         `json:"attempts"`
	MaxAttempts   int         `json:"max_attempts"`
	LastAttemptAt *time.Time  `json:"last_attempt_at,omitempty"`
	LastError     *string     `json:"last_error,omitempty"`
	ScheduledFor  time.Time   `json:"scheduled_for"`
	CreatedAt     time.Time   `json:"created_at"`
}

type EmailAttachment struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	Content     []byte `json:"content"`
}

func (e *EmailQueue) Validate() error {
	if e.ToEmail == "" {
		return &ValidationError{
			Field:   "to_email",
			Message: "to_email is required",
		}
	}

	if e.Subject == "" {
		return &ValidationError{
			Field:   "subject",
			Message: "subject is required",
		}
	}

	if e.BodyText == "" {
		return &ValidationError{
			Field:   "body_text",
			Message: "body_text is required",
		}
	}

	if !e.Status.Valid() {
		return &ValidationError{
			Field:   "status",
			Message: "status must be pending, sending, sent, failed, or cancelled",
		}
	}

	if e.MaxAttempts < 0 {
		return &ValidationError{
			Field:   "max_attempts",
			Message: "max_attempts must be non-negative",
		}
	}

	return nil
}

func (e *EmailQueue) SetAttachments(attachments []EmailAttachment) error {
	if len(attachments) == 0 {
		e.Attachments = nil
		return nil
	}

	data, err := json.Marshal(attachments)
	if err != nil {
		return fmt.Errorf("failed to marshal attachments: %w", err)
	}

	e.Attachments = data
	return nil
}

func (e *EmailQueue) GetAttachments() ([]EmailAttachment, error) {
	if len(e.Attachments) == 0 {
		return []EmailAttachment{}, nil
	}

	var attachments []EmailAttachment
	if err := json.Unmarshal(e.Attachments, &attachments); err != nil {
		return nil, fmt.Errorf("failed to unmarshal attachments: %w", err)
	}

	return attachments, nil
}

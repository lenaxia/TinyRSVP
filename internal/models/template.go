package models

import (
	"time"
)

type TemplateType string

const (
	TemplateTypeInviteEmail      TemplateType = "invite_email"
	TemplateTypeRSVPPage         TemplateType = "rsvp_page"
	TemplateTypeConfirmationPage TemplateType = "confirmation_page"
)

type Template struct {
	ID          int64        `json:"id"`
	EventID     *int64       `json:"event_id,omitempty"`
	Name        string       `json:"name"`
	Type        TemplateType `json:"type"`
	Description string       `json:"description"`

	HTMLContent string  `json:"html_content"`
	TextContent *string `json:"text_content,omitempty"`
	CSSContent  *string `json:"css_content,omitempty"`

	IsDefault bool      `json:"is_default"`
	IsActive  bool      `json:"is_active"`
	Version   int       `json:"version"`
	CreatedBy int64     `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (t *Template) Validate() error {
	if t.Name == "" {
		return &ValidationError{Field: "name", Message: "Template name is required"}
	}

	if len(t.Name) < 3 || len(t.Name) > 100 {
		return &ValidationError{Field: "name", Message: "Template name must be 3-100 characters"}
	}

	if !t.Type.IsValid() {
		return &ValidationError{Field: "type", Message: "Invalid template type"}
	}

	if t.HTMLContent == "" {
		return &ValidationError{Field: "html_content", Message: "HTML content is required"}
	}

	if t.Type == TemplateTypeInviteEmail && t.TextContent == nil {
		return &ValidationError{Field: "text_content", Message: "Text content required for email templates"}
	}

	if t.CreatedBy == 0 {
		return &ValidationError{Field: "created_by", Message: "Created by is required"}
	}

	return nil
}

func (tt TemplateType) IsValid() bool {
	switch tt {
	case TemplateTypeInviteEmail, TemplateTypeRSVPPage, TemplateTypeConfirmationPage:
		return true
	default:
		return false
	}
}

func (tt TemplateType) String() string {
	return string(tt)
}

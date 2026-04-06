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

type TemplateCategory string

const (
	CategoryPlain               TemplateCategory = "plain"
	CategoryCard                TemplateCategory = "card"
	CategoryModern              TemplateCategory = "modern"
	CategoryClassic             TemplateCategory = "classic"
	CategoryFun                 TemplateCategory = "fun"
	CategoryWeddingElegance     TemplateCategory = "wedding-elegance"
	CategoryBirthdayCelebration TemplateCategory = "birthday-celebration"
	CategoryCorporatePro        TemplateCategory = "corporate-professional"
	CategoryHolidayFestive      TemplateCategory = "holiday-festive"
	CategoryGardenParty         TemplateCategory = "garden-party"
	CategoryModernMinimalist    TemplateCategory = "modern-minimalist"
	CategoryPlainText           TemplateCategory = "plain-text"
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

	Category     TemplateCategory `json:"category"`
	ThumbnailURL *string          `json:"thumbnail_url,omitempty"`
	ImageURL     *string          `json:"image_url,omitempty"`
	Tags         []string         `json:"tags"`
	SortOrder    int              `json:"sort_order"`

	ComponentConfig *string `json:"component_config,omitempty"`
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

	if t.Category == "" {
		t.Category = CategoryPlain
	}

	if !t.Category.IsValid() {
		return &ValidationError{Field: "category", Message: "Invalid template category"}
	}

	if len(t.Description) > 500 {
		return &ValidationError{Field: "description", Message: "Description cannot exceed 500 characters"}
	}

	if t.SortOrder < 0 {
		return &ValidationError{Field: "sort_order", Message: "Sort order must be >= 0"}
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

func (tc TemplateCategory) IsValid() bool {
	switch tc {
	case CategoryPlain, CategoryCard, CategoryModern, CategoryClassic, CategoryFun,
		CategoryWeddingElegance, CategoryBirthdayCelebration, CategoryCorporatePro,
		CategoryHolidayFestive, CategoryGardenParty, CategoryModernMinimalist, CategoryPlainText:
		return true
	default:
		return false
	}
}

func (tc TemplateCategory) String() string {
	return string(tc)
}

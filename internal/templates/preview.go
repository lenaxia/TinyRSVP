package templates

import (
	"time"

	"github.com/lenaxia/tinyrsvp/internal/models"
)

type PreviewRequest struct {
	Type        models.TemplateType `json:"type"`
	HTMLContent string              `json:"html_content"`
	TextContent *string             `json:"text_content,omitempty"`
	CSSContent  *string             `json:"css_content,omitempty"`
}

type PreviewResponse struct {
	HTMLPreview string `json:"html_preview"`
	TextPreview string `json:"text_preview,omitempty"`
}

func CreateTestData(templateType models.TemplateType) interface{} {
	now := time.Now()
	startTime := now.Add(30 * 24 * time.Hour)
	endTime := startTime.Add(3 * time.Hour)
	deadline := startTime.Add(-7 * 24 * time.Hour)

	switch templateType {
	case models.TemplateTypeInviteEmail:
		return &InviteEmailData{
			Event: struct {
				Title        string
				Description  string
				StartTime    time.Time
				EndTime      *time.Time
				Timezone     string
				Location     string
				RSVPDeadline *time.Time
			}{
				Title:        "Sample Event",
				Description:  "This is a sample event for template preview",
				StartTime:    startTime,
				EndTime:      &endTime,
				Timezone:     "America/Los_Angeles",
				Location:     "123 Main Street, City, State 12345",
				RSVPDeadline: &deadline,
			},
			Invite: struct {
				Name  string
				Email string
			}{
				Name:  "John Doe",
				Email: "john@example.com",
			},
			RSVPURL:     "https://rsvp.example.com/rsvp/sample-token-preview",
			MaxPlusOnes: 2,
		}

	case models.TemplateTypeRSVPPage:
		return &RSVPPageData{
			Event: struct {
				Title        string
				Description  string
				StartTime    time.Time
				EndTime      *time.Time
				Timezone     string
				Location     string
				RSVPDeadline *time.Time
			}{
				Title:        "Sample Event",
				Description:  "This is a sample event for template preview",
				StartTime:    startTime,
				EndTime:      &endTime,
				Timezone:     "America/Los_Angeles",
				Location:     "123 Main Street, City, State 12345",
				RSVPDeadline: &deadline,
			},
			Token:       "sample-token-preview",
			MaxPlusOnes: 2,
			Questions: []QuestionData{
				{
					ID:           1,
					QuestionText: "Dietary restrictions?",
					QuestionType: "text",
					Required:     false,
					HelpText:     "Please let us know if you have any dietary restrictions",
				},
				{
					ID:           2,
					QuestionText: "Preferred meal",
					QuestionType: "select",
					Options: []OptionData{
						{Value: "chicken", Label: "Chicken"},
						{Value: "fish", Label: "Fish"},
						{Value: "vegetarian", Label: "Vegetarian"},
					},
					Required: true,
					HelpText: "Choose your preferred meal option",
				},
			},
		}

	case models.TemplateTypeConfirmationPage:
		return &ConfirmationPageData{
			Event: struct {
				Title       string
				Description string
				StartTime   time.Time
				EndTime     *time.Time
				Timezone    string
				Location    string
			}{
				Title:       "Sample Event",
				Description: "This is a sample event for template preview",
				StartTime:   startTime,
				EndTime:     &endTime,
				Timezone:    "America/Los_Angeles",
				Location:    "123 Main Street, City, State 12345",
			},
			Token: "sample-token-preview",
			RSVP: struct {
				Response string
				PlusOnes int
				Notes    string
			}{
				Response: "yes",
				PlusOnes: 2,
				Notes:    "Looking forward to the event!",
			},
			Answers: []AnswerData{
				{
					QuestionText:  "Dietary restrictions?",
					AnswerDisplay: "Vegetarian",
				},
				{
					QuestionText:  "Preferred meal",
					AnswerDisplay: "Vegetarian",
				},
			},
		}

	default:
		return nil
	}
}

package events

import (
	"context"
	"strings"
	"testing"

	"github.com/lenaxia/tinyrsvp/internal/models"
)

func TestQuestionValidator_ValidateCreate(t *testing.T) {
	tests := []struct {
		name     string
		question *models.PreferenceQuestion
		wantErr  bool
		errMsg   string
	}{
		{
			name: "valid text question",
			question: &models.PreferenceQuestion{
				EventID:      1,
				QuestionText: "What is your dietary preference?",
				QuestionType: models.QuestionTypeText,
				Required:     true,
			},
			wantErr: false,
		},
		{
			name: "valid single choice",
			question: &models.PreferenceQuestion{
				EventID:      1,
				QuestionText: "Will you attend?",
				QuestionType: models.QuestionTypeSingleChoice,
				Required:     true,
				Options:      stringPtr(`["Yes", "No", "Maybe"]`),
			},
			wantErr: false,
		},
		{
			name: "valid multiple choice",
			question: &models.PreferenceQuestion{
				EventID:      1,
				QuestionText: "Select dietary restrictions",
				QuestionType: models.QuestionTypeMultipleChoice,
				Required:     false,
				Options:      stringPtr(`["Vegetarian", "Vegan", "Gluten-free", "None"]`),
			},
			wantErr: false,
		},
		{
			name: "missing event ID",
			question: &models.PreferenceQuestion{
				QuestionText: "What is your dietary preference?",
				QuestionType: models.QuestionTypeText,
			},
			wantErr: true,
			errMsg:  "event_id is required",
		},
		{
			name: "question text too short",
			question: &models.PreferenceQuestion{
				EventID:      1,
				QuestionText: "Why?",
				QuestionType: models.QuestionTypeText,
			},
			wantErr: true,
			errMsg:  "question text must be between 5 and 500 characters",
		},
		{
			name: "question text too long",
			question: &models.PreferenceQuestion{
				EventID:      1,
				QuestionText: strings.Repeat("a", 501),
				QuestionType: models.QuestionTypeText,
			},
			wantErr: true,
			errMsg:  "question text must be between 5 and 500 characters",
		},
		{
			name: "question text with leading whitespace",
			question: &models.PreferenceQuestion{
				EventID:      1,
				QuestionText: "  What is your name?",
				QuestionType: models.QuestionTypeText,
			},
			wantErr: true,
			errMsg:  "question text cannot have leading or trailing whitespace",
		},
		{
			name: "question text with trailing whitespace",
			question: &models.PreferenceQuestion{
				EventID:      1,
				QuestionText: "What is your name?  ",
				QuestionType: models.QuestionTypeText,
			},
			wantErr: true,
			errMsg:  "question text cannot have leading or trailing whitespace",
		},
		{
			name: "empty question text",
			question: &models.PreferenceQuestion{
				EventID:      1,
				QuestionText: "",
				QuestionType: models.QuestionTypeText,
			},
			wantErr: true,
			errMsg:  "question text is required",
		},
		{
			name: "invalid question type",
			question: &models.PreferenceQuestion{
				EventID:      1,
				QuestionText: "What is your name?",
				QuestionType: models.QuestionType("invalid"),
			},
			wantErr: true,
			errMsg:  "invalid question type",
		},
		{
			name: "empty question type",
			question: &models.PreferenceQuestion{
				EventID:      1,
				QuestionText: "What is your name?",
				QuestionType: models.QuestionType(""),
			},
			wantErr: true,
			errMsg:  "question type is required",
		},
		{
			name: "single choice without options",
			question: &models.PreferenceQuestion{
				EventID:      1,
				QuestionText: "Choose one",
				QuestionType: models.QuestionTypeSingleChoice,
			},
			wantErr: true,
			errMsg:  "options required for choice questions",
		},
		{
			name: "multiple choice without options",
			question: &models.PreferenceQuestion{
				EventID:      1,
				QuestionText: "Choose multiple",
				QuestionType: models.QuestionTypeMultipleChoice,
			},
			wantErr: true,
			errMsg:  "options required for choice questions",
		},
		{
			name: "text question with options",
			question: &models.PreferenceQuestion{
				EventID:      1,
				QuestionText: "What is your name?",
				QuestionType: models.QuestionTypeText,
				Options:      stringPtr(`["Option 1", "Option 2"]`),
			},
			wantErr: true,
			errMsg:  "options not allowed for text questions",
		},
		{
			name: "too few options",
			question: &models.PreferenceQuestion{
				EventID:      1,
				QuestionText: "Choose one",
				QuestionType: models.QuestionTypeSingleChoice,
				Options:      stringPtr(`["Only one"]`),
			},
			wantErr: true,
			errMsg:  "must have 2-10 options",
		},
		{
			name: "too many options",
			question: &models.PreferenceQuestion{
				EventID:      1,
				QuestionText: "Choose one",
				QuestionType: models.QuestionTypeSingleChoice,
				Options:      stringPtr(`["1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11"]`),
			},
			wantErr: true,
			errMsg:  "must have 2-10 options",
		},
		{
			name: "invalid JSON options",
			question: &models.PreferenceQuestion{
				EventID:      1,
				QuestionText: "Choose one",
				QuestionType: models.QuestionTypeSingleChoice,
				Options:      stringPtr(`invalid json`),
			},
			wantErr: true,
			errMsg:  "invalid options JSON",
		},
		{
			name: "option text too short",
			question: &models.PreferenceQuestion{
				EventID:      1,
				QuestionText: "Choose one",
				QuestionType: models.QuestionTypeSingleChoice,
				Options:      stringPtr(`["", "Valid"]`),
			},
			wantErr: true,
			errMsg:  "each option must be 1-200 characters",
		},
		{
			name: "option text too long",
			question: &models.PreferenceQuestion{
				EventID:      1,
				QuestionText: "Choose one",
				QuestionType: models.QuestionTypeSingleChoice,
				Options:      stringPtr(`["Valid", "` + strings.Repeat("a", 201) + `"]`),
			},
			wantErr: true,
			errMsg:  "each option must be 1-200 characters",
		},
		{
			name: "duplicate options",
			question: &models.PreferenceQuestion{
				EventID:      1,
				QuestionText: "Choose one",
				QuestionType: models.QuestionTypeSingleChoice,
				Options:      stringPtr(`["Option 1", "Option 1"]`),
			},
			wantErr: true,
			errMsg:  "duplicate options not allowed",
		},
	}

	validator := NewQuestionValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateCreate(context.Background(), tt.question)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCreate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil && tt.errMsg != "" {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Error message = %q, want to contain %q", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestQuestionValidator_ValidateUpdate(t *testing.T) {
	tests := []struct {
		name     string
		question *models.PreferenceQuestion
		wantErr  bool
		errMsg   string
	}{
		{
			name: "valid update",
			question: &models.PreferenceQuestion{
				ID:           1,
				EventID:      1,
				QuestionText: "Updated question text",
				QuestionType: models.QuestionTypeText,
				Required:     true,
			},
			wantErr: false,
		},
		{
			name: "missing ID",
			question: &models.PreferenceQuestion{
				EventID:      1,
				QuestionText: "What is your name?",
				QuestionType: models.QuestionTypeText,
			},
			wantErr: true,
			errMsg:  "question ID is required",
		},
		{
			name: "invalid question text",
			question: &models.PreferenceQuestion{
				ID:           1,
				EventID:      1,
				QuestionText: "Why?",
				QuestionType: models.QuestionTypeText,
			},
			wantErr: true,
			errMsg:  "question text must be between 5 and 500 characters",
		},
	}

	validator := NewQuestionValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateUpdate(context.Background(), tt.question)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateUpdate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil && tt.errMsg != "" {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Error message = %q, want to contain %q", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

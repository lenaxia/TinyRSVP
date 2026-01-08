package events

import (
	"context"
	"fmt"
	"strings"

	"github.com/lenaxia/tinyrsvp/internal/models"
)

type questionValidator struct{}

func NewQuestionValidator() QuestionValidator {
	return &questionValidator{}
}

func (v *questionValidator) ValidateCreate(ctx context.Context, question *models.PreferenceQuestion) error {
	if question.EventID == 0 {
		return &models.ValidationError{
			Field:   "event_id",
			Message: "event_id is required",
		}
	}

	if err := v.validateQuestionText(question.QuestionText); err != nil {
		return err
	}

	if err := v.validateQuestionType(question.QuestionType); err != nil {
		return err
	}

	if err := v.validateOptions(question); err != nil {
		return err
	}

	return nil
}

func (v *questionValidator) ValidateUpdate(ctx context.Context, question *models.PreferenceQuestion) error {
	if question.ID == 0 {
		return &models.ValidationError{
			Field:   "id",
			Message: "question ID is required",
		}
	}

	if question.EventID == 0 {
		return &models.ValidationError{
			Field:   "event_id",
			Message: "event_id is required",
		}
	}

	if err := v.validateQuestionText(question.QuestionText); err != nil {
		return err
	}

	if err := v.validateQuestionType(question.QuestionType); err != nil {
		return err
	}

	if err := v.validateOptions(question); err != nil {
		return err
	}

	return nil
}

func (v *questionValidator) validateQuestionText(text string) error {
	if text == "" {
		return &models.ValidationError{
			Field:   "question_text",
			Message: "question text is required",
		}
	}

	if text != strings.TrimSpace(text) {
		return &models.ValidationError{
			Field:   "question_text",
			Message: "question text cannot have leading or trailing whitespace",
		}
	}

	if len(text) < 5 || len(text) > 500 {
		return &models.ValidationError{
			Field:   "question_text",
			Message: "question text must be between 5 and 500 characters",
		}
	}

	return nil
}

func (v *questionValidator) validateQuestionType(qtype models.QuestionType) error {
	if qtype == "" {
		return &models.ValidationError{
			Field:   "question_type",
			Message: "question type is required",
		}
	}

	if qtype != models.QuestionTypeText &&
		qtype != models.QuestionTypeSingleChoice &&
		qtype != models.QuestionTypeMultipleChoice {
		return &models.ValidationError{
			Field:   "question_type",
			Message: "invalid question type",
		}
	}

	return nil
}

func (v *questionValidator) validateOptions(question *models.PreferenceQuestion) error {
	isChoiceQuestion := question.QuestionType == models.QuestionTypeSingleChoice ||
		question.QuestionType == models.QuestionTypeMultipleChoice

	if isChoiceQuestion && (question.Options == nil || *question.Options == "") {
		return &models.ValidationError{
			Field:   "options",
			Message: "options required for choice questions",
		}
	}

	if question.QuestionType == models.QuestionTypeText && question.Options != nil && *question.Options != "" {
		return &models.ValidationError{
			Field:   "options",
			Message: "options not allowed for text questions",
		}
	}

	if question.Options != nil && *question.Options != "" {
		options, err := question.ParseOptions()
		if err != nil {
			return &models.ValidationError{
				Field:   "options",
				Message: "invalid options JSON",
			}
		}

		if len(options) < 2 || len(options) > 10 {
			return &models.ValidationError{
				Field:   "options",
				Message: "must have 2-10 options",
			}
		}

		seen := make(map[string]bool)
		for _, opt := range options {
			if len(opt) < 1 || len(opt) > 200 {
				return &models.ValidationError{
					Field:   "options",
					Message: "each option must be 1-200 characters",
				}
			}

			if seen[opt] {
				return &models.ValidationError{
					Field:   "options",
					Message: fmt.Sprintf("duplicate options not allowed: %q", opt),
				}
			}
			seen[opt] = true
		}
	}

	return nil
}

package models

import (
	"encoding/json"
	"time"
)

type QuestionType string

const (
	QuestionTypeText           QuestionType = "text"
	QuestionTypeSingleChoice   QuestionType = "single_choice"
	QuestionTypeMultipleChoice QuestionType = "multiple_choice"
)

type PreferenceQuestion struct {
	ID           int64        `db:"id" json:"id"`
	EventID      int64        `db:"event_id" json:"event_id"`
	QuestionText string       `db:"question_text" json:"question_text"`
	QuestionType QuestionType `db:"question_type" json:"question_type"`
	Required     bool         `db:"required" json:"required"`
	DisplayOrder int          `db:"display_order" json:"display_order"`
	Options      *string      `db:"options" json:"options,omitempty"`
	CreatedAt    time.Time    `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time    `db:"updated_at" json:"updated_at"`
}

func (q *PreferenceQuestion) ParseOptions() ([]string, error) {
	if q.Options == nil || *q.Options == "" {
		return nil, nil
	}

	var options []string
	if err := json.Unmarshal([]byte(*q.Options), &options); err != nil {
		return nil, err
	}

	return options, nil
}

func (q *PreferenceQuestion) SetOptions(options []string) error {
	if options == nil {
		q.Options = nil
		return nil
	}

	data, err := json.Marshal(options)
	if err != nil {
		return err
	}

	optionsStr := string(data)
	q.Options = &optionsStr
	return nil
}

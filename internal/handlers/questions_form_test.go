package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/lenaxia/tinyrsvp/internal/models"
)

func TestParseQuestionsFromForm(t *testing.T) {
	form := url.Values{
		"questions[0][text]":    {"Dietary restrictions?"},
		"questions[0][type]":    {"text"},
		"questions[0][required]": {"on"},
		"questions[1][text]":    {"Meal choice?"},
		"questions[1][type]":    {"single_choice"},
		"questions[1][options]": {"Chicken\nBeef\nVegan"},
	}

	questions := parseQuestionsFromForm(form)

	if len(questions) != 2 {
		t.Fatalf("expected 2 questions, got %d", len(questions))
	}

	if questions[0].QuestionText != "Dietary restrictions?" {
		t.Errorf("question[0] text = %q", questions[0].QuestionText)
	}
	if questions[0].QuestionType != models.QuestionTypeText {
		t.Errorf("question[0] type = %q", questions[0].QuestionType)
	}
	if !questions[0].Required {
		t.Error("question[0] should be required")
	}

	if questions[1].QuestionType != models.QuestionTypeSingleChoice {
		t.Errorf("question[1] type = %q", questions[1].QuestionType)
	}
	if questions[1].Options == nil {
		t.Fatal("question[1] options should be set")
	}
	opts, err := questions[1].ParseOptions()
	if err != nil {
		t.Fatalf("ParseOptions failed: %v", err)
	}
	if len(opts) != 3 || opts[0] != "Chicken" || opts[2] != "Vegan" {
		t.Errorf("parsed options = %v", opts)
	}
}

func TestParseQuestionsFromForm_IgnoresEmpty(t *testing.T) {
	form := url.Values{
		"questions[0][text]": {""},
		"questions[0][type]": {"text"},
	}
	questions := parseQuestionsFromForm(form)
	if len(questions) != 0 {
		t.Errorf("expected no questions for empty text, got %d", len(questions))
	}
}

func TestCreateEventFromForm_CreatesQuestions(t *testing.T) {
	created := 0
	qsvc := &mockQuestionService{
		addQuestionFunc: func(ctx context.Context, q *models.PreferenceQuestion) error {
			created++
			if q.EventID == 0 {
				t.Error("question EventID should be set from created event")
			}
			return nil
		},
	}

	form := url.Values{
		"title":                   {"Test Event"},
		"description":             {"desc"},
		"start_time":              {"2026-09-01T18:00"},
		"timezone":                {"America/Los_Angeles"},
		"questions[0][text]":      {"Question one?"},
		"questions[0][type]":      {"text"},
	}
	req := httptest.NewRequest(http.MethodPost, "/events", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	h := &EventWebHandlers{
		service: &mockEventService{
			CreateEventFunc: func(ctx context.Context, event *models.Event) error {
				event.ID = 42
				event.Status = models.EventStatusDraft
				return nil
			},
		},
		questionService: qsvc,
	}

	w := httptest.NewRecorder()
	h.CreateEventFromForm(w, req)

	if created != 1 {
		t.Errorf("expected 1 question created, got %d", created)
	}
}

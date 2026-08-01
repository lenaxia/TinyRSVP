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

func TestSyncQuestions_UpdateExisting(t *testing.T) {
	updated := false
	deleted := false
	qsvc := &mockQuestionService{
		getQuestionsFunc: func(ctx context.Context, eventID int64) ([]*models.PreferenceQuestion, error) {
			return []*models.PreferenceQuestion{
				{ID: 10, EventID: eventID, QuestionText: "Old text", QuestionType: models.QuestionTypeText},
			}, nil
		},
		updateQuestionFunc: func(ctx context.Context, q *models.PreferenceQuestion) error {
			if q.ID == 10 && q.QuestionText == "Updated text" {
				updated = true
			}
			return nil
		},
		deleteQuestionFunc: func(ctx context.Context, id int64) error {
			deleted = true
			return nil
		},
	}

	h := &EventWebHandlers{questionService: qsvc}
	submitted := []*models.PreferenceQuestion{
		{ID: 10, QuestionText: "Updated text", QuestionType: models.QuestionTypeText},
	}

	if err := h.syncQuestions(context.Background(), 1, submitted); err != nil {
		t.Fatalf("syncQuestions failed: %v", err)
	}

	if !updated {
		t.Error("expected existing question to be updated")
	}
	if deleted {
		t.Error("should not have deleted (question 10 was in submitted list)")
	}
}

func TestSyncQuestions_DeleteRemoved(t *testing.T) {
	deletedIDs := []int64{}
	qsvc := &mockQuestionService{
		getQuestionsFunc: func(ctx context.Context, eventID int64) ([]*models.PreferenceQuestion, error) {
			return []*models.PreferenceQuestion{
				{ID: 10, EventID: eventID, QuestionText: "Q1"},
				{ID: 20, EventID: eventID, QuestionText: "Q2"},
			}, nil
		},
		addQuestionFunc: func(ctx context.Context, q *models.PreferenceQuestion) error {
			return nil
		},
		deleteQuestionFunc: func(ctx context.Context, id int64) error {
			deletedIDs = append(deletedIDs, id)
			return nil
		},
	}

	h := &EventWebHandlers{questionService: qsvc}
	submitted := []*models.PreferenceQuestion{
		{ID: 10, QuestionText: "Q1"}, // Q2 (id=20) removed
	}

	if err := h.syncQuestions(context.Background(), 1, submitted); err != nil {
		t.Fatalf("syncQuestions failed: %v", err)
	}

	if len(deletedIDs) != 1 || deletedIDs[0] != 20 {
		t.Errorf("expected id=20 to be deleted, got %v", deletedIDs)
	}
}

func TestSyncQuestions_ReorderByPosition(t *testing.T) {
	var orders []int
	qsvc := &mockQuestionService{
		getQuestionsFunc: func(ctx context.Context, eventID int64) ([]*models.PreferenceQuestion, error) {
			return []*models.PreferenceQuestion{
				{ID: 10, QuestionText: "A", DisplayOrder: 0},
				{ID: 20, QuestionText: "B", DisplayOrder: 1},
			}, nil
		},
		updateQuestionFunc: func(ctx context.Context, q *models.PreferenceQuestion) error {
			orders = append(orders, q.DisplayOrder)
			return nil
		},
	}

	h := &EventWebHandlers{questionService: qsvc}
	submitted := []*models.PreferenceQuestion{
		{ID: 20, QuestionText: "B"}, // now first
		{ID: 10, QuestionText: "A"}, // now second
	}

	if err := h.syncQuestions(context.Background(), 1, submitted); err != nil {
		t.Fatalf("syncQuestions failed: %v", err)
	}

	if len(orders) != 2 {
		t.Fatalf("expected 2 updates, got %d", len(orders))
	}
	if orders[0] != 0 || orders[1] != 1 {
		t.Errorf("expected display orders [0,1] for reordered questions, got %v", orders)
	}
}

func TestUpdateEventFromForm_SkipsQuestionsForPublishedEvent(t *testing.T) {
	addCalled := false
	qsvc := &mockQuestionService{
		addQuestionFunc: func(ctx context.Context, q *models.PreferenceQuestion) error {
			addCalled = true
			return nil
		},
	}

	form := url.Values{
		"title":              {"Updated"},
		"version":            {"1"},
		"questions[0][text]": {"Should be ignored"},
		"questions[0][type]": {"text"},
	}
	req := httptest.NewRequest(http.MethodPost, "/events/1", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	mockSvc := &mockEventService{
		GetEventFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			return &models.Event{
				ID:     1,
				Title:  "Test",
				Status: models.EventStatusPublished,
			}, nil
		},
		UpdateEventFunc: func(ctx context.Context, event *models.Event) error {
			return nil
		},
	}

	h := &EventWebHandlers{
		service:         mockSvc,
		questionService: qsvc,
	}

	w := httptest.NewRecorder()
	h.UpdateEventFromForm(w, req)

	if addCalled {
		t.Error("should not call AddQuestion for a published event")
	}
}

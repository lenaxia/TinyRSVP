package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/db"
	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/events"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

func setupQuestionIntegrationTest(t *testing.T) (db.Database, *QuestionHandlers, *EventHandlers, *models.User) {
	t.Helper()

	database, err := db.NewDatabase(db.Config{
		Type:         "sqlite",
		Path:         ":memory:",
		MaxOpenConns: 1,
		MaxIdleConns: 1,
		MaxLifetime:  time.Hour,
	})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	migrator, err := db.NewMigrator(database.DB(), "../../migrations/sqlite")
	if err != nil {
		t.Fatalf("Failed to create migrator: %v", err)
	}

	ctx := context.Background()
	if err := migrator.Up(ctx); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	userRepo := repositories.NewUserRepository(database)
	user := &models.User{
		Email: "test@example.com",
		Name:  "Test User",
		Role:  models.RoleEventManager,
	}
	if err := userRepo.Create(ctx, user); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	eventRepo := repositories.NewEventRepository(database)
	questionRepo := repositories.NewQuestionRepository(database)

	eventValidator := events.NewValidator(events.NewTimezoneValidator())
	questionValidator := events.NewQuestionValidator()
	authz := auth.NewAuthorizationChecker()

	eventService := events.NewService(eventRepo, nil, eventValidator, authz)
	questionService := events.NewQuestionService(eventRepo, questionRepo, questionValidator, authz)

	eventHandlers := NewEventHandlers(eventService)
	questionHandlers := NewQuestionHandlers(questionService)

	return database, questionHandlers, eventHandlers, user
}

func TestQuestionIntegration_FullLifecycle(t *testing.T) {
	database, questionHandlers, eventHandlers, user := setupQuestionIntegrationTest(t)
	defer database.Close()

	ctx := auth.WithUser(context.Background(), user)

	eventReq := CreateEventRequest{
		Title:       "Test Event",
		StartTime:   time.Now().Add(24 * time.Hour),
		Timezone:    "America/Los_Angeles",
		MaxPlusOnes: 2,
	}
	eventBody, _ := json.Marshal(eventReq)

	req := httptest.NewRequest(http.MethodPost, "/api/events", bytes.NewReader(eventBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	eventHandlers.CreateEvent(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("Failed to create event: %d - %s", rec.Code, rec.Body.String())
	}

	var event EventResponse
	json.NewDecoder(rec.Body).Decode(&event)

	t.Run("create text question", func(t *testing.T) {
		questionReq := map[string]interface{}{
			"question_text": "What is your dietary preference?",
			"question_type": "text",
			"required":      true,
		}
		body, _ := json.Marshal(questionReq)

		req := httptest.NewRequest(http.MethodPost, "/api/events/1/questions", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req = req.WithContext(ctx)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rec := httptest.NewRecorder()
		questionHandlers.CreateQuestion(rec, req)

		if rec.Code != http.StatusCreated {
			t.Errorf("Expected status 201, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("create single choice question", func(t *testing.T) {
		questionReq := map[string]interface{}{
			"question_text": "Will you attend?",
			"question_type": "single_choice",
			"required":      true,
			"options":       []string{"Yes", "No", "Maybe"},
		}
		body, _ := json.Marshal(questionReq)

		req := httptest.NewRequest(http.MethodPost, "/api/events/1/questions", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req = req.WithContext(ctx)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rec := httptest.NewRecorder()
		questionHandlers.CreateQuestion(rec, req)

		if rec.Code != http.StatusCreated {
			t.Errorf("Expected status 201, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("list questions", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/events/1/questions", nil)
		req = req.WithContext(ctx)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rec := httptest.NewRecorder()
		questionHandlers.GetQuestions(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Expected status 200, got %d: %s", rec.Code, rec.Body.String())
		}

		var questions []*models.PreferenceQuestion
		json.NewDecoder(rec.Body).Decode(&questions)

		if len(questions) != 2 {
			t.Errorf("Expected 2 questions, got %d", len(questions))
		}
	})

	t.Run("update question", func(t *testing.T) {
		updateReq := map[string]interface{}{
			"question_text": "Updated dietary question",
			"question_type": "text",
			"required":      false,
		}
		body, _ := json.Marshal(updateReq)

		req := httptest.NewRequest(http.MethodPut, "/api/events/1/questions/1", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req = req.WithContext(ctx)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "1")
		rctx.URLParams.Add("qid", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rec := httptest.NewRecorder()
		questionHandlers.UpdateQuestion(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("reorder questions", func(t *testing.T) {
		reorderReq := map[string]interface{}{
			"question_ids": []int64{2, 1},
		}
		body, _ := json.Marshal(reorderReq)

		req := httptest.NewRequest(http.MethodPost, "/api/events/1/questions/reorder", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req = req.WithContext(ctx)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rec := httptest.NewRecorder()
		questionHandlers.ReorderQuestions(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d: %s", rec.Code, rec.Body.String())
		}

		req = httptest.NewRequest(http.MethodGet, "/api/events/1/questions", nil)
		req = req.WithContext(ctx)
		rctx = chi.NewRouteContext()
		rctx.URLParams.Add("id", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rec = httptest.NewRecorder()
		questionHandlers.GetQuestions(rec, req)

		var questions []*models.PreferenceQuestion
		json.NewDecoder(rec.Body).Decode(&questions)

		if len(questions) != 2 {
			t.Fatalf("Expected 2 questions, got %d", len(questions))
		}

		if questions[0].ID != 2 {
			t.Errorf("First question should be ID 2, got %d", questions[0].ID)
		}
		if questions[1].ID != 1 {
			t.Errorf("Second question should be ID 1, got %d", questions[1].ID)
		}
	})

	t.Run("delete question", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/events/1/questions/1", nil)
		req = req.WithContext(ctx)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "1")
		rctx.URLParams.Add("qid", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rec := httptest.NewRecorder()
		questionHandlers.DeleteQuestion(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Errorf("Expected status 204, got %d: %s", rec.Code, rec.Body.String())
		}

		req = httptest.NewRequest(http.MethodGet, "/api/events/1/questions", nil)
		req = req.WithContext(ctx)
		rctx = chi.NewRouteContext()
		rctx.URLParams.Add("id", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rec = httptest.NewRecorder()
		questionHandlers.GetQuestions(rec, req)

		var questions []*models.PreferenceQuestion
		json.NewDecoder(rec.Body).Decode(&questions)

		if len(questions) != 1 {
			t.Errorf("Expected 1 question after delete, got %d", len(questions))
		}
	})
}

func TestQuestionIntegration_PublishedEventRestrictions(t *testing.T) {
	database, questionHandlers, eventHandlers, user := setupQuestionIntegrationTest(t)
	defer database.Close()

	ctx := auth.WithUser(context.Background(), user)

	eventReq := CreateEventRequest{
		Title:       "Test Event",
		StartTime:   time.Now().Add(24 * time.Hour),
		Timezone:    "America/Los_Angeles",
		MaxPlusOnes: 2,
	}
	eventBody, _ := json.Marshal(eventReq)

	req := httptest.NewRequest(http.MethodPost, "/api/events", bytes.NewReader(eventBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	eventHandlers.CreateEvent(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("Failed to create event: %d - %s", rec.Code, rec.Body.String())
	}

	questionReq := map[string]interface{}{
		"question_text": "Test question",
		"question_type": "text",
		"required":      true,
	}
	body, _ := json.Marshal(questionReq)

	req = httptest.NewRequest(http.MethodPost, "/api/events/1/questions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req = req.WithContext(ctx)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rec = httptest.NewRecorder()
	questionHandlers.CreateQuestion(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("Failed to create question: %d - %s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodPost, "/api/events/1/publish", nil)
	req = req.WithContext(ctx)
	rctx = chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rec = httptest.NewRecorder()
	eventHandlers.PublishEvent(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Failed to publish event: %d - %s", rec.Code, rec.Body.String())
	}

	t.Run("cannot add question to published event", func(t *testing.T) {
		questionReq := map[string]interface{}{
			"question_text": "Another question",
			"question_type": "text",
		}
		body, _ := json.Marshal(questionReq)

		req := httptest.NewRequest(http.MethodPost, "/api/events/1/questions", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req = req.WithContext(ctx)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rec := httptest.NewRecorder()
		questionHandlers.CreateQuestion(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("cannot update question on published event", func(t *testing.T) {
		updateReq := map[string]interface{}{
			"question_text": "Updated text",
			"question_type": "text",
		}
		body, _ := json.Marshal(updateReq)

		req := httptest.NewRequest(http.MethodPut, "/api/events/1/questions/1", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req = req.WithContext(ctx)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "1")
		rctx.URLParams.Add("qid", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rec := httptest.NewRecorder()
		questionHandlers.UpdateQuestion(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("cannot delete question from published event", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/events/1/questions/1", nil)
		req = req.WithContext(ctx)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "1")
		rctx.URLParams.Add("qid", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rec := httptest.NewRecorder()
		questionHandlers.DeleteQuestion(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("cannot reorder questions on published event", func(t *testing.T) {
		reorderReq := map[string]interface{}{
			"question_ids": []int64{1},
		}
		body, _ := json.Marshal(reorderReq)

		req := httptest.NewRequest(http.MethodPost, "/api/events/1/questions/reorder", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req = req.WithContext(ctx)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rec := httptest.NewRecorder()
		questionHandlers.ReorderQuestions(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d: %s", rec.Code, rec.Body.String())
		}
	})
}

func TestQuestionIntegration_ValidationErrors(t *testing.T) {
	database, questionHandlers, eventHandlers, user := setupQuestionIntegrationTest(t)
	defer database.Close()

	ctx := auth.WithUser(context.Background(), user)

	eventReq := CreateEventRequest{
		Title:       "Test Event",
		StartTime:   time.Now().Add(24 * time.Hour),
		Timezone:    "America/Los_Angeles",
		MaxPlusOnes: 2,
	}
	eventBody, _ := json.Marshal(eventReq)

	req := httptest.NewRequest(http.MethodPost, "/api/events", bytes.NewReader(eventBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	eventHandlers.CreateEvent(rec, req)

	t.Run("question text too short", func(t *testing.T) {
		questionReq := map[string]interface{}{
			"question_text": "Why?",
			"question_type": "text",
		}
		body, _ := json.Marshal(questionReq)

		req := httptest.NewRequest(http.MethodPost, "/api/events/1/questions", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req = req.WithContext(ctx)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rec := httptest.NewRecorder()
		questionHandlers.CreateQuestion(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("choice question without options", func(t *testing.T) {
		questionReq := map[string]interface{}{
			"question_text": "Choose one",
			"question_type": "single_choice",
		}
		body, _ := json.Marshal(questionReq)

		req := httptest.NewRequest(http.MethodPost, "/api/events/1/questions", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req = req.WithContext(ctx)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rec := httptest.NewRecorder()
		questionHandlers.CreateQuestion(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("too few options", func(t *testing.T) {
		questionReq := map[string]interface{}{
			"question_text": "Choose one",
			"question_type": "single_choice",
			"options":       []string{"Only one"},
		}
		body, _ := json.Marshal(questionReq)

		req := httptest.NewRequest(http.MethodPost, "/api/events/1/questions", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req = req.WithContext(ctx)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rec := httptest.NewRecorder()
		questionHandlers.CreateQuestion(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d: %s", rec.Code, rec.Body.String())
		}
	})
}

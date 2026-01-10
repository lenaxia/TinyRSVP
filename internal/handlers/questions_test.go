package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

type mockQuestionService struct {
	addQuestionFunc      func(ctx context.Context, question *models.PreferenceQuestion) error
	updateQuestionFunc   func(ctx context.Context, question *models.PreferenceQuestion) error
	deleteQuestionFunc   func(ctx context.Context, id int64) error
	getQuestionsFunc     func(ctx context.Context, eventID int64) ([]*models.PreferenceQuestion, error)
	reorderQuestionsFunc func(ctx context.Context, eventID int64, questionIDs []int64) error
}

func (m *mockQuestionService) AddQuestion(ctx context.Context, question *models.PreferenceQuestion) error {
	if m.addQuestionFunc != nil {
		return m.addQuestionFunc(ctx, question)
	}
	question.ID = 1
	return nil
}

func (m *mockQuestionService) UpdateQuestion(ctx context.Context, question *models.PreferenceQuestion) error {
	if m.updateQuestionFunc != nil {
		return m.updateQuestionFunc(ctx, question)
	}
	return nil
}

func (m *mockQuestionService) DeleteQuestion(ctx context.Context, id int64) error {
	if m.deleteQuestionFunc != nil {
		return m.deleteQuestionFunc(ctx, id)
	}
	return nil
}

func (m *mockQuestionService) GetQuestions(ctx context.Context, eventID int64) ([]*models.PreferenceQuestion, error) {
	if m.getQuestionsFunc != nil {
		return m.getQuestionsFunc(ctx, eventID)
	}
	return []*models.PreferenceQuestion{}, nil
}

func (m *mockQuestionService) ReorderQuestions(ctx context.Context, eventID int64, questionIDs []int64) error {
	if m.reorderQuestionsFunc != nil {
		return m.reorderQuestionsFunc(ctx, eventID, questionIDs)
	}
	return nil
}

func TestQuestionHandlers_CreateQuestion(t *testing.T) {
	tests := []struct {
		name           string
		eventID        string
		body           interface{}
		setupMock      func(*mockQuestionService)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:    "create text question",
			eventID: "1",
			body: map[string]interface{}{
				"question_text": "What is your dietary preference?",
				"question_type": "text",
				"required":      true,
			},
			setupMock:      func(m *mockQuestionService) {},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var resp models.PreferenceQuestion
				if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}
				if resp.ID == 0 {
					t.Error("Expected ID to be set")
				}
			},
		},
		{
			name:    "create single choice question",
			eventID: "1",
			body: map[string]interface{}{
				"question_text": "Will you attend?",
				"question_type": "single_choice",
				"required":      true,
				"options":       []string{"Yes", "No", "Maybe"},
			},
			setupMock:      func(m *mockQuestionService) {},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "invalid event ID",
			eventID:        "invalid",
			body:           map[string]interface{}{},
			setupMock:      func(m *mockQuestionService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:    "validation error",
			eventID: "1",
			body: map[string]interface{}{
				"question_text": "Bad",
				"question_type": "text",
			},
			setupMock: func(m *mockQuestionService) {
				m.addQuestionFunc = func(ctx context.Context, question *models.PreferenceQuestion) error {
					return &models.ValidationError{
						Field:   "question_text",
						Message: "question text must be between 5 and 500 characters",
					}
				}
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:    "permission denied",
			eventID: "1",
			body: map[string]interface{}{
				"question_text": "What is your name?",
				"question_type": "text",
			},
			setupMock: func(m *mockQuestionService) {
				m.addQuestionFunc = func(ctx context.Context, question *models.PreferenceQuestion) error {
					return &models.PermissionDeniedError{
						Action:   "add question",
						Resource: "PreferenceQuestion",
					}
				}
			},
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockQuestionService{}
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			handlers := NewQuestionHandlers(mockService)

			var body []byte
			var err error
			if tt.body != nil {
				body, err = json.Marshal(tt.body)
				if err != nil {
					t.Fatalf("Failed to marshal request body: %v", err)
				}
			}

			req := httptest.NewRequest(http.MethodPost, "/api/events/"+tt.eventID+"/questions", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Accept", "application/json")

			user := &models.User{ID: 1, Role: models.RoleEventManager}
			ctx := auth.WithUser(req.Context(), user)
			req = req.WithContext(ctx)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.eventID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rec := httptest.NewRecorder()

			handlers.CreateQuestion(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Body: %s", tt.expectedStatus, rec.Code, rec.Body.String())
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, rec)
			}
		})
	}
}

func TestQuestionHandlers_GetQuestions(t *testing.T) {
	tests := []struct {
		name           string
		eventID        string
		setupMock      func(*mockQuestionService)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:    "get questions for event",
			eventID: "1",
			setupMock: func(m *mockQuestionService) {
				m.getQuestionsFunc = func(ctx context.Context, eventID int64) ([]*models.PreferenceQuestion, error) {
					return []*models.PreferenceQuestion{
						{ID: 1, EventID: eventID, QuestionText: "Question 1", QuestionType: models.QuestionTypeText},
						{ID: 2, EventID: eventID, QuestionText: "Question 2", QuestionType: models.QuestionTypeSingleChoice},
					}, nil
				}
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var resp []*models.PreferenceQuestion
				if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}
				if len(resp) != 2 {
					t.Errorf("Expected 2 questions, got %d", len(resp))
				}
			},
		},
		{
			name:           "invalid event ID",
			eventID:        "invalid",
			setupMock:      func(m *mockQuestionService) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockQuestionService{}
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			handlers := NewQuestionHandlers(mockService)

			req := httptest.NewRequest(http.MethodGet, "/api/events/"+tt.eventID+"/questions", nil)

			user := &models.User{ID: 1, Role: models.RoleEventManager}
			ctx := auth.WithUser(req.Context(), user)
			req = req.WithContext(ctx)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.eventID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rec := httptest.NewRecorder()

			handlers.GetQuestions(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Body: %s", tt.expectedStatus, rec.Code, rec.Body.String())
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, rec)
			}
		})
	}
}

func TestQuestionHandlers_UpdateQuestion(t *testing.T) {
	tests := []struct {
		name           string
		eventID        string
		questionID     string
		body           interface{}
		setupMock      func(*mockQuestionService)
		expectedStatus int
	}{
		{
			name:       "update question",
			eventID:    "1",
			questionID: "1",
			body: map[string]interface{}{
				"question_text": "Updated question",
				"question_type": "text",
				"required":      true,
			},
			setupMock:      func(m *mockQuestionService) {},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid event ID",
			eventID:        "invalid",
			questionID:     "1",
			body:           map[string]interface{}{},
			setupMock:      func(m *mockQuestionService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid question ID",
			eventID:        "1",
			questionID:     "invalid",
			body:           map[string]interface{}{},
			setupMock:      func(m *mockQuestionService) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockQuestionService{}
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			handlers := NewQuestionHandlers(mockService)

			var body []byte
			var err error
			if tt.body != nil {
				body, err = json.Marshal(tt.body)
				if err != nil {
					t.Fatalf("Failed to marshal request body: %v", err)
				}
			}

			req := httptest.NewRequest(http.MethodPut, "/api/events/"+tt.eventID+"/questions/"+tt.questionID, bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Accept", "application/json")

			user := &models.User{ID: 1, Role: models.RoleEventManager}
			ctx := auth.WithUser(req.Context(), user)
			req = req.WithContext(ctx)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.eventID)
			rctx.URLParams.Add("qid", tt.questionID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rec := httptest.NewRecorder()

			handlers.UpdateQuestion(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Body: %s", tt.expectedStatus, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestQuestionHandlers_DeleteQuestion(t *testing.T) {
	tests := []struct {
		name           string
		eventID        string
		questionID     string
		setupMock      func(*mockQuestionService)
		expectedStatus int
	}{
		{
			name:           "delete question",
			eventID:        "1",
			questionID:     "1",
			setupMock:      func(m *mockQuestionService) {},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "invalid event ID",
			eventID:        "invalid",
			questionID:     "1",
			setupMock:      func(m *mockQuestionService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid question ID",
			eventID:        "1",
			questionID:     "invalid",
			setupMock:      func(m *mockQuestionService) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockQuestionService{}
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			handlers := NewQuestionHandlers(mockService)

			req := httptest.NewRequest(http.MethodDelete, "/api/events/"+tt.eventID+"/questions/"+tt.questionID, nil)

			user := &models.User{ID: 1, Role: models.RoleEventManager}
			ctx := auth.WithUser(req.Context(), user)
			req = req.WithContext(ctx)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.eventID)
			rctx.URLParams.Add("qid", tt.questionID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rec := httptest.NewRecorder()

			handlers.DeleteQuestion(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Body: %s", tt.expectedStatus, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestQuestionHandlers_ReorderQuestions(t *testing.T) {
	tests := []struct {
		name           string
		eventID        string
		body           interface{}
		setupMock      func(*mockQuestionService)
		expectedStatus int
	}{
		{
			name:    "reorder questions",
			eventID: "1",
			body: map[string]interface{}{
				"question_ids": []int64{3, 1, 2},
			},
			setupMock:      func(m *mockQuestionService) {},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid event ID",
			eventID:        "invalid",
			body:           map[string]interface{}{},
			setupMock:      func(m *mockQuestionService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:    "empty question IDs",
			eventID: "1",
			body: map[string]interface{}{
				"question_ids": []int64{},
			},
			setupMock: func(m *mockQuestionService) {
				m.reorderQuestionsFunc = func(ctx context.Context, eventID int64, questionIDs []int64) error {
					return &models.ValidationError{
						Field:   "question_ids",
						Message: "at least one question ID required",
					}
				}
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockQuestionService{}
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			handlers := NewQuestionHandlers(mockService)

			var body []byte
			var err error
			if tt.body != nil {
				body, err = json.Marshal(tt.body)
				if err != nil {
					t.Fatalf("Failed to marshal request body: %v", err)
				}
			}

			req := httptest.NewRequest(http.MethodPost, "/api/events/"+tt.eventID+"/questions/reorder", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Accept", "application/json")

			user := &models.User{ID: 1, Role: models.RoleEventManager}
			ctx := auth.WithUser(req.Context(), user)
			req = req.WithContext(ctx)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.eventID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rec := httptest.NewRecorder()

			handlers.ReorderQuestions(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Body: %s", tt.expectedStatus, rec.Code, rec.Body.String())
			}
		})
	}
}

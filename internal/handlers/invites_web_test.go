package handlers

import (
	"context"
	"html/template"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/invites"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

func TestInviteWebHandlers_ListInvitesPage_Success(t *testing.T) {
	mockService := &FullMockInviteService{}
	mockEventRepo := &mockEventRepository{}

	funcMap := template.FuncMap{
		"sub": func(a, b int) int { return a - b },
		"add": func(a, b int) int { return a + b },
		"div": func(a, b int) int {
			if b == 0 {
				return 0
			}
			return a / b
		},
		"until": func(count int) []int {
			result := make([]int, count)
			for i := 0; i < count; i++ {
				result[i] = i
			}
			return result
		},
	}

	tmpl, err := template.New("invite_list.html").Funcs(funcMap).ParseFiles("../../templates/web/invite_list.html")
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	handler := NewInviteWebHandlers(mockService, mockEventRepo)
	handler.SetTemplates(tmpl)

	user := &models.User{
		ID:    1,
		Email: "test@example.com",
		Role:  models.RoleAdmin,
	}

	event := &models.Event{
		ID:        1,
		Title:     "Test Event",
		CreatedBy: user.ID,
	}

	johnName := "John Doe"
	johnEmail := "john@example.com"
	janeName := "Jane Smith"
	janeEmail := "jane@example.com"
	sentAt := time.Now()

	inviteList := []*models.Invite{
		{
			ID:          1,
			EventID:     1,
			Name:        &johnName,
			Email:       &johnEmail,
			Status:      "draft",
			MaxPlusOnes: 2,
			CreatedAt:   time.Now(),
		},
		{
			ID:          2,
			EventID:     1,
			Name:        &janeName,
			Email:       &janeEmail,
			Status:      "sent",
			MaxPlusOnes: 1,
			SentAt:      &sentAt,
			CreatedAt:   time.Now(),
		},
	}

	stats := &repositories.InviteStats{
		Total:     2,
		Draft:     1,
		Sent:      1,
		Viewed:    0,
		Responded: 0,
		Revoked:   0,
	}

	mockEventRepo.getByIDFunc = func(ctx context.Context, id int64) (*models.Event, error) {
		return event, nil
	}

	mockService.ListInvitesFunc = func(ctx context.Context, req *invites.ListInvitesRequest) (*invites.ListInvitesResponse, error) {
		return &invites.ListInvitesResponse{
			Invites: inviteList,
			Total:   2,
			Stats:   stats,
		}, nil
	}

	req := httptest.NewRequest(http.MethodGet, "/events/1/invites", nil)
	req = req.WithContext(auth.WithUser(context.Background(), user))
	
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("eventId", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler.ListInvitesPage(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	body := w.Body.String()
	if body == "" {
		t.Error("Expected non-empty response body")
	}

	expectedStrings := []string{
		"Test Event",
		"John Doe",
		"jane@example.com",
		"Total",
		"Draft",
		"Sent",
	}

	for _, expected := range expectedStrings {
		if !contains(body, expected) {
			t.Errorf("Expected response to contain %q", expected)
		}
	}
}

func TestInviteWebHandlers_ListInvitesPage_InvalidEventID(t *testing.T) {
	mockService := &FullMockInviteService{}
	mockEventRepo := &mockEventRepository{}

	handler := NewInviteWebHandlers(mockService, mockEventRepo)

	user := &models.User{
		ID:    1,
		Email: "test@example.com",
		Role:  models.RoleAdmin,
	}

	req := httptest.NewRequest(http.MethodGet, "/events/invalid/invites", nil)
	req = req.WithContext(auth.WithUser(context.Background(), user))
	
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("eventId", "invalid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler.ListInvitesPage(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestInviteWebHandlers_ListInvitesPage_Unauthorized(t *testing.T) {
	mockService := &FullMockInviteService{}
	mockEventRepo := &mockEventRepository{}

	handler := NewInviteWebHandlers(mockService, mockEventRepo)

	req := httptest.NewRequest(http.MethodGet, "/events/1/invites", nil)
	
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("eventId", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler.ListInvitesPage(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestInviteWebHandlers_ListInvitesPage_EventNotFound(t *testing.T) {
	mockService := &FullMockInviteService{}
	mockEventRepo := &mockEventRepository{}

	handler := NewInviteWebHandlers(mockService, mockEventRepo)

	user := &models.User{
		ID:    1,
		Email: "test@example.com",
		Role:  models.RoleAdmin,
	}

	mockEventRepo.getByIDFunc = func(ctx context.Context, id int64) (*models.Event, error) {
		return nil, &models.NotFoundError{Resource: "event", ID: id}
	}

	req := httptest.NewRequest(http.MethodGet, "/events/999/invites", nil)
	req = req.WithContext(auth.WithUser(context.Background(), user))
	
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("eventId", "999")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler.ListInvitesPage(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestInviteWebHandlers_ListInvitesPage_PermissionDenied(t *testing.T) {
	mockService := &FullMockInviteService{}
	mockEventRepo := &mockEventRepository{}

	handler := NewInviteWebHandlers(mockService, mockEventRepo)

	user := &models.User{
		ID:    1,
		Email: "test@example.com",
		Role:  models.RoleEventManager,
	}

	event := &models.Event{
		ID:        1,
		Title:     "Test Event",
		CreatedBy: 999,
	}

	mockEventRepo.getByIDFunc = func(ctx context.Context, id int64) (*models.Event, error) {
		return event, nil
	}

	req := httptest.NewRequest(http.MethodGet, "/events/1/invites", nil)
	req = req.WithContext(auth.WithUser(context.Background(), user))
	
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("eventId", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler.ListInvitesPage(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status %d, got %d", http.StatusForbidden, w.Code)
	}
}

func TestInviteWebHandlers_ListInvitesPage_WithFilters(t *testing.T) {
	mockService := &FullMockInviteService{}
	mockEventRepo := &mockEventRepository{}

	funcMap := template.FuncMap{
		"sub": func(a, b int) int { return a - b },
		"add": func(a, b int) int { return a + b },
		"div": func(a, b int) int {
			if b == 0 {
				return 0
			}
			return a / b
		},
		"until": func(count int) []int {
			result := make([]int, count)
			for i := 0; i < count; i++ {
				result[i] = i
			}
			return result
		},
	}

	tmpl, err := template.New("invite_list.html").Funcs(funcMap).ParseFiles("../../templates/web/invite_list.html")
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	handler := NewInviteWebHandlers(mockService, mockEventRepo)
	handler.SetTemplates(tmpl)

	user := &models.User{
		ID:    1,
		Email: "test@example.com",
		Role:  models.RoleAdmin,
	}

	event := &models.Event{
		ID:        1,
		Title:     "Test Event",
		CreatedBy: user.ID,
	}

	mockEventRepo.getByIDFunc = func(ctx context.Context, id int64) (*models.Event, error) {
		return event, nil
	}

	mockService.ListInvitesFunc = func(ctx context.Context, req *invites.ListInvitesRequest) (*invites.ListInvitesResponse, error) {
		if req.Status != nil && *req.Status != "sent" {
			t.Errorf("Expected status filter 'sent', got %v", *req.Status)
		}
		if req.Search != nil && *req.Search != "john" {
			t.Errorf("Expected search 'john', got %v", *req.Search)
		}
		return &invites.ListInvitesResponse{
			Invites: []*models.Invite{},
			Total:   0,
			Stats:   &repositories.InviteStats{},
		}, nil
	}

	req := httptest.NewRequest(http.MethodGet, "/events/1/invites?status=sent&search=john", nil)
	req = req.WithContext(auth.WithUser(context.Background(), user))
	
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("eventId", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler.ListInvitesPage(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestInviteWebHandlers_SetTemplates(t *testing.T) {
	mockService := &FullMockInviteService{}
	mockEventRepo := &mockEventRepository{}

	handler := NewInviteWebHandlers(mockService, mockEventRepo)

	tmpl := template.New("test")
	handler.SetTemplates(tmpl)

	if handler.templates != tmpl {
		t.Error("Expected templates to be set")
	}
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) >= len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

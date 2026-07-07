package handlers

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/events"
	"github.com/lenaxia/tinyrsvp/internal/middleware"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

func TestEventWebHandlers_FullWebUIFlow_Integration(t *testing.T) {
	database := setupIntegrationTestDB(t)
	defer database.Close()

	userRepo := repositories.NewUserRepository(database)
	eventRepo := repositories.NewEventRepository(database)
	authChecker := auth.NewAuthorizationChecker()
	validator := events.NewValidator(events.NewTimezoneValidator())
	eventService := events.NewService(eventRepo, nil, validator, authChecker)

	tmpl := template.Must(template.New("test").Parse(`
		{{define "event_list.html"}}
		<!DOCTYPE html>
		<html><body>
			{{if .Error}}
				<div>Error: {{.Error}}</div>
			{{else if eq (len .Events) 0}}
				<div>No Events Found</div>
			{{else}}
				{{range .Events}}
					<div>{{.Title}}</div>
				{{end}}
			{{end}}
		</body></html>
		{{end}}

		{{define "event_form.html"}}
		<!DOCTYPE html>
		<html><body>
			<h1>{{if .Event.ID}}Edit{{else}}Create{{end}} Event</h1>
			<form>
				<input type="hidden" name="csrf_token" value="{{.CSRFToken}}">
				<input type="text" name="title" value="{{.Event.Title}}">
			</form>
		</body></html>
		{{end}}

		{{define "event_detail.html"}}
		<!DOCTYPE html>
		<html><body>
			<h1>{{.Event.Title}}</h1>
			<p>Status: {{.Event.Status}}</p>
		</body></html>
		{{end}}
	`))

	handlers := NewEventWebHandlers(eventService, nil, tmpl, tmpl, tmpl)

	ctx := context.Background()

	manager := &models.User{
		Email: "manager@example.com",
		Name:  "Event Manager",
		Role:  models.RoleEventManager,
	}
	if err := userRepo.Create(ctx, manager); err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	req := httptest.NewRequest("GET", "/events", nil)
	ctx = auth.WithUser(req.Context(), manager)
	ctx = context.WithValue(ctx, middleware.CSRFTokenKey, "test-csrf-token")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handlers.ListEventsPage(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200 for list page, got %d: %s", w.Code, w.Body.String())
	}

	if !strings.Contains(w.Body.String(), "No Events Found") {
		t.Error("Expected empty events message")
	}

	req = httptest.NewRequest("GET", "/events/new", nil)
	ctx = auth.WithUser(req.Context(), manager)
	ctx = context.WithValue(ctx, middleware.CSRFTokenKey, "test-csrf-token")
	req = req.WithContext(ctx)
	w = httptest.NewRecorder()

	handlers.NewEventForm(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200 for new form, got %d: %s", w.Code, w.Body.String())
	}

	if !strings.Contains(w.Body.String(), "Create Event") {
		t.Error("Expected 'Create Event' in form")
	}

	if !strings.Contains(w.Body.String(), "test-csrf-token") {
		t.Error("CSRF token not injected into form")
	}

	futureTime := time.Now().Add(24 * time.Hour).Format("2006-01-02T15:04")

	formData := url.Values{
		"title":         []string{"Integration Test Event"},
		"description":   []string{"Testing full web UI flow"},
		"start_time":    []string{futureTime},
		"timezone":      []string{"America/Los_Angeles"},
		"location":      []string{"Test Location"},
		"max_plus_ones": []string{"2"},
		"csrf_token":    []string{"test-csrf-token"},
	}

	req = httptest.NewRequest("POST", "/events", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	ctx = auth.WithUser(req.Context(), manager)
	ctx = context.WithValue(ctx, middleware.CSRFTokenKey, "test-csrf-token")
	req = req.WithContext(ctx)
	w = httptest.NewRecorder()

	handlers.CreateEventFromForm(w, req)

	if w.Code != http.StatusSeeOther {
		t.Fatalf("Expected status 303 for create redirect, got %d: %s", w.Code, w.Body.String())
	}

	location := w.Header().Get("Location")
	if !strings.HasPrefix(location, "/events/") {
		t.Errorf("Expected redirect to /events/{id}, got %s", location)
	}

	eventID := strings.TrimPrefix(location, "/events/")

	req = httptest.NewRequest("GET", location, nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", eventID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	ctx = auth.WithUser(req.Context(), manager)
	ctx = context.WithValue(ctx, middleware.CSRFTokenKey, "test-csrf-token")
	req = req.WithContext(ctx)
	w = httptest.NewRecorder()

	handlers.GetEventPage(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200 for event detail, got %d: %s", w.Code, w.Body.String())
	}

	if !strings.Contains(w.Body.String(), "Integration Test Event") {
		t.Error("Expected event title in detail page")
	}

	if !strings.Contains(w.Body.String(), "draft") {
		t.Error("Expected draft status in detail page")
	}

	req = httptest.NewRequest("GET", location+"/edit", nil)
	rctx = chi.NewRouteContext()
	rctx.URLParams.Add("id", eventID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	ctx = auth.WithUser(req.Context(), manager)
	ctx = context.WithValue(ctx, middleware.CSRFTokenKey, "test-csrf-token")
	req = req.WithContext(ctx)
	w = httptest.NewRecorder()

	handlers.EditEventForm(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200 for edit form, got %d: %s", w.Code, w.Body.String())
	}

	if !strings.Contains(w.Body.String(), "Edit Event") {
		t.Error("Expected 'Edit Event' in form")
	}

	if !strings.Contains(w.Body.String(), "Integration Test Event") {
		t.Error("Expected event title in edit form")
	}

	updateFormData := url.Values{
		"title":      []string{"Updated Event Title"},
		"version":    []string{"1"},
		"csrf_token": []string{"test-csrf-token"},
	}

	req = httptest.NewRequest("POST", location, strings.NewReader(updateFormData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rctx = chi.NewRouteContext()
	rctx.URLParams.Add("id", eventID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	ctx = auth.WithUser(req.Context(), manager)
	ctx = context.WithValue(ctx, middleware.CSRFTokenKey, "test-csrf-token")
	req = req.WithContext(ctx)
	w = httptest.NewRecorder()

	handlers.UpdateEventFromForm(w, req)

	if w.Code != http.StatusSeeOther {
		t.Fatalf("Expected status 303 for update redirect, got %d: %s", w.Code, w.Body.String())
	}

	publishFormData := url.Values{
		"csrf_token": []string{"test-csrf-token"},
	}

	req = httptest.NewRequest("POST", location+"/publish", strings.NewReader(publishFormData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rctx = chi.NewRouteContext()
	rctx.URLParams.Add("id", eventID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	ctx = auth.WithUser(req.Context(), manager)
	ctx = context.WithValue(ctx, middleware.CSRFTokenKey, "test-csrf-token")
	req = req.WithContext(ctx)
	w = httptest.NewRecorder()

	handlers.PublishEventAction(w, req)

	if w.Code != http.StatusSeeOther {
		t.Fatalf("Expected status 303 for publish redirect, got %d: %s", w.Code, w.Body.String())
	}

	req = httptest.NewRequest("GET", location, nil)
	rctx = chi.NewRouteContext()
	rctx.URLParams.Add("id", eventID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	ctx = auth.WithUser(req.Context(), manager)
	ctx = context.WithValue(ctx, middleware.CSRFTokenKey, "test-csrf-token")
	req = req.WithContext(ctx)
	w = httptest.NewRecorder()

	handlers.GetEventPage(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200 after publish, got %d: %s", w.Code, w.Body.String())
	}

	if !strings.Contains(w.Body.String(), "published") {
		t.Error("Expected published status after publish action")
	}

	cancelFormData := url.Values{
		"reason":     []string{"Testing cancellation in integration test"},
		"csrf_token": []string{"test-csrf-token"},
	}

	req = httptest.NewRequest("POST", location+"/cancel", strings.NewReader(cancelFormData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rctx = chi.NewRouteContext()
	rctx.URLParams.Add("id", eventID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	ctx = auth.WithUser(req.Context(), manager)
	ctx = context.WithValue(ctx, middleware.CSRFTokenKey, "test-csrf-token")
	req = req.WithContext(ctx)
	w = httptest.NewRecorder()

	handlers.CancelEventAction(w, req)

	if w.Code != http.StatusSeeOther {
		t.Fatalf("Expected status 303 for cancel redirect, got %d: %s", w.Code, w.Body.String())
	}

	req = httptest.NewRequest("GET", location, nil)
	rctx = chi.NewRouteContext()
	rctx.URLParams.Add("id", eventID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	ctx = auth.WithUser(req.Context(), manager)
	ctx = context.WithValue(ctx, middleware.CSRFTokenKey, "test-csrf-token")
	req = req.WithContext(ctx)
	w = httptest.NewRecorder()

	handlers.GetEventPage(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200 after cancel, got %d: %s", w.Code, w.Body.String())
	}

	if !strings.Contains(w.Body.String(), "cancelled") {
		t.Error("Expected cancelled status after cancel action")
	}
}

func TestEventWebHandlers_PermissionEnforcement_Integration(t *testing.T) {
	database := setupIntegrationTestDB(t)
	defer database.Close()

	userRepo := repositories.NewUserRepository(database)
	eventRepo := repositories.NewEventRepository(database)
	authChecker := auth.NewAuthorizationChecker()
	validator := events.NewValidator(events.NewTimezoneValidator())
	eventService := events.NewService(eventRepo, nil, validator, authChecker)

	tmpl := template.Must(template.New("test").Parse(`
		{{define "event_form.html"}}
		<html><body><h1>Edit Event</h1></body></html>
		{{end}}
	`))

	handlers := NewEventWebHandlers(eventService, nil, tmpl, tmpl, tmpl)

	ctx := context.Background()

	manager1 := &models.User{
		Email: "manager1@example.com",
		Name:  "Event Manager 1",
		Role:  models.RoleEventManager,
	}
	if err := userRepo.Create(ctx, manager1); err != nil {
		t.Fatalf("Failed to create manager1: %v", err)
	}

	manager2 := &models.User{
		Email: "manager2@example.com",
		Name:  "Event Manager 2",
		Role:  models.RoleEventManager,
	}
	if err := userRepo.Create(ctx, manager2); err != nil {
		t.Fatalf("Failed to create manager2: %v", err)
	}

	admin := &models.User{
		Email: "admin@example.com",
		Name:  "Admin User",
		Role:  models.RoleAdmin,
	}
	if err := userRepo.Create(ctx, admin); err != nil {
		t.Fatalf("Failed to create admin: %v", err)
	}

	event := &models.Event{
		Title:       "Permission Test Event",
		StartTime:   time.Now().Add(24 * time.Hour),
		Timezone:    "America/Los_Angeles",
		MaxPlusOnes: 2,
		CreatedBy:   manager1.ID,
		Status:      models.EventStatusDraft,
		Version:     1,
	}
	if err := eventRepo.Create(ctx, event); err != nil {
		t.Fatalf("Failed to create event: %v", err)
	}

	formData := url.Values{
		"title":      []string{"Unauthorized Update"},
		"version":    []string{"1"},
		"csrf_token": []string{"test-csrf-token"},
	}

	req := httptest.NewRequest("POST", "/events/1", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	ctx = auth.WithUser(req.Context(), manager2)
	ctx = context.WithValue(ctx, middleware.CSRFTokenKey, "test-csrf-token")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handlers.UpdateEventFromForm(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403 for non-owner manager, got %d: %s", w.Code, w.Body.String())
	}

	req = httptest.NewRequest("POST", "/events/1", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rctx = chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	ctx = auth.WithUser(req.Context(), admin)
	ctx = context.WithValue(ctx, middleware.CSRFTokenKey, "test-csrf-token")
	req = req.WithContext(ctx)
	w = httptest.NewRecorder()

	handlers.UpdateEventFromForm(w, req)

	if w.Code != http.StatusSeeOther {
		t.Errorf("Expected status 303 for admin update, got %d: %s", w.Code, w.Body.String())
	}
}

func TestEventWebHandlers_RouterIntegration(t *testing.T) {
	database := setupIntegrationTestDB(t)
	defer database.Close()

	userRepo := repositories.NewUserRepository(database)
	eventRepo := repositories.NewEventRepository(database)
	authChecker := auth.NewAuthorizationChecker()
	validator := events.NewValidator(events.NewTimezoneValidator())
	eventService := events.NewService(eventRepo, nil, validator, authChecker)

	tmpl := template.Must(template.New("test").Parse(`
		{{define "event_list.html"}}<html><body>Events</body></html>{{end}}
		{{define "event_form.html"}}<html><body>Form</body></html>{{end}}
		{{define "event_detail.html"}}<html><body>Detail</body></html>{{end}}
	`))

	eventWebHandlers := NewEventWebHandlers(eventService, nil, tmpl, tmpl, tmpl)

	authMiddleware := &mockAuthMiddleware{}

	router := NewRouter(&RouterHandlers{
		EventWebHandlers: eventWebHandlers,
		AuthMiddleware:   authMiddleware,
	})

	ctx := context.Background()

	manager := &models.User{
		Email: "manager@example.com",
		Name:  "Event Manager",
		Role:  models.RoleEventManager,
	}
	if err := userRepo.Create(ctx, manager); err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	event := &models.Event{
		Title:       "Router Test Event",
		StartTime:   time.Now().Add(24 * time.Hour),
		Timezone:    "America/Los_Angeles",
		MaxPlusOnes: 2,
		CreatedBy:   manager.ID,
		Status:      models.EventStatusDraft,
		Version:     1,
	}
	if err := eventRepo.Create(ctx, event); err != nil {
		t.Fatalf("Failed to create event: %v", err)
	}

	routes := []struct {
		method string
		path   string
	}{
		{"GET", "/events"},
		{"GET", "/events/new"},
		{"POST", "/events"},
		{"GET", "/events/1"},
		{"GET", "/events/1/edit"},
		{"POST", "/events/1"},
		{"POST", "/events/1/publish"},
		{"POST", "/events/1/cancel"},
		{"POST", "/events/1/delete"},
	}

	for _, route := range routes {
		t.Run(fmt.Sprintf("%s %s", route.method, route.path), func(t *testing.T) {
			var req *http.Request
			if route.method == "POST" {
				formData := url.Values{
					"csrf_token": []string{"test-csrf-token"},
					"title":      []string{"Test"},
					"start_time": []string{"2026-06-15T14:00"},
					"timezone":   []string{"America/Los_Angeles"},
					"reason":     []string{"Testing cancellation reason for integration"},
				}
				req = httptest.NewRequest(route.method, route.path, strings.NewReader(formData.Encode()))
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			} else {
				req = httptest.NewRequest(route.method, route.path, nil)
			}

			ctx := auth.WithUser(req.Context(), manager)
			ctx = context.WithValue(ctx, middleware.CSRFTokenKey, "test-csrf-token")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code == http.StatusMethodNotAllowed {
				t.Errorf("Route %s %s not registered", route.method, route.path)
			}

			if w.Code == http.StatusNotFound {
				t.Errorf("Route %s %s not found", route.method, route.path)
			}
		})
	}
}

func TestEventWebHandlers_CSRFProtection_Integration(t *testing.T) {
	database := setupIntegrationTestDB(t)
	defer database.Close()

	userRepo := repositories.NewUserRepository(database)
	eventRepo := repositories.NewEventRepository(database)
	authChecker := auth.NewAuthorizationChecker()
	validator := events.NewValidator(events.NewTimezoneValidator())
	eventService := events.NewService(eventRepo, nil, validator, authChecker)

	tmpl := template.New("test")
	handlers := NewEventWebHandlers(eventService, nil, tmpl, tmpl, tmpl)

	ctx := context.Background()

	manager := &models.User{
		Email: "manager@example.com",
		Name:  "Event Manager",
		Role:  models.RoleEventManager,
	}
	if err := userRepo.Create(ctx, manager); err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	event := &models.Event{
		Title:       "CSRF Test Event",
		StartTime:   time.Now().Add(24 * time.Hour),
		Timezone:    "America/Los_Angeles",
		MaxPlusOnes: 2,
		CreatedBy:   manager.ID,
		Status:      models.EventStatusDraft,
		Version:     1,
	}
	if err := eventRepo.Create(ctx, event); err != nil {
		t.Fatalf("Failed to create event: %v", err)
	}

	tests := []struct {
		name      string
		path      string
		csrfToken string
		wantPass  bool
	}{
		{
			name:      "valid CSRF token",
			path:      "/events/1/publish",
			csrfToken: "test-csrf-token",
			wantPass:  true,
		},
		{
			name:      "missing CSRF token",
			path:      "/events/1/publish",
			csrfToken: "",
			wantPass:  false,
		},
		{
			name:      "invalid CSRF token",
			path:      "/events/1/publish",
			csrfToken: "wrong-token",
			wantPass:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formData := url.Values{}
			if tt.csrfToken != "" {
				formData.Set("csrf_token", tt.csrfToken)
			}

			req := httptest.NewRequest("POST", tt.path, strings.NewReader(formData.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", "1")
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			ctx := auth.WithUser(req.Context(), manager)
			ctx = context.WithValue(ctx, middleware.CSRFTokenKey, "test-csrf-token")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			handlers.PublishEventAction(w, req)

			if tt.wantPass {
				if w.Code != http.StatusSeeOther {
					t.Errorf("Expected success (303), got %d: %s", w.Code, w.Body.String())
				}
			}
		})
	}
}

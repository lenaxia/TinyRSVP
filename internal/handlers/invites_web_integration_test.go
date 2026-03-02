package handlers

import (
	"context"
	"encoding/hex"
	"fmt"
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/invites"
	"github.com/lenaxia/tinyrsvp/internal/middleware"
	"github.com/lenaxia/tinyrsvp/internal/models"
	"github.com/lenaxia/tinyrsvp/pkg/token"
)

func TestInviteWebHandlers_FullWebUIFlow_Integration(t *testing.T) {
	database := setupIntegrationTestDB(t)
	defer database.Close()

	userRepo := repositories.NewUserRepository(database)
	eventRepo := repositories.NewEventRepository(database)
	inviteRepo := repositories.NewInviteRepository(database)

	tokenSecret := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	tokenSecretBytes, _ := hex.DecodeString(tokenSecret)
	tokenGenerator := token.NewGenerator(tokenSecretBytes)
	inviteService := invites.NewInviteService(tokenGenerator, inviteRepo)

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

	tmpl, err := template.New("invite_list.html").Funcs(funcMap).ParseFiles(
		"../../templates/web/invite_list.html",
		"../../templates/web/partials/base.html",
		"../../templates/web/partials/navigation.html",
	)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	handlers := NewInviteWebHandlers(inviteService, eventRepo)
	handlers.SetTemplates(tmpl)

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
		Title:       "Integration Test Event",
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

	johnName := "John Doe"
	johnEmail := "john@example.com"
	janeName := "Jane Smith"
	janeEmail := "jane@example.com"

	_, _, err = inviteService.CreateInvite(ctx, event.ID, &johnName, &johnEmail, 2, time.Now().Add(30*24*time.Hour))
	if err != nil {
		t.Fatalf("Failed to create invite 1: %v", err)
	}

	invite2, _, err := inviteService.CreateInvite(ctx, event.ID, &janeName, &janeEmail, 1, time.Now().Add(30*24*time.Hour))
	if err != nil {
		t.Fatalf("Failed to create invite 2: %v", err)
	}

	if err := inviteService.MarkInviteSent(ctx, invite2.ID); err != nil {
		t.Fatalf("Failed to mark invite 2 as sent: %v", err)
	}

	req := httptest.NewRequest("GET", fmt.Sprintf("/events/%d/invites", event.ID), nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("eventId", fmt.Sprintf("%d", event.ID))
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	ctx = auth.WithUser(req.Context(), manager)
	ctx = context.WithValue(ctx, middleware.CSRFTokenKey, "test-csrf-token")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handlers.ListInvitesPage(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200 for list page, got %d: %s", w.Code, w.Body.String())
	}

	body := w.Body.String()

	expectedStrings := []string{
		"Integration Test Event",
		"John Doe",
		"john@example.com",
		"Jane Smith",
		"jane@example.com",
		"Total",
		"Draft",
		"Sent",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(body, expected) {
			t.Errorf("Expected response to contain %q", expected)
		}
	}

	if !strings.Contains(body, `data-invite-id="1"`) {
		t.Error("Expected invite 1 data attribute")
	}

	if !strings.Contains(body, `data-invite-id="2"`) {
		t.Error("Expected invite 2 data attribute")
	}

	req = httptest.NewRequest("GET", fmt.Sprintf("/events/%d/invites?status=draft", event.ID), nil)
	rctx = chi.NewRouteContext()
	rctx.URLParams.Add("eventId", fmt.Sprintf("%d", event.ID))
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	ctx = auth.WithUser(req.Context(), manager)
	ctx = context.WithValue(ctx, middleware.CSRFTokenKey, "test-csrf-token")
	req = req.WithContext(ctx)
	w = httptest.NewRecorder()

	handlers.ListInvitesPage(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200 for filtered list, got %d: %s", w.Code, w.Body.String())
	}

	body = w.Body.String()

	if !strings.Contains(body, "John Doe") {
		t.Error("Expected draft invite (John Doe) in filtered results")
	}

	req = httptest.NewRequest("GET", fmt.Sprintf("/events/%d/invites?search=jane", event.ID), nil)
	rctx = chi.NewRouteContext()
	rctx.URLParams.Add("eventId", fmt.Sprintf("%d", event.ID))
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	ctx = auth.WithUser(req.Context(), manager)
	ctx = context.WithValue(ctx, middleware.CSRFTokenKey, "test-csrf-token")
	req = req.WithContext(ctx)
	w = httptest.NewRecorder()

	handlers.ListInvitesPage(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200 for search results, got %d: %s", w.Code, w.Body.String())
	}

	body = w.Body.String()

	if !strings.Contains(body, "Jane Smith") {
		t.Error("Expected search to find Jane Smith")
	}
}

func TestInviteWebHandlers_PermissionEnforcement_Integration(t *testing.T) {
	database := setupIntegrationTestDB(t)
	defer database.Close()

	userRepo := repositories.NewUserRepository(database)
	eventRepo := repositories.NewEventRepository(database)
	inviteRepo := repositories.NewInviteRepository(database)

	tokenSecret := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	tokenSecretBytes, _ := hex.DecodeString(tokenSecret)
	tokenGenerator := token.NewGenerator(tokenSecretBytes)
	inviteService := invites.NewInviteService(tokenGenerator, inviteRepo)

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

	tmpl, err := template.New("invite_list.html").Funcs(funcMap).ParseFiles(
		"../../templates/web/invite_list.html",
		"../../templates/web/partials/base.html",
		"../../templates/web/partials/navigation.html",
	)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	handlers := NewInviteWebHandlers(inviteService, eventRepo)
	handlers.SetTemplates(tmpl)

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

	req := httptest.NewRequest("GET", fmt.Sprintf("/events/%d/invites", event.ID), nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("eventId", fmt.Sprintf("%d", event.ID))
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	ctx = auth.WithUser(req.Context(), manager2)
	ctx = context.WithValue(ctx, middleware.CSRFTokenKey, "test-csrf-token")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handlers.ListInvitesPage(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403 for non-owner manager, got %d: %s", w.Code, w.Body.String())
	}

	req = httptest.NewRequest("GET", fmt.Sprintf("/events/%d/invites", event.ID), nil)
	rctx = chi.NewRouteContext()
	rctx.URLParams.Add("eventId", fmt.Sprintf("%d", event.ID))
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	ctx = auth.WithUser(req.Context(), admin)
	ctx = context.WithValue(ctx, middleware.CSRFTokenKey, "test-csrf-token")
	req = req.WithContext(ctx)
	w = httptest.NewRecorder()

	handlers.ListInvitesPage(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for admin access, got %d: %s", w.Code, w.Body.String())
	}
}

func TestInviteWebHandlers_RouterIntegration(t *testing.T) {
	database := setupIntegrationTestDB(t)
	defer database.Close()

	userRepo := repositories.NewUserRepository(database)
	eventRepo := repositories.NewEventRepository(database)
	inviteRepo := repositories.NewInviteRepository(database)

	tokenSecret := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	tokenSecretBytes, _ := hex.DecodeString(tokenSecret)
	tokenGenerator := token.NewGenerator(tokenSecretBytes)
	inviteService := invites.NewInviteService(tokenGenerator, inviteRepo)

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

	tmpl, err := template.New("invite_list.html").Funcs(funcMap).ParseFiles(
		"../../templates/web/invite_list.html",
		"../../templates/web/partials/base.html",
		"../../templates/web/partials/navigation.html",
	)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	inviteWebHandlers := NewInviteWebHandlers(inviteService, eventRepo)
	inviteWebHandlers.SetTemplates(tmpl)

	authMiddleware := &mockAuthMiddleware{}

	router := NewRouter(&RouterHandlers{
		InviteWebHandlers: inviteWebHandlers,
		AuthMiddleware:    authMiddleware,
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
		{"GET", fmt.Sprintf("/events/%d/invites", event.ID)},
	}

	for _, route := range routes {
		t.Run(fmt.Sprintf("%s %s", route.method, route.path), func(t *testing.T) {
			req := httptest.NewRequest(route.method, route.path, nil)

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

			if w.Code != http.StatusOK {
				t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
			}
		})
	}
}

func TestInviteWebHandlers_EmptyState_Integration(t *testing.T) {
	database := setupIntegrationTestDB(t)
	defer database.Close()

	userRepo := repositories.NewUserRepository(database)
	eventRepo := repositories.NewEventRepository(database)
	inviteRepo := repositories.NewInviteRepository(database)

	tokenSecret := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	tokenSecretBytes, _ := hex.DecodeString(tokenSecret)
	tokenGenerator := token.NewGenerator(tokenSecretBytes)
	inviteService := invites.NewInviteService(tokenGenerator, inviteRepo)

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

	tmpl, err := template.New("invite_list.html").Funcs(funcMap).ParseFiles(
		"../../templates/web/invite_list.html",
		"../../templates/web/partials/base.html",
		"../../templates/web/partials/navigation.html",
	)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	handlers := NewInviteWebHandlers(inviteService, eventRepo)
	handlers.SetTemplates(tmpl)

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
		Title:       "Empty Invite List Event",
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

	req := httptest.NewRequest("GET", fmt.Sprintf("/events/%d/invites", event.ID), nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("eventId", fmt.Sprintf("%d", event.ID))
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	ctx = auth.WithUser(req.Context(), manager)
	ctx = context.WithValue(ctx, middleware.CSRFTokenKey, "test-csrf-token")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handlers.ListInvitesPage(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200 for empty list, got %d: %s", w.Code, w.Body.String())
	}

	body := w.Body.String()

	if !strings.Contains(body, "No Invites Found") {
		t.Error("Expected empty state message")
	}

	if !strings.Contains(body, "Add Guest") {
		t.Error("Expected 'Add Guest' button in empty state")
	}
}

func TestInviteWebHandlers_StatsDisplay_Integration(t *testing.T) {
	database := setupIntegrationTestDB(t)
	defer database.Close()

	userRepo := repositories.NewUserRepository(database)
	eventRepo := repositories.NewEventRepository(database)
	inviteRepo := repositories.NewInviteRepository(database)

	tokenSecret := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	tokenSecretBytes, _ := hex.DecodeString(tokenSecret)
	tokenGenerator := token.NewGenerator(tokenSecretBytes)
	inviteService := invites.NewInviteService(tokenGenerator, inviteRepo)

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

	tmpl, err := template.New("invite_list.html").Funcs(funcMap).ParseFiles(
		"../../templates/web/invite_list.html",
		"../../templates/web/partials/base.html",
		"../../templates/web/partials/navigation.html",
	)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	handlers := NewInviteWebHandlers(inviteService, eventRepo)
	handlers.SetTemplates(tmpl)

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
		Title:       "Stats Test Event",
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

	name1 := "Guest 1"
	email1 := "guest1@example.com"
	name2 := "Guest 2"
	email2 := "guest2@example.com"
	name3 := "Guest 3"
	email3 := "guest3@example.com"

	_, _, err = inviteService.CreateInvite(ctx, event.ID, &name1, &email1, 2, time.Now().Add(30*24*time.Hour))
	if err != nil {
		t.Fatalf("Failed to create invite 1: %v", err)
	}

	invite2, _, err := inviteService.CreateInvite(ctx, event.ID, &name2, &email2, 1, time.Now().Add(30*24*time.Hour))
	if err != nil {
		t.Fatalf("Failed to create invite 2: %v", err)
	}

	invite3, _, err := inviteService.CreateInvite(ctx, event.ID, &name3, &email3, 0, time.Now().Add(30*24*time.Hour))
	if err != nil {
		t.Fatalf("Failed to create invite 3: %v", err)
	}

	if err := inviteService.MarkInviteSent(ctx, invite2.ID); err != nil {
		t.Fatalf("Failed to mark invite 2 as sent: %v", err)
	}

	if err := inviteService.MarkInviteSent(ctx, invite3.ID); err != nil {
		t.Fatalf("Failed to mark invite 3 as sent: %v", err)
	}

	if err := inviteService.MarkInviteViewed(ctx, invite3.ID); err != nil {
		t.Fatalf("Failed to mark invite 3 as viewed: %v", err)
	}

	req := httptest.NewRequest("GET", fmt.Sprintf("/events/%d/invites", event.ID), nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("eventId", fmt.Sprintf("%d", event.ID))
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	ctx = auth.WithUser(req.Context(), manager)
	ctx = context.WithValue(ctx, middleware.CSRFTokenKey, "test-csrf-token")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handlers.ListInvitesPage(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	body := w.Body.String()

	statsChecks := []struct {
		label string
		value string
	}{
		{"Total", "3"},
		{"Draft", "1"},
		{"Sent", "2"},
		{"Viewed", "1"},
	}

	for _, check := range statsChecks {
		if !strings.Contains(body, check.label) {
			t.Errorf("Expected stats to contain label %q", check.label)
		}
	}
}

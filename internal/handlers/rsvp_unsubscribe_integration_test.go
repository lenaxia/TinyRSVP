package handlers

import (
	"context"
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/invites"
	"github.com/lenaxia/tinyrsvp/internal/models"
	"github.com/lenaxia/tinyrsvp/pkg/token"
)

func TestUnsubscribeHandler_Integration_Success(t *testing.T) {
	database := setupIntegrationTestDB(t)
	defer database.Close()

	user := createTestUser(t, database)
	event := createTestEventForRSVP(t, database, user.ID)
	invite, inviteToken := createTestInviteForRSVP(t, database, event.ID)

	secret := []byte("test-secret-key-32-bytes-long!!")
	generator := token.NewGenerator(secret)
	inviteRepo := repositories.NewInviteRepository(database)
	inviteService := invites.NewInviteService(generator, inviteRepo)
	eventRepo := repositories.NewEventRepository(database)
	rsvpRepo := repositories.NewRSVPRepository(database)
	questionRepo := repositories.NewQuestionRepository(database)

	handler := NewRSVPHandler(inviteService, eventRepo, rsvpRepo, questionRepo)

	tmpl, err := template.ParseFiles("../../templates/web/unsubscribe.html")
	if err != nil {
		t.Logf("Warning: Failed to load template: %v. Using fallback HTML.", err)
	} else {
		handler.SetTemplates(tmpl)
	}

	r := chi.NewRouter()
	r.Get("/unsubscribe/{token}", handler.Unsubscribe)

	req := httptest.NewRequest("GET", "/unsubscribe/"+inviteToken, nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	body := w.Body.String()
	if !strings.Contains(body, event.Title) {
		t.Errorf("Response should contain event title. Body: %s", body)
	}

	updatedInvite, err := inviteRepo.GetByID(context.Background(), invite.ID)
	if err != nil {
		t.Fatalf("Failed to get updated invite: %v", err)
	}

	if !updatedInvite.Unsubscribed {
		t.Error("Expected invite to be marked as unsubscribed")
	}
}

func TestUnsubscribeHandler_Integration_AlreadyUnsubscribed(t *testing.T) {
	database := setupIntegrationTestDB(t)
	defer database.Close()

	user := createTestUser(t, database)
	event := createTestEventForRSVP(t, database, user.ID)
	invite, inviteToken := createTestInviteForRSVP(t, database, event.ID)

	inviteRepo := repositories.NewInviteRepository(database)
	invite.Unsubscribed = true
	if err := inviteRepo.Update(context.Background(), invite); err != nil {
		t.Fatalf("Failed to mark invite as unsubscribed: %v", err)
	}

	secret := []byte("test-secret-key-32-bytes-long!!")
	generator := token.NewGenerator(secret)
	inviteService := invites.NewInviteService(generator, inviteRepo)
	eventRepo := repositories.NewEventRepository(database)
	rsvpRepo := repositories.NewRSVPRepository(database)
	questionRepo := repositories.NewQuestionRepository(database)

	handler := NewRSVPHandler(inviteService, eventRepo, rsvpRepo, questionRepo)

	r := chi.NewRouter()
	r.Get("/unsubscribe/{token}", handler.Unsubscribe)

	req := httptest.NewRequest("GET", "/unsubscribe/"+inviteToken, nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	body := w.Body.String()
	if !strings.Contains(strings.ToLower(body), "unsubscribed") || !strings.Contains(body, event.Title) {
		t.Errorf("Response should indicate successful unsubscribe. Body: %s", body)
	}
}

func TestUnsubscribeHandler_Integration_InvalidToken(t *testing.T) {
	database := setupIntegrationTestDB(t)
	defer database.Close()

	secret := []byte("test-secret-key-32-bytes-long!!")
	generator := token.NewGenerator(secret)
	inviteRepo := repositories.NewInviteRepository(database)
	inviteService := invites.NewInviteService(generator, inviteRepo)
	eventRepo := repositories.NewEventRepository(database)
	rsvpRepo := repositories.NewRSVPRepository(database)
	questionRepo := repositories.NewQuestionRepository(database)

	handler := NewRSVPHandler(inviteService, eventRepo, rsvpRepo, questionRepo)

	r := chi.NewRouter()
	r.Get("/unsubscribe/{token}", handler.Unsubscribe)

	req := httptest.NewRequest("GET", "/unsubscribe/invalidtoken123456789012345678901234567890", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "not found") && !strings.Contains(body, "revoked") {
		t.Error("Response should indicate invite not found")
	}
}

func TestUnsubscribeHandler_Integration_ExpiredToken(t *testing.T) {
	database := setupIntegrationTestDB(t)
	defer database.Close()

	user := createTestUser(t, database)
	event := createTestEventForRSVP(t, database, user.ID)
	invite, plainToken := createTestInviteForRSVP(t, database, event.ID)

	inviteRepo := repositories.NewInviteRepository(database)
	invite.ExpiresAt = time.Now().Add(-1 * time.Hour)

	ctx := context.Background()
	query := `UPDATE invites SET expires_at = ? WHERE id = ?`
	if _, err := database.Exec(ctx, query, invite.ExpiresAt, invite.ID); err != nil {
		t.Fatalf("Failed to update invite expiration: %v", err)
	}

	secret := []byte("test-secret-key-32-bytes-long!!")
	generator := token.NewGenerator(secret)
	inviteService := invites.NewInviteService(generator, inviteRepo)
	eventRepo := repositories.NewEventRepository(database)
	rsvpRepo := repositories.NewRSVPRepository(database)
	questionRepo := repositories.NewQuestionRepository(database)

	handler := NewRSVPHandler(inviteService, eventRepo, rsvpRepo, questionRepo)

	r := chi.NewRouter()
	r.Get("/unsubscribe/{token}", handler.Unsubscribe)

	req := httptest.NewRequest("GET", "/unsubscribe/"+plainToken, nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "expired") {
		t.Error("Response should indicate invite has expired")
	}
}

func TestUnsubscribeHandler_Integration_RevokedInvite(t *testing.T) {
	database := setupIntegrationTestDB(t)
	defer database.Close()

	user := createTestUser(t, database)
	event := createTestEventForRSVP(t, database, user.ID)
	invite, inviteToken := createTestInviteForRSVP(t, database, event.ID)

	inviteRepo := repositories.NewInviteRepository(database)
	invite.Status = models.InviteStatusRevoked
	reason := "Test revocation"
	invite.RevocationReason = &reason
	if err := inviteRepo.Update(context.Background(), invite); err != nil {
		t.Fatalf("Failed to revoke invite: %v", err)
	}

	secret := []byte("test-secret-key-32-bytes-long!!")
	generator := token.NewGenerator(secret)
	inviteService := invites.NewInviteService(generator, inviteRepo)
	eventRepo := repositories.NewEventRepository(database)
	rsvpRepo := repositories.NewRSVPRepository(database)
	questionRepo := repositories.NewQuestionRepository(database)

	handler := NewRSVPHandler(inviteService, eventRepo, rsvpRepo, questionRepo)

	r := chi.NewRouter()
	r.Get("/unsubscribe/{token}", handler.Unsubscribe)

	req := httptest.NewRequest("GET", "/unsubscribe/"+inviteToken, nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "revoked") {
		t.Error("Response should indicate invite has been revoked")
	}
}

func TestUnsubscribeHandler_Integration_Idempotent(t *testing.T) {
	database := setupIntegrationTestDB(t)
	defer database.Close()

	user := createTestUser(t, database)
	event := createTestEventForRSVP(t, database, user.ID)
	invite, inviteToken := createTestInviteForRSVP(t, database, event.ID)

	secret := []byte("test-secret-key-32-bytes-long!!")
	generator := token.NewGenerator(secret)
	inviteRepo := repositories.NewInviteRepository(database)
	inviteService := invites.NewInviteService(generator, inviteRepo)
	eventRepo := repositories.NewEventRepository(database)
	rsvpRepo := repositories.NewRSVPRepository(database)
	questionRepo := repositories.NewQuestionRepository(database)

	handler := NewRSVPHandler(inviteService, eventRepo, rsvpRepo, questionRepo)

	r := chi.NewRouter()
	r.Get("/unsubscribe/{token}", handler.Unsubscribe)

	req1 := httptest.NewRequest("GET", "/unsubscribe/"+inviteToken, nil)
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)

	if w1.Code != http.StatusOK {
		t.Errorf("First unsubscribe: Expected status 200, got %d", w1.Code)
	}

	updatedInvite, err := inviteRepo.GetByID(context.Background(), invite.ID)
	if err != nil {
		t.Fatalf("Failed to get updated invite: %v", err)
	}

	if !updatedInvite.Unsubscribed {
		t.Error("Expected invite to be marked as unsubscribed after first call")
	}

	req2 := httptest.NewRequest("GET", "/unsubscribe/"+inviteToken, nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("Second unsubscribe: Expected status 200, got %d", w2.Code)
	}

	finalInvite, err := inviteRepo.GetByID(context.Background(), invite.ID)
	if err != nil {
		t.Fatalf("Failed to get final invite: %v", err)
	}

	if !finalInvite.Unsubscribed {
		t.Error("Expected invite to still be marked as unsubscribed after second call")
	}
}

// TestUnsubscribeHandler_ProductionTemplateSet proves that the production
// template set (rsvpPageTemplates as constructed in main.go) contains
// "unsubscribe.html" and renders it correctly with UnsubscribePageData.
// This is the regression test for the bug where unsubscribe.html was missing
// from the set and ExecuteTemplate returned an error.
func TestUnsubscribeHandler_ProductionTemplateSet_RendersUnsubscribePage(t *testing.T) {
	// Mirror the ParseFiles call from cmd/server/main.go exactly.
	tmpl, err := template.New("rsvp_page.html").ParseFiles(
		"../../templates/web/partials/base.html",
		"../../templates/web/partials/navigation.html",
		"../../templates/web/rsvp_page.html",
		"../../templates/web/unsubscribe.html",
	)
	if err != nil {
		t.Fatalf("Failed to parse production template set: %v", err)
	}

	database := setupIntegrationTestDB(t)
	defer database.Close()

	user := createTestUser(t, database)
	event := createTestEventForRSVP(t, database, user.ID)
	_, inviteToken := createTestInviteForRSVP(t, database, event.ID)

	secret := []byte("test-secret-key-32-bytes-long!!")
	generator := token.NewGenerator(secret)
	inviteRepo := repositories.NewInviteRepository(database)
	inviteService := invites.NewInviteService(generator, inviteRepo)
	eventRepo := repositories.NewEventRepository(database)
	rsvpRepo := repositories.NewRSVPRepository(database)
	questionRepo := repositories.NewQuestionRepository(database)

	handler := NewRSVPHandler(inviteService, eventRepo, rsvpRepo, questionRepo)
	handler.SetTemplates(tmpl)

	r := chi.NewRouter()
	r.Get("/unsubscribe/{token}", handler.Unsubscribe)

	req := httptest.NewRequest("GET", "/unsubscribe/"+inviteToken, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	body := w.Body.String()

	// Must not be the error fallback string.
	if strings.Contains(body, "Failed to render page") {
		t.Error("Response must not contain fallback error — template was not found in set")
	}
	// Must contain the styled page content from unsubscribe.html.
	if !strings.Contains(body, "Unsubscribed Successfully") {
		t.Errorf("Expected styled success content from unsubscribe.html, got: %s", body)
	}
	if !strings.Contains(body, event.Title) {
		t.Errorf("Expected event title %q in response, got: %s", event.Title, body)
	}
}

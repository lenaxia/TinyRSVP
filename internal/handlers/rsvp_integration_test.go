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
	"github.com/lenaxia/tinyrsvp/internal/db"
	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/invites"
	"github.com/lenaxia/tinyrsvp/internal/models"
	"github.com/lenaxia/tinyrsvp/pkg/token"
)

func createTestUser(t *testing.T, database db.Database) *models.User {
	t.Helper()

	userRepo := repositories.NewUserRepository(database)
	user := &models.User{
		Email: "admin@example.com",
		Name:  "Admin User",
		Role:  models.RoleAdmin,
	}

	if err := userRepo.Create(context.Background(), user); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	return user
}

func createTestEventForRSVP(t *testing.T, database db.Database, userID int64) *models.Event {
	t.Helper()

	eventRepo := repositories.NewEventRepository(database)
	startTime := time.Now().Add(30 * 24 * time.Hour)
	endTime := startTime.Add(2 * time.Hour)
	rsvpDeadline := startTime.Add(-7 * 24 * time.Hour)
	desc := "Test event for RSVP"
	loc := "123 Test St, Test City"

	event := &models.Event{
		Title:        "Test Event",
		Description:  &desc,
		StartTime:    startTime,
		EndTime:      &endTime,
		Timezone:     "America/Los_Angeles",
		Location:     &loc,
		Status:       models.EventStatusPublished,
		CreatedBy:    userID,
		MaxPlusOnes:  3,
		RSVPDeadline: &rsvpDeadline,
	}

	if err := eventRepo.Create(context.Background(), event); err != nil {
		t.Fatalf("Failed to create test event: %v", err)
	}

	return event
}

func createTestInviteForRSVP(t *testing.T, database db.Database, eventID int64) (*models.Invite, string) {
	t.Helper()

	secret := []byte("test-secret-key-32-bytes-long!!")
	generator := token.NewGenerator(secret)
	inviteRepo := repositories.NewInviteRepository(database)
	inviteService := invites.NewInviteService(generator, inviteRepo)

	email := "guest@example.com"
	name := "Test Guest"
	expiresAt := time.Now().Add(60 * 24 * time.Hour)

	invite, plainToken, err := inviteService.CreateInvite(
		context.Background(),
		eventID,
		&name,
		&email,
		2,
		expiresAt,
	)
	if err != nil {
		t.Fatalf("Failed to create test invite: %v", err)
	}

	invite.Status = models.InviteStatusSent
	now := time.Now()
	invite.SentAt = &now
	if err := inviteRepo.Update(context.Background(), invite); err != nil {
		t.Fatalf("Failed to update invite status: %v", err)
	}

	return invite, plainToken
}

func TestRSVPHandler_Integration_ValidToken(t *testing.T) {
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
	
	tmpl, err := template.ParseFiles("../../templates/web/rsvp_page.html")
	if err != nil {
		t.Logf("Warning: Failed to load template: %v. Using fallback HTML.", err)
	} else {
		handler.SetTemplates(tmpl)
	}

	r := chi.NewRouter()
	r.Get("/rsvp/{token}", handler.GetRSVPPage)

	req := httptest.NewRequest("GET", "/rsvp/"+inviteToken, nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	body := w.Body.String()

	if !strings.Contains(body, event.Title) {
		t.Errorf("Response should contain event title. Body: %s", body)
	}
}

func TestRSVPHandler_Integration_WithExistingRSVP(t *testing.T) {
	database := setupIntegrationTestDB(t)
	defer database.Close()

	user := createTestUser(t, database)
	event := createTestEventForRSVP(t, database, user.ID)
	invite, inviteToken := createTestInviteForRSVP(t, database, event.ID)

	rsvpRepo := repositories.NewRSVPRepository(database)
	rsvp := &models.RSVP{
		InviteID: invite.ID,
		Response: models.RSVPResponseYes,
		PlusOnes: 2,
	}
	if err := rsvpRepo.Create(context.Background(), rsvp); err != nil {
		t.Fatalf("Failed to create RSVP: %v", err)
	}

	invite.Status = models.InviteStatusResponded
	inviteRepo := repositories.NewInviteRepository(database)
	if err := inviteRepo.Update(context.Background(), invite); err != nil {
		t.Fatalf("Failed to update invite: %v", err)
	}

	secret := []byte("test-secret-key-32-bytes-long!!")
	generator := token.NewGenerator(secret)
	inviteService := invites.NewInviteService(generator, inviteRepo)
	eventRepo := repositories.NewEventRepository(database)
	questionRepo := repositories.NewQuestionRepository(database)

	handler := NewRSVPHandler(inviteService, eventRepo, rsvpRepo, questionRepo)
	
	tmpl, err := template.ParseFiles("../../templates/web/rsvp_page.html")
	if err != nil {
		t.Logf("Warning: Failed to load template: %v. Using fallback HTML.", err)
	} else {
		handler.SetTemplates(tmpl)
	}

	r := chi.NewRouter()
	r.Get("/rsvp/{token}", handler.GetRSVPPage)

	req := httptest.NewRequest("GET", "/rsvp/"+inviteToken, nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	body := w.Body.String()

	if !strings.Contains(body, "Already Responded") {
		t.Error("Response should indicate guest has already responded")
	}

	if !strings.Contains(body, "yes") {
		t.Error("Response should show the existing RSVP response")
	}
}

func TestRSVPHandler_Integration_ExpiredToken(t *testing.T) {
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
	r.Get("/rsvp/{token}", handler.GetRSVPPage)

	req := httptest.NewRequest("GET", "/rsvp/"+plainToken, nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusGone {
		t.Errorf("Expected status 410, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "expired") {
		t.Error("Response should indicate invite has expired")
	}
}

func TestRSVPHandler_Integration_RevokedInvite(t *testing.T) {
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
	r.Get("/rsvp/{token}", handler.GetRSVPPage)

	req := httptest.NewRequest("GET", "/rsvp/"+inviteToken, nil)
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

func TestRSVPHandler_Integration_CancelledEvent(t *testing.T) {
	database := setupIntegrationTestDB(t)
	defer database.Close()

	user := createTestUser(t, database)
	event := createTestEventForRSVP(t, database, user.ID)
	_, inviteToken := createTestInviteForRSVP(t, database, event.ID)

	eventRepo := repositories.NewEventRepository(database)
	if err := eventRepo.UpdateStatus(context.Background(), event.ID, models.EventStatusCancelled); err != nil {
		t.Fatalf("Failed to cancel event: %v", err)
	}

	secret := []byte("test-secret-key-32-bytes-long!!")
	generator := token.NewGenerator(secret)
	inviteRepo := repositories.NewInviteRepository(database)
	inviteService := invites.NewInviteService(generator, inviteRepo)
	rsvpRepo := repositories.NewRSVPRepository(database)
	questionRepo := repositories.NewQuestionRepository(database)

	handler := NewRSVPHandler(inviteService, eventRepo, rsvpRepo, questionRepo)

	r := chi.NewRouter()
	r.Get("/rsvp/{token}", handler.GetRSVPPage)

	req := httptest.NewRequest("GET", "/rsvp/"+inviteToken, nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusGone {
		t.Errorf("Expected status 410, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "cancelled") {
		t.Error("Response should indicate event has been cancelled")
	}
}

func TestRSVPHandler_Integration_InvalidToken(t *testing.T) {
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
	r.Get("/rsvp/{token}", handler.GetRSVPPage)

	req := httptest.NewRequest("GET", "/rsvp/invalidtoken123456789012345678901234567890", nil)
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

func TestRSVPHandler_Integration_MarkInviteViewed(t *testing.T) {
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
	r.Get("/rsvp/{token}", handler.GetRSVPPage)

	req := httptest.NewRequest("GET", "/rsvp/"+inviteToken, nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	updatedInvite, err := inviteRepo.GetByID(context.Background(), invite.ID)
	if err != nil {
		t.Fatalf("Failed to get updated invite: %v", err)
	}

	if updatedInvite.Status != models.InviteStatusViewed {
		t.Errorf("Expected invite status to be viewed, got %s", updatedInvite.Status)
	}

	if updatedInvite.ViewedAt == nil {
		t.Error("Expected ViewedAt to be set")
	}
}

func TestRSVPHandler_Integration_WithPreferenceQuestions(t *testing.T) {
	database := setupIntegrationTestDB(t)
	defer database.Close()

	user := createTestUser(t, database)
	event := createTestEventForRSVP(t, database, user.ID)
	_, inviteToken := createTestInviteForRSVP(t, database, event.ID)

	questionRepo := repositories.NewQuestionRepository(database)
	questionText := "What is your dietary preference?"
	question := &models.PreferenceQuestion{
		EventID:      event.ID,
		QuestionText: questionText,
		QuestionType: models.QuestionTypeText,
		Required:     true,
		DisplayOrder: 1,
	}
	if err := questionRepo.Create(context.Background(), question); err != nil {
		t.Fatalf("Failed to create question: %v", err)
	}

	secret := []byte("test-secret-key-32-bytes-long!!")
	generator := token.NewGenerator(secret)
	inviteRepo := repositories.NewInviteRepository(database)
	inviteService := invites.NewInviteService(generator, inviteRepo)
	eventRepo := repositories.NewEventRepository(database)
	rsvpRepo := repositories.NewRSVPRepository(database)

	handler := NewRSVPHandler(inviteService, eventRepo, rsvpRepo, questionRepo)

	r := chi.NewRouter()
	r.Get("/rsvp/{token}", handler.GetRSVPPage)

	req := httptest.NewRequest("GET", "/rsvp/"+inviteToken, nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

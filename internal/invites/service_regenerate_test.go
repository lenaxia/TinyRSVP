package invites

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/models"
	"github.com/lenaxia/tinyrsvp/pkg/token"
)

func strPtr(s string) *string {
	return &s
}

func TestRegenerateToken_Success(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := repositories.NewInviteRepository(db)
	secret := []byte("test-secret-key-32-bytes-long!")
	generator := token.NewGenerator(secret)
	service := NewInviteService(generator, repo)

	ctx := context.Background()

	eventID := createTestEvent(t, db)
	expiresAt := time.Now().Add(24 * time.Hour)

	invite, originalToken, err := service.CreateInvite(ctx, eventID, strPtr("Test User"), strPtr("test@example.com"), 2, expiresAt)
	if err != nil {
		t.Fatalf("Failed to create invite: %v", err)
	}

	originalTokenHash := invite.TokenHash

	result, err := service.RegenerateToken(ctx, invite.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	if result.Token == "" {
		t.Error("Expected non-empty token")
	}

	if result.Token == originalToken {
		t.Error("Expected new token to be different from original")
	}

	if result.RSVPURL == "" {
		t.Error("Expected non-empty RSVP URL")
	}

	expectedURL := "/rsvp/" + result.Token
	if result.RSVPURL != expectedURL {
		t.Errorf("Expected RSVP URL %s, got %s", expectedURL, result.RSVPURL)
	}

	updatedInvite, err := service.GetInviteByID(ctx, invite.ID)
	if err != nil {
		t.Fatalf("Failed to get updated invite: %v", err)
	}

	if updatedInvite.TokenHash == originalTokenHash {
		t.Error("Expected token hash to be updated")
	}

	if updatedInvite.Status != invite.Status {
		t.Errorf("Expected status to remain %s, got %s", invite.Status, updatedInvite.Status)
	}

	if updatedInvite.Name == nil || *updatedInvite.Name != *invite.Name {
		t.Error("Expected name to be preserved")
	}

	if updatedInvite.Email == nil || *updatedInvite.Email != *invite.Email {
		t.Error("Expected email to be preserved")
	}

	if updatedInvite.MaxPlusOnes != invite.MaxPlusOnes {
		t.Error("Expected max_plus_ones to be preserved")
	}
}

func TestRegenerateToken_OldTokenInvalidated(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := repositories.NewInviteRepository(db)
	secret := []byte("test-secret-key-32-bytes-long!")
	generator := token.NewGenerator(secret)
	service := NewInviteService(generator, repo)

	ctx := context.Background()

	eventID := createTestEvent(t, db)
	expiresAt := time.Now().Add(24 * time.Hour)

	invite, originalToken, err := service.CreateInvite(ctx, eventID, strPtr("Test User"), strPtr("test@example.com"), 2, expiresAt)
	if err != nil {
		t.Fatalf("Failed to create invite: %v", err)
	}

	_, err = service.RegenerateToken(ctx, invite.ID)
	if err != nil {
		t.Fatalf("Failed to regenerate token: %v", err)
	}

	_, err = service.GetInviteByToken(ctx, originalToken)
	if err == nil {
		t.Error("Expected error when using old token, got nil")
	}
}

func TestRegenerateToken_NewTokenWorks(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := repositories.NewInviteRepository(db)
	secret := []byte("test-secret-key-32-bytes-long!")
	generator := token.NewGenerator(secret)
	service := NewInviteService(generator, repo)

	ctx := context.Background()

	eventID := createTestEvent(t, db)
	expiresAt := time.Now().Add(24 * time.Hour)

	invite, _, err := service.CreateInvite(ctx, eventID, strPtr("Test User"), strPtr("test@example.com"), 2, expiresAt)
	if err != nil {
		t.Fatalf("Failed to create invite: %v", err)
	}

	result, err := service.RegenerateToken(ctx, invite.ID)
	if err != nil {
		t.Fatalf("Failed to regenerate token: %v", err)
	}

	retrievedInvite, err := service.GetInviteByToken(ctx, result.Token)
	if err != nil {
		t.Errorf("Expected no error when using new token, got %v", err)
	}

	if retrievedInvite.ID != invite.ID {
		t.Errorf("Expected invite ID %d, got %d", invite.ID, retrievedInvite.ID)
	}
}

func TestRegenerateToken_CannotRegenerateRevoked(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := repositories.NewInviteRepository(db)
	secret := []byte("test-secret-key-32-bytes-long!")
	generator := token.NewGenerator(secret)
	service := NewInviteService(generator, repo)

	ctx := context.Background()

	eventID := createTestEvent(t, db)
	expiresAt := time.Now().Add(24 * time.Hour)

	invite, _, err := service.CreateInvite(ctx, eventID, strPtr("Test User"), strPtr("test@example.com"), 2, expiresAt)
	if err != nil {
		t.Fatalf("Failed to create invite: %v", err)
	}

	err = service.RevokeInvite(ctx, &RevokeInviteRequest{
		InviteID: invite.ID,
		Reason:   strPtr("Test revocation"),
	})
	if err != nil {
		t.Fatalf("Failed to revoke invite: %v", err)
	}

	_, err = service.RegenerateToken(ctx, invite.ID)
	if err == nil {
		t.Error("Expected error when regenerating revoked invite, got nil")
	}

	if !strings.Contains(err.Error(), "revoked") {
		t.Errorf("Expected error to contain 'revoked', got %v", err)
	}
}

func TestRegenerateToken_CannotRegenerateResponded(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := repositories.NewInviteRepository(db)
	secret := []byte("test-secret-key-32-bytes-long!")
	generator := token.NewGenerator(secret)
	service := NewInviteService(generator, repo)

	ctx := context.Background()

	eventID := createTestEvent(t, db)
	expiresAt := time.Now().Add(24 * time.Hour)

	invite, _, err := service.CreateInvite(ctx, eventID, strPtr("Test User"), strPtr("test@example.com"), 2, expiresAt)
	if err != nil {
		t.Fatalf("Failed to create invite: %v", err)
	}

	invite.Status = models.InviteStatusResponded
	invite.UpdatedAt = time.Now()
	err = repo.Update(ctx, invite)
	if err != nil {
		t.Fatalf("Failed to update invite status: %v", err)
	}

	_, err = service.RegenerateToken(ctx, invite.ID)
	if err == nil {
		t.Error("Expected error when regenerating responded invite, got nil")
	}

	if !strings.Contains(err.Error(), "responded") {
		t.Errorf("Expected error to contain 'responded', got %v", err)
	}
}

func TestRegenerateToken_InviteNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := repositories.NewInviteRepository(db)
	secret := []byte("test-secret-key-32-bytes-long!")
	generator := token.NewGenerator(secret)
	service := NewInviteService(generator, repo)

	ctx := context.Background()

	_, err := service.RegenerateToken(ctx, 99999)
	if err == nil {
		t.Error("Expected error when invite not found, got nil")
	}
}

func TestRegenerateToken_PreservesSentStatus(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := repositories.NewInviteRepository(db)
	secret := []byte("test-secret-key-32-bytes-long!")
	generator := token.NewGenerator(secret)
	service := NewInviteService(generator, repo)

	ctx := context.Background()

	eventID := createTestEvent(t, db)
	expiresAt := time.Now().Add(24 * time.Hour)

	invite, _, err := service.CreateInvite(ctx, eventID, strPtr("Test User"), strPtr("test@example.com"), 2, expiresAt)
	if err != nil {
		t.Fatalf("Failed to create invite: %v", err)
	}

	invite.Status = models.InviteStatusSent
	now := time.Now()
	invite.SentAt = &now
	invite.UpdatedAt = now
	err = repo.Update(ctx, invite)
	if err != nil {
		t.Fatalf("Failed to update invite status: %v", err)
	}

	_, err = service.RegenerateToken(ctx, invite.ID)
	if err != nil {
		t.Fatalf("Failed to regenerate token: %v", err)
	}

	updatedInvite, err := service.GetInviteByID(ctx, invite.ID)
	if err != nil {
		t.Fatalf("Failed to get updated invite: %v", err)
	}

	if updatedInvite.Status != models.InviteStatusSent {
		t.Errorf("Expected status to remain %s, got %s", models.InviteStatusSent, updatedInvite.Status)
	}

	if updatedInvite.SentAt == nil {
		t.Error("Expected SentAt to be preserved")
	}
}

func TestRegenerateToken_PreservesViewedStatus(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := repositories.NewInviteRepository(db)
	secret := []byte("test-secret-key-32-bytes-long!")
	generator := token.NewGenerator(secret)
	service := NewInviteService(generator, repo)

	ctx := context.Background()

	eventID := createTestEvent(t, db)
	expiresAt := time.Now().Add(24 * time.Hour)

	invite, _, err := service.CreateInvite(ctx, eventID, strPtr("Test User"), strPtr("test@example.com"), 2, expiresAt)
	if err != nil {
		t.Fatalf("Failed to create invite: %v", err)
	}

	invite.Status = models.InviteStatusViewed
	now := time.Now()
	invite.ViewedAt = &now
	invite.UpdatedAt = now
	err = repo.Update(ctx, invite)
	if err != nil {
		t.Fatalf("Failed to update invite status: %v", err)
	}

	_, err = service.RegenerateToken(ctx, invite.ID)
	if err != nil {
		t.Fatalf("Failed to regenerate token: %v", err)
	}

	updatedInvite, err := service.GetInviteByID(ctx, invite.ID)
	if err != nil {
		t.Fatalf("Failed to get updated invite: %v", err)
	}

	if updatedInvite.Status != models.InviteStatusViewed {
		t.Errorf("Expected status to remain %s, got %s", models.InviteStatusViewed, updatedInvite.Status)
	}

	if updatedInvite.ViewedAt == nil {
		t.Error("Expected ViewedAt to be preserved")
	}
}

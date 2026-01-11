package repositories

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/models"
)

func TestInviteRepository_Create_WithToken(t *testing.T) {
	database := setupInviteTestDB(t)
	defer database.Close()

	repo := NewInviteRepository(database)

	plainToken := "test-token-12345"
	tokenHash := strings.Repeat("a", 43)
	futureTime := time.Now().Add(30 * 24 * time.Hour)

	invite := &models.Invite{
		EventID:     1,
		Token:       &plainToken,
		TokenHash:   tokenHash,
		MaxPlusOnes: 2,
		Status:      models.InviteStatusDraft,
		ExpiresAt:   futureTime,
	}

	err := repo.Create(context.Background(), invite)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	retrieved, err := repo.GetByID(context.Background(), invite.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if retrieved.Token == nil {
		t.Fatal("Expected token to be set, got nil")
	}

	if *retrieved.Token != plainToken {
		t.Errorf("Token = %s, want %s", *retrieved.Token, plainToken)
	}
}

func TestInviteRepository_CreateBatch_WithToken(t *testing.T) {
	database := setupInviteTestDB(t)
	defer database.Close()

	repo := NewInviteRepository(database)

	futureTime := time.Now().Add(30 * 24 * time.Hour)
	token1 := "token-1"
	token2 := "token-2"

	invites := []*models.Invite{
		{
			EventID:     1,
			Token:       &token1,
			TokenHash:   strings.Repeat("a", 43),
			MaxPlusOnes: 2,
			Status:      models.InviteStatusDraft,
			ExpiresAt:   futureTime,
		},
		{
			EventID:     1,
			Token:       &token2,
			TokenHash:   strings.Repeat("b", 43),
			MaxPlusOnes: 1,
			Status:      models.InviteStatusDraft,
			ExpiresAt:   futureTime,
		},
	}

	err := repo.CreateBatch(context.Background(), invites)
	if err != nil {
		t.Fatalf("CreateBatch() error = %v", err)
	}

	for i, invite := range invites {
		retrieved, err := repo.GetByID(context.Background(), invite.ID)
		if err != nil {
			t.Fatalf("GetByID() error = %v for invite %d", err, i)
		}

		if retrieved.Token == nil {
			t.Fatalf("Expected token to be set for invite %d, got nil", i)
		}

		expectedToken := token1
		if i == 1 {
			expectedToken = token2
		}

		if *retrieved.Token != expectedToken {
			t.Errorf("Invite %d: Token = %s, want %s", i, *retrieved.Token, expectedToken)
		}
	}
}

func TestInviteRepository_GetByTokenHash_ReturnsToken(t *testing.T) {
	database := setupInviteTestDB(t)
	defer database.Close()

	repo := NewInviteRepository(database)

	plainToken := "test-token-xyz"
	tokenHash := strings.Repeat("a", 43)
	futureTime := time.Now().Add(30 * 24 * time.Hour)

	invite := &models.Invite{
		EventID:     1,
		Token:       &plainToken,
		TokenHash:   tokenHash,
		MaxPlusOnes: 2,
		Status:      models.InviteStatusDraft,
		ExpiresAt:   futureTime,
	}

	err := repo.Create(context.Background(), invite)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	retrieved, err := repo.GetByTokenHash(context.Background(), tokenHash)
	if err != nil {
		t.Fatalf("GetByTokenHash() error = %v", err)
	}

	if retrieved.Token == nil {
		t.Fatal("Expected token to be set, got nil")
	}

	if *retrieved.Token != plainToken {
		t.Errorf("Token = %s, want %s", *retrieved.Token, plainToken)
	}
}

func TestInviteRepository_ListByEventID_ReturnsTokens(t *testing.T) {
	database := setupInviteTestDB(t)
	defer database.Close()

	repo := NewInviteRepository(database)

	futureTime := time.Now().Add(30 * 24 * time.Hour)
	token1 := "list-token-1"
	token2 := "list-token-2"

	invites := []*models.Invite{
		{
			EventID:     1,
			Token:       &token1,
			TokenHash:   strings.Repeat("a", 43),
			MaxPlusOnes: 2,
			Status:      models.InviteStatusDraft,
			ExpiresAt:   futureTime,
		},
		{
			EventID:     1,
			Token:       &token2,
			TokenHash:   strings.Repeat("b", 43),
			MaxPlusOnes: 1,
			Status:      models.InviteStatusDraft,
			ExpiresAt:   futureTime,
		},
	}

	for _, inv := range invites {
		if err := repo.Create(context.Background(), inv); err != nil {
			t.Fatalf("Failed to create invite: %v", err)
		}
	}

	results, err := repo.ListByEventID(context.Background(), 1, InviteFilters{})
	if err != nil {
		t.Fatalf("ListByEventID() error = %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("Expected 2 invites, got %d", len(results))
	}

	for i, result := range results {
		if result.Token == nil {
			t.Errorf("Invite %d: Expected token to be set, got nil", i)
		}
	}
}

func TestInviteRepository_Update_PreservesToken(t *testing.T) {
	database := setupInviteTestDB(t)
	defer database.Close()

	repo := NewInviteRepository(database)

	plainToken := "original-token"
	tokenHash := strings.Repeat("a", 43)
	futureTime := time.Now().Add(30 * 24 * time.Hour)
	email := "test@example.com"

	invite := &models.Invite{
		EventID:     1,
		Token:       &plainToken,
		TokenHash:   tokenHash,
		MaxPlusOnes: 2,
		Status:      models.InviteStatusDraft,
		ExpiresAt:   futureTime,
	}

	err := repo.Create(context.Background(), invite)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	invite.Email = &email
	invite.Status = models.InviteStatusSent
	now := time.Now()
	invite.SentAt = &now

	err = repo.Update(context.Background(), invite)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	retrieved, err := repo.GetByID(context.Background(), invite.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if retrieved.Token == nil {
		t.Fatal("Expected token to be preserved after update, got nil")
	}

	if *retrieved.Token != plainToken {
		t.Errorf("Token = %s, want %s (token should be preserved)", *retrieved.Token, plainToken)
	}
}

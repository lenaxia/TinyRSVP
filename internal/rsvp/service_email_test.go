package rsvp

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/email"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

func TestService_SubmitRSVP_SendsConfirmationEmail(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	ctx := context.Background()
	future := time.Now().Add(48 * time.Hour)
	rsvpDeadline := time.Now().Add(24 * time.Hour)

	eventRepo := repositories.NewEventRepository(database)
	event := &models.Event{
		Title:        "Test Event",
		Description:  strPtr("Test Description"),
		StartTime:    future,
		Timezone:     "UTC",
		Status:       models.EventStatusPublished,
		RSVPDeadline: &rsvpDeadline,
		CreatedBy:    1,
	}
	if err := eventRepo.Create(ctx, event); err != nil {
		t.Fatalf("Failed to create test event: %v", err)
	}

	inviteRepo := repositories.NewInviteRepository(database)
	invite := &models.Invite{
		EventID:     event.ID,
		TokenHash:   hashToken("validtoken"),
		Email:       strPtr("guest@example.com"),
		Name:        strPtr("Test Guest"),
		Status:      models.InviteStatusSent,
		MaxPlusOnes: 2,
		ExpiresAt:   future,
	}
	if err := inviteRepo.Create(ctx, invite); err != nil {
		t.Fatalf("Failed to create test invite: %v", err)
	}

	inviteService := &mockInviteService{
		getInviteByTokenFunc: func(ctx context.Context, token string) (*models.Invite, error) {
			return invite, nil
		},
	}

	rsvpRepo := repositories.NewRSVPRepository(database)
	answerRepo := repositories.NewAnswerRepository(database)
	questionRepo := repositories.NewQuestionRepository(database)

	var wg sync.WaitGroup
	wg.Add(1)
	mockEmail := &email.MockService{
		SendConfirmationEmailFunc: func(ctx context.Context, token string, rsvp *models.RSVP, invite *models.Invite, event *models.Event, answers []*models.RSVPAnswer) error {
			defer wg.Done()
			return nil
		},
	}
	service := NewServiceWithEmail(database, inviteService, inviteRepo, eventRepo, rsvpRepo, answerRepo, questionRepo, mockEmail)

	req := &SubmitRSVPRequest{
		Response: "yes",
		PlusOnes: 2,
		Answers:  []AnswerRequest{},
	}

	rsvp, err := service.SubmitRSVP(ctx, "validtoken", req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if rsvp == nil {
		t.Fatal("Expected RSVP to be returned")
	}

	wg.Wait()

	if mockEmail.SendConfirmationEmailCalls != 1 {
		t.Errorf("Expected SendConfirmationEmail to be called once, got %d calls", mockEmail.SendConfirmationEmailCalls)
	}
}

func TestService_SubmitRSVP_EmailFailureDoesNotBlockRSVP(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	ctx := context.Background()
	future := time.Now().Add(48 * time.Hour)
	rsvpDeadline := time.Now().Add(24 * time.Hour)

	eventRepo := repositories.NewEventRepository(database)
	event := &models.Event{
		Title:        "Test Event",
		Description:  strPtr("Test Description"),
		StartTime:    future,
		Timezone:     "UTC",
		Status:       models.EventStatusPublished,
		RSVPDeadline: &rsvpDeadline,
		CreatedBy:    1,
	}
	if err := eventRepo.Create(ctx, event); err != nil {
		t.Fatalf("Failed to create test event: %v", err)
	}

	inviteRepo := repositories.NewInviteRepository(database)
	invite := &models.Invite{
		EventID:     event.ID,
		TokenHash:   hashToken("validtoken"),
		Email:       strPtr("guest@example.com"),
		Name:        strPtr("Test Guest"),
		Status:      models.InviteStatusSent,
		MaxPlusOnes: 2,
		ExpiresAt:   future,
	}
	if err := inviteRepo.Create(ctx, invite); err != nil {
		t.Fatalf("Failed to create test invite: %v", err)
	}

	inviteService := &mockInviteService{
		getInviteByTokenFunc: func(ctx context.Context, token string) (*models.Invite, error) {
			return invite, nil
		},
	}

	rsvpRepo := repositories.NewRSVPRepository(database)
	answerRepo := repositories.NewAnswerRepository(database)
	questionRepo := repositories.NewQuestionRepository(database)

	var wg sync.WaitGroup
	wg.Add(1)
	mockEmail := &email.MockService{
		SendConfirmationEmailFunc: func(ctx context.Context, token string, rsvp *models.RSVP, invite *models.Invite, event *models.Event, answers []*models.RSVPAnswer) error {
			defer wg.Done()
			return errors.New("email service unavailable")
		},
	}
	service := NewServiceWithEmail(database, inviteService, inviteRepo, eventRepo, rsvpRepo, answerRepo, questionRepo, mockEmail)

	req := &SubmitRSVPRequest{
		Response: "yes",
		PlusOnes: 1,
		Answers:  []AnswerRequest{},
	}

	rsvp, err := service.SubmitRSVP(ctx, "validtoken", req)

	if err != nil {
		t.Fatalf("Expected RSVP to succeed even if email fails, got error: %v", err)
	}

	if rsvp == nil {
		t.Fatal("Expected RSVP to be returned")
	}

	if rsvp.Response != models.RSVPResponseYes {
		t.Errorf("Expected response 'yes', got '%s'", rsvp.Response)
	}

	wg.Wait()
}

func TestService_UpdateRSVP_SendsConfirmationEmail(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	ctx := context.Background()
	future := time.Now().Add(48 * time.Hour)
	rsvpDeadline := time.Now().Add(24 * time.Hour)

	eventRepo := repositories.NewEventRepository(database)
	event := &models.Event{
		Title:          "Test Event",
		Description:    strPtr("Test Description"),
		StartTime:      future,
		Timezone:       "UTC",
		Status:         models.EventStatusPublished,
		RSVPDeadline:   &rsvpDeadline,
		AllowMaybeRSVP: true,
		CreatedBy:      1,
	}
	if err := eventRepo.Create(ctx, event); err != nil {
		t.Fatalf("Failed to create test event: %v", err)
	}

	inviteRepo := repositories.NewInviteRepository(database)
	invite := &models.Invite{
		EventID:     event.ID,
		TokenHash:   hashToken("validtoken"),
		Email:       strPtr("guest@example.com"),
		Name:        strPtr("Test Guest"),
		Status:      models.InviteStatusResponded,
		MaxPlusOnes: 2,
		ExpiresAt:   future,
	}
	if err := inviteRepo.Create(ctx, invite); err != nil {
		t.Fatalf("Failed to create test invite: %v", err)
	}

	rsvpRepo := repositories.NewRSVPRepository(database)
	existingRSVP := &models.RSVP{
		InviteID: invite.ID,
		Response: models.RSVPResponseYes,
		PlusOnes: 1,
	}
	if err := rsvpRepo.Create(ctx, existingRSVP); err != nil {
		t.Fatalf("Failed to create existing RSVP: %v", err)
	}

	inviteService := &mockInviteService{
		getInviteByTokenFunc: func(ctx context.Context, token string) (*models.Invite, error) {
			return invite, nil
		},
	}

	answerRepo := repositories.NewAnswerRepository(database)
	questionRepo := repositories.NewQuestionRepository(database)

	var wg sync.WaitGroup
	wg.Add(1)
	mockEmail := &email.MockService{
		SendConfirmationEmailFunc: func(ctx context.Context, token string, rsvp *models.RSVP, invite *models.Invite, event *models.Event, answers []*models.RSVPAnswer) error {
			defer wg.Done()
			return nil
		},
	}
	service := NewServiceWithEmail(database, inviteService, inviteRepo, eventRepo, rsvpRepo, answerRepo, questionRepo, mockEmail)

	req := &SubmitRSVPRequest{
		Response: "maybe",
		PlusOnes: 2,
		Answers:  []AnswerRequest{},
	}

	rsvp, err := service.UpdateRSVP(ctx, "validtoken", req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if rsvp == nil {
		t.Fatal("Expected RSVP to be returned")
	}

	wg.Wait()

	if mockEmail.SendConfirmationEmailCalls != 1 {
		t.Errorf("Expected SendConfirmationEmail to be called once, got %d calls", mockEmail.SendConfirmationEmailCalls)
	}
}

func TestService_SubmitRSVP_NoEmailIfInviteHasNoEmail(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	ctx := context.Background()
	future := time.Now().Add(48 * time.Hour)
	rsvpDeadline := time.Now().Add(24 * time.Hour)

	eventRepo := repositories.NewEventRepository(database)
	event := &models.Event{
		Title:        "Test Event",
		Description:  strPtr("Test Description"),
		StartTime:    future,
		Timezone:     "UTC",
		Status:       models.EventStatusPublished,
		RSVPDeadline: &rsvpDeadline,
		CreatedBy:    1,
	}
	if err := eventRepo.Create(ctx, event); err != nil {
		t.Fatalf("Failed to create test event: %v", err)
	}

	inviteRepo := repositories.NewInviteRepository(database)
	invite := &models.Invite{
		EventID:     event.ID,
		TokenHash:   hashToken("validtoken"),
		Email:       nil,
		Name:        strPtr("Test Guest"),
		Status:      models.InviteStatusDraft,
		MaxPlusOnes: 2,
		ExpiresAt:   future,
	}
	if err := inviteRepo.Create(ctx, invite); err != nil {
		t.Fatalf("Failed to create test invite: %v", err)
	}

	inviteService := &mockInviteService{
		getInviteByTokenFunc: func(ctx context.Context, token string) (*models.Invite, error) {
			return invite, nil
		},
	}

	rsvpRepo := repositories.NewRSVPRepository(database)
	answerRepo := repositories.NewAnswerRepository(database)
	questionRepo := repositories.NewQuestionRepository(database)

	mockEmail := &email.MockService{}
	service := NewServiceWithEmail(database, inviteService, inviteRepo, eventRepo, rsvpRepo, answerRepo, questionRepo, mockEmail)

	req := &SubmitRSVPRequest{
		Response: "yes",
		PlusOnes: 1,
		Answers:  []AnswerRequest{},
	}

	rsvp, err := service.SubmitRSVP(ctx, "validtoken", req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if rsvp == nil {
		t.Fatal("Expected RSVP to be returned")
	}

	if mockEmail.SendConfirmationEmailCalls != 0 {
		t.Errorf("Expected SendConfirmationEmail not to be called when invite has no email, got %d calls", mockEmail.SendConfirmationEmailCalls)
	}
}

package repositories

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/db"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

func setupIntegrationTestDB(t *testing.T) db.Database {
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

	migrator, err := db.NewMigrator(database.DB(), "../../../migrations/sqlite")
	if err != nil {
		t.Fatalf("Failed to create migrator: %v", err)
	}

	ctx := context.Background()
	if err := migrator.Up(ctx); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	_, err = database.Exec(ctx, `
		INSERT INTO users (id, email, name, role, created_at, updated_at)
		VALUES (1, 'test@example.com', 'Test User', 'admin', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	return database
}

func TestRSVPIntegration_CreateWithAnswers(t *testing.T) {
	database := setupIntegrationTestDB(t)
	defer database.Close()

	rsvpRepo := NewRSVPRepository(database)
	answerRepo := NewAnswerRepository(database)

	ctx := context.Background()

	result, _ := database.Exec(ctx, `
		INSERT INTO events (title, start_time, timezone, status, created_by, max_plus_ones)
		VALUES ('Integration Test Event', ?, 'America/Los_Angeles', 'published', 1, 2)
	`, time.Now().Add(24*time.Hour))
	eventID, _ := result.LastInsertId()

	result, _ = database.Exec(ctx, `
		INSERT INTO invites (event_id, name, email, token_hash, max_plus_ones, status, expires_at)
		VALUES (?, 'Guest Name', 'guest@example.com', ?, 2, 'sent', ?)
	`, eventID, fmt.Sprintf("token_%d", time.Now().UnixNano()), time.Now().Add(48*time.Hour))
	inviteID, _ := result.LastInsertId()

	result, _ = database.Exec(ctx, `
		INSERT INTO preference_questions (event_id, question_text, question_type, required, display_order)
		VALUES (?, 'Dietary restrictions?', 'text', 1, 0)
	`, eventID)
	q1ID, _ := result.LastInsertId()

	result, _ = database.Exec(ctx, `
		INSERT INTO preference_questions (event_id, question_text, question_type, required, display_order)
		VALUES (?, 'Will you attend the after party?', 'single_choice', 0, 1)
	`, eventID)
	q2ID, _ := result.LastInsertId()

	rsvp := &models.RSVP{
		InviteID: inviteID,
		Response: models.RSVPResponseYes,
		PlusOnes: 2,
	}

	err := rsvpRepo.Create(ctx, rsvp)
	if err != nil {
		t.Fatalf("Failed to create RSVP: %v", err)
	}

	dietaryText := "Vegetarian"
	answer1 := &models.RSVPAnswer{
		RSVPID:     rsvp.ID,
		QuestionID: q1ID,
		AnswerText: &dietaryText,
	}

	err = answerRepo.Create(ctx, answer1)
	if err != nil {
		t.Fatalf("Failed to create answer 1: %v", err)
	}

	afterPartyOption := "yes"
	answer2 := &models.RSVPAnswer{
		RSVPID:       rsvp.ID,
		QuestionID:   q2ID,
		AnswerOption: &afterPartyOption,
	}

	err = answerRepo.Create(ctx, answer2)
	if err != nil {
		t.Fatalf("Failed to create answer 2: %v", err)
	}

	answers, err := answerRepo.GetByRSVPID(ctx, rsvp.ID)
	if err != nil {
		t.Fatalf("Failed to get answers: %v", err)
	}

	if len(answers) != 2 {
		t.Errorf("Expected 2 answers, got %d", len(answers))
	}
}

func TestRSVPIntegration_UpdateRSVPAndAnswers(t *testing.T) {
	database := setupIntegrationTestDB(t)
	defer database.Close()

	rsvpRepo := NewRSVPRepository(database)
	answerRepo := NewAnswerRepository(database)

	ctx := context.Background()

	result, _ := database.Exec(ctx, `
		INSERT INTO events (title, start_time, timezone, status, created_by, max_plus_ones)
		VALUES ('Update Test Event', ?, 'America/Los_Angeles', 'published', 1, 3)
	`, time.Now().Add(24*time.Hour))
	eventID, _ := result.LastInsertId()

	result, _ = database.Exec(ctx, `
		INSERT INTO invites (event_id, name, email, token_hash, max_plus_ones, status, expires_at)
		VALUES (?, 'Guest Name', 'guest@example.com', ?, 3, 'sent', ?)
	`, eventID, fmt.Sprintf("token_%d", time.Now().UnixNano()), time.Now().Add(48*time.Hour))
	inviteID, _ := result.LastInsertId()

	result, _ = database.Exec(ctx, `
		INSERT INTO preference_questions (event_id, question_text, question_type, required, display_order)
		VALUES (?, 'Meal preference?', 'text', 1, 0)
	`, eventID)
	questionID, _ := result.LastInsertId()

	rsvp := &models.RSVP{
		InviteID: inviteID,
		Response: models.RSVPResponseMaybe,
		PlusOnes: 1,
	}
	rsvpRepo.Create(ctx, rsvp)

	mealText := "Chicken"
	answer := &models.RSVPAnswer{
		RSVPID:     rsvp.ID,
		QuestionID: questionID,
		AnswerText: &mealText,
	}
	answerRepo.Create(ctx, answer)

	time.Sleep(1100 * time.Millisecond)

	rsvp.Response = models.RSVPResponseYes
	rsvp.PlusOnes = 3
	err := rsvpRepo.Update(ctx, rsvp)
	if err != nil {
		t.Fatalf("Failed to update RSVP: %v", err)
	}

	newMealText := "Fish"
	answer.AnswerText = &newMealText
	err = answerRepo.Update(ctx, answer)
	if err != nil {
		t.Fatalf("Failed to update answer: %v", err)
	}

	updatedRSVP, _ := rsvpRepo.GetByID(ctx, rsvp.ID)
	if updatedRSVP.Response != models.RSVPResponseYes {
		t.Errorf("RSVP response = %s, want yes", updatedRSVP.Response)
	}
	if updatedRSVP.PlusOnes != 3 {
		t.Errorf("RSVP plus_ones = %d, want 3", updatedRSVP.PlusOnes)
	}

	updatedAnswers, _ := answerRepo.GetByRSVPID(ctx, rsvp.ID)
	if len(updatedAnswers) != 1 {
		t.Fatalf("Expected 1 answer, got %d", len(updatedAnswers))
	}
	if *updatedAnswers[0].AnswerText != "Fish" {
		t.Errorf("Answer text = %s, want Fish", *updatedAnswers[0].AnswerText)
	}
}

func TestRSVPIntegration_GetStatsWithMultipleRSVPs(t *testing.T) {
	database := setupIntegrationTestDB(t)
	defer database.Close()

	rsvpRepo := NewRSVPRepository(database)

	ctx := context.Background()

	result, _ := database.Exec(ctx, `
		INSERT INTO events (title, start_time, timezone, status, created_by, max_plus_ones)
		VALUES ('Stats Test Event', ?, 'America/Los_Angeles', 'published', 1, 5)
	`, time.Now().Add(24*time.Hour))
	eventID, _ := result.LastInsertId()

	invites := []struct {
		maxPlusOnes int
		response    *models.RSVPResponse
		plusOnes    int
	}{
		{5, func() *models.RSVPResponse { r := models.RSVPResponseYes; return &r }(), 3},
		{5, func() *models.RSVPResponse { r := models.RSVPResponseYes; return &r }(), 2},
		{5, func() *models.RSVPResponse { r := models.RSVPResponseYes; return &r }(), 0},
		{5, func() *models.RSVPResponse { r := models.RSVPResponseNo; return &r }(), 0},
		{5, func() *models.RSVPResponse { r := models.RSVPResponseNo; return &r }(), 0},
		{5, func() *models.RSVPResponse { r := models.RSVPResponseMaybe; return &r }(), 1},
		{5, nil, 0},
		{5, nil, 0},
	}

	for i, inv := range invites {
		result, _ = database.Exec(ctx, `
			INSERT INTO invites (event_id, name, email, token_hash, max_plus_ones, status, expires_at)
			VALUES (?, ?, ?, ?, ?, 'sent', ?)
		`, eventID, fmt.Sprintf("Guest %d", i), fmt.Sprintf("guest%d@example.com", i),
			fmt.Sprintf("token_%d_%d", eventID, i), inv.maxPlusOnes, time.Now().Add(48*time.Hour))
		inviteID, _ := result.LastInsertId()

		if inv.response != nil {
			rsvp := &models.RSVP{
				InviteID: inviteID,
				Response: *inv.response,
				PlusOnes: inv.plusOnes,
			}
			rsvpRepo.Create(ctx, rsvp)
		}
	}

	stats, err := rsvpRepo.GetStats(ctx, eventID)
	if err != nil {
		t.Fatalf("GetStats() error = %v", err)
	}

	if stats.TotalInvites != 8 {
		t.Errorf("TotalInvites = %d, want 8", stats.TotalInvites)
	}
	if stats.YesCount != 3 {
		t.Errorf("YesCount = %d, want 3", stats.YesCount)
	}
	if stats.NoCount != 2 {
		t.Errorf("NoCount = %d, want 2", stats.NoCount)
	}
	if stats.MaybeCount != 1 {
		t.Errorf("MaybeCount = %d, want 1", stats.MaybeCount)
	}
	if stats.NoResponse != 2 {
		t.Errorf("NoResponse = %d, want 2", stats.NoResponse)
	}
	if stats.TotalGuests != 8 {
		t.Errorf("TotalGuests = %d, want 8 (3 yes: 1+3, 1+2, 1+0)", stats.TotalGuests)
	}
}

func TestRSVPIntegration_DeleteAnswersOnRSVPUpdate(t *testing.T) {
	database := setupIntegrationTestDB(t)
	defer database.Close()

	rsvpRepo := NewRSVPRepository(database)
	answerRepo := NewAnswerRepository(database)

	ctx := context.Background()

	result, _ := database.Exec(ctx, `
		INSERT INTO events (title, start_time, timezone, status, created_by, max_plus_ones)
		VALUES ('Delete Test Event', ?, 'America/Los_Angeles', 'published', 1, 2)
	`, time.Now().Add(24*time.Hour))
	eventID, _ := result.LastInsertId()

	result, _ = database.Exec(ctx, `
		INSERT INTO invites (event_id, name, email, token_hash, max_plus_ones, status, expires_at)
		VALUES (?, 'Guest Name', 'guest@example.com', ?, 2, 'sent', ?)
	`, eventID, fmt.Sprintf("token_%d", time.Now().UnixNano()), time.Now().Add(48*time.Hour))
	inviteID, _ := result.LastInsertId()

	result, _ = database.Exec(ctx, `
		INSERT INTO preference_questions (event_id, question_text, question_type, required, display_order)
		VALUES (?, 'Question 1', 'text', 0, 0)
	`, eventID)
	q1ID, _ := result.LastInsertId()

	result, _ = database.Exec(ctx, `
		INSERT INTO preference_questions (event_id, question_text, question_type, required, display_order)
		VALUES (?, 'Question 2', 'single_choice', 0, 1)
	`, eventID)
	q2ID, _ := result.LastInsertId()

	rsvp := &models.RSVP{
		InviteID: inviteID,
		Response: models.RSVPResponseYes,
		PlusOnes: 1,
	}
	rsvpRepo.Create(ctx, rsvp)

	text1 := "Answer 1"
	answer1 := &models.RSVPAnswer{
		RSVPID:     rsvp.ID,
		QuestionID: q1ID,
		AnswerText: &text1,
	}
	answerRepo.Create(ctx, answer1)

	option2 := "option1"
	answer2 := &models.RSVPAnswer{
		RSVPID:       rsvp.ID,
		QuestionID:   q2ID,
		AnswerOption: &option2,
	}
	answerRepo.Create(ctx, answer2)

	answers, _ := answerRepo.GetByRSVPID(ctx, rsvp.ID)
	if len(answers) != 2 {
		t.Fatalf("Expected 2 answers before delete, got %d", len(answers))
	}

	err := answerRepo.DeleteByRSVPID(ctx, rsvp.ID)
	if err != nil {
		t.Fatalf("DeleteByRSVPID() error = %v", err)
	}

	answers, _ = answerRepo.GetByRSVPID(ctx, rsvp.ID)
	if len(answers) != 0 {
		t.Errorf("Expected 0 answers after delete, got %d", len(answers))
	}

	retrievedRSVP, err := rsvpRepo.GetByID(ctx, rsvp.ID)
	if err != nil {
		t.Fatalf("RSVP should still exist after deleting answers: %v", err)
	}
	if retrievedRSVP.ID != rsvp.ID {
		t.Error("RSVP ID mismatch")
	}
}

func TestRSVPIntegration_CascadeDeleteOnInviteDelete(t *testing.T) {
	database := setupIntegrationTestDB(t)
	defer database.Close()

	rsvpRepo := NewRSVPRepository(database)
	answerRepo := NewAnswerRepository(database)

	ctx := context.Background()

	result, _ := database.Exec(ctx, `
		INSERT INTO events (title, start_time, timezone, status, created_by, max_plus_ones)
		VALUES ('Cascade Test Event', ?, 'America/Los_Angeles', 'published', 1, 2)
	`, time.Now().Add(24*time.Hour))
	eventID, _ := result.LastInsertId()

	result, _ = database.Exec(ctx, `
		INSERT INTO invites (event_id, name, email, token_hash, max_plus_ones, status, expires_at)
		VALUES (?, 'Guest Name', 'guest@example.com', ?, 2, 'sent', ?)
	`, eventID, fmt.Sprintf("token_%d", time.Now().UnixNano()), time.Now().Add(48*time.Hour))
	inviteID, _ := result.LastInsertId()

	result, _ = database.Exec(ctx, `
		INSERT INTO preference_questions (event_id, question_text, question_type, required, display_order)
		VALUES (?, 'Test Question', 'text', 0, 0)
	`, eventID)
	questionID, _ := result.LastInsertId()

	rsvp := &models.RSVP{
		InviteID: inviteID,
		Response: models.RSVPResponseYes,
		PlusOnes: 1,
	}
	rsvpRepo.Create(ctx, rsvp)

	text := "Test Answer"
	answer := &models.RSVPAnswer{
		RSVPID:     rsvp.ID,
		QuestionID: questionID,
		AnswerText: &text,
	}
	answerRepo.Create(ctx, answer)

	_, err := database.Exec(ctx, `DELETE FROM invites WHERE id = ?`, inviteID)
	if err != nil {
		t.Fatalf("Failed to delete invite: %v", err)
	}

	_, err = rsvpRepo.GetByID(ctx, rsvp.ID)
	if err == nil {
		t.Error("Expected RSVP to be deleted via cascade, but it still exists")
	}

	answers, _ := answerRepo.GetByRSVPID(ctx, rsvp.ID)
	if len(answers) != 0 {
		t.Errorf("Expected 0 answers after cascade delete, got %d", len(answers))
	}
}

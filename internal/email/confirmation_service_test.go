package email

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/models"
	"github.com/lenaxia/tinyrsvp/pkg/ics"
)

type mockTemplateRenderer struct {
	renderHTMLFunc func(ctx context.Context, templateName string, data interface{}) (string, error)
	renderTextFunc func(ctx context.Context, templateName string, data interface{}) (string, error)
	renderHTMLCalls int
	renderTextCalls int
}

func (m *mockTemplateRenderer) RenderHTML(ctx context.Context, templateName string, data interface{}) (string, error) {
	m.renderHTMLCalls++
	if m.renderHTMLFunc != nil {
		return m.renderHTMLFunc(ctx, templateName, data)
	}
	return "<html>test</html>", nil
}

func (m *mockTemplateRenderer) RenderText(ctx context.Context, templateName string, data interface{}) (string, error) {
	m.renderTextCalls++
	if m.renderTextFunc != nil {
		return m.renderTextFunc(ctx, templateName, data)
	}
	return "test text", nil
}

func (m *mockTemplateRenderer) LoadTemplates() error {
	return nil
}

func (m *mockTemplateRenderer) ReloadTemplates() error {
	return nil
}

type mockEmailQueueRepository struct {
	createFunc func(ctx context.Context, email *models.EmailQueue) error
	createCalls int
	lastEmail *models.EmailQueue
}

func (m *mockEmailQueueRepository) Create(ctx context.Context, email *models.EmailQueue) error {
	m.createCalls++
	m.lastEmail = email
	if m.createFunc != nil {
		return m.createFunc(ctx, email)
	}
	email.ID = 1
	email.CreatedAt = time.Now()
	return nil
}

func (m *mockEmailQueueRepository) GetByID(ctx context.Context, id int64) (*models.EmailQueue, error) {
	return nil, nil
}

func (m *mockEmailQueueRepository) GetPending(ctx context.Context, maxCount int) ([]*models.EmailQueue, error) {
	return nil, nil
}

func (m *mockEmailQueueRepository) GetByStatus(ctx context.Context, status models.EmailStatus, limit int) ([]*models.EmailQueue, error) {
	return nil, nil
}

func (m *mockEmailQueueRepository) GetByRecipient(ctx context.Context, email string, limit int) ([]*models.EmailQueue, error) {
	return nil, nil
}

func (m *mockEmailQueueRepository) UpdateStatus(ctx context.Context, id int64, status models.EmailStatus) error {
	return nil
}

func (m *mockEmailQueueRepository) IncrementAttempts(ctx context.Context, id int64, errorMsg string) error {
	return nil
}

func (m *mockEmailQueueRepository) MarkSending(ctx context.Context, id int64) error {
	return nil
}

func (m *mockEmailQueueRepository) MarkSent(ctx context.Context, id int64) error {
	return nil
}

func (m *mockEmailQueueRepository) MarkFailed(ctx context.Context, id int64, errorMsg string) error {
	return nil
}

func (m *mockEmailQueueRepository) MarkCancelled(ctx context.Context, id int64) error {
	return nil
}

func (m *mockEmailQueueRepository) Reschedule(ctx context.Context, id int64, scheduledFor time.Time) error {
	return nil
}

func (m *mockEmailQueueRepository) GetStats(ctx context.Context) (*repositories.EmailQueueStats, error) {
	return &repositories.EmailQueueStats{}, nil
}

type mockICSGenerator struct {
	generateFunc func(event *models.Event, rsvpURL string) ([]byte, error)
	generateCalls int
}

func (m *mockICSGenerator) Generate(event *models.Event, rsvpURL string) ([]byte, error) {
	m.generateCalls++
	if m.generateFunc != nil {
		return m.generateFunc(event, rsvpURL)
	}
	return []byte("BEGIN:VCALENDAR\nEND:VCALENDAR"), nil
}

func TestNewConfirmationService(t *testing.T) {
	renderer := &mockTemplateRenderer{}
	repo := &mockEmailQueueRepository{}
	generator := &mockICSGenerator{}

	service := NewConfirmationService(renderer, repo, generator)

	if service == nil {
		t.Fatal("Expected service to be created, got nil")
	}
}

func TestSendConfirmationEmail_HappyPath_Attending(t *testing.T) {
	renderer := &mockTemplateRenderer{}
	repo := &mockEmailQueueRepository{}
	generator := &mockICSGenerator{}

	service := NewConfirmationService(renderer, repo, generator)

	ctx := context.Background()
	startTime := time.Now().Add(24 * time.Hour)
	endTime := startTime.Add(2 * time.Hour)
	location := "123 Main St"
	guestName := "John Doe"

	event := &models.Event{
		ID:          1,
		Title:       "Test Event",
		StartTime:   startTime,
		EndTime:     &endTime,
		Location:    &location,
		Timezone:    "America/Los_Angeles",
		ICSSequence: 0,
	}

	email := "test@example.com"
	invite := &models.Invite{
		ID:      1,
		EventID: 1,
		Email:   &email,
		Name:    &guestName,
	}

	rsvp := &models.RSVP{
		ID:       1,
		InviteID: 1,
		Response: models.RSVPResponseYes,
		PlusOnes: 2,
	}

	answerText := "Vegetarian"
	answers := []*models.RSVPAnswer{
		{
			ID:         1,
			RSVPID:     1,
			QuestionID: 1,
			AnswerText: &answerText,
		},
	}

	err := service.SendConfirmationEmail(ctx, "test-token", rsvp, invite, event, answers)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if renderer.renderHTMLCalls != 1 {
		t.Errorf("Expected RenderHTML to be called once, got %d", renderer.renderHTMLCalls)
	}

	if renderer.renderTextCalls != 1 {
		t.Errorf("Expected RenderText to be called once, got %d", renderer.renderTextCalls)
	}

	if generator.generateCalls != 1 {
		t.Errorf("Expected Generate to be called once, got %d", generator.generateCalls)
	}

	if repo.createCalls != 1 {
		t.Errorf("Expected Create to be called once, got %d", repo.createCalls)
	}

	if repo.lastEmail == nil {
		t.Fatal("Expected email to be created")
	}

	if repo.lastEmail.ToEmail != *invite.Email {
		t.Errorf("Expected ToEmail to be %s, got %s", *invite.Email, repo.lastEmail.ToEmail)
	}

	if repo.lastEmail.ToName == nil || *repo.lastEmail.ToName != guestName {
		t.Errorf("Expected ToName to be %s, got %v", guestName, repo.lastEmail.ToName)
	}

	if repo.lastEmail.Subject != "RSVP Confirmed: Test Event" {
		t.Errorf("Expected subject to be 'RSVP Confirmed: Test Event', got %s", repo.lastEmail.Subject)
	}

	if repo.lastEmail.Status != models.EmailStatusPending {
		t.Errorf("Expected status to be pending, got %s", repo.lastEmail.Status)
	}

	if repo.lastEmail.MaxAttempts != 3 {
		t.Errorf("Expected max attempts to be 3, got %d", repo.lastEmail.MaxAttempts)
	}

	attachments, err := repo.lastEmail.GetAttachments()
	if err != nil {
		t.Fatalf("Failed to get attachments: %v", err)
	}

	if len(attachments) != 1 {
		t.Fatalf("Expected 1 attachment, got %d", len(attachments))
	}

	if attachments[0].Filename != "event.ics" {
		t.Errorf("Expected attachment filename to be 'event.ics', got %s", attachments[0].Filename)
	}

	if attachments[0].ContentType != "text/calendar" {
		t.Errorf("Expected attachment content type to be 'text/calendar', got %s", attachments[0].ContentType)
	}
}

func TestSendConfirmationEmail_HappyPath_Declined(t *testing.T) {
	renderer := &mockTemplateRenderer{}
	repo := &mockEmailQueueRepository{}
	generator := &mockICSGenerator{}

	service := NewConfirmationService(renderer, repo, generator)

	ctx := context.Background()
	startTime := time.Now().Add(24 * time.Hour)
	guestName := "Jane Smith"

	event := &models.Event{
		ID:        1,
		Title:     "Test Event",
		StartTime: startTime,
		Timezone:  "America/Los_Angeles",
	}

	email := "jane@example.com"
	invite := &models.Invite{
		ID:      1,
		EventID: 1,
		Email:   &email,
		Name:    &guestName,
	}

	rsvp := &models.RSVP{
		ID:       2,
		InviteID: 1,
		Response: models.RSVPResponseNo,
		PlusOnes: 0,
	}

	err := service.SendConfirmationEmail(ctx, "test-token-2", rsvp, invite, event, nil)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if repo.createCalls != 1 {
		t.Errorf("Expected Create to be called once, got %d", repo.createCalls)
	}
}

func TestSendConfirmationEmail_HappyPath_Tentative(t *testing.T) {
	renderer := &mockTemplateRenderer{}
	repo := &mockEmailQueueRepository{}
	generator := &mockICSGenerator{}

	service := NewConfirmationService(renderer, repo, generator)

	ctx := context.Background()
	startTime := time.Now().Add(24 * time.Hour)

	event := &models.Event{
		ID:        1,
		Title:     "Test Event",
		StartTime: startTime,
		Timezone:  "America/Los_Angeles",
	}

	email := "maybe@example.com"
	invite := &models.Invite{
		ID:      1,
		EventID: 1,
		Email:   &email,
	}

	rsvp := &models.RSVP{
		ID:       3,
		InviteID: 1,
		Response: models.RSVPResponseMaybe,
		PlusOnes: 1,
	}

	err := service.SendConfirmationEmail(ctx, "test-token-3", rsvp, invite, event, nil)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestSendConfirmationEmail_TemplateRenderingError_HTML(t *testing.T) {
	renderer := &mockTemplateRenderer{
		renderHTMLFunc: func(ctx context.Context, templateName string, data interface{}) (string, error) {
			return "", fmt.Errorf("template rendering failed")
		},
	}
	repo := &mockEmailQueueRepository{}
	generator := &mockICSGenerator{}

	service := NewConfirmationService(renderer, repo, generator)

	ctx := context.Background()
	startTime := time.Now().Add(24 * time.Hour)

	event := &models.Event{
		ID:        1,
		Title:     "Test Event",
		StartTime: startTime,
		Timezone:  "America/Los_Angeles",
	}

	email := "test@example.com"
	invite := &models.Invite{
		ID:      1,
		EventID: 1,
		Email:   &email,
	}

	rsvp := &models.RSVP{
		ID:       1,
		InviteID: 1,
		Response: models.RSVPResponseYes,
	}

	err := service.SendConfirmationEmail(ctx, "test-token", rsvp, invite, event, nil)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if repo.createCalls != 0 {
		t.Errorf("Expected Create not to be called, got %d calls", repo.createCalls)
	}
}

func TestSendConfirmationEmail_TemplateRenderingError_Text(t *testing.T) {
	renderer := &mockTemplateRenderer{
		renderTextFunc: func(ctx context.Context, templateName string, data interface{}) (string, error) {
			return "", fmt.Errorf("text template rendering failed")
		},
	}
	repo := &mockEmailQueueRepository{}
	generator := &mockICSGenerator{}

	service := NewConfirmationService(renderer, repo, generator)

	ctx := context.Background()
	startTime := time.Now().Add(24 * time.Hour)

	event := &models.Event{
		ID:        1,
		Title:     "Test Event",
		StartTime: startTime,
		Timezone:  "America/Los_Angeles",
	}

	email := "test@example.com"
	invite := &models.Invite{
		ID:      1,
		EventID: 1,
		Email:   &email,
	}

	rsvp := &models.RSVP{
		ID:       1,
		InviteID: 1,
		Response: models.RSVPResponseYes,
	}

	err := service.SendConfirmationEmail(ctx, "test-token", rsvp, invite, event, nil)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if repo.createCalls != 0 {
		t.Errorf("Expected Create not to be called, got %d calls", repo.createCalls)
	}
}

func TestSendConfirmationEmail_ICSGenerationError(t *testing.T) {
	renderer := &mockTemplateRenderer{}
	repo := &mockEmailQueueRepository{}
	generator := &mockICSGenerator{
		generateFunc: func(event *models.Event, rsvpURL string) ([]byte, error) {
			return nil, fmt.Errorf("ICS generation failed")
		},
	}

	service := NewConfirmationService(renderer, repo, generator)

	ctx := context.Background()
	startTime := time.Now().Add(24 * time.Hour)

	event := &models.Event{
		ID:        1,
		Title:     "Test Event",
		StartTime: startTime,
		Timezone:  "America/Los_Angeles",
	}

	email := "test@example.com"
	invite := &models.Invite{
		ID:      1,
		EventID: 1,
		Email:   &email,
	}

	rsvp := &models.RSVP{
		ID:       1,
		InviteID: 1,
		Response: models.RSVPResponseYes,
	}

	err := service.SendConfirmationEmail(ctx, "test-token", rsvp, invite, event, nil)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if repo.createCalls != 0 {
		t.Errorf("Expected Create not to be called, got %d calls", repo.createCalls)
	}
}

func TestSendConfirmationEmail_EmailQueueError(t *testing.T) {
	renderer := &mockTemplateRenderer{}
	repo := &mockEmailQueueRepository{
		createFunc: func(ctx context.Context, email *models.EmailQueue) error {
			return fmt.Errorf("database error")
		},
	}
	generator := &mockICSGenerator{}

	service := NewConfirmationService(renderer, repo, generator)

	ctx := context.Background()
	startTime := time.Now().Add(24 * time.Hour)

	event := &models.Event{
		ID:        1,
		Title:     "Test Event",
		StartTime: startTime,
		Timezone:  "America/Los_Angeles",
	}

	email := "test@example.com"
	invite := &models.Invite{
		ID:      1,
		EventID: 1,
		Email:   &email,
	}

	rsvp := &models.RSVP{
		ID:       1,
		InviteID: 1,
		Response: models.RSVPResponseYes,
	}

	err := service.SendConfirmationEmail(ctx, "test-token", rsvp, invite, event, nil)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if repo.createCalls != 1 {
		t.Errorf("Expected Create to be called once, got %d", repo.createCalls)
	}
}

func TestSendConfirmationEmail_WithAnswers(t *testing.T) {
	var capturedData interface{}
	renderer := &mockTemplateRenderer{
		renderHTMLFunc: func(ctx context.Context, templateName string, data interface{}) (string, error) {
			capturedData = data
			return "<html>test</html>", nil
		},
	}
	repo := &mockEmailQueueRepository{}
	generator := &mockICSGenerator{}

	service := NewConfirmationService(renderer, repo, generator)

	ctx := context.Background()
	startTime := time.Now().Add(24 * time.Hour)
	guestName := "Test User"

	event := &models.Event{
		ID:        1,
		Title:     "Test Event",
		StartTime: startTime,
		Timezone:  "America/Los_Angeles",
	}

	email := "test@example.com"
	invite := &models.Invite{
		ID:      1,
		EventID: 1,
		Email:   &email,
		Name:    &guestName,
	}

	rsvp := &models.RSVP{
		ID:       1,
		InviteID: 1,
		Response: models.RSVPResponseYes,
		PlusOnes: 1,
	}

	answer1 := "Vegetarian"
	answer2 := "No allergies"
	answers := []*models.RSVPAnswer{
		{
			ID:         1,
			RSVPID:     1,
			QuestionID: 1,
			AnswerText: &answer1,
		},
		{
			ID:         2,
			RSVPID:     1,
			QuestionID: 2,
			AnswerText: &answer2,
		},
	}

	err := service.SendConfirmationEmail(ctx, "test-token", rsvp, invite, event, answers)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if capturedData == nil {
		t.Fatal("Expected template data to be captured")
	}

	dataMap, ok := capturedData.(map[string]interface{})
	if !ok {
		t.Fatal("Expected template data to be a map")
	}

	if dataMap["GuestName"] != guestName {
		t.Errorf("Expected GuestName to be %s, got %v", guestName, dataMap["GuestName"])
	}

	if dataMap["Response"] != "yes" {
		t.Errorf("Expected Response to be 'yes', got %v", dataMap["Response"])
	}

	if dataMap["PlusOnes"] != 1 {
		t.Errorf("Expected PlusOnes to be 1, got %v", dataMap["PlusOnes"])
	}

	answersData, ok := dataMap["Answers"].([]map[string]string)
	if !ok {
		t.Fatal("Expected Answers to be a slice of maps")
	}

	if len(answersData) != 2 {
		t.Fatalf("Expected 2 answers, got %d", len(answersData))
	}
}

func TestSendConfirmationEmail_NilGuestName(t *testing.T) {
	renderer := &mockTemplateRenderer{}
	repo := &mockEmailQueueRepository{}
	generator := &mockICSGenerator{}

	service := NewConfirmationService(renderer, repo, generator)

	ctx := context.Background()
	startTime := time.Now().Add(24 * time.Hour)

	event := &models.Event{
		ID:        1,
		Title:     "Test Event",
		StartTime: startTime,
		Timezone:  "America/Los_Angeles",
	}

	email := "test@example.com"
	invite := &models.Invite{
		ID:      1,
		EventID: 1,
		Email:   &email,
		Name:    nil,
	}

	rsvp := &models.RSVP{
		ID:       1,
		InviteID: 1,
		Response: models.RSVPResponseYes,
	}

	err := service.SendConfirmationEmail(ctx, "test-token", rsvp, invite, event, nil)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if repo.lastEmail.ToName != nil {
		t.Errorf("Expected ToName to be nil, got %v", repo.lastEmail.ToName)
	}
}

func TestSendConfirmationEmail_ContextCancellation(t *testing.T) {
	renderer := &mockTemplateRenderer{
		renderHTMLFunc: func(ctx context.Context, templateName string, data interface{}) (string, error) {
			return "", ctx.Err()
		},
	}
	repo := &mockEmailQueueRepository{}
	generator := &mockICSGenerator{}

	service := NewConfirmationService(renderer, repo, generator)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	startTime := time.Now().Add(24 * time.Hour)

	event := &models.Event{
		ID:        1,
		Title:     "Test Event",
		StartTime: startTime,
		Timezone:  "America/Los_Angeles",
	}

	email := "test@example.com"
	invite := &models.Invite{
		ID:      1,
		EventID: 1,
		Email:   &email,
	}

	rsvp := &models.RSVP{
		ID:       1,
		InviteID: 1,
		Response: models.RSVPResponseYes,
	}

	err := service.SendConfirmationEmail(ctx, "test-token", rsvp, invite, event, nil)

	if err == nil {
		t.Fatal("Expected error due to context cancellation, got nil")
	}
}

func TestConfirmationService_Integration(t *testing.T) {
	config := &TemplateConfig{
		TemplateDir:  "../../templates/email",
		CacheEnabled: true,
	}

	renderer, err := NewTemplateRenderer(config)
	if err != nil {
		t.Skipf("Skipping integration test: %v", err)
		return
	}

	repo := &mockEmailQueueRepository{}
	generator := ics.NewGenerator()

	service := NewConfirmationService(renderer, repo, generator)

	ctx := context.Background()
	startTime := time.Now().Add(24 * time.Hour)
	endTime := startTime.Add(2 * time.Hour)
	location := "123 Main St"
	description := "Test event description"
	guestName := "Integration Test User"

	event := &models.Event{
		ID:          1,
		Title:       "Integration Test Event",
		Description: &description,
		StartTime:   startTime,
		EndTime:     &endTime,
		Location:    &location,
		Timezone:    "America/Los_Angeles",
		ICSSequence: 0,
	}

	email := "integration@example.com"
	invite := &models.Invite{
		ID:      1,
		EventID: 1,
		Email:   &email,
		Name:    &guestName,
	}

	rsvp := &models.RSVP{
		ID:       1,
		InviteID: 1,
		Response: models.RSVPResponseYes,
		PlusOnes: 2,
	}

	answerText := "Vegetarian"
	answers := []*models.RSVPAnswer{
		{
			ID:         1,
			RSVPID:     1,
			QuestionID: 1,
			AnswerText: &answerText,
		},
	}

	err = service.SendConfirmationEmail(ctx, "integration-token", rsvp, invite, event, answers)

	if err != nil {
		t.Fatalf("Integration test failed: %v", err)
	}

	if repo.createCalls != 1 {
		t.Errorf("Expected Create to be called once, got %d", repo.createCalls)
	}

	if repo.lastEmail == nil {
		t.Fatal("Expected email to be created")
	}

	if repo.lastEmail.BodyHTML == nil || *repo.lastEmail.BodyHTML == "" {
		t.Error("Expected HTML body to be rendered")
	}

	if repo.lastEmail.BodyText == "" {
		t.Error("Expected text body to be rendered")
	}

	attachments, err := repo.lastEmail.GetAttachments()
	if err != nil {
		t.Fatalf("Failed to get attachments: %v", err)
	}

	if len(attachments) != 1 {
		t.Fatalf("Expected 1 attachment, got %d", len(attachments))
	}

	if len(attachments[0].Content) == 0 {
		t.Error("Expected ICS attachment to have content")
	}
}

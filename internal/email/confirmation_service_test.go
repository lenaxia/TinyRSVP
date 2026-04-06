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
	renderHTMLFunc  func(ctx context.Context, templateName string, data interface{}) (string, error)
	renderTextFunc  func(ctx context.Context, templateName string, data interface{}) (string, error)
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
	createFunc  func(ctx context.Context, email *models.EmailQueue) error
	createCalls int
	lastEmail   *models.EmailQueue
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
	generateFunc  func(event *models.Event, rsvpURL string) ([]byte, error)
	generateCalls int
}

func (m *mockICSGenerator) Generate(event *models.Event, rsvpURL string) ([]byte, error) {
	m.generateCalls++
	if m.generateFunc != nil {
		return m.generateFunc(event, rsvpURL)
	}
	return []byte("BEGIN:VCALENDAR\nEND:VCALENDAR"), nil
}

// mockQuestionRepository is a minimal stub for QuestionRepository used in
// confirmation service tests. GetByEventID is the primary method exercised —
// it filters the questions map by EventID, matching production behaviour.
type mockQuestionRepository struct {
	questions map[int64]*models.PreferenceQuestion
}

func (m *mockQuestionRepository) GetByID(ctx context.Context, id int64) (*models.PreferenceQuestion, error) {
	if q, ok := m.questions[id]; ok {
		return q, nil
	}
	return nil, fmt.Errorf("question %d not found", id)
}

func (m *mockQuestionRepository) Create(ctx context.Context, q *models.PreferenceQuestion) error {
	return nil
}
func (m *mockQuestionRepository) GetByEventID(ctx context.Context, eventID int64) ([]*models.PreferenceQuestion, error) {
	var result []*models.PreferenceQuestion
	for _, q := range m.questions {
		if q.EventID == eventID {
			result = append(result, q)
		}
	}
	return result, nil
}
func (m *mockQuestionRepository) Update(ctx context.Context, q *models.PreferenceQuestion) error {
	return nil
}
func (m *mockQuestionRepository) Delete(ctx context.Context, id int64) error { return nil }
func (m *mockQuestionRepository) Reorder(ctx context.Context, eventID int64, ids []int64) error {
	return nil
}

func TestNewConfirmationService(t *testing.T) {
	renderer := &mockTemplateRenderer{}
	repo := &mockEmailQueueRepository{}
	generator := &mockICSGenerator{}

	service := NewConfirmationService(renderer, repo, generator, "https://rsvp.example.com")

	if service == nil {
		t.Fatal("Expected service to be created, got nil")
	}
}

func TestSendConfirmationEmail_HappyPath_Attending(t *testing.T) {
	renderer := &mockTemplateRenderer{}
	repo := &mockEmailQueueRepository{}
	generator := &mockICSGenerator{}

	service := NewConfirmationService(renderer, repo, generator, "https://rsvp.example.com")

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

	service := NewConfirmationService(renderer, repo, generator, "https://rsvp.example.com")

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

	service := NewConfirmationService(renderer, repo, generator, "https://rsvp.example.com")

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

	service := NewConfirmationService(renderer, repo, generator, "https://rsvp.example.com")

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

	service := NewConfirmationService(renderer, repo, generator, "https://rsvp.example.com")

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

	service := NewConfirmationService(renderer, repo, generator, "https://rsvp.example.com")

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

	service := NewConfirmationService(renderer, repo, generator, "https://rsvp.example.com")

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
	questionRepo := &mockQuestionRepository{
		questions: map[int64]*models.PreferenceQuestion{
			1: {ID: 1, EventID: 1, QuestionText: "Dietary requirements?"},
			2: {ID: 2, EventID: 1, QuestionText: "T-shirt size?"},
		},
	}

	service := NewConfirmationServiceWithQuestions(renderer, repo, generator, "https://rsvp.example.com", questionRepo)

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
		{ID: 1, RSVPID: 1, QuestionID: 1, AnswerText: &answer1},
		{ID: 2, RSVPID: 1, QuestionID: 2, AnswerText: &answer2},
	}

	err := service.SendConfirmationEmail(ctx, "test-token", rsvp, invite, event, answers)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if capturedData == nil {
		t.Fatal("Expected template data to be captured")
	}

	templateData, ok := capturedData.(*RSVPConfirmationTemplateData)
	if !ok {
		t.Fatalf("Expected template data to be *RSVPConfirmationTemplateData, got %T", capturedData)
	}

	if templateData.GuestName != guestName {
		t.Errorf("Expected GuestName to be %s, got %s", guestName, templateData.GuestName)
	}

	if templateData.Response != "yes" {
		t.Errorf("Expected Response to be 'yes', got %s", templateData.Response)
	}

	if templateData.PlusOnes != 1 {
		t.Errorf("Expected PlusOnes to be 1, got %d", templateData.PlusOnes)
	}

	if len(templateData.Answers) != 2 {
		t.Fatalf("Expected 2 answers, got %d", len(templateData.Answers))
	}

	if templateData.Answers[0].Question != "Dietary requirements?" {
		t.Errorf("Expected first answer question to be 'Dietary requirements?', got %s", templateData.Answers[0].Question)
	}

	if templateData.Answers[0].Answer != answer1 {
		t.Errorf("Expected first answer to be %s, got %s", answer1, templateData.Answers[0].Answer)
	}

	if templateData.Answers[1].Question != "T-shirt size?" {
		t.Errorf("Expected second answer question to be 'T-shirt size?', got %s", templateData.Answers[1].Question)
	}

	if templateData.Answers[1].Answer != answer2 {
		t.Errorf("Expected second answer to be %s, got %s", answer2, templateData.Answers[1].Answer)
	}
}

func TestSendConfirmationEmail_NilGuestName(t *testing.T) {
	renderer := &mockTemplateRenderer{}
	repo := &mockEmailQueueRepository{}
	generator := &mockICSGenerator{}

	service := NewConfirmationService(renderer, repo, generator, "https://rsvp.example.com")

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

	service := NewConfirmationService(renderer, repo, generator, "https://rsvp.example.com")

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

	service := NewConfirmationService(renderer, repo, generator, "https://rsvp.example.com")

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

func TestSendConfirmationEmail_UsesBaseURL(t *testing.T) {
	baseURL := "https://rsvp.myserver.com"

	var capturedICSURL string
	generator := &mockICSGenerator{
		generateFunc: func(event *models.Event, rsvpURL string) ([]byte, error) {
			capturedICSURL = rsvpURL
			return []byte("BEGIN:VCALENDAR\nEND:VCALENDAR"), nil
		},
	}

	var capturedTemplateData interface{}
	renderer := &mockTemplateRenderer{
		renderHTMLFunc: func(ctx context.Context, templateName string, data interface{}) (string, error) {
			capturedTemplateData = data
			return "<html>test</html>", nil
		},
	}
	repo := &mockEmailQueueRepository{}

	service := NewConfirmationService(renderer, repo, generator, baseURL)

	ctx := context.Background()
	startTime := time.Now().Add(24 * time.Hour)
	email := "guest@example.com"
	name := "Guest"

	event := &models.Event{ID: 1, Title: "My Event", StartTime: startTime, Timezone: "UTC"}
	invite := &models.Invite{ID: 1, EventID: 1, Email: &email, Name: &name}
	rsvp := &models.RSVP{ID: 1, InviteID: 1, Response: models.RSVPResponseYes}

	err := service.SendConfirmationEmail(ctx, "abc123token", rsvp, invite, event, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	expectedURL := "https://rsvp.myserver.com/rsvp/abc123token"

	if capturedICSURL != expectedURL {
		t.Errorf("ICS rsvpURL = %q, want %q", capturedICSURL, expectedURL)
	}

	if td, ok := capturedTemplateData.(*RSVPConfirmationTemplateData); ok {
		if td.UpdateURL != expectedURL {
			t.Errorf("UpdateURL = %q, want %q", td.UpdateURL, expectedURL)
		}
	} else {
		t.Errorf("Template data was not *RSVPConfirmationTemplateData, got %T", capturedTemplateData)
	}
}

func TestSendConfirmationEmail_WithAnswers_FallsBackWhenQuestionNotFound(t *testing.T) {
	var capturedData interface{}
	renderer := &mockTemplateRenderer{
		renderHTMLFunc: func(ctx context.Context, templateName string, data interface{}) (string, error) {
			capturedData = data
			return "<html>test</html>", nil
		},
	}

	// Question repo that knows about question 1 (EventID 1) but not question 99.
	// Answer for question 99 should be silently omitted.
	questionRepo := &mockQuestionRepository{
		questions: map[int64]*models.PreferenceQuestion{
			1: {ID: 1, EventID: 1, QuestionText: "Dietary requirements?"},
		},
	}

	service := NewConfirmationServiceWithQuestions(renderer, &mockEmailQueueRepository{}, &mockICSGenerator{}, "https://rsvp.example.com", questionRepo)

	startTime := time.Now().Add(24 * time.Hour)
	email := "test@example.com"
	name := "Alice"
	answer1 := "Vegan"
	answer2 := "Medium"

	answers := []*models.RSVPAnswer{
		{ID: 1, RSVPID: 1, QuestionID: 1, AnswerText: &answer1},
		{ID: 2, RSVPID: 1, QuestionID: 99, AnswerText: &answer2}, // question 99 doesn't exist
	}

	err := service.SendConfirmationEmail(
		context.Background(), "tok",
		&models.RSVP{ID: 1, Response: models.RSVPResponseYes},
		&models.Invite{ID: 1, EventID: 1, Email: &email, Name: &name},
		&models.Event{ID: 1, Title: "Test Event", StartTime: startTime},
		answers,
	)
	if err != nil {
		t.Fatalf("Expected no error on missing question, got %v", err)
	}

	td, ok := capturedData.(*RSVPConfirmationTemplateData)
	if !ok {
		t.Fatalf("Expected *RSVPConfirmationTemplateData, got %T", capturedData)
	}
	// Only the known question should appear; the deleted one is omitted.
	if len(td.Answers) != 1 {
		t.Fatalf("Expected 1 answer (deleted question omitted), got %d", len(td.Answers))
	}
	if td.Answers[0].Question != "Dietary requirements?" {
		t.Errorf("Expected 'Dietary requirements?', got %q", td.Answers[0].Question)
	}
}

func TestSendConfirmationEmail_WithAnswers_NoQuestionRepo_OmitsAnswers(t *testing.T) {
	var capturedData interface{}
	renderer := &mockTemplateRenderer{
		renderHTMLFunc: func(ctx context.Context, templateName string, data interface{}) (string, error) {
			capturedData = data
			return "<html>test</html>", nil
		},
	}

	// NewConfirmationService — no question repo.
	// Answers cannot be labelled so they should be omitted entirely.
	service := NewConfirmationService(renderer, &mockEmailQueueRepository{}, &mockICSGenerator{}, "https://rsvp.example.com")

	startTime := time.Now().Add(24 * time.Hour)
	email := "test@example.com"
	name := "Bob"
	answer := "Chicken"

	answers := []*models.RSVPAnswer{
		{ID: 1, RSVPID: 1, QuestionID: 5, AnswerText: &answer},
	}

	err := service.SendConfirmationEmail(
		context.Background(), "tok",
		&models.RSVP{ID: 1, Response: models.RSVPResponseYes},
		&models.Invite{ID: 1, EventID: 1, Email: &email, Name: &name},
		&models.Event{ID: 1, Title: "Dinner", StartTime: startTime},
		answers,
	)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	td, ok := capturedData.(*RSVPConfirmationTemplateData)
	if !ok {
		t.Fatalf("Expected *RSVPConfirmationTemplateData, got %T", capturedData)
	}
	if len(td.Answers) != 0 {
		t.Errorf("Expected answers to be omitted when no question repo is configured, got %d answers", len(td.Answers))
	}
}

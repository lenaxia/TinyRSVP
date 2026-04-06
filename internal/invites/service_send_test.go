package invites

import (
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/models"
	"github.com/lenaxia/tinyrsvp/internal/templates"
	"github.com/lenaxia/tinyrsvp/internal/testutil"
	"github.com/lenaxia/tinyrsvp/pkg/token"
)

type mockSendInviteRepo struct {
	getByIDFunc func(ctx context.Context, id int64) (*models.Invite, error)
	updateFunc  func(ctx context.Context, invite *models.Invite) error
}

func (m *mockSendInviteRepo) GetByID(ctx context.Context, id int64) (*models.Invite, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, &models.NotFoundError{Resource: "invite"}
}

func (m *mockSendInviteRepo) Update(ctx context.Context, invite *models.Invite) error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, invite)
	}
	return nil
}

func (m *mockSendInviteRepo) Create(ctx context.Context, invite *models.Invite) error {
	return nil
}

func (m *mockSendInviteRepo) CreateBatch(ctx context.Context, invites []*models.Invite) error {
	return nil
}

func (m *mockSendInviteRepo) Delete(ctx context.Context, id int64) error {
	return nil
}

func (m *mockSendInviteRepo) GetByTokenHash(ctx context.Context, tokenHash string) (*models.Invite, error) {
	return nil, &models.NotFoundError{Resource: "invite"}
}

func (m *mockSendInviteRepo) ListByEventID(ctx context.Context, eventID int64, filters repositories.InviteFilters) ([]*models.Invite, error) {
	return nil, nil
}

func (m *mockSendInviteRepo) CountByEventID(ctx context.Context, eventID int64) (int, error) {
	return 0, nil
}

func (m *mockSendInviteRepo) CountByEventIDWithFilters(ctx context.Context, eventID int64, filters repositories.InviteFilters) (int, error) {
	return 0, nil
}

func (m *mockSendInviteRepo) GetStats(ctx context.Context, eventID int64) (*repositories.InviteStats, error) {
	return nil, nil
}

func (m *mockSendInviteRepo) FindDuplicateEmails(ctx context.Context, eventID int64, emails []string) ([]string, error) {
	return nil, nil
}

func (m *mockSendInviteRepo) DeleteExpired(ctx context.Context, before time.Time) (int64, error) {
	return 0, nil
}

func (m *mockSendInviteRepo) GetByEventIDs(ctx context.Context, eventIDs []int64) ([]*models.Invite, error) {
	return nil, nil
}

func (m *mockSendInviteRepo) CountInvites(ctx context.Context) (int, error) {
	return 0, nil
}

func (m *mockSendInviteRepo) UpdateExpiresAtByEventID(ctx context.Context, eventID int64, expiresAt time.Time) error {
	return nil
}

type mockEmailQueueRepo struct {
	createFunc func(ctx context.Context, email *models.EmailQueue) error
}

func (m *mockEmailQueueRepo) Create(ctx context.Context, email *models.EmailQueue) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, email)
	}
	return nil
}

func (m *mockEmailQueueRepo) GetByID(ctx context.Context, id int64) (*models.EmailQueue, error) {
	return nil, nil
}

func (m *mockEmailQueueRepo) GetPending(ctx context.Context, maxCount int) ([]*models.EmailQueue, error) {
	return nil, nil
}

func (m *mockEmailQueueRepo) GetByStatus(ctx context.Context, status models.EmailStatus, limit int) ([]*models.EmailQueue, error) {
	return nil, nil
}

func (m *mockEmailQueueRepo) GetByRecipient(ctx context.Context, email string, limit int) ([]*models.EmailQueue, error) {
	return nil, nil
}

func (m *mockEmailQueueRepo) UpdateStatus(ctx context.Context, id int64, status models.EmailStatus) error {
	return nil
}

func (m *mockEmailQueueRepo) IncrementAttempts(ctx context.Context, id int64, errorMsg string) error {
	return nil
}

func (m *mockEmailQueueRepo) MarkSending(ctx context.Context, id int64) error {
	return nil
}

func (m *mockEmailQueueRepo) MarkSent(ctx context.Context, id int64) error {
	return nil
}

func (m *mockEmailQueueRepo) MarkFailed(ctx context.Context, id int64, errorMsg string) error {
	return nil
}

func (m *mockEmailQueueRepo) Delete(ctx context.Context, id int64) error {
	return nil
}

func (m *mockEmailQueueRepo) DeleteOlderThan(ctx context.Context, before time.Time) (int64, error) {
	return 0, nil
}

func (m *mockEmailQueueRepo) MarkCancelled(ctx context.Context, id int64) error {
	return nil
}

func (m *mockEmailQueueRepo) Reschedule(ctx context.Context, id int64, scheduledFor time.Time) error {
	return nil
}

func (m *mockEmailQueueRepo) GetStats(ctx context.Context) (*repositories.EmailQueueStats, error) {
	return &repositories.EmailQueueStats{}, nil
}

func TestSendInvite_Success(t *testing.T) {
	now := time.Now()
	expiresAt := now.Add(30 * 24 * time.Hour)
	email := "test@example.com"
	existingToken := "existing-plain-token-abc123"

	invite := &models.Invite{
		ID:          1,
		EventID:     100,
		Email:       &email,
		Name:        testutil.StringPtr("Test User"),
		Token:       &existingToken,
		TokenHash:   "dGVzdF90b2tlbl9oYXNoXzEyMzQ1Njc4OTBhYmNkZWZnaGlqa2xtbm9wcXJzdHV2d3h5eg==",
		MaxPlusOnes: 2,
		Status:      models.InviteStatusDraft,
		ExpiresAt:   expiresAt,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	updateCalled := false
	var updatedInvite *models.Invite
	repo := &mockSendInviteRepo{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Invite, error) {
			return invite, nil
		},
		updateFunc: func(ctx context.Context, inv *models.Invite) error {
			updateCalled = true
			updatedInvite = inv
			return nil
		},
	}

	var queuedEmail *models.EmailQueue
	emailRepo := &mockEmailQueueRepo{
		createFunc: func(ctx context.Context, email *models.EmailQueue) error {
			queuedEmail = email
			return nil
		},
	}

	generator := token.NewGenerator([]byte("test-secret-key-32-bytes-long!!"))
	service := &inviteService{
		generator: generator,
		repo:      repo,
	}

	req := &SendInviteRequest{
		InviteID: 1,
		BaseURL:  "https://rsvp.example.com",
	}

	err := service.SendInvite(context.Background(), req, emailRepo)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Update should be called (to mark as sent), but token must not change
	if !updateCalled {
		t.Error("Expected repo.Update to be called to mark invite as sent")
	}
	if updatedInvite != nil && updatedInvite.Token != nil && *updatedInvite.Token != existingToken {
		t.Errorf("Token must not change on send: got %q, want %q", *updatedInvite.Token, existingToken)
	}

	if queuedEmail == nil {
		t.Fatal("Expected email to be queued")
	}

	expectedURL := "https://rsvp.example.com/rsvp/" + existingToken
	if !contains(queuedEmail.BodyText, expectedURL) {
		t.Errorf("Expected email body to contain %q, got: %s", expectedURL, queuedEmail.BodyText)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestSendInvite_NoEmail(t *testing.T) {
	now := time.Now()
	expiresAt := now.Add(30 * 24 * time.Hour)

	invite := &models.Invite{
		ID:          1,
		EventID:     100,
		Email:       nil,
		Name:        testutil.StringPtr("Test User"),
		TokenHash:   "dGVzdF90b2tlbl9oYXNoXzEyMzQ1Njc4OTBhYmNkZWZnaGlqa2xtbm9wcXJzdHV2d3h5eg==",
		MaxPlusOnes: 2,
		Status:      models.InviteStatusDraft,
		ExpiresAt:   expiresAt,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	repo := &mockSendInviteRepo{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Invite, error) {
			return invite, nil
		},
	}

	emailRepo := &mockEmailQueueRepo{}

	generator := token.NewGenerator([]byte("test-secret-key-32-bytes-long!!"))
	service := &inviteService{
		generator: generator,
		repo:      repo,
	}

	req := &SendInviteRequest{
		InviteID: 1,
		BaseURL:  "https://rsvp.example.com",
	}

	err := service.SendInvite(context.Background(), req, emailRepo)
	if err == nil {
		t.Error("Expected error for invite without email, got nil")
	}
}

func TestSendInvite_RevokedInvite(t *testing.T) {
	now := time.Now()
	expiresAt := now.Add(30 * 24 * time.Hour)
	email := "test@example.com"

	invite := &models.Invite{
		ID:          1,
		EventID:     100,
		Email:       &email,
		Name:        testutil.StringPtr("Test User"),
		TokenHash:   "dGVzdF90b2tlbl9oYXNoXzEyMzQ1Njc4OTBhYmNkZWZnaGlqa2xtbm9wcXJzdHV2d3h5eg==",
		MaxPlusOnes: 2,
		Status:      models.InviteStatusRevoked,
		ExpiresAt:   expiresAt,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	repo := &mockSendInviteRepo{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Invite, error) {
			return invite, nil
		},
	}

	emailRepo := &mockEmailQueueRepo{}

	generator := token.NewGenerator([]byte("test-secret-key-32-bytes-long!!"))
	service := &inviteService{
		generator: generator,
		repo:      repo,
	}

	req := &SendInviteRequest{
		InviteID: 1,
		BaseURL:  "https://rsvp.example.com",
	}

	err := service.SendInvite(context.Background(), req, emailRepo)
	if err == nil {
		t.Error("Expected error for revoked invite, got nil")
	}
}

func TestSendInvite_NilToken_ReturnsError(t *testing.T) {
	now := time.Now()
	expiresAt := now.Add(30 * 24 * time.Hour)
	email := "test@example.com"

	// Invite with no plain token stored (e.g. created via old code path)
	invite := &models.Invite{
		ID:          1,
		EventID:     100,
		Email:       &email,
		Name:        testutil.StringPtr("Test User"),
		Token:       nil,
		TokenHash:   "dGVzdF90b2tlbl9oYXNoXzEyMzQ1Njc4OTBhYmNkZWZnaGlqa2xtbm9wcXJzdHV2d3h5eg==",
		MaxPlusOnes: 2,
		Status:      models.InviteStatusDraft,
		ExpiresAt:   expiresAt,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	repo := &mockSendInviteRepo{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Invite, error) {
			return invite, nil
		},
	}

	emailRepo := &mockEmailQueueRepo{}

	generator := token.NewGenerator([]byte("test-secret-key-32-bytes-long!!"))
	service := &inviteService{
		generator: generator,
		repo:      repo,
	}

	req := &SendInviteRequest{
		InviteID: 1,
		BaseURL:  "https://rsvp.example.com",
	}

	err := service.SendInvite(context.Background(), req, emailRepo)
	if err == nil {
		t.Error("Expected error for invite with nil token, got nil")
	}
}

// --- template rendering tests ---

// stubTemplateService is a minimal stub of templates.Service used only in
// send tests. Only RenderEmailTemplate is implemented; all other methods
// return zero values or errors.
type stubTemplateService struct {
	renderEmailTemplateFn func(ctx context.Context, eventID int64, templateType string, data interface{}) (string, string, error)
}

func (s *stubTemplateService) RenderEmailTemplate(ctx context.Context, eventID int64, templateType models.TemplateType, data interface{}) (string, string, error) {
	if s.renderEmailTemplateFn != nil {
		return s.renderEmailTemplateFn(ctx, eventID, string(templateType), data)
	}
	return "", "", nil
}

// Unused interface methods — zero-value stubs.
func (s *stubTemplateService) CreateTemplate(ctx context.Context, t *models.Template) error {
	return nil
}
func (s *stubTemplateService) GetTemplate(ctx context.Context, id int64) (*models.Template, error) {
	return nil, nil
}
func (s *stubTemplateService) GetTemplateForEvent(ctx context.Context, eventID int64, templateType models.TemplateType) (*models.Template, error) {
	return nil, nil
}
func (s *stubTemplateService) GetDefaultTemplate(ctx context.Context, templateType models.TemplateType) (*models.Template, error) {
	return nil, nil
}
func (s *stubTemplateService) UpdateTemplate(ctx context.Context, t *models.Template) error {
	return nil
}
func (s *stubTemplateService) DeleteTemplate(ctx context.Context, id int64) error { return nil }
func (s *stubTemplateService) SetActive(ctx context.Context, id int64, active bool) error {
	return nil
}
func (s *stubTemplateService) SetDefault(ctx context.Context, id int64) error { return nil }
func (s *stubTemplateService) ListTemplates(ctx context.Context, filters *repositories.TemplateFilters) ([]*models.Template, error) {
	return nil, nil
}
func (s *stubTemplateService) PreviewTemplate(ctx context.Context, req *templates.PreviewRequest) (*templates.PreviewResponse, error) {
	return nil, nil
}
func (s *stubTemplateService) GetComponentRenderer() *templates.ComponentRenderer { return nil }
func (s *stubTemplateService) RenderRSVPPage(w io.Writer, event *models.Event, tmpl *models.Template) error {
	return nil
}

func TestSendInvite_WithTemplateService_RendersHTMLAndText(t *testing.T) {
	now := time.Now()
	expiresAt := now.Add(30 * 24 * time.Hour)
	emailAddr := "guest@example.com"
	plainToken := "test-token-abc123"

	invite := &models.Invite{
		ID:          1,
		EventID:     42,
		Email:       &emailAddr,
		Name:        testutil.StringPtr("Alice"),
		Token:       &plainToken,
		TokenHash:   "somehash",
		MaxPlusOnes: 1,
		Status:      models.InviteStatusDraft,
		ExpiresAt:   expiresAt,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	event := &models.Event{
		ID:    42,
		Title: "Alice's Birthday",
	}

	repo := &mockSendInviteRepo{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Invite, error) { return invite, nil },
		updateFunc:  func(ctx context.Context, inv *models.Invite) error { return nil },
	}

	var queuedEmail *models.EmailQueue
	emailRepo := &mockEmailQueueRepo{
		createFunc: func(ctx context.Context, e *models.EmailQueue) error {
			queuedEmail = e
			return nil
		},
	}

	renderCalled := false
	var capturedEventID int64
	var capturedType string

	tmplSvc := &stubTemplateService{
		renderEmailTemplateFn: func(ctx context.Context, eventID int64, templateType string, data interface{}) (string, string, error) {
			renderCalled = true
			capturedEventID = eventID
			capturedType = templateType
			return "<html>You're invited to Alice's Birthday!</html>", "You're invited to Alice's Birthday!", nil
		},
	}

	generator := token.NewGenerator([]byte("test-secret-key-32-bytes-long!!"))
	service := &inviteService{
		generator:       generator,
		repo:            repo,
		templateService: tmplSvc,
	}

	req := &SendInviteRequest{
		InviteID: 1,
		BaseURL:  "https://rsvp.example.com",
		Event:    event,
	}

	if err := service.SendInvite(context.Background(), req, emailRepo); err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !renderCalled {
		t.Error("Expected template service RenderEmailTemplate to be called")
	}
	if capturedEventID != 42 {
		t.Errorf("Expected eventID 42, got %d", capturedEventID)
	}
	if capturedType != string(models.TemplateTypeInviteEmail) {
		t.Errorf("Expected template type %q, got %q", models.TemplateTypeInviteEmail, capturedType)
	}

	if queuedEmail == nil {
		t.Fatal("Expected email to be queued")
	}
	if queuedEmail.BodyHTML == nil || *queuedEmail.BodyHTML == "" {
		t.Error("Expected BodyHTML to be populated from template")
	}
	if queuedEmail.BodyText == "" {
		t.Error("Expected BodyText to be populated from template")
	}
	// The stub returns a fixed string — verify the service used it as-is
	// (the template content itself is tested in the templates package).
	if queuedEmail.BodyText != "You're invited to Alice's Birthday!" {
		t.Errorf("Expected BodyText from stub renderer, got: %s", queuedEmail.BodyText)
	}
	expectedSubject := "You're Invited: Alice's Birthday"
	if queuedEmail.Subject != expectedSubject {
		t.Errorf("Expected Subject %q, got %q", expectedSubject, queuedEmail.Subject)
	}
}

func TestSendInvite_WithTemplateService_FallsBackToPlaintextOnRenderError(t *testing.T) {
	now := time.Now()
	expiresAt := now.Add(30 * 24 * time.Hour)
	emailAddr := "guest@example.com"
	plainToken := "test-token-abc123"

	invite := &models.Invite{
		ID:          1,
		EventID:     42,
		Email:       &emailAddr,
		Name:        testutil.StringPtr("Bob"),
		Token:       &plainToken,
		TokenHash:   "somehash",
		MaxPlusOnes: 0,
		Status:      models.InviteStatusDraft,
		ExpiresAt:   expiresAt,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	event := &models.Event{ID: 42, Title: "Bob's Party"}

	repo := &mockSendInviteRepo{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Invite, error) { return invite, nil },
		updateFunc:  func(ctx context.Context, inv *models.Invite) error { return nil },
	}

	var queuedEmail *models.EmailQueue
	emailRepo := &mockEmailQueueRepo{
		createFunc: func(ctx context.Context, e *models.EmailQueue) error {
			queuedEmail = e
			return nil
		},
	}

	tmplSvc := &stubTemplateService{
		renderEmailTemplateFn: func(ctx context.Context, eventID int64, templateType string, data interface{}) (string, string, error) {
			return "", "", fmt.Errorf("template not found")
		},
	}

	generator := token.NewGenerator([]byte("test-secret-key-32-bytes-long!!"))
	service := &inviteService{
		generator:       generator,
		repo:            repo,
		templateService: tmplSvc,
	}

	req := &SendInviteRequest{
		InviteID: 1,
		BaseURL:  "https://rsvp.example.com",
		Event:    event,
	}

	if err := service.SendInvite(context.Background(), req, emailRepo); err != nil {
		t.Fatalf("Expected no error on render failure (should fall back), got %v", err)
	}

	if queuedEmail == nil {
		t.Fatal("Expected email to be queued even when template render fails")
	}
	if queuedEmail.BodyText == "" {
		t.Error("Expected plaintext fallback body to be non-empty")
	}
	expectedURL := "https://rsvp.example.com/rsvp/" + plainToken
	if !contains(queuedEmail.BodyText, expectedURL) {
		t.Errorf("Expected fallback body to contain RSVP URL %q, got: %s", expectedURL, queuedEmail.BodyText)
	}
}

func TestSendInvite_WithoutTemplateService_UsesPlaintext(t *testing.T) {
	now := time.Now()
	expiresAt := now.Add(30 * 24 * time.Hour)
	emailAddr := "guest@example.com"
	plainToken := "test-token-abc123"

	invite := &models.Invite{
		ID:          2,
		EventID:     10,
		Email:       &emailAddr,
		Name:        testutil.StringPtr("Carol"),
		Token:       &plainToken,
		TokenHash:   "somehash",
		MaxPlusOnes: 0,
		Status:      models.InviteStatusDraft,
		ExpiresAt:   expiresAt,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	repo := &mockSendInviteRepo{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Invite, error) { return invite, nil },
		updateFunc:  func(ctx context.Context, inv *models.Invite) error { return nil },
	}

	var queuedEmail *models.EmailQueue
	emailRepo := &mockEmailQueueRepo{
		createFunc: func(ctx context.Context, e *models.EmailQueue) error {
			queuedEmail = e
			return nil
		},
	}

	generator := token.NewGenerator([]byte("test-secret-key-32-bytes-long!!"))
	// No templateService — uses NewInviteService constructor path.
	service := &inviteService{generator: generator, repo: repo}

	req := &SendInviteRequest{InviteID: 2, BaseURL: "https://rsvp.example.com"}

	if err := service.SendInvite(context.Background(), req, emailRepo); err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if queuedEmail == nil {
		t.Fatal("Expected email to be queued")
	}
	if queuedEmail.BodyHTML != nil {
		t.Error("Expected BodyHTML to be nil when no template service is configured")
	}
	expectedURL := "https://rsvp.example.com/rsvp/" + plainToken
	if !contains(queuedEmail.BodyText, expectedURL) {
		t.Errorf("Expected plaintext body to contain RSVP URL %q, got: %s", expectedURL, queuedEmail.BodyText)
	}
}

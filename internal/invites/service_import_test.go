package invites

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

type mockInviteRepo struct {
	invites         []*models.Invite
	createBatchFunc func(ctx context.Context, invites []*models.Invite) error
	findDuplicates  func(ctx context.Context, eventID int64, emails []string) ([]string, error)
}

func (m *mockInviteRepo) Create(ctx context.Context, invite *models.Invite) error {
	m.invites = append(m.invites, invite)
	return nil
}

func (m *mockInviteRepo) CreateBatch(ctx context.Context, invites []*models.Invite) error {
	if m.createBatchFunc != nil {
		return m.createBatchFunc(ctx, invites)
	}
	m.invites = append(m.invites, invites...)
	return nil
}

func (m *mockInviteRepo) GetByID(ctx context.Context, id int64) (*models.Invite, error) {
	return nil, nil
}

func (m *mockInviteRepo) GetByTokenHash(ctx context.Context, tokenHash string) (*models.Invite, error) {
	return nil, nil
}

func (m *mockInviteRepo) Update(ctx context.Context, invite *models.Invite) error {
	return nil
}

func (m *mockInviteRepo) Delete(ctx context.Context, id int64) error {
	return nil
}

func (m *mockInviteRepo) ListByEventID(ctx context.Context, eventID int64, filters repositories.InviteFilters) ([]*models.Invite, error) {
	return nil, nil
}

func (m *mockInviteRepo) CountByEventID(ctx context.Context, eventID int64) (int, error) {
	return 0, nil
}

func (m *mockInviteRepo) GetStats(ctx context.Context, eventID int64) (*repositories.InviteStats, error) {
	return nil, nil
}

func (m *mockInviteRepo) FindDuplicateEmails(ctx context.Context, eventID int64, emails []string) ([]string, error) {
	if m.findDuplicates != nil {
		return m.findDuplicates(ctx, eventID, emails)
	}
	return []string{}, nil
}

func (m *mockInviteRepo) DeleteExpired(ctx context.Context, before time.Time) (int64, error) {
	return 0, nil
}

func TestImportCSV_Success(t *testing.T) {
	csvData := `email,name,max_plus_ones
john@example.com,John Doe,2
jane@example.com,Jane Smith,1
bob@example.com,Bob Johnson,0`

	repo := &mockInviteRepo{}
	generator := &mockGenerator{}
	service := NewInviteService(generator, repo).(*inviteService)

	expiresAt := time.Now().Add(30 * 24 * time.Hour)
	result, err := service.ImportCSV(context.Background(), 1, []byte(csvData), 5, expiresAt)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.Total != 3 {
		t.Errorf("Expected total 3, got %d", result.Total)
	}

	if result.Created != 3 {
		t.Errorf("Expected created 3, got %d", result.Created)
	}

	if result.Failed != 0 {
		t.Errorf("Expected failed 0, got %d", result.Failed)
	}

	if result.Duplicates != 0 {
		t.Errorf("Expected duplicates 0, got %d", result.Duplicates)
	}

	if len(result.Errors) != 0 {
		t.Errorf("Expected 0 errors, got %d", len(result.Errors))
	}

	if len(repo.invites) != 3 {
		t.Errorf("Expected 3 invites created, got %d", len(repo.invites))
	}
}

func TestImportCSV_DuplicatesWithinCSV(t *testing.T) {
	csvData := `email
john@example.com
jane@example.com
john@example.com`

	repo := &mockInviteRepo{}
	generator := &mockGenerator{}
	service := NewInviteService(generator, repo).(*inviteService)

	expiresAt := time.Now().Add(30 * 24 * time.Hour)
	result, err := service.ImportCSV(context.Background(), 1, []byte(csvData), 5, expiresAt)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.Total != 3 {
		t.Errorf("Expected total 3, got %d", result.Total)
	}

	if result.Created != 2 {
		t.Errorf("Expected created 2, got %d", result.Created)
	}

	if result.Duplicates != 1 {
		t.Errorf("Expected duplicates 1, got %d", result.Duplicates)
	}

	if len(result.Errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(result.Errors))
	}

	if len(repo.invites) != 2 {
		t.Errorf("Expected 2 invites created, got %d", len(repo.invites))
	}
}

func TestImportCSV_DuplicatesInDatabase(t *testing.T) {
	csvData := `email
john@example.com
jane@example.com
bob@example.com`

	repo := &mockInviteRepo{
		findDuplicates: func(ctx context.Context, eventID int64, emails []string) ([]string, error) {
			return []string{"jane@example.com"}, nil
		},
	}
	generator := &mockGenerator{}
	service := NewInviteService(generator, repo).(*inviteService)

	expiresAt := time.Now().Add(30 * 24 * time.Hour)
	result, err := service.ImportCSV(context.Background(), 1, []byte(csvData), 5, expiresAt)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.Total != 3 {
		t.Errorf("Expected total 3, got %d", result.Total)
	}

	if result.Created != 2 {
		t.Errorf("Expected created 2, got %d", result.Created)
	}

	if result.Duplicates != 1 {
		t.Errorf("Expected duplicates 1, got %d", result.Duplicates)
	}

	if len(result.Errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(result.Errors))
	}

	if len(repo.invites) != 2 {
		t.Errorf("Expected 2 invites created, got %d", len(repo.invites))
	}
}

func TestImportCSV_InvalidEmails(t *testing.T) {
	csvData := `email,name
john@example.com,John Doe
invalid-email,Jane Smith
bob@example.com,Bob Johnson`

	repo := &mockInviteRepo{}
	generator := &mockGenerator{}
	service := NewInviteService(generator, repo).(*inviteService)

	expiresAt := time.Now().Add(30 * 24 * time.Hour)
	result, err := service.ImportCSV(context.Background(), 1, []byte(csvData), 5, expiresAt)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.Total != 3 {
		t.Errorf("Expected total 3, got %d", result.Total)
	}

	if result.Created != 2 {
		t.Errorf("Expected created 2, got %d", result.Created)
	}

	if result.Failed != 1 {
		t.Errorf("Expected failed 1, got %d", result.Failed)
	}

	if len(result.Errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(result.Errors))
	}

	if result.Errors[0].Row != 3 {
		t.Errorf("Expected error on row 3, got row %d", result.Errors[0].Row)
	}

	if len(repo.invites) != 2 {
		t.Errorf("Expected 2 invites created, got %d", len(repo.invites))
	}
}

func TestImportCSV_PartialSuccess(t *testing.T) {
	csvData := `email,name
john@example.com,John Doe
invalid-email,Invalid User
jane@example.com,Jane Smith
john@example.com,Duplicate`

	repo := &mockInviteRepo{}
	generator := &mockGenerator{}
	service := NewInviteService(generator, repo).(*inviteService)

	expiresAt := time.Now().Add(30 * 24 * time.Hour)
	result, err := service.ImportCSV(context.Background(), 1, []byte(csvData), 5, expiresAt)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.Total != 4 {
		t.Errorf("Expected total 4, got %d", result.Total)
	}

	if result.Created != 2 {
		t.Errorf("Expected created 2, got %d", result.Created)
	}

	if result.Failed != 1 {
		t.Errorf("Expected failed 1, got %d", result.Failed)
	}

	if result.Duplicates != 1 {
		t.Errorf("Expected duplicates 1, got %d", result.Duplicates)
	}

	if len(result.Errors) != 2 {
		t.Errorf("Expected 2 errors, got %d", len(result.Errors))
	}

	if len(repo.invites) != 2 {
		t.Errorf("Expected 2 invites created, got %d", len(repo.invites))
	}
}

func TestImportCSV_EmptyCSV(t *testing.T) {
	csvData := ``

	repo := &mockInviteRepo{}
	generator := &mockGenerator{}
	service := NewInviteService(generator, repo).(*inviteService)

	expiresAt := time.Now().Add(30 * 24 * time.Hour)
	_, err := service.ImportCSV(context.Background(), 1, []byte(csvData), 5, expiresAt)

	if err == nil {
		t.Fatal("Expected error for empty CSV, got nil")
	}
}

func TestImportCSV_MissingEmailColumn(t *testing.T) {
	csvData := `name,max_plus_ones
John Doe,2
Jane Smith,1`

	repo := &mockInviteRepo{}
	generator := &mockGenerator{}
	service := NewInviteService(generator, repo).(*inviteService)

	expiresAt := time.Now().Add(30 * 24 * time.Hour)
	_, err := service.ImportCSV(context.Background(), 1, []byte(csvData), 5, expiresAt)

	if err == nil {
		t.Fatal("Expected error for missing email column, got nil")
	}

	if !strings.Contains(err.Error(), "email") {
		t.Errorf("Expected error message to mention 'email', got '%s'", err.Error())
	}
}

func TestImportCSV_ExceedsRowLimit(t *testing.T) {
	var sb strings.Builder
	sb.WriteString("email\n")
	for i := 0; i < 501; i++ {
		sb.WriteString("user")
		sb.WriteString(strings.Repeat("0", 3))
		sb.WriteString("@example.com\n")
	}

	repo := &mockInviteRepo{}
	generator := &mockGenerator{}
	service := NewInviteService(generator, repo).(*inviteService)

	expiresAt := time.Now().Add(30 * 24 * time.Hour)
	_, err := service.ImportCSV(context.Background(), 1, []byte(sb.String()), 5, expiresAt)

	if err == nil {
		t.Fatal("Expected error for exceeding row limit, got nil")
	}

	if !strings.Contains(err.Error(), "500") {
		t.Errorf("Expected error message to mention row limit, got '%s'", err.Error())
	}
}

func TestImportCSV_DefaultMaxPlusOnes(t *testing.T) {
	csvData := `email
john@example.com
jane@example.com`

	repo := &mockInviteRepo{}
	generator := &mockGenerator{}
	service := NewInviteService(generator, repo).(*inviteService)

	expiresAt := time.Now().Add(30 * 24 * time.Hour)
	result, err := service.ImportCSV(context.Background(), 1, []byte(csvData), 3, expiresAt)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.Created != 2 {
		t.Errorf("Expected created 2, got %d", result.Created)
	}

	if len(repo.invites) != 2 {
		t.Errorf("Expected 2 invites created, got %d", len(repo.invites))
	}

	if repo.invites[0].MaxPlusOnes != 3 {
		t.Errorf("Expected max_plus_ones 3, got %d", repo.invites[0].MaxPlusOnes)
	}
}

func TestImportCSV_CaseInsensitiveDuplicates(t *testing.T) {
	csvData := `email
john@example.com
JOHN@EXAMPLE.COM
Jane@Example.Com`

	repo := &mockInviteRepo{}
	generator := &mockGenerator{}
	service := NewInviteService(generator, repo).(*inviteService)

	expiresAt := time.Now().Add(30 * 24 * time.Hour)
	result, err := service.ImportCSV(context.Background(), 1, []byte(csvData), 5, expiresAt)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.Total != 3 {
		t.Errorf("Expected total 3, got %d", result.Total)
	}

	if result.Created != 2 {
		t.Errorf("Expected created 2 (case-insensitive duplicates), got %d", result.Created)
	}

	if result.Duplicates != 1 {
		t.Errorf("Expected duplicates 1, got %d", result.Duplicates)
	}
}

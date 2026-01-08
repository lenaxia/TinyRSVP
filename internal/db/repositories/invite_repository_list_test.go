package repositories

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/db"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

func TestInviteRepository_ListByEventID_WithSearch(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	repo := NewInviteRepository(database)
	ctx := context.Background()

	eventID := createTestEvent(t, database)

	invites := []*models.Invite{
		createTestInviteWithDetails(t, repo, eventID, "john@example.com", "John Doe"),
		createTestInviteWithDetails(t, repo, eventID, "jane@example.com", "Jane Smith"),
		createTestInviteWithDetails(t, repo, eventID, "bob@test.com", "Bob Johnson"),
	}

	tests := []struct {
		name          string
		search        string
		expectedCount int
		expectedIDs   []int64
	}{
		{
			name:          "search by email substring",
			search:        "example",
			expectedCount: 2,
			expectedIDs:   []int64{invites[0].ID, invites[1].ID},
		},
		{
			name:          "search by name substring",
			search:        "John",
			expectedCount: 2,
			expectedIDs:   []int64{invites[0].ID, invites[2].ID},
		},
		{
			name:          "search case insensitive",
			search:        "JANE",
			expectedCount: 1,
			expectedIDs:   []int64{invites[1].ID},
		},
		{
			name:          "search no match",
			search:        "nonexistent",
			expectedCount: 0,
			expectedIDs:   []int64{},
		},
		{
			name:          "empty search returns all",
			search:        "",
			expectedCount: 3,
			expectedIDs:   []int64{invites[0].ID, invites[1].ID, invites[2].ID},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filters := InviteFilters{
				Search: &tt.search,
			}

			results, err := repo.ListByEventID(ctx, eventID, filters)
			if err != nil {
				t.Fatalf("ListByEventID failed: %v", err)
			}

			if len(results) != tt.expectedCount {
				t.Errorf("expected %d results, got %d", tt.expectedCount, len(results))
			}

			resultIDs := make(map[int64]bool)
			for _, invite := range results {
				resultIDs[invite.ID] = true
			}

			for _, expectedID := range tt.expectedIDs {
				if !resultIDs[expectedID] {
					t.Errorf("expected invite ID %d not found in results", expectedID)
				}
			}
		})
	}
}

func TestInviteRepository_ListByEventID_WithSorting(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	repo := NewInviteRepository(database)
	ctx := context.Background()

	eventID := createTestEvent(t, database)

	now := time.Now()
	invite1 := createTestInviteWithDetails(t, repo, eventID, "a@example.com", "Alice")
	time.Sleep(10 * time.Millisecond)
	invite2 := createTestInviteWithDetails(t, repo, eventID, "b@example.com", "Bob")
	time.Sleep(10 * time.Millisecond)
	invite3 := createTestInviteWithDetails(t, repo, eventID, "c@example.com", "Charlie")

	sentAt := now.Add(1 * time.Hour)
	invite2.SentAt = &sentAt
	invite2.Status = models.InviteStatusSent
	if err := repo.Update(ctx, invite2); err != nil {
		t.Fatalf("failed to update invite2: %v", err)
	}

	viewedAt := now.Add(2 * time.Hour)
	invite3.ViewedAt = &viewedAt
	invite3.Status = models.InviteStatusViewed
	if err := repo.Update(ctx, invite3); err != nil {
		t.Fatalf("failed to update invite3: %v", err)
	}

	tests := []struct {
		name        string
		sortBy      string
		sortOrder   string
		expectedIDs []int64
	}{
		{
			name:        "sort by created_at desc",
			sortBy:      "created_at",
			sortOrder:   "desc",
			expectedIDs: []int64{invite3.ID, invite2.ID, invite1.ID},
		},
		{
			name:        "sort by created_at asc",
			sortBy:      "created_at",
			sortOrder:   "asc",
			expectedIDs: []int64{invite1.ID, invite2.ID, invite3.ID},
		},
		{
			name:        "sort by email asc",
			sortBy:      "email",
			sortOrder:   "asc",
			expectedIDs: []int64{invite1.ID, invite2.ID, invite3.ID},
		},
		{
			name:        "sort by email desc",
			sortBy:      "email",
			sortOrder:   "desc",
			expectedIDs: []int64{invite3.ID, invite2.ID, invite1.ID},
		},
		{
			name:        "sort by name asc",
			sortBy:      "name",
			sortOrder:   "asc",
			expectedIDs: []int64{invite1.ID, invite2.ID, invite3.ID},
		},
		{
			name:        "sort by status asc",
			sortBy:      "status",
			sortOrder:   "asc",
			expectedIDs: []int64{invite1.ID, invite2.ID, invite3.ID},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filters := InviteFilters{
				SortBy:    &tt.sortBy,
				SortOrder: &tt.sortOrder,
			}

			results, err := repo.ListByEventID(ctx, eventID, filters)
			if err != nil {
				t.Fatalf("ListByEventID failed: %v", err)
			}

			if len(results) != 3 {
				t.Fatalf("expected 3 results, got %d", len(results))
			}

			for i, expectedID := range tt.expectedIDs {
				if results[i].ID != expectedID {
					t.Errorf("position %d: expected ID %d, got %d", i, expectedID, results[i].ID)
				}
			}
		})
	}
}

func TestInviteRepository_ListByEventID_WithPagination(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	repo := NewInviteRepository(database)
	ctx := context.Background()

	eventID := createTestEvent(t, database)

	for i := 0; i < 10; i++ {
		email := "user" + string(rune('0'+i)) + "@example.com"
		createTestInviteWithDetails(t, repo, eventID, email, "User")
	}

	tests := []struct {
		name          string
		limit         int
		offset        int
		expectedCount int
	}{
		{
			name:          "first page",
			limit:         3,
			offset:        0,
			expectedCount: 3,
		},
		{
			name:          "second page",
			limit:         3,
			offset:        3,
			expectedCount: 3,
		},
		{
			name:          "last page partial",
			limit:         3,
			offset:        9,
			expectedCount: 1,
		},
		{
			name:          "offset beyond results",
			limit:         3,
			offset:        20,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filters := InviteFilters{
				Limit:  tt.limit,
				Offset: tt.offset,
			}

			results, err := repo.ListByEventID(ctx, eventID, filters)
			if err != nil {
				t.Fatalf("ListByEventID failed: %v", err)
			}

			if len(results) != tt.expectedCount {
				t.Errorf("expected %d results, got %d", tt.expectedCount, len(results))
			}
		})
	}
}

func TestInviteRepository_ListByEventID_CombinedFilters(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	repo := NewInviteRepository(database)
	ctx := context.Background()

	eventID := createTestEvent(t, database)

	for i := 0; i < 5; i++ {
		email := "user" + string(rune('0'+i)) + "@example.com"
		name := "User " + string(rune('A'+i))
		invite := createTestInviteWithDetails(t, repo, eventID, email, name)

		if i%2 == 0 {
			now := time.Now()
			invite.SentAt = &now
			invite.Status = models.InviteStatusSent
			if err := repo.Update(ctx, invite); err != nil {
				t.Fatalf("failed to update invite: %v", err)
			}
		}
	}

	search := "user"
	status := models.InviteStatusSent
	sortBy := "email"
	sortOrder := "asc"

	filters := InviteFilters{
		Status:    &status,
		Search:    &search,
		SortBy:    &sortBy,
		SortOrder: &sortOrder,
		Limit:     2,
		Offset:    0,
	}

	results, err := repo.ListByEventID(ctx, eventID, filters)
	if err != nil {
		t.Fatalf("ListByEventID failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}

	for _, invite := range results {
		if invite.Status != models.InviteStatusSent {
			t.Errorf("expected status sent, got %s", invite.Status)
		}
	}
}

func createTestInviteWithDetails(t *testing.T, repo InviteRepository, eventID int64, email, name string) *models.Invite {
	t.Helper()

	tokenHash := strings.Repeat("a", 43)
	if len(email) > 0 {
		for i, c := range email {
			if i < 43 {
				tokenHash = tokenHash[:i] + string(c) + tokenHash[i+1:]
			}
		}
	}

	invite := &models.Invite{
		EventID:     eventID,
		Email:       &email,
		Name:        &name,
		TokenHash:   tokenHash,
		MaxPlusOnes: 2,
		Status:      models.InviteStatusDraft,
		ExpiresAt:   time.Now().Add(30 * 24 * time.Hour),
	}

	if err := repo.Create(context.Background(), invite); err != nil {
		t.Fatalf("failed to create test invite: %v", err)
	}

	return invite
}

func createTestEvent(t *testing.T, database db.Database) int64 {
	t.Helper()

	ctx := context.Background()
	
	_, err := database.Exec(ctx, `
		INSERT INTO users (id, email, name, role, created_at, updated_at)
		VALUES (1, 'test@example.com', 'Test User', 'admin', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT(id) DO NOTHING
	`)
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	query := `
		INSERT INTO events (
			title, description, location, start_time, end_time, timezone,
			max_plus_ones, status, created_by, version, ics_sequence, created_at, updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	startTime := now.Add(30 * 24 * time.Hour)
	endTime := startTime.Add(4 * time.Hour)

	result, err := database.Exec(ctx, query,
		"Test Event",
		"Test Description",
		"Test Location",
		startTime,
		endTime,
		"America/Los_Angeles",
		2,
		"published",
		1,
		1,
		0,
		now,
		now,
	)

	if err != nil {
		t.Fatalf("failed to create test event: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("failed to get event ID: %v", err)
	}

	return id
}

func TestInviteRepository_ListByEventID_SortBySentAtWithNulls(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	repo := NewInviteRepository(database)
	ctx := context.Background()

	eventID := createTestEvent(t, database)

	now := time.Now()
	invite1 := createTestInviteWithDetails(t, repo, eventID, "a@example.com", "Alice")
	invite2 := createTestInviteWithDetails(t, repo, eventID, "b@example.com", "Bob")
	_ = createTestInviteWithDetails(t, repo, eventID, "c@example.com", "Charlie")

	sentAt1 := now.Add(-2 * time.Hour)
	invite1.SentAt = &sentAt1
	invite1.Status = models.InviteStatusSent
	if err := repo.Update(ctx, invite1); err != nil {
		t.Fatalf("failed to update invite1: %v", err)
	}

	sentAt2 := now.Add(-1 * time.Hour)
	invite2.SentAt = &sentAt2
	invite2.Status = models.InviteStatusSent
	if err := repo.Update(ctx, invite2); err != nil {
		t.Fatalf("failed to update invite2: %v", err)
	}

	tests := []struct {
		name      string
		sortOrder string
	}{
		{
			name:      "sort by sent_at asc with nulls",
			sortOrder: "asc",
		},
		{
			name:      "sort by sent_at desc with nulls",
			sortOrder: "desc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sortBy := "sent_at"
			filters := InviteFilters{
				SortBy:    &sortBy,
				SortOrder: &tt.sortOrder,
			}

			results, err := repo.ListByEventID(ctx, eventID, filters)
			if err != nil {
				t.Fatalf("ListByEventID failed: %v", err)
			}

			if len(results) != 3 {
				t.Errorf("expected 3 results, got %d", len(results))
			}

			if strings.ToUpper(tt.sortOrder) == "ASC" {
				if results[0].SentAt != nil && results[1].SentAt != nil {
					if results[0].SentAt.After(*results[1].SentAt) {
						t.Error("sent_at not sorted ascending")
					}
				}
			} else {
				if results[0].SentAt != nil && results[1].SentAt != nil {
					if results[0].SentAt.Before(*results[1].SentAt) {
						t.Error("sent_at not sorted descending")
					}
				}
			}
		})
	}
}

func TestInviteRepository_ListByEventID_SortByViewedAtWithNulls(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	repo := NewInviteRepository(database)
	ctx := context.Background()

	eventID := createTestEvent(t, database)

	now := time.Now()
	invite1 := createTestInviteWithDetails(t, repo, eventID, "a@example.com", "Alice")
	invite2 := createTestInviteWithDetails(t, repo, eventID, "b@example.com", "Bob")
	_ = createTestInviteWithDetails(t, repo, eventID, "c@example.com", "Charlie")

	viewedAt1 := now.Add(-2 * time.Hour)
	invite1.ViewedAt = &viewedAt1
	invite1.Status = models.InviteStatusViewed
	if err := repo.Update(ctx, invite1); err != nil {
		t.Fatalf("failed to update invite1: %v", err)
	}

	viewedAt2 := now.Add(-1 * time.Hour)
	invite2.ViewedAt = &viewedAt2
	invite2.Status = models.InviteStatusViewed
	if err := repo.Update(ctx, invite2); err != nil {
		t.Fatalf("failed to update invite2: %v", err)
	}

	tests := []struct {
		name      string
		sortOrder string
	}{
		{
			name:      "sort by viewed_at asc with nulls",
			sortOrder: "asc",
		},
		{
			name:      "sort by viewed_at desc with nulls",
			sortOrder: "desc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sortBy := "viewed_at"
			filters := InviteFilters{
				SortBy:    &sortBy,
				SortOrder: &tt.sortOrder,
			}

			results, err := repo.ListByEventID(ctx, eventID, filters)
			if err != nil {
				t.Fatalf("ListByEventID failed: %v", err)
			}

			if len(results) != 3 {
				t.Errorf("expected 3 results, got %d", len(results))
			}

			if strings.ToUpper(tt.sortOrder) == "ASC" {
				if results[0].ViewedAt != nil && results[1].ViewedAt != nil {
					if results[0].ViewedAt.After(*results[1].ViewedAt) {
						t.Error("viewed_at not sorted ascending")
					}
				}
			} else {
				if results[0].ViewedAt != nil && results[1].ViewedAt != nil {
					if results[0].ViewedAt.Before(*results[1].ViewedAt) {
						t.Error("viewed_at not sorted descending")
					}
				}
			}
		})
	}
}

func TestInviteRepository_GetStats_Accuracy(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	repo := NewInviteRepository(database)
	ctx := context.Background()

	eventID := createTestEvent(t, database)

	statuses := []models.InviteStatus{
		models.InviteStatusDraft,
		models.InviteStatusDraft,
		models.InviteStatusSent,
		models.InviteStatusSent,
		models.InviteStatusSent,
		models.InviteStatusViewed,
		models.InviteStatusResponded,
		models.InviteStatusRevoked,
	}

	for i, status := range statuses {
		email := "user" + string(rune('0'+i)) + "@example.com"
		invite := createTestInviteWithDetails(t, repo, eventID, email, "User")
		invite.Status = status
		if err := repo.Update(ctx, invite); err != nil {
			t.Fatalf("failed to update invite: %v", err)
		}
	}

	stats, err := repo.GetStats(ctx, eventID)
	if err != nil {
		t.Fatalf("GetStats failed: %v", err)
	}

	if stats.Total != 8 {
		t.Errorf("Total = %d, want 8", stats.Total)
	}
	if stats.Draft != 2 {
		t.Errorf("Draft = %d, want 2", stats.Draft)
	}
	if stats.Sent != 3 {
		t.Errorf("Sent = %d, want 3", stats.Sent)
	}
	if stats.Viewed != 1 {
		t.Errorf("Viewed = %d, want 1", stats.Viewed)
	}
	if stats.Responded != 1 {
		t.Errorf("Responded = %d, want 1", stats.Responded)
	}
	if stats.Revoked != 1 {
		t.Errorf("Revoked = %d, want 1", stats.Revoked)
	}
}

func TestInviteRepository_ListByEventID_EmptyResults(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	repo := NewInviteRepository(database)
	ctx := context.Background()

	eventID := createTestEvent(t, database)

	filters := InviteFilters{
		Limit: 50,
	}

	results, err := repo.ListByEventID(ctx, eventID, filters)
	if err != nil {
		t.Fatalf("ListByEventID failed: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("expected 0 results for event with no invites, got %d", len(results))
	}
}

func TestInviteRepository_CountByEventIDWithFilters(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	repo := NewInviteRepository(database)
	ctx := context.Background()

	eventID := createTestEvent(t, database)

	for i := 0; i < 10; i++ {
		email := "user" + string(rune('0'+i)) + "@example.com"
		invite := createTestInviteWithDetails(t, repo, eventID, email, "User")

		if i < 5 {
			invite.Status = models.InviteStatusSent
			if err := repo.Update(ctx, invite); err != nil {
				t.Fatalf("failed to update invite: %v", err)
			}
		}
	}

	tests := []struct {
		name          string
		filters       InviteFilters
		expectedCount int
	}{
		{
			name: "count all",
			filters: InviteFilters{
				Limit: 100,
			},
			expectedCount: 10,
		},
		{
			name: "count with status filter",
			filters: InviteFilters{
				Status: func() *models.InviteStatus { s := models.InviteStatusSent; return &s }(),
				Limit:  100,
			},
			expectedCount: 5,
		},
		{
			name: "count with search filter",
			filters: InviteFilters{
				Search: func() *string { s := "user0"; return &s }(),
				Limit:  100,
			},
			expectedCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count, err := repo.CountByEventIDWithFilters(ctx, eventID, tt.filters)
			if err != nil {
				t.Fatalf("CountByEventIDWithFilters failed: %v", err)
			}

			if count != tt.expectedCount {
				t.Errorf("expected count %d, got %d", tt.expectedCount, count)
			}
		})
	}
}

func TestInviteRepository_ListByEventID_SQLInjectionDefense(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	repo := NewInviteRepository(database)
	ctx := context.Background()

	eventID := createTestEvent(t, database)
	_ = createTestInviteWithDetails(t, repo, eventID, "test@example.com", "Test User")

	tests := []struct {
		name      string
		sortBy    string
		sortOrder string
	}{
		{
			name:      "SQL injection in sort field",
			sortBy:    "email; DROP TABLE invites; --",
			sortOrder: "asc",
		},
		{
			name:      "SQL injection in sort order",
			sortBy:    "email",
			sortOrder: "asc; DROP TABLE invites; --",
		},
		{
			name:      "SQL injection with UNION",
			sortBy:    "email UNION SELECT * FROM users",
			sortOrder: "asc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filters := InviteFilters{
				SortBy:    &tt.sortBy,
				SortOrder: &tt.sortOrder,
			}

			results, err := repo.ListByEventID(ctx, eventID, filters)
			if err != nil {
				t.Logf("Query rejected (expected): %v", err)
			}

			if results != nil && len(results) > 0 {
				if results[0].Email != nil && *results[0].Email == "test@example.com" {
					t.Log("Query executed safely with invalid sort field")
				}
			}

			var count int
			err = database.QueryRow(ctx, "SELECT COUNT(*) FROM invites WHERE event_id = ?", eventID).Scan(&count)
			if err != nil {
				t.Fatalf("failed to verify table integrity: %v", err)
			}
			if count != 1 {
				t.Errorf("table was modified, expected 1 invite, got %d", count)
			}
		})
	}
}

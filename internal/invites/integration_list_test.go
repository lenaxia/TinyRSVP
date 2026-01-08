package invites

import (
	"context"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/models"
	"github.com/lenaxia/tinyrsvp/pkg/token"
)

func TestIntegration_ListInvites(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	eventID := createTestEvent(t, db)

	secret := []byte("test-secret-key-32-bytes-long!")
	generator := token.NewGenerator(secret)
	repo := repositories.NewInviteRepository(db)
	service := NewInviteService(generator, repo)

	ctx := context.Background()
	expiresAt := time.Now().Add(30 * 24 * time.Hour)

	testInvites := []struct {
		name   string
		email  string
		status models.InviteStatus
	}{
		{"Alice", "alice@example.com", models.InviteStatusDraft},
		{"Bob", "bob@example.com", models.InviteStatusSent},
		{"Charlie", "charlie@example.com", models.InviteStatusViewed},
		{"Diana", "diana@example.com", models.InviteStatusResponded},
		{"Eve", "eve@example.com", models.InviteStatusRevoked},
		{"Frank", "frank@example.com", models.InviteStatusSent},
		{"Grace", "grace@example.com", models.InviteStatusViewed},
		{"Henry", "henry@example.com", models.InviteStatusSent},
	}

	createdInvites := make([]*models.Invite, 0, len(testInvites))
	for _, ti := range testInvites {
		name := ti.name
		email := ti.email
		invite, _, err := service.CreateInvite(ctx, eventID, &name, &email, 2, expiresAt)
		if err != nil {
			t.Fatalf("CreateInvite() error = %v", err)
		}

		switch ti.status {
		case models.InviteStatusSent:
			if err := service.MarkInviteSent(ctx, invite.ID); err != nil {
				t.Fatalf("MarkInviteSent() error = %v", err)
			}
		case models.InviteStatusViewed:
			if err := service.MarkInviteSent(ctx, invite.ID); err != nil {
				t.Fatalf("MarkInviteSent() error = %v", err)
			}
			if err := service.MarkInviteViewed(ctx, invite.ID); err != nil {
				t.Fatalf("MarkInviteViewed() error = %v", err)
			}
		case models.InviteStatusResponded:
			if err := service.MarkInviteSent(ctx, invite.ID); err != nil {
				t.Fatalf("MarkInviteSent() error = %v", err)
			}
			if err := service.MarkInviteViewed(ctx, invite.ID); err != nil {
				t.Fatalf("MarkInviteViewed() error = %v", err)
			}
			if err := service.MarkInviteResponded(ctx, invite.ID); err != nil {
				t.Fatalf("MarkInviteResponded() error = %v", err)
			}
		case models.InviteStatusRevoked:
			req := &RevokeInviteRequest{InviteID: invite.ID}
			if err := service.RevokeInvite(ctx, req); err != nil {
				t.Fatalf("RevokeInvite() error = %v", err)
			}
		}

		refreshed, err := service.GetInviteByID(ctx, invite.ID)
		if err != nil {
			t.Fatalf("GetInviteByID() error = %v", err)
		}
		createdInvites = append(createdInvites, refreshed)
	}

	t.Run("list all invites", func(t *testing.T) {
		req := &ListInvitesRequest{
			EventID: eventID,
			Limit:   100,
			Offset:  0,
		}

		resp, err := service.ListInvites(ctx, req)
		if err != nil {
			t.Fatalf("ListInvites() error = %v", err)
		}

		if len(resp.Invites) != len(testInvites) {
			t.Errorf("got %d invites, want %d", len(resp.Invites), len(testInvites))
		}

		if resp.Total != len(testInvites) {
			t.Errorf("Total = %d, want %d", resp.Total, len(testInvites))
		}

		if resp.Stats == nil {
			t.Fatal("Stats is nil")
		}

		if resp.Stats.Total != len(testInvites) {
			t.Errorf("Stats.Total = %d, want %d", resp.Stats.Total, len(testInvites))
		}
	})

	t.Run("filter by status sent", func(t *testing.T) {
		status := string(models.InviteStatusSent)
		req := &ListInvitesRequest{
			EventID: eventID,
			Status:  &status,
			Limit:   100,
			Offset:  0,
		}

		resp, err := service.ListInvites(ctx, req)
		if err != nil {
			t.Fatalf("ListInvites() error = %v", err)
		}

		expectedCount := 3
		if len(resp.Invites) != expectedCount {
			t.Errorf("got %d sent invites, want %d", len(resp.Invites), expectedCount)
		}

		for _, inv := range resp.Invites {
			if inv.Status != models.InviteStatusSent {
				t.Errorf("invite %d has status %s, want %s", inv.ID, inv.Status, models.InviteStatusSent)
			}
		}

		if resp.Total != expectedCount {
			t.Errorf("Total = %d, want %d", resp.Total, expectedCount)
		}
	})

	t.Run("search by email", func(t *testing.T) {
		search := "alice"
		req := &ListInvitesRequest{
			EventID: eventID,
			Search:  &search,
			Limit:   100,
			Offset:  0,
		}

		resp, err := service.ListInvites(ctx, req)
		if err != nil {
			t.Fatalf("ListInvites() error = %v", err)
		}

		if len(resp.Invites) != 1 {
			t.Errorf("got %d invites, want 1", len(resp.Invites))
		}

		if len(resp.Invites) > 0 && resp.Invites[0].Email != nil {
			if *resp.Invites[0].Email != "alice@example.com" {
				t.Errorf("email = %s, want alice@example.com", *resp.Invites[0].Email)
			}
		}
	})

	t.Run("search by name", func(t *testing.T) {
		search := "Bob"
		req := &ListInvitesRequest{
			EventID: eventID,
			Search:  &search,
			Limit:   100,
			Offset:  0,
		}

		resp, err := service.ListInvites(ctx, req)
		if err != nil {
			t.Fatalf("ListInvites() error = %v", err)
		}

		if len(resp.Invites) != 1 {
			t.Errorf("got %d invites, want 1", len(resp.Invites))
		}

		if len(resp.Invites) > 0 && resp.Invites[0].Name != nil {
			if *resp.Invites[0].Name != "Bob" {
				t.Errorf("name = %s, want Bob", *resp.Invites[0].Name)
			}
		}
	})

	t.Run("sort by email ascending", func(t *testing.T) {
		sortBy := "email"
		sortOrder := "asc"
		req := &ListInvitesRequest{
			EventID:   eventID,
			SortBy:    &sortBy,
			SortOrder: &sortOrder,
			Limit:     100,
			Offset:    0,
		}

		resp, err := service.ListInvites(ctx, req)
		if err != nil {
			t.Fatalf("ListInvites() error = %v", err)
		}

		if len(resp.Invites) < 2 {
			t.Skip("need at least 2 invites to test sorting")
		}

		for i := 1; i < len(resp.Invites); i++ {
			if resp.Invites[i].Email == nil || resp.Invites[i-1].Email == nil {
				continue
			}
			if *resp.Invites[i].Email < *resp.Invites[i-1].Email {
				t.Errorf("invites not sorted by email ascending: %s comes before %s",
					*resp.Invites[i].Email, *resp.Invites[i-1].Email)
			}
		}
	})

	t.Run("sort by sent_at descending", func(t *testing.T) {
		sortBy := "sent_at"
		sortOrder := "desc"
		req := &ListInvitesRequest{
			EventID:   eventID,
			SortBy:    &sortBy,
			SortOrder: &sortOrder,
			Limit:     100,
			Offset:    0,
		}

		resp, err := service.ListInvites(ctx, req)
		if err != nil {
			t.Fatalf("ListInvites() error = %v", err)
		}

		for i := 1; i < len(resp.Invites); i++ {
			if resp.Invites[i].SentAt == nil || resp.Invites[i-1].SentAt == nil {
				continue
			}
			if resp.Invites[i].SentAt.After(*resp.Invites[i-1].SentAt) {
				t.Errorf("invites not sorted by sent_at descending")
			}
		}
	})

	t.Run("sort by viewed_at ascending", func(t *testing.T) {
		sortBy := "viewed_at"
		sortOrder := "asc"
		req := &ListInvitesRequest{
			EventID:   eventID,
			SortBy:    &sortBy,
			SortOrder: &sortOrder,
			Limit:     100,
			Offset:    0,
		}

		resp, err := service.ListInvites(ctx, req)
		if err != nil {
			t.Fatalf("ListInvites() error = %v", err)
		}

		for i := 1; i < len(resp.Invites); i++ {
			if resp.Invites[i].ViewedAt == nil || resp.Invites[i-1].ViewedAt == nil {
				continue
			}
			if resp.Invites[i].ViewedAt.Before(*resp.Invites[i-1].ViewedAt) {
				t.Errorf("invites not sorted by viewed_at ascending")
			}
		}
	})

	t.Run("pagination first page", func(t *testing.T) {
		req := &ListInvitesRequest{
			EventID: eventID,
			Limit:   3,
			Offset:  0,
		}

		resp, err := service.ListInvites(ctx, req)
		if err != nil {
			t.Fatalf("ListInvites() error = %v", err)
		}

		if len(resp.Invites) != 3 {
			t.Errorf("got %d invites, want 3", len(resp.Invites))
		}

		if resp.Total != len(testInvites) {
			t.Errorf("Total = %d, want %d", resp.Total, len(testInvites))
		}
	})

	t.Run("pagination second page", func(t *testing.T) {
		req := &ListInvitesRequest{
			EventID: eventID,
			Limit:   3,
			Offset:  3,
		}

		resp, err := service.ListInvites(ctx, req)
		if err != nil {
			t.Fatalf("ListInvites() error = %v", err)
		}

		if len(resp.Invites) != 3 {
			t.Errorf("got %d invites, want 3", len(resp.Invites))
		}

		if resp.Total != len(testInvites) {
			t.Errorf("Total = %d, want %d", resp.Total, len(testInvites))
		}
	})

	t.Run("combined filters status and search", func(t *testing.T) {
		status := string(models.InviteStatusSent)
		search := "frank"
		req := &ListInvitesRequest{
			EventID: eventID,
			Status:  &status,
			Search:  &search,
			Limit:   100,
			Offset:  0,
		}

		resp, err := service.ListInvites(ctx, req)
		if err != nil {
			t.Fatalf("ListInvites() error = %v", err)
		}

		if len(resp.Invites) != 1 {
			t.Errorf("got %d invites, want 1", len(resp.Invites))
		}

		if len(resp.Invites) > 0 {
			if resp.Invites[0].Status != models.InviteStatusSent {
				t.Errorf("status = %s, want %s", resp.Invites[0].Status, models.InviteStatusSent)
			}
			if resp.Invites[0].Name != nil && *resp.Invites[0].Name != "Frank" {
				t.Errorf("name = %s, want Frank", *resp.Invites[0].Name)
			}
		}

		if resp.Total != 1 {
			t.Errorf("Total = %d, want 1", resp.Total)
		}
	})

	t.Run("statistics accuracy", func(t *testing.T) {
		req := &ListInvitesRequest{
			EventID: eventID,
			Limit:   100,
			Offset:  0,
		}

		resp, err := service.ListInvites(ctx, req)
		if err != nil {
			t.Fatalf("ListInvites() error = %v", err)
		}

		if resp.Stats == nil {
			t.Fatal("Stats is nil")
		}

		expectedStats := map[models.InviteStatus]int{
			models.InviteStatusDraft:     1,
			models.InviteStatusSent:      3,
			models.InviteStatusViewed:    2,
			models.InviteStatusResponded: 1,
			models.InviteStatusRevoked:   1,
		}

		if resp.Stats.Draft != expectedStats[models.InviteStatusDraft] {
			t.Errorf("Stats.Draft = %d, want %d", resp.Stats.Draft, expectedStats[models.InviteStatusDraft])
		}
		if resp.Stats.Sent != expectedStats[models.InviteStatusSent] {
			t.Errorf("Stats.Sent = %d, want %d", resp.Stats.Sent, expectedStats[models.InviteStatusSent])
		}
		if resp.Stats.Viewed != expectedStats[models.InviteStatusViewed] {
			t.Errorf("Stats.Viewed = %d, want %d", resp.Stats.Viewed, expectedStats[models.InviteStatusViewed])
		}
		if resp.Stats.Responded != expectedStats[models.InviteStatusResponded] {
			t.Errorf("Stats.Responded = %d, want %d", resp.Stats.Responded, expectedStats[models.InviteStatusResponded])
		}
		if resp.Stats.Revoked != expectedStats[models.InviteStatusRevoked] {
			t.Errorf("Stats.Revoked = %d, want %d", resp.Stats.Revoked, expectedStats[models.InviteStatusRevoked])
		}
	})

	t.Run("empty results", func(t *testing.T) {
		search := "nonexistent"
		req := &ListInvitesRequest{
			EventID: eventID,
			Search:  &search,
			Limit:   100,
			Offset:  0,
		}

		resp, err := service.ListInvites(ctx, req)
		if err != nil {
			t.Fatalf("ListInvites() error = %v", err)
		}

		if len(resp.Invites) != 0 {
			t.Errorf("got %d invites, want 0", len(resp.Invites))
		}

		if resp.Total != 0 {
			t.Errorf("Total = %d, want 0", resp.Total)
		}
	})

	t.Run("permissions check - different event", func(t *testing.T) {
		otherEventID := createTestEvent(t, db)

		req := &ListInvitesRequest{
			EventID: otherEventID,
			Limit:   100,
			Offset:  0,
		}

		resp, err := service.ListInvites(ctx, req)
		if err != nil {
			t.Fatalf("ListInvites() error = %v", err)
		}

		if len(resp.Invites) != 0 {
			t.Errorf("got %d invites for different event, want 0", len(resp.Invites))
		}

		if resp.Total != 0 {
			t.Errorf("Total = %d for different event, want 0", resp.Total)
		}
	})
}

func TestIntegration_ListInvites_FilteredCount(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	eventID := createTestEvent(t, db)

	secret := []byte("test-secret-key-32-bytes-long!")
	generator := token.NewGenerator(secret)
	repo := repositories.NewInviteRepository(db)
	service := NewInviteService(generator, repo)

	ctx := context.Background()
	expiresAt := time.Now().Add(30 * 24 * time.Hour)

	for i := 0; i < 10; i++ {
		name := "User"
		email := "user@example.com"
		invite, _, err := service.CreateInvite(ctx, eventID, &name, &email, 2, expiresAt)
		if err != nil {
			t.Fatalf("CreateInvite() error = %v", err)
		}

		if i < 5 {
			if err := service.MarkInviteSent(ctx, invite.ID); err != nil {
				t.Fatalf("MarkInviteSent() error = %v", err)
			}
		}
	}

	t.Run("filtered count matches filtered results", func(t *testing.T) {
		status := string(models.InviteStatusSent)
		req := &ListInvitesRequest{
			EventID: eventID,
			Status:  &status,
			Limit:   100,
			Offset:  0,
		}

		resp, err := service.ListInvites(ctx, req)
		if err != nil {
			t.Fatalf("ListInvites() error = %v", err)
		}

		if resp.Total != len(resp.Invites) {
			t.Errorf("Total = %d, but got %d invites", resp.Total, len(resp.Invites))
		}

		if resp.Total != 5 {
			t.Errorf("Total = %d, want 5", resp.Total)
		}
	})

	t.Run("paginated filtered count is consistent", func(t *testing.T) {
		status := string(models.InviteStatusSent)
		req1 := &ListInvitesRequest{
			EventID: eventID,
			Status:  &status,
			Limit:   2,
			Offset:  0,
		}

		resp1, err := service.ListInvites(ctx, req1)
		if err != nil {
			t.Fatalf("ListInvites() error = %v", err)
		}

		req2 := &ListInvitesRequest{
			EventID: eventID,
			Status:  &status,
			Limit:   2,
			Offset:  2,
		}

		resp2, err := service.ListInvites(ctx, req2)
		if err != nil {
			t.Fatalf("ListInvites() error = %v", err)
		}

		if resp1.Total != resp2.Total {
			t.Errorf("Total inconsistent: page1=%d, page2=%d", resp1.Total, resp2.Total)
		}

		if resp1.Total != 5 {
			t.Errorf("Total = %d, want 5", resp1.Total)
		}
	})
}

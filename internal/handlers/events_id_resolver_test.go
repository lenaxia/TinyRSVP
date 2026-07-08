package handlers

import (
	"context"
	"testing"

	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

func TestResolveEventID(t *testing.T) {
	publicID := "aBcD123456"
	friendlyName := "summer-party"

	mockRepo := &mockEventIDResolverRepo{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			if id == 123 {
				return &models.Event{ID: 123, Title: "Event 123"}, nil
			}
			return nil, &models.NotFoundError{Resource: "Event", ID: id}
		},
		getByPublicIDFunc: func(ctx context.Context, pid string) (*models.Event, error) {
			if pid == publicID {
				return &models.Event{ID: 456, PublicID: &publicID, Title: "Event by PublicID"}, nil
			}
			return nil, &models.NotFoundError{Resource: "Event", ID: pid}
		},
		getByFriendlyNameFunc: func(ctx context.Context, fname string) (*models.Event, error) {
			if fname == friendlyName {
				return &models.Event{ID: 789, FriendlyName: &friendlyName, Title: "Event by FriendlyName"}, nil
			}
			return nil, &models.NotFoundError{Resource: "Event", ID: fname}
		},
	}

	tests := []struct {
		name      string
		idParam   string
		wantID    int64
		wantErr   bool
		wantTitle string
	}{
		{
			name:      "numeric ID",
			idParam:   "123",
			wantID:    123,
			wantErr:   false,
			wantTitle: "Event 123",
		},
		{
			name:      "public ID",
			idParam:   publicID,
			wantID:    456,
			wantErr:   false,
			wantTitle: "Event by PublicID",
		},
		{
			name:      "friendly name",
			idParam:   friendlyName,
			wantID:    789,
			wantErr:   false,
			wantTitle: "Event by FriendlyName",
		},
		{
			name:    "non-existent numeric ID",
			idParam: "999",
			wantErr: true,
		},
		{
			name:    "non-existent public ID",
			idParam: "nonexist1",
			wantErr: true,
		},
		{
			name:    "non-existent friendly name",
			idParam: "nonexistent-event",
			wantErr: true,
		},
		{
			name:    "empty string",
			idParam: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event, err := resolveEventID(context.Background(), mockRepo, tt.idParam)

			if (err != nil) != tt.wantErr {
				t.Errorf("resolveEventID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if event == nil {
					t.Fatal("Expected event, got nil")
				}
				if event.ID != tt.wantID {
					t.Errorf("Expected ID %d, got %d", tt.wantID, event.ID)
				}
				if event.Title != tt.wantTitle {
					t.Errorf("Expected title %s, got %s", tt.wantTitle, event.Title)
				}
			}
		})
	}
}

func TestResolveEventID_PriorityOrder(t *testing.T) {
	// Test that public_id is tried before friendly_name when both could match
	// Use a valid 10-character public_id format
	ambiguousID := "abc1234567"

	mockRepo := &mockEventIDResolverRepo{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			return nil, &models.NotFoundError{Resource: "Event", ID: id}
		},
		getByPublicIDFunc: func(ctx context.Context, pid string) (*models.Event, error) {
			if pid == ambiguousID {
				return &models.Event{ID: 100, Title: "Found by PublicID"}, nil
			}
			return nil, &models.NotFoundError{Resource: "Event", ID: pid}
		},
		getByFriendlyNameFunc: func(ctx context.Context, fname string) (*models.Event, error) {
			if fname == ambiguousID {
				return &models.Event{ID: 200, Title: "Found by FriendlyName"}, nil
			}
			return nil, &models.NotFoundError{Resource: "Event", ID: fname}
		},
	}

	event, err := resolveEventID(context.Background(), mockRepo, ambiguousID)
	if err != nil {
		t.Fatalf("resolveEventID() error = %v", err)
	}

	// "abc1234567" is a valid 10-character public_id, so it should be found by PublicID first
	if event.ID != 100 {
		t.Errorf("Expected to find by PublicID (ID 100), got ID %d", event.ID)
	}

	if event.Title != "Found by PublicID" {
		t.Errorf("Expected title 'Found by PublicID', got '%s'", event.Title)
	}
}

type mockEventIDResolverRepo struct {
	getByIDFunc           func(ctx context.Context, id int64) (*models.Event, error)
	getByPublicIDFunc     func(ctx context.Context, publicID string) (*models.Event, error)
	getByFriendlyNameFunc func(ctx context.Context, friendlyName string) (*models.Event, error)
}

func (m *mockEventIDResolverRepo) GetByID(ctx context.Context, id int64) (*models.Event, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, &models.NotFoundError{Resource: "Event", ID: id}
}

func (m *mockEventIDResolverRepo) GetByPublicID(ctx context.Context, publicID string) (*models.Event, error) {
	if m.getByPublicIDFunc != nil {
		return m.getByPublicIDFunc(ctx, publicID)
	}
	return nil, &models.NotFoundError{Resource: "Event", ID: publicID}
}

func (m *mockEventIDResolverRepo) GetByFriendlyName(ctx context.Context, friendlyName string) (*models.Event, error) {
	if m.getByFriendlyNameFunc != nil {
		return m.getByFriendlyNameFunc(ctx, friendlyName)
	}
	return nil, &models.NotFoundError{Resource: "Event", ID: friendlyName}
}

func (m *mockEventIDResolverRepo) Create(ctx context.Context, event *models.Event) error {
	return nil
}

func (m *mockEventIDResolverRepo) Update(ctx context.Context, event *models.Event) error {
	return nil
}

func (m *mockEventIDResolverRepo) UpdateWithVersion(ctx context.Context, event *models.Event, expectedVersion int) error {
	return nil
}

func (m *mockEventIDResolverRepo) UpdateStatus(ctx context.Context, id int64, status models.EventStatus) error {
	return nil
}

func (m *mockEventIDResolverRepo) Delete(ctx context.Context, id int64) error {
	return nil
}

func (m *mockEventIDResolverRepo) List(ctx context.Context, filters repositories.ListFilters) ([]*models.Event, error) {
	return nil, nil
}

func (m *mockEventIDResolverRepo) GetByStatus(ctx context.Context, status models.EventStatus) ([]*models.Event, error) {
	return nil, nil
}

func (m *mockEventIDResolverRepo) GetEventsToArchive(ctx context.Context, daysAfterEvent int) ([]*models.Event, error) {
	return nil, nil
}

func (m *mockEventIDResolverRepo) GetByCreatorID(ctx context.Context, creatorID int64) ([]*models.Event, error) {
	return nil, nil
}

func (m *mockEventIDResolverRepo) CountEvents(ctx context.Context) (int, error) {
	return 0, nil
}

func (m * mockEventIDResolverRepo) GetDashboardStatsByCreator(ctx context.Context, creatorID int64) (*models.DashboardStats, error) {
	return &models.DashboardStats{}, nil
}

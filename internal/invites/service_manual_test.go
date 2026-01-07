package invites

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/models"
)

func TestCreateManualInvite(t *testing.T) {
	tests := []struct {
		name          string
		eventID       int64
		guestName     *string
		maxPlusOnes   *int
		setupMocks    func(*mockGenerator, *mockInviteRepository)
		wantErr       bool
		wantErrMsg    string
		validateToken bool
		validateURL   bool
	}{
		{
			name:    "successful manual invite with name",
			eventID: 1,
			guestName: func() *string {
				s := "John Doe"
				return &s
			}(),
			maxPlusOnes: func() *int {
				i := 2
				return &i
			}(),
			setupMocks: func(mg *mockGenerator, mr *mockInviteRepository) {
				mg.generateFunc = func() (string, error) {
					return strings.Repeat("a", 43), nil
				}
				mg.hashFunc = func(token string) (string, error) {
					return strings.Repeat("b", 43), nil
				}
				mr.createFunc = func(ctx context.Context, invite *models.Invite) error {
					invite.ID = 1
					invite.CreatedAt = time.Now()
					invite.UpdatedAt = time.Now()
					return nil
				}
			},
			wantErr:       false,
			validateToken: true,
			validateURL:   true,
		},
		{
			name:        "successful manual invite without name",
			eventID:     1,
			guestName:   nil,
			maxPlusOnes: nil,
			setupMocks: func(mg *mockGenerator, mr *mockInviteRepository) {
				mg.generateFunc = func() (string, error) {
					return strings.Repeat("c", 43), nil
				}
				mg.hashFunc = func(token string) (string, error) {
					return strings.Repeat("d", 43), nil
				}
				mr.createFunc = func(ctx context.Context, invite *models.Invite) error {
					invite.ID = 2
					invite.CreatedAt = time.Now()
					invite.UpdatedAt = time.Now()
					return nil
				}
			},
			wantErr:       false,
			validateToken: true,
			validateURL:   true,
		},
		{
			name:    "successful manual invite with zero plus ones",
			eventID: 1,
			guestName: func() *string {
				s := "Jane Smith"
				return &s
			}(),
			maxPlusOnes: func() *int {
				i := 0
				return &i
			}(),
			setupMocks: func(mg *mockGenerator, mr *mockInviteRepository) {
				mg.generateFunc = func() (string, error) {
					return strings.Repeat("e", 43), nil
				}
				mg.hashFunc = func(token string) (string, error) {
					return strings.Repeat("f", 43), nil
				}
				mr.createFunc = func(ctx context.Context, invite *models.Invite) error {
					invite.ID = 3
					invite.CreatedAt = time.Now()
					invite.UpdatedAt = time.Now()
					return nil
				}
			},
			wantErr:       false,
			validateToken: true,
			validateURL:   true,
		},
		{
			name:        "token generation fails",
			eventID:     1,
			guestName:   nil,
			maxPlusOnes: nil,
			setupMocks: func(mg *mockGenerator, mr *mockInviteRepository) {
				mg.generateFunc = func() (string, error) {
					return "", errors.New("random generation failed")
				}
			},
			wantErr:    true,
			wantErrMsg: "failed to generate token",
		},
		{
			name:    "token hashing fails",
			eventID: 1,
			guestName: func() *string {
				s := "Test User"
				return &s
			}(),
			maxPlusOnes: nil,
			setupMocks: func(mg *mockGenerator, mr *mockInviteRepository) {
				mg.generateFunc = func() (string, error) {
					return strings.Repeat("g", 43), nil
				}
				mg.hashFunc = func(token string) (string, error) {
					return "", errors.New("hashing failed")
				}
			},
			wantErr:    true,
			wantErrMsg: "failed to hash token",
		},
		{
			name:    "invalid max plus ones",
			eventID: 1,
			guestName: func() *string {
				s := "Test User"
				return &s
			}(),
			maxPlusOnes: func() *int {
				i := -1
				return &i
			}(),
			setupMocks: func(mg *mockGenerator, mr *mockInviteRepository) {
				mg.generateFunc = func() (string, error) {
					return strings.Repeat("h", 43), nil
				}
				mg.hashFunc = func(token string) (string, error) {
					return strings.Repeat("i", 43), nil
				}
			},
			wantErr:    true,
			wantErrMsg: "max_plus_ones must be between 0 and 10",
		},
		{
			name:        "database create fails",
			eventID:     1,
			guestName:   nil,
			maxPlusOnes: nil,
			setupMocks: func(mg *mockGenerator, mr *mockInviteRepository) {
				mg.generateFunc = func() (string, error) {
					return strings.Repeat("j", 43), nil
				}
				mg.hashFunc = func(token string) (string, error) {
					return strings.Repeat("k", 43), nil
				}
				mr.createFunc = func(ctx context.Context, invite *models.Invite) error {
					return errors.New("database error")
				}
			},
			wantErr:    true,
			wantErrMsg: "failed to create invite",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGen := &mockGenerator{}
			mockRepo := &mockInviteRepository{}
			tt.setupMocks(mockGen, mockRepo)

			service := NewInviteService(mockGen, mockRepo)

			expiresAt := time.Now().Add(30 * 24 * time.Hour)
			req := &CreateManualInviteRequest{
				EventID:     tt.eventID,
				Name:        tt.guestName,
				MaxPlusOnes: tt.maxPlusOnes,
			}

			resp, err := service.CreateManualInvite(context.Background(), req, expiresAt)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.wantErrMsg)
					return
				}
				if tt.wantErrMsg != "" && !strings.Contains(err.Error(), tt.wantErrMsg) {
					t.Errorf("expected error containing %q, got %q", tt.wantErrMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if resp == nil {
				t.Fatal("expected response, got nil")
			}

			if resp.Invite == nil {
				t.Fatal("expected invite in response, got nil")
			}

			if resp.Invite.EventID != tt.eventID {
				t.Errorf("expected event ID %d, got %d", tt.eventID, resp.Invite.EventID)
			}

			if tt.guestName != nil {
				if resp.Invite.Name == nil {
					t.Error("expected name in invite, got nil")
				} else if *resp.Invite.Name != *tt.guestName {
					t.Errorf("expected name %q, got %q", *tt.guestName, *resp.Invite.Name)
				}
			} else {
				if resp.Invite.Name != nil {
					t.Errorf("expected nil name, got %q", *resp.Invite.Name)
				}
			}

			if resp.Invite.Email != nil {
				t.Errorf("expected nil email for manual invite, got %q", *resp.Invite.Email)
			}

			if resp.Invite.Status != models.InviteStatusDraft {
				t.Errorf("expected status %q, got %q", models.InviteStatusDraft, resp.Invite.Status)
			}

			expectedMaxPlusOnes := 0
			if tt.maxPlusOnes != nil {
				expectedMaxPlusOnes = *tt.maxPlusOnes
			}
			if resp.Invite.MaxPlusOnes != expectedMaxPlusOnes {
				t.Errorf("expected max plus ones %d, got %d", expectedMaxPlusOnes, resp.Invite.MaxPlusOnes)
			}

			if tt.validateToken {
				if resp.Token == "" {
					t.Error("expected non-empty token")
				}
			}

			if tt.validateURL {
				if resp.RSVPURL == "" {
					t.Error("expected non-empty RSVP URL")
				}
				if !strings.HasPrefix(resp.RSVPURL, "/rsvp/") {
					t.Errorf("expected RSVP URL to start with /rsvp/, got %q", resp.RSVPURL)
				}
			}
		})
	}
}

func TestCreateManualInvite_MultipleInvites(t *testing.T) {
	inviteCounter := int64(0)
	tokenCounter := 0
	mockGen := &mockGenerator{
		generateFunc: func() (string, error) {
			tokenCounter++
			char := rune('a' + (tokenCounter % 26))
			return strings.Repeat(string(char), 43), nil
		},
		hashFunc: func(token string) (string, error) {
			return strings.Repeat("x", 43), nil
		},
	}
	mockRepo := &mockInviteRepository{
		createFunc: func(ctx context.Context, invite *models.Invite) error {
			inviteCounter++
			invite.ID = inviteCounter
			invite.CreatedAt = time.Now()
			invite.UpdatedAt = time.Now()
			return nil
		},
	}

	service := NewInviteService(mockGen, mockRepo)

	expiresAt := time.Now().Add(30 * 24 * time.Hour)
	tokens := make(map[string]bool)

	for i := 0; i < 10; i++ {
		req := &CreateManualInviteRequest{
			EventID:     1,
			Name:        nil,
			MaxPlusOnes: nil,
		}

		resp, err := service.CreateManualInvite(context.Background(), req, expiresAt)
		if err != nil {
			t.Fatalf("iteration %d: unexpected error: %v", i, err)
		}

		if resp.Token == "" {
			t.Fatalf("iteration %d: expected non-empty token", i)
		}

		if tokens[resp.Token] {
			t.Fatalf("iteration %d: duplicate token generated: %s", i, resp.Token)
		}
		tokens[resp.Token] = true

		if resp.Invite.ID != int64(i+1) {
			t.Errorf("iteration %d: expected invite ID %d, got %d", i, i+1, resp.Invite.ID)
		}
	}

	if len(tokens) != 10 {
		t.Errorf("expected 10 unique tokens, got %d", len(tokens))
	}
}

package handlers

import (
	"testing"

	"github.com/lenaxia/tinyrsvp/internal/models"
)

func TestHidePrivateGuestListStats(t *testing.T) {
	tests := []struct {
		name           string
		events         []*models.EventWithStats
		user           *models.User
		wantHidden     []bool
	}{
		{
			name: "non-owner sees zeros for private event",
			events: []*models.EventWithStats{
				{Event: models.Event{ID: 1, CreatedBy: 10, PrivateGuestList: true}, InviteCount: 5, RSVPCount: 3, AcceptCount: 2},
				{Event: models.Event{ID: 2, CreatedBy: 20, PrivateGuestList: false}, InviteCount: 8, RSVPCount: 6, AcceptCount: 4},
			},
			user:       &models.User{ID: 20, Role: models.RoleEventManager},
			wantHidden: []bool{true, false},
		},
		{
			name: "owner sees full stats for own private event",
			events: []*models.EventWithStats{
				{Event: models.Event{ID: 1, CreatedBy: 10, PrivateGuestList: true}, InviteCount: 5, RSVPCount: 3, AcceptCount: 2},
			},
			user:       &models.User{ID: 10, Role: models.RoleEventManager},
			wantHidden: []bool{false},
		},
		{
			name: "admin sees full stats even for private events",
			events: []*models.EventWithStats{
				{Event: models.Event{ID: 1, CreatedBy: 10, PrivateGuestList: true}, InviteCount: 5, RSVPCount: 3, AcceptCount: 2},
			},
			user:       &models.User{ID: 99, Role: models.RoleAdmin},
			wantHidden: []bool{false},
		},
		{
			name: "non-private events are never hidden",
			events: []*models.EventWithStats{
				{Event: models.Event{ID: 1, CreatedBy: 10, PrivateGuestList: false}, InviteCount: 5, RSVPCount: 3, AcceptCount: 2},
			},
			user:       &models.User{ID: 20, Role: models.RoleEventManager},
			wantHidden: []bool{false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hidePrivateGuestListStats(tt.events, tt.user)

			for i, e := range tt.events {
				hidden := e.InviteCount == 0 && e.RSVPCount == 0 && e.AcceptCount == 0
				if hidden != tt.wantHidden[i] {
					t.Errorf("event[%d]: stats hidden=%v, want %v (Invite=%d RSVP=%d Accept=%d)",
						i, hidden, tt.wantHidden[i], e.InviteCount, e.RSVPCount, e.AcceptCount)
				}
			}
		})
	}
}

#!/bin/bash

# Add GetByCreatorID to mockRegenerateEventRepository
if grep -q "type mockRegenerateEventRepository struct" internal/handlers/invites_regenerate_test.go; then
    if ! grep -q "GetByCreatorID" internal/handlers/invites_regenerate_test.go; then
        sed -i '/func (m \*mockRegenerateEventRepository) GetEventsToArchive/a\\nfunc (m *mockRegenerateEventRepository) GetByCreatorID(ctx context.Context, creatorID int64) ([]*models.Event, error) {\n\treturn []*models.Event{}, nil\n}' internal/handlers/invites_regenerate_test.go
    fi
fi

# Add GetByCreatorID to mockRevokeEventRepository
if grep -q "type mockRevokeEventRepository struct" internal/handlers/invites_revoke_test.go; then
    if ! grep -q "GetByCreatorID" internal/handlers/invites_revoke_test.go; then
        sed -i '/func (m \*mockRevokeEventRepository) GetEventsToArchive/a\\nfunc (m *mockRevokeEventRepository) GetByCreatorID(ctx context.Context, creatorID int64) ([]*models.Event, error) {\n\treturn []*models.Event{}, nil\n}' internal/handlers/invites_revoke_test.go
    fi
fi

# Add GetByCreatorID to mockRSVPEventRepository
if grep -q "type mockRSVPEventRepository struct" internal/handlers/rsvp_confirmation_test.go; then
    if ! grep -q "GetByCreatorID" internal/handlers/rsvp_confirmation_test.go; then
        sed -i '/func (m \*mockRSVPEventRepository) GetEventsToArchive/a\\nfunc (m *mockRSVPEventRepository) GetByCreatorID(ctx context.Context, creatorID int64) ([]*models.Event, error) {\n\treturn []*models.Event{}, nil\n}' internal/handlers/rsvp_confirmation_test.go
    fi
fi

# Add GetByInviteIDs to mockRSVPRSVPRepository  
if grep -q "type mockRSVPRSVPRepository struct" internal/handlers/rsvp_confirmation_test.go; then
    if ! grep -q "GetByInviteIDs" internal/handlers/rsvp_confirmation_test.go; then
        sed -i '/func (m \*mockRSVPRSVPRepository) GetStats/a\\nfunc (m *mockRSVPRSVPRepository) GetByInviteIDs(ctx context.Context, inviteIDs []int64) ([]*models.RSVP, error) {\n\treturn []*models.RSVP{}, nil\n}' internal/handlers/rsvp_confirmation_test.go
    fi
fi

echo "Mock repositories updated"

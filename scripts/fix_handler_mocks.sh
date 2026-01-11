#!/bin/bash

# Script to add component override methods to all mock event repositories in handlers tests

FILES=(
    "internal/handlers/invites_get_test.go"
    "internal/handlers/invites_revoke_test.go"
    "internal/handlers/rsvp_summary_test.go"
    "internal/handlers/invites_regenerate_test.go"
    "internal/handlers/invites_delete_test.go"
    "internal/handlers/invites_list_test.go"
    "internal/handlers/invites_import_permission_test.go"
    "internal/handlers/rsvp_test.go"
    "internal/handlers/invites_send_test.go"
    "internal/handlers/invites_update_test.go"
)

METHODS='
func (m *mockRSVPEventRepository) GetComponentOverrides(ctx context.Context, eventID int64) (*models.ComponentOverrides, error) {
	return nil, nil
}

func (m *mockRSVPEventRepository) UpdateComponentOverrides(ctx context.Context, eventID int64, overrides *models.ComponentOverrides) error {
	return nil
}

func (m *mockRSVPEventRepository) DeleteComponentOverrides(ctx context.Context, eventID int64) error {
	return nil
}
'

for file in "${FILES[@]}"; do
    if [ -f "$file" ]; then
        # Find the mock struct and add methods after CountEvents if not already present
        if ! grep -q "DeleteComponentOverrides" "$file"; then
            # Find the line number of CountEvents method
            line=$(grep -n "func (m \*mock.*EventRepo.*) CountEvents" "$file" | head -1 | cut -d: -f1)
            if [ -n "$line" ]; then
                # Find the end of CountEvents method (next blank line or next func)
                end_line=$(tail -n +$((line + 1)) "$file" | grep -n "^func\|^$" | head -1 | cut -d: -f1)
                if [ -n "$end_line" ]; then
                    insert_line=$((line + end_line))
                    # Insert the new methods
                    sed -i "${insert_line}i\\${METHODS}" "$file"
                    echo "Updated $file"
                fi
            fi
        fi
    fi
done

echo "All mock repositories updated"

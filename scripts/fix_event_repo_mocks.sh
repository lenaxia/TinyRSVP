#!/bin/bash

# Script to add missing GetByPublicID and GetByFriendlyName methods to mock event repositories

FILES=(
    "internal/handlers/invites_get_test.go"
    "internal/handlers/invites_revoke_test.go"
    "internal/handlers/invites_regenerate_test.go"
    "internal/handlers/invites_delete_test.go"
    "internal/handlers/invites_list_test.go"
    "internal/handlers/invites_import_permission_test.go"
    "internal/handlers/invites_send_test.go"
    "internal/handlers/invites_update_test.go"
)

METHODS='
func (m *mock%sEventRepo) GetByPublicID(ctx context.Context, publicID string) (*models.Event, error) {
	return nil, nil
}

func (m *mock%sEventRepo) GetByFriendlyName(ctx context.Context, friendlyName string) (*models.Event, error) {
	return nil, nil
}
'

for file in "${FILES[@]}"; do
    if [ -f "$file" ]; then
        echo "Processing $file..."
        
        # Extract the mock type name from the file
        mock_name=$(grep -o "type mock[A-Za-z]*EventRepo" "$file" | head -1 | sed 's/type mock//' | sed 's/EventRepo//')
        
        if [ -n "$mock_name" ]; then
            # Format the methods with the correct type name
            formatted_methods=$(printf "$METHODS" "$mock_name" "$mock_name")
            
            # Find the line with CountEvents and add after it
            if grep -q "func (m \*mock${mock_name}EventRepo) CountEvents" "$file"; then
                # Add after CountEvents
                sed -i "/func (m \*mock${mock_name}EventRepo) CountEvents/,/^}$/a\\
\\
$formatted_methods" "$file"
            else
                echo "  Warning: CountEvents not found in $file"
            fi
        fi
    fi
done

echo "Done!"

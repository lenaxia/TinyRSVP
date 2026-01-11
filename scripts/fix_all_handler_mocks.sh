#!/bin/bash

# Add component override methods to all mock event repositories in handlers

cd /home/mikekao/personal/TinyRSVP

# Find all test files with mock event repositories
for file in internal/handlers/*_test.go; do
    if grep -q "type mock.*EventRepo" "$file" && ! grep -q "DeleteComponentOverrides" "$file"; then
        echo "Fixing $file..."
        
        # Find the CountEvents method and add new methods after it
        awk '
        /^func \(m \*mock.*EventRepo.*\) CountEvents/ {
            in_count_events = 1
        }
        {
            print
        }
        in_count_events && /^}$/ {
            print ""
            print "func (m *" mock_name ") GetComponentOverrides(ctx context.Context, eventID int64) (*models.ComponentOverrides, error) {"
            print "\treturn nil, nil"
            print "}"
            print ""
            print "func (m *" mock_name ") UpdateComponentOverrides(ctx context.Context, eventID int64, overrides *models.ComponentOverrides) error {"
            print "\treturn nil"
            print "}"
            print ""
            print "func (m *" mock_name ") DeleteComponentOverrides(ctx context.Context, eventID int64) error {"
            print "\treturn nil"
            print "}"
            in_count_events = 0
        }
        /^type (mock.*EventRepo.*)struct/ {
            mock_name = $2
        }
        ' "$file" > "$file.tmp" && mv "$file.tmp" "$file"
    fi
done

echo "Done!"

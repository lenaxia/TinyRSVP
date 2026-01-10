#!/bin/bash
# Script to add Accept: application/json headers to test requests
# This ensures tests expecting JSON responses get JSON, not HTML

set -e

# Find all test files in handlers
TEST_FILES=$(find internal/handlers -name "*_test.go" -type f)

for file in $TEST_FILES; do
    echo "Processing $file..."
    
    # Create backup
    cp "$file" "$file.bak"
    
    # Add Accept header after httptest.NewRequest calls that don't already have it
    # This is a complex sed operation that:
    # 1. Finds httptest.NewRequest lines
    # 2. Checks if the next few lines already have req.Header.Set("Accept"
    # 3. If not, adds it after the NewRequest line
    
    # For now, let's use a simpler approach with awk
    awk '
    /httptest\.NewRequest/ {
        print
        # Read next line
        getline
        # If it is not already setting Accept header, add it
        if ($0 !~ /req\.Header\.Set\("Accept"/) {
            # Check if this is setting Content-Type instead
            if ($0 ~ /req\.Header\.Set\("Content-Type"/) {
                print
                print "\t\t\treq.Header.Set(\"Accept\", \"application/json\")"
            } else {
                print "\t\t\treq.Header.Set(\"Accept\", \"application/json\")"
                print
            }
        } else {
            print
        }
        next
    }
    { print }
    ' "$file" > "$file.tmp" && mv "$file.tmp" "$file"
    
    echo "Completed $file"
done

echo "All test files processed!"
echo "Backups created with .bak extension"

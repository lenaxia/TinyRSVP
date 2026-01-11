#!/bin/bash

# Script to add Category: models.CategoryPlain to Template struct initializations
# This script processes Go test files that create models.Template structs

set -e

echo "Finding all Go files with models.Template struct initializations..."

# Find all Go files in internal/ that might have Template structs
files=$(find internal/ -name "*.go" -type f)

for file in $files; do
    # Check if file contains models.Template{
    if grep -q "models\.Template{" "$file"; then
        echo "Processing: $file"
        
        # Create a backup
        cp "$file" "$file.bak"
        
        # Use awk to add Category field after CreatedBy in Template structs
        # This is more reliable than sed for multi-line patterns
        awk '
        /models\.Template\{/ { in_template = 1 }
        in_template && /CreatedBy:/ { 
            print
            # Check if next line already has Category
            getline nextline
            if (nextline !~ /Category:/) {
                print "\t\t\tCategory:    models.CategoryPlain,"
            }
            print nextline
            next
        }
        { print }
        /\}/ && in_template { in_template = 0 }
        ' "$file.bak" > "$file"
        
        # Remove backup if successful
        rm "$file.bak"
        echo "  ✓ Updated $file"
    fi
done

echo ""
echo "Done! Running tests to verify..."
go test -timeout 30s ./internal/templates/... 2>&1 | head -50

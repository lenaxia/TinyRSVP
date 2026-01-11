#!/bin/bash

# Script to add Category: models.CategoryPlain to all Template struct initializations in test files

# Files to update
files=(
    "internal/db/repositories/template_repository_test.go"
    "internal/db/repositories/template_repository_integration_test.go"
)

for file in "${files[@]}"; do
    echo "Processing $file..."
    
    # Use sed to add Category field after CreatedBy field in Template struct initializations
    # This is a complex sed operation that looks for CreatedBy followed by a comma and adds Category
    sed -i '/CreatedBy:.*,$/a\			Category:    models.CategoryPlain,' "$file"
    
    echo "Updated $file"
done

echo "Done! Please review the changes and run tests."

#!/usr/bin/env python3
"""
Script to add Category: models.CategoryPlain to all models.Template struct initializations
that don't already have it.
"""

import re
import sys
from pathlib import Path

def add_category_to_template(content):
    """Add Category field to Template structs that don't have it."""
    
    # Pattern to match Template struct initialization
    # Looks for models.Template{ ... CreatedBy: ... } without Category
    pattern = r'(models\.Template\{[^}]*?CreatedBy:\s*[^,\n]+,)(\s*\n)(\s*)((?!Category:)[^}]*?\})'
    
    def replacer(match):
        before_created_by = match.group(1)
        newline = match.group(2)
        indent = match.group(3)
        after_created_by = match.group(4)
        
        # Check if Category already exists in the after part
        if 'Category:' in after_created_by:
            return match.group(0)
        
        # Add Category field
        return f"{before_created_by}{newline}{indent}Category:    models.CategoryPlain,{newline}{indent}{after_created_by}"
    
    # Apply the replacement
    modified = re.sub(pattern, replacer, content, flags=re.DOTALL)
    
    return modified

def process_file(filepath):
    """Process a single file."""
    try:
        with open(filepath, 'r') as f:
            content = f.read()
        
        # Check if file has models.Template
        if 'models.Template{' not in content:
            return False
        
        modified = add_category_to_template(content)
        
        if modified != content:
            with open(filepath, 'w') as f:
                f.write(modified)
            return True
        
        return False
    except Exception as e:
        print(f"Error processing {filepath}: {e}", file=sys.stderr)
        return False

def main():
    """Main function."""
    files_to_process = [
        'internal/templates/css_sanitizer_integration_test.go',
        'internal/templates/xss_integration_test.go',
        'internal/templates/service_integration_test.go',
        'internal/templates/validator_test.go',
        'internal/templates/service_test.go',
        'internal/templates/validator_integration_test.go',
    ]
    
    modified_count = 0
    for filepath in files_to_process:
        path = Path(filepath)
        if path.exists():
            if process_file(path):
                print(f"✓ Modified: {filepath}")
                modified_count += 1
            else:
                print(f"  Skipped: {filepath} (no changes needed)")
        else:
            print(f"  Not found: {filepath}")
    
    print(f"\nModified {modified_count} files")

if __name__ == '__main__':
    main()

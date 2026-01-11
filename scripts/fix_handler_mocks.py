#!/usr/bin/env python3
"""
Script to add component override methods to all mock event repositories in handler tests.
"""

import re
import sys

def add_methods_to_mock(content, mock_name):
    """Add component override methods after CountEvents method for a specific mock."""
    
    # Pattern to find CountEvents method for this specific mock
    pattern = rf'(func \(m \*{re.escape(mock_name)}\) CountEvents\(ctx context\.Context\) \(int, error\) \{{[^}}]*\}})'
    
    # Methods to add
    methods = f'''

func (m *{mock_name}) GetComponentOverrides(ctx context.Context, eventID int64) (*models.ComponentOverrides, error) {{
\treturn nil, nil
}}

func (m *{mock_name}) UpdateComponentOverrides(ctx context.Context, eventID int64, overrides *models.ComponentOverrides) error {{
\treturn nil
}}

func (m *{mock_name}) DeleteComponentOverrides(ctx context.Context, eventID int64) error {{
\treturn nil
}}'''
    
    # Check if methods already exist
    if f'func (m *{mock_name}) GetComponentOverrides' in content:
        return content
    
    # Add methods after CountEvents
    new_content = re.sub(pattern, r'\1' + methods, content)
    
    return new_content

def process_file(filepath):
    """Process a single test file."""
    try:
        with open(filepath, 'r') as f:
            content = f.read()
        
        # Find all mock event repository types in this file
        mock_pattern = r'type (mock\w*Event\w*Repo\w*) struct'
        mocks = re.findall(mock_pattern, content)
        
        if not mocks:
            return False
        
        print(f"Processing {filepath}...")
        print(f"  Found mocks: {', '.join(mocks)}")
        
        # Add methods for each mock
        modified = False
        for mock_name in mocks:
            new_content = add_methods_to_mock(content, mock_name)
            if new_content != content:
                content = new_content
                modified = True
                print(f"  Added methods to {mock_name}")
        
        if modified:
            with open(filepath, 'w') as f:
                f.write(content)
            print(f"  ✓ Updated {filepath}")
            return True
        else:
            print(f"  - No changes needed for {filepath}")
            return False
            
    except Exception as e:
        print(f"  ✗ Error processing {filepath}: {e}")
        return False

def main():
    files = [
        "internal/handlers/rsvp_test.go",
        "internal/handlers/color_override_integration_test.go",
        "internal/handlers/invites_delete_test.go",
        "internal/handlers/invites_get_test.go",
        "internal/handlers/invites_list_test.go",
        "internal/handlers/invites_regenerate_test.go",
        "internal/handlers/invites_revoke_test.go",
        "internal/handlers/invites_send_test.go",
        "internal/handlers/invites_update_test.go",
        "internal/handlers/rsvp_summary_test.go",
        "internal/handlers/invites_import_permission_test.go",
    ]
    
    updated_count = 0
    for filepath in files:
        if process_file(filepath):
            updated_count += 1
    
    print(f"\nDone! Updated {updated_count} files.")

if __name__ == "__main__":
    main()

#!/usr/bin/env python3
"""Add Category field to Template structs."""

import re
import sys

def fix_template_structs(content):
    """Add Category: models.CategoryPlain to Template structs that don't have it."""
    
    # Find all models.Template{ ... } blocks
    pattern = r'(models\.Template\{)([^}]+)(\})'
    
    def add_category(match):
        opening = match.group(1)
        body = match.group(2)
        closing = match.group(3)
        
        # Check if Category already exists
        if 'Category:' in body:
            return match.group(0)
        
        # Find the last field and add Category before closing
        # Split by lines and find last non-empty line
        lines = body.split('\n')
        
        # Find the last line with content (not just whitespace)
        last_field_idx = -1
        for i in range(len(lines) - 1, -1, -1):
            if lines[i].strip() and not lines[i].strip().startswith('//'):
                last_field_idx = i
                break
        
        if last_field_idx >= 0:
            # Get indentation from last field line
            last_line = lines[last_field_idx]
            indent = len(last_line) - len(last_line.lstrip())
            indent_str = '\t' * (indent // 4) if '\t' in last_line else ' ' * indent
            
            # Add Category field after last field
            category_line = f"{indent_str}Category:    models.CategoryPlain,"
            lines.insert(last_field_idx + 1, category_line)
        
        new_body = '\n'.join(lines)
        return f"{opening}{new_body}{closing}"
    
    return re.sub(pattern, add_category, content, flags=re.DOTALL)

def main():
    files = [
        'internal/templates/service_test.go',
        'internal/templates/service_integration_test.go',
        'internal/templates/css_sanitizer_integration_test.go',
        'internal/templates/validator_test.go',
        'internal/templates/validator_integration_test.go',
        'internal/templates/xss_integration_test.go',
    ]
    
    for filepath in files:
        try:
            with open(filepath, 'r') as f:
                content = f.read()
            
            modified = fix_template_structs(content)
            
            with open(filepath, 'w') as f:
                f.write(modified)
            
            print(f"✓ Processed: {filepath}")
        except Exception as e:
            print(f"✗ Error processing {filepath}: {e}", file=sys.stderr)
            return 1
    
    return 0

if __name__ == '__main__':
    sys.exit(main())

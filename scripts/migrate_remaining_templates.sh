#!/bin/bash

# This script migrates the remaining template files to use the base template system
# It creates backups and can be reverted if needed

echo "Creating backups..."
cp templates/web/event_detail.html templates/web/event_detail.html.backup
cp templates/web/invite_list.html templates/web/invite_list.html.backup
cp templates/web/user_management.html templates/web/user_management.html.backup
cp templates/web/rsvp_summary.html templates/web/rsvp_summary.html.backup

echo "Backups created. Proceeding with migrations..."
echo "Note: This is a placeholder script. Actual migrations need to be done carefully."
echo "Remaining pages to migrate:"
echo "  - event_detail.html (187 lines)"
echo "  - invite_list.html (406 lines)"
echo "  - user_management.html (131 lines)"
echo "  - rsvp_summary.html (230 lines)"
echo ""
echo "Each page needs:"
echo "  1. Remove <!DOCTYPE>, <html>, <head>, <body> boilerplate"
echo "  2. Add {{template \"base\" .}} at top"
echo "  3. Define {{define \"title\"}}...{{end}}"
echo "  4. Define {{define \"content\"}}...{{end}}"
echo "  5. Move page-specific CSS to {{define \"css-extra\"}}...{{end}}"
echo "  6. Move page-specific JS to {{define \"js-extra\"}}...{{end}}"
echo ""
echo "Template loading in main.go also needs updating to include:"
echo "  - templates/web/partials/base.html"
echo "  - templates/web/partials/datetime_picker_panel.html (if needed)"
echo ""
echo "To restore backups if needed:"
echo "  mv templates/web/*.backup templates/web/"

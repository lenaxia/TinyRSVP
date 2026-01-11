#!/usr/bin/env python3

import re

# Read the file
with open('cmd/server/main.go', 'r') as f:
    content = f.read()

# Update event web templates to include datetime_picker_panel
content = re.sub(
    r'(eventWebTemplates, err := template\.New\("events"\)\.Funcs\(funcMap\)\.ParseFiles\(\s*"templates/web/partials/base\.html",\s*"templates/web/partials/navigation\.html",\s*"templates/web/partials/theme_picker\.html",)',
    r'eventWebTemplates, err := template.New("events").Funcs(funcMap).ParseFiles(\n\t\t"templates/web/partials/base.html",\n\t\t"templates/web/partials/navigation.html",\n\t\t"templates/web/partials/datetime_picker_panel.html",\n\t\t"templates/web/partials/theme_picker.html",',
    content
)

# Write the file back
with open('cmd/server/main.go', 'w') as f:
    f.write(content)

print("Template loading updated to include datetime_picker_panel.html")

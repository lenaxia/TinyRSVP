#!/usr/bin/env python3

import re

# Read the file
with open('cmd/server/main.go', 'r') as f:
    content = f.read()

# Update dashboard templates
content = re.sub(
    r'(dashboardTemplates, err := template\.New\("dashboard\.html"\)\.ParseFiles\(\s*"templates/web/partials/navigation\.html",)',
    r'dashboardTemplates, err := template.New("dashboard.html").ParseFiles(\n\t\t"templates/web/partials/base.html",\n\t\t"templates/web/partials/navigation.html",',
    content
)

# Update event web templates
content = re.sub(
    r'(eventWebTemplates, err := template\.New\("events"\)\.Funcs\(funcMap\)\.ParseFiles\(\s*"templates/web/partials/navigation\.html",)',
    r'eventWebTemplates, err := template.New("events").Funcs(funcMap).ParseFiles(\n\t\t"templates/web/partials/base.html",\n\t\t"templates/web/partials/navigation.html",',
    content
)

# Update admin dashboard templates
content = re.sub(
    r'(adminDashboardTemplates, err := template\.New\("admin_dashboard\.html"\)\.ParseFiles\(\s*"templates/web/partials/navigation\.html",)',
    r'adminDashboardTemplates, err := template.New("admin_dashboard.html").ParseFiles(\n\t\t"templates/web/partials/base.html",\n\t\t"templates/web/partials/navigation.html",',
    content
)

# Write the file back
with open('cmd/server/main.go', 'w') as f:
    f.write(content)

print("Template loading updated successfully")

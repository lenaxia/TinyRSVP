#!/bin/bash

# Update dashboard templates
sed -i '/dashboardTemplates, err := template.New("dashboard.html").ParseFiles(/,/)/ {
    /ParseFiles(/a\
\t\t"templates/web/partials/base.html",
}' cmd/server/main.go

# Update event web templates  
sed -i '/eventWebTemplates, err := template.New("events").Funcs(funcMap).ParseFiles(/,/)/ {
    /ParseFiles(/a\
\t\t"templates/web/partials/base.html",
}' cmd/server/main.go

# Update admin dashboard templates
sed -i '/adminDashboardTemplates, err := template.New("admin_dashboard.html").ParseFiles(/,/)/ {
    /ParseFiles(/a\
\t\t"templates/web/partials/base.html",
}' cmd/server/main.go

echo "Template loading updated successfully"

# RSVP Theme Templates

This directory contains HTML templates for RSVP page themes.

## Structure

```
rsvp_themes/
├── plain-text.html                 # Plain text theme (no header image)
├── wedding-elegance.html           # Wedding theme
├── birthday-celebration.html       # Birthday theme
├── corporate-professional.html     # Corporate theme
├── holiday-festive.html            # Holiday theme
├── garden-party.html               # Garden theme
├── modern-minimalist.html          # Modern theme
└── theme_templates_test.go         # Template validation tests
```

## Template Structure

All templates follow this structure:

```html
<!DOCTYPE html>
<html lang="en" data-event-theme="theme-name">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>RSVP - {{.Event.Title}}</title>
    <link rel="stylesheet" href="/static/css/variables.css">
    <link rel="stylesheet" href="/static/css/themes/theme-name.css">
    <script src="/static/js/theme_controller.js" defer></script>
</head>
<body>
    <div class="rsvp-container">
        <div class="rsvp-card">
            <!-- Card-based themes only: -->
            <div class="rsvp-card-header">
                <img src="/static/images/themes/theme-name-header.svg" 
                     alt="Theme description"
                     class="theme-header-image">
            </div>
            
            <div class="rsvp-card-content">
                <!-- Event details and RSVP form -->
            </div>
        </div>
    </div>
</body>
</html>
```

## Go Template Variables

### Event Data

- `{{.Event.Title}}` - Event title
- `{{.Event.StartTime}}` - Event start time
- `{{.Event.EndTime}}` - Event end time (optional)
- `{{.Event.Location}}` - Event location (optional)
- `{{.Event.Timezone}}` - Event timezone (optional)
- `{{.Event.Description}}` - Event description (optional)
- `{{.Event.RSVPDeadline}}` - RSVP deadline (optional)
- `{{.Event.CustomThemeImageURL}}` - Custom header image (optional)

### Form Data

- `{{.Token}}` - Guest invite token
- `{{.CSRFToken}}` - CSRF protection token
- `{{.MaxPlusOnes}}` - Maximum additional guests
- `{{.Questions}}` - Preference questions array

### Template Functions

- `{{formatDateTime .Event.StartTime}}` - Format date and time
- `{{formatTime .Event.EndTime}}` - Format time only
- `{{formatDate .Event.RSVPDeadline "Monday, January 2, 2006"}}` - Format date
- `{{if gt .MaxPlusOnes 0}}` - Conditional rendering
- `{{range .Questions}}` - Loop over questions

## Testing

Run template validation tests:

```bash
cd templates/web/rsvp_themes
go test -timeout 30s -v
```

Tests verify:
- All templates exist
- Required HTML structure present
- Go template variables present
- Correct CSS file linked
- Correct data-event-theme attribute
- Card-based themes have header images
- Plain text theme has no header image

## Adding New Theme Template

1. Copy existing theme template
2. Update `data-event-theme` attribute
3. Update CSS file link
4. Update header image path (if card-based)
5. Update alt text
6. Update `theme_templates_test.go`
7. Run tests

## Design Guidelines

See [docs/THEME_DESIGN_SYSTEM.md](../../../docs/THEME_DESIGN_SYSTEM.md) for complete specifications.

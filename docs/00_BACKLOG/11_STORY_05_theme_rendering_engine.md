# User Story 11.05: Theme Rendering Engine

**Epic:** [11_EPIC_rsvp_themes.md](11_EPIC_rsvp_themes.md)  
**Priority:** High  
**Status:** Not Started  
**Estimated Effort:** 2 days  
**Owner:** Unassigned

---

## User Story

As a **guest**,  
I want to **see the RSVP page styled with the event manager's selected theme**,  
So that **I have a visually appealing and branded invitation experience**.

---

## Context

This story implements the core rendering logic that applies the selected theme to RSVP pages when guests visit. The rendering engine must:
- Load the correct theme template for the event
- Apply theme-specific CSS and images
- Respect guest's light/dark mode preference
- Handle custom theme overrides (images, colors)
- Work on all devices and browsers

---

## Acceptance Criteria

### Theme Loading
- [ ] RSVP handler loads event's selected theme
- [ ] Falls back to default theme if none selected
- [ ] Handles missing theme gracefully
- [ ] Loads theme CSS file
- [ ] Loads theme images

### Theme Application
- [ ] Theme HTML template rendered with event data
- [ ] Theme CSS variables applied via data attribute
- [ ] Theme images displayed correctly
- [ ] Custom theme image used if provided
- [ ] Custom theme color used if provided

### Light/Dark Mode Support
- [ ] Guest's light/dark preference detected
- [ ] Theme adapts to guest's preference
- [ ] Theme controller script included
- [ ] Theme toggle button available
- [ ] Preference persists across page loads

### Data Binding
- [ ] Event title rendered
- [ ] Event date/time rendered with timezone
- [ ] Event location rendered
- [ ] Event description rendered (Markdown supported)
- [ ] Preference questions rendered
- [ ] RSVP form rendered
- [ ] All template variables work

### Performance
- [ ] Page loads in <2 seconds
- [ ] Images lazy loaded
- [ ] CSS minified (production)
- [ ] No layout shift during load
- [ ] Caching headers set appropriately

### Error Handling
- [ ] Missing theme → use default
- [ ] Invalid theme ID → use default
- [ ] Missing theme image → show placeholder
- [ ] Template rendering error → show error page
- [ ] Graceful degradation

---

## Technical Details

### RSVP Handler Update

**File:** `internal/handlers/rsvp_web.go`

```go
func (h *Handler) HandleRSVPPage(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    // Extract and validate token
    token := chi.URLParam(r, "token")
    if token == "" {
        HandleError(w, r, &models.ValidationError{
            Field:   "token",
            Message: "Invalid invite link",
        })
        return
    }
    
    // Validate token and get invite
    invite, err := h.inviteService.ValidateToken(ctx, token)
    if err != nil {
        HandleError(w, r, err)
        return
    }
    
    // Get event
    event, err := h.eventService.GetEvent(ctx, invite.EventID)
    if err != nil {
        HandleError(w, r, err)
        return
    }
    
    // Get theme template
    theme, err := h.getEventTheme(ctx, event)
    if err != nil {
        HandleError(w, r, err)
        return
    }
    
    // Get preference questions
    questions, err := h.eventService.GetPreferenceQuestions(ctx, event.ID)
    if err != nil {
        HandleError(w, r, err)
        return
    }
    
    // Get existing RSVP if any
    existingRSVP, _ := h.rsvpService.GetRSVPByInvite(ctx, invite.ID)
    
    // Prepare template data
    data := RSVPPageData{
        Event:             event,
        Invite:            invite,
        Questions:         questions,
        ExistingRSVP:      existingRSVP,
        ThemeCategory:     theme.Category,
        ThemeImageURL:     h.getThemeImageURL(event, theme),
        ThemeColor:        h.getThemeColor(event, theme),
        CSRFToken:         GetCSRFToken(r),
    }
    
    // Render with theme
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    
    // Add theme data attributes
    html := fmt.Sprintf(`<!DOCTYPE html><html lang="en" data-event-theme="%s">`, theme.Category)
    w.Write([]byte(html))
    
    if err := h.renderer.RenderToWriter(w, theme.HTMLContent, data); err != nil {
        HandleError(w, r, fmt.Errorf("failed to render theme: %w", err))
        return
    }
    
    w.Write([]byte(`</html>`))
}

func (h *Handler) getEventTheme(ctx context.Context, event *models.Event) (*models.Template, error) {
    // Use event's selected theme if set
    if event.TemplateID != nil {
        theme, err := h.templateService.GetTemplate(ctx, *event.TemplateID)
        if err == nil {
            return theme, nil
        }
        // Log error but continue to default
        log.Printf("Failed to load event theme %d: %v", *event.TemplateID, err)
    }
    
    // Fall back to default RSVP page theme
    return h.templateService.GetDefaultTemplate(ctx, models.TemplateTypeRSVPPage)
}

func (h *Handler) getThemeImageURL(event *models.Event, theme *models.Template) string {
    // Custom image takes precedence
    if event.CustomThemeImageURL != nil && *event.CustomThemeImageURL != "" {
        return *event.CustomThemeImageURL
    }
    
    // Use theme's default image
    if theme.ImageURL != nil {
        return *theme.ImageURL
    }
    
    return ""
}

func (h *Handler) getThemeColor(event *models.Event, theme *models.Template) string {
    // Custom color takes precedence
    if event.CustomThemeColor != nil && *event.CustomThemeColor != "" {
        return *event.CustomThemeColor
    }
    
    return ""
}
```

### Template Data Structure

**File:** `internal/handlers/rsvp_data.go`

```go
type RSVPPageData struct {
    Event             *models.Event
    Invite            *models.Invite
    Questions         []*models.PreferenceQuestion
    ExistingRSVP      *models.RSVP
    ThemeCategory     string
    ThemeImageURL     string
    ThemeColor        string
    CSRFToken         string
}
```

### Theme CSS Loading

**Update theme templates to include CSS:**

```html
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>RSVP: {{.Event.Title}}</title>
    
    <!-- Base styles -->
    <link rel="stylesheet" href="/static/css/variables.css">
    <link rel="stylesheet" href="/static/css/reset.css">
    <link rel="stylesheet" href="/static/css/typography.css">
    <link rel="stylesheet" href="/static/css/forms.css">
    <link rel="stylesheet" href="/static/css/buttons.css">
    
    <!-- Theme-specific CSS -->
    <link rel="stylesheet" href="/static/css/themes/{{.ThemeCategory}}.css">
    
    <!-- Custom color override if provided -->
    {{if .ThemeColor}}
    <style>
        [data-event-theme] {
            --theme-primary: {{.ThemeColor}} !important;
        }
    </style>
    {{end}}
    
    <!-- Theme controller for light/dark switching -->
    <script src="/static/js/theme_controller.js" defer></script>
</head>
```

---

## Tasks

### Handler Updates
- [ ] Update `HandleRSVPPage` to load theme
- [ ] Implement `getEventTheme()` helper
- [ ] Implement `getThemeImageURL()` helper
- [ ] Implement `getThemeColor()` helper
- [ ] Add fallback to default theme
- [ ] Handle theme loading errors
- [ ] Write handler tests

### Data Structure
- [ ] Create `RSVPPageData` struct
- [ ] Add theme-related fields
- [ ] Document struct fields
- [ ] Write validation tests

### Template Updates
- [ ] Update theme templates to include CSS links
- [ ] Add data-event-theme attribute
- [ ] Add custom color override style block
- [ ] Include theme controller script
- [ ] Test template rendering

### Error Handling
- [ ] Handle missing theme gracefully
- [ ] Handle invalid theme ID
- [ ] Handle missing theme images
- [ ] Handle template rendering errors
- [ ] Log errors appropriately
- [ ] Write error handling tests

### Performance Optimization
- [ ] Add caching headers for theme assets
- [ ] Lazy load theme images
- [ ] Minify CSS in production
- [ ] Test page load times
- [ ] Optimize critical rendering path

### Testing
- [ ] Unit tests for helper functions
- [ ] Unit tests for theme loading
- [ ] Integration tests for RSVP rendering
- [ ] Test with each theme
- [ ] Test in light and dark modes
- [ ] Test with custom overrides
- [ ] Test error scenarios
- [ ] Test on mobile/tablet/desktop

---

## Definition of Done

- [ ] All acceptance criteria met
- [ ] RSVP handler updated
- [ ] Theme loading implemented
- [ ] Custom overrides supported
- [ ] Error handling implemented
- [ ] All unit tests passing
- [ ] All integration tests passing
- [ ] Performance targets met
- [ ] Mobile-responsive
- [ ] Changes committed to git

---

## Dependencies

**Depends on:**
- Story 11.01: Theme Model Extension
- Story 11.02: Theme Asset Creation
- Story 11.03: Theme Picker UI
- Story 11.04: Theme Preview Modal

**Blocks:**
- Story 11.06: Theme Seeding System
- Story 11.07: Theme Integration Testing

---

## References

- **Epic:** [11_EPIC_rsvp_themes.md](11_EPIC_rsvp_themes.md)
- **Analysis:** [11_ANALYSIS_rsvp_page_themes.md](11_ANALYSIS_rsvp_page_themes.md)
- **RSVP Handler:** `internal/handlers/rsvp_web.go`
- **Template Service:** `internal/templates/service.go`
- **Current RSVP Page:** `templates/web/rsvp_page.html`

---

## Notes

- Theme rendering is the core of the theme system
- Must be fast and reliable
- Graceful degradation is critical
- Consider caching rendered themes (v1)
- Monitor performance metrics
- Log theme loading for debugging

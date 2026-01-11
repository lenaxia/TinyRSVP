# User Story 11.07: Theme Integration Testing

**Epic:** [11_EPIC_rsvp_themes.md](11_EPIC_rsvp_themes.md)  
**Priority:** High  
**Status:** Not Started  
**Estimated Effort:** 2-3 days  
**Owner:** Unassigned

---

## User Story

As a **quality assurance engineer**,  
I want **comprehensive integration tests for the theme system**,  
So that **we can confidently deploy themes knowing they work correctly across all scenarios**.

---

## Context

The theme system involves multiple components working together:
- Theme model and database
- Theme picker UI
- Theme preview modal
- Theme rendering engine
- Light/dark mode integration
- Event creation flow

Integration tests must verify the complete end-to-end flow and all component interactions.

---

## Acceptance Criteria

### End-to-End Flow Tests
- [ ] Test complete flow: select theme → preview → create event → view RSVP
- [ ] Test theme selection persists to database
- [ ] Test RSVP page renders with correct theme
- [ ] Test theme works in light mode
- [ ] Test theme works in dark mode
- [ ] Test theme switching between light/dark

### Theme Picker Tests
- [ ] Test theme gallery displays all themes
- [ ] Test theme filtering by category
- [ ] Test theme selection updates form
- [ ] Test keyboard navigation
- [ ] Test screen reader compatibility
- [ ] Test mobile responsiveness

### Theme Preview Tests
- [ ] Test preview modal opens
- [ ] Test preview loads correct theme
- [ ] Test preview uses event form data
- [ ] Test preview theme toggle
- [ ] Test preview select button
- [ ] Test preview close functionality

### Theme Rendering Tests
- [ ] Test each theme renders correctly
- [ ] Test theme with minimal event data
- [ ] Test theme with complete event data
- [ ] Test theme with long text content
- [ ] Test theme with many preference questions
- [ ] Test custom image override
- [ ] Test custom color override

### Cross-Browser Tests
- [ ] Test on Chrome/Chromium
- [ ] Test on Firefox
- [ ] Test on Safari (if available)
- [ ] Test on mobile browsers
- [ ] Test on different screen sizes

### Performance Tests
- [ ] Test page load time <2 seconds
- [ ] Test theme switching performance
- [ ] Test with slow network (throttling)
- [ ] Test memory usage
- [ ] Test no memory leaks

### Accessibility Tests
- [ ] Test keyboard-only navigation
- [ ] Test screen reader announcements
- [ ] Test focus management
- [ ] Test color contrast (WCAG AA)
- [ ] Test with zoom (200%)
- [ ] Test with high contrast mode

### Error Scenario Tests
- [ ] Test with missing theme
- [ ] Test with invalid theme ID
- [ ] Test with missing theme images
- [ ] Test with corrupted theme data
- [ ] Test database errors
- [ ] Test network errors

---

## Technical Details

### Integration Test Structure

**File:** `internal/handlers/theme_integration_test.go`

```go
package handlers_test

import (
    "context"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"
    
    "github.com/lenaxia/tinyrsvp/internal/models"
)

func TestThemeSystemIntegration(t *testing.T) {
    // Setup test environment
    db := setupTestDB(t)
    defer db.Close()
    
    handler := setupTestHandler(t, db)
    
    // Seed themes
    seeder := templates.NewSeeder(handler.templateService)
    if err := seeder.SeedThemes(context.Background()); err != nil {
        t.Fatalf("Failed to seed themes: %v", err)
    }
    
    t.Run("complete theme selection flow", func(t *testing.T) {
        // 1. Get event creation page with themes
        req := httptest.NewRequest("GET", "/events/new", nil)
        req = addAuthContext(req, testAdmin)
        w := httptest.NewRecorder()
        
        handler.HandleEventCreatePage(w, req)
        
        if w.Code != http.StatusOK {
            t.Errorf("Expected 200, got %d", w.Code)
        }
        
        // Verify themes in response
        body := w.Body.String()
        if !strings.Contains(body, "Wedding Elegance") {
            t.Error("Wedding theme not found in page")
        }
        
        // 2. Create event with selected theme
        themeID := getThemeIDByName(t, db, "Wedding Elegance")
        
        formData := url.Values{
            "title":       {"Test Wedding"},
            "start_time":  {time.Now().Add(24 * time.Hour).Format(time.RFC3339)},
            "timezone":    {"America/Los_Angeles"},
            "template_id": {fmt.Sprintf("%d", themeID)},
        }
        
        req = httptest.NewRequest("POST", "/events", strings.NewReader(formData.Encode()))
        req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
        req = addAuthContext(req, testAdmin)
        w = httptest.NewRecorder()
        
        handler.HandleEventCreate(w, req)
        
        if w.Code != http.StatusFound {
            t.Errorf("Expected 302, got %d", w.Code)
        }
        
        // 3. Get event from database
        eventID := extractEventIDFromLocation(t, w.Header().Get("Location"))
        event, err := handler.eventService.GetEvent(context.Background(), eventID)
        if err != nil {
            t.Fatalf("Failed to get event: %v", err)
        }
        
        // Verify theme selected
        if event.TemplateID == nil || *event.TemplateID != themeID {
            t.Errorf("Expected theme ID %d, got %v", themeID, event.TemplateID)
        }
        
        // 4. Create invite for event
        invite, token := createTestInvite(t, handler, event.ID)
        
        // 5. View RSVP page as guest
        req = httptest.NewRequest("GET", "/rsvp/"+token, nil)
        w = httptest.NewRecorder()
        
        handler.HandleRSVPPage(w, req)
        
        if w.Code != http.StatusOK {
            t.Errorf("Expected 200, got %d", w.Code)
        }
        
        // Verify theme applied
        body = w.Body.String()
        if !strings.Contains(body, `data-event-theme="card"`) {
            t.Error("Theme not applied to RSVP page")
        }
        if !strings.Contains(body, "wedding-elegance-header.jpg") {
            t.Error("Theme image not found in RSVP page")
        }
    })
}

func TestThemePreviewIntegration(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    
    handler := setupTestHandler(t, db)
    
    // Seed themes
    seeder := templates.NewSeeder(handler.templateService)
    seeder.SeedThemes(context.Background())
    
    t.Run("preview endpoint returns themed page", func(t *testing.T) {
        themeID := getThemeIDByName(t, db, "Birthday Celebration")
        
        req := httptest.NewRequest("GET", fmt.Sprintf("/api/themes/preview?theme_id=%d&title=Test+Party&theme_mode=light", themeID), nil)
        req = addAuthContext(req, testAdmin)
        w := httptest.NewRecorder()
        
        handler.HandleThemePreview(w, req)
        
        if w.Code != http.StatusOK {
            t.Errorf("Expected 200, got %d", w.Code)
        }
        
        body := w.Body.String()
        
        // Verify theme applied
        if !strings.Contains(body, `data-event-theme="card"`) {
            t.Error("Theme not applied in preview")
        }
        
        // Verify event data used
        if !strings.Contains(body, "Test Party") {
            t.Error("Event title not found in preview")
        }
        
        // Verify light mode
        if !strings.Contains(body, `data-theme="light"`) {
            t.Error("Light mode not applied in preview")
        }
    })
    
    t.Run("preview with dark mode", func(t *testing.T) {
        themeID := getThemeIDByName(t, db, "Birthday Celebration")
        
        req := httptest.NewRequest("GET", fmt.Sprintf("/api/themes/preview?theme_id=%d&theme_mode=dark", themeID), nil)
        req = addAuthContext(req, testAdmin)
        w := httptest.NewRecorder()
        
        handler.HandleThemePreview(w, req)
        
        body := w.Body.String()
        
        // Verify dark mode
        if !strings.Contains(body, `data-theme="dark"`) {
            t.Error("Dark mode not applied in preview")
        }
    })
}

func TestAllThemesRender(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    
    handler := setupTestHandler(t, db)
    
    // Seed themes
    seeder := templates.NewSeeder(handler.templateService)
    seeder.SeedThemes(context.Background())
    
    // Get all themes
    themes, err := handler.templateService.ListThemes(context.Background(), models.TemplateTypeRSVPPage, nil)
    if err != nil {
        t.Fatalf("Failed to list themes: %v", err)
    }
    
    // Test each theme
    for _, theme := range themes {
        t.Run(theme.Name, func(t *testing.T) {
            // Test in light mode
            t.Run("light mode", func(t *testing.T) {
                testThemeRendering(t, handler, theme.ID, "light")
            })
            
            // Test in dark mode
            t.Run("dark mode", func(t *testing.T) {
                testThemeRendering(t, handler, theme.ID, "dark")
            })
        })
    }
}

func testThemeRendering(t *testing.T, handler *Handler, themeID int64, mode string) {
    req := httptest.NewRequest("GET", fmt.Sprintf("/api/themes/preview?theme_id=%d&theme_mode=%s", themeID, mode), nil)
    w := httptest.NewRecorder()
    
    handler.HandleThemePreview(w, req)
    
    if w.Code != http.StatusOK {
        t.Errorf("Expected 200, got %d", w.Code)
    }
    
    body := w.Body.String()
    
    // Verify basic structure
    if !strings.Contains(body, "<!DOCTYPE html>") {
        t.Error("Missing DOCTYPE")
    }
    if !strings.Contains(body, `data-theme="`+mode+`"`) {
        t.Errorf("Theme mode %s not applied", mode)
    }
    
    // Verify no rendering errors
    if strings.Contains(body, "template error") {
        t.Error("Template rendering error found")
    }
}
```

### Visual Regression Tests

**File:** `templates/web/theme_visual_test.go`

```go
package web_test

import (
    "testing"
)

func TestThemeVisualRegression(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping visual regression tests in short mode")
    }
    
    themes := []string{
        "plain-text",
        "wedding-elegance",
        "birthday-celebration",
        "corporate-professional",
        "holiday-festive",
        "garden-party",
        "modern-minimalist",
    }
    
    modes := []string{"light", "dark"}
    viewports := []struct {
        name   string
        width  int
        height int
    }{
        {"mobile", 375, 667},
        {"tablet", 768, 1024},
        {"desktop", 1920, 1080},
    }
    
    for _, theme := range themes {
        for _, mode := range modes {
            for _, viewport := range viewports {
                t.Run(fmt.Sprintf("%s_%s_%s", theme, mode, viewport.name), func(t *testing.T) {
                    // Take screenshot
                    // Compare with baseline
                    // Report differences
                    testVisualRegression(t, theme, mode, viewport.width, viewport.height)
                })
            }
        }
    }
}
```

---

## Tasks

### Integration Test Suite
- [ ] Create `theme_integration_test.go`
- [ ] Test complete theme selection flow
- [ ] Test theme preview flow
- [ ] Test theme rendering on RSVP pages
- [ ] Test all 7 themes render correctly
- [ ] Test light/dark mode for each theme
- [ ] Test custom overrides
- [ ] Test error scenarios

### Visual Regression Tests
- [ ] Create `theme_visual_test.go`
- [ ] Set up screenshot comparison
- [ ] Create baseline screenshots
- [ ] Test each theme in light/dark
- [ ] Test on mobile/tablet/desktop
- [ ] Document visual test process

### Performance Tests
- [ ] Test page load times
- [ ] Test theme switching performance
- [ ] Test with network throttling
- [ ] Test memory usage
- [ ] Test concurrent users

### Accessibility Tests
- [ ] Test keyboard navigation
- [ ] Test screen reader compatibility
- [ ] Test color contrast
- [ ] Test focus management
- [ ] Test ARIA attributes
- [ ] Run automated accessibility audit

### Cross-Browser Tests
- [ ] Test on Chrome
- [ ] Test on Firefox
- [ ] Test on Safari (if available)
- [ ] Test on mobile browsers
- [ ] Document browser compatibility

### Error Scenario Tests
- [ ] Test missing theme
- [ ] Test invalid theme ID
- [ ] Test missing theme images
- [ ] Test database errors
- [ ] Test network errors
- [ ] Test graceful degradation

### Documentation
- [ ] Document test setup
- [ ] Document how to run tests
- [ ] Document test coverage
- [ ] Document known issues
- [ ] Create testing checklist

---

## Definition of Done

- [ ] All acceptance criteria met
- [ ] Integration test suite created
- [ ] Visual regression tests created
- [ ] Performance tests created
- [ ] Accessibility tests created
- [ ] All tests passing
- [ ] Test coverage >80%
- [ ] Documentation complete
- [ ] Changes committed to git

---

## Dependencies

**Depends on:**
- Story 11.01: Theme Model Extension
- Story 11.02: Theme Asset Creation
- Story 11.03: Theme Picker UI
- Story 11.04: Theme Preview Modal
- Story 11.05: Theme Rendering Engine
- Story 11.06: Theme Seeding System

**Blocks:**
- Phase 1 completion
- Phase 2 stories

---

## References

- **Epic:** [11_EPIC_rsvp_themes.md](11_EPIC_rsvp_themes.md)
- **Analysis:** [11_ANALYSIS_rsvp_page_themes.md](11_ANALYSIS_rsvp_page_themes.md)
- **Testing Guide:** README-LLM.md Section 10

---

## Notes

- Integration tests are critical for theme system confidence
- Visual regression tests catch UI regressions
- Performance tests ensure no degradation
- Accessibility tests ensure WCAG compliance
- Document test failures for debugging
- Consider CI/CD integration

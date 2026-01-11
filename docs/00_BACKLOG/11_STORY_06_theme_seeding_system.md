# User Story 11.06: Theme Seeding System

**Epic:** [11_EPIC_rsvp_themes.md](11_EPIC_rsvp_themes.md)  
**Priority:** High  
**Status:** Not Started  
**Estimated Effort:** 1 day  
**Owner:** Unassigned

---

## User Story

As a **system administrator**,  
I want **themes automatically seeded into the database on first startup**,  
So that **event managers have themes available immediately without manual setup**.

---

## Context

The 7 pre-designed themes need to be inserted into the database when the application first starts or when new themes are added. This should be:
- Automatic (no manual SQL required)
- Idempotent (safe to run multiple times)
- Version-aware (only seed new themes)
- Testable

---

## Acceptance Criteria

### Seeding Mechanism
- [ ] Themes seeded on application startup
- [ ] Seeding is idempotent (safe to run multiple times)
- [ ] Existing themes not duplicated
- [ ] Existing themes updated if changed
- [ ] New themes added automatically
- [ ] Seeding completes in <5 seconds

### Theme Data
- [ ] All 7 themes seeded (1 plain + 6 card)
- [ ] Theme metadata complete (name, description, category, tags)
- [ ] Image URLs correct
- [ ] Sort order set appropriately
- [ ] One theme marked as default
- [ ] Created_by is NULL (system themes)

### Validation
- [ ] Seeded themes pass validation
- [ ] Image URLs point to existing files
- [ ] CSS files exist for all themes
- [ ] HTML templates exist for all themes
- [ ] No broken references

### Error Handling
- [ ] Database errors logged and reported
- [ ] Missing files logged as warnings
- [ ] Seeding failure doesn't prevent startup (uses existing themes)
- [ ] Partial seeding handled gracefully

### Testing
- [ ] Unit tests for seeding logic
- [ ] Integration tests with clean database
- [ ] Integration tests with existing themes
- [ ] Test idempotency (run twice)
- [ ] Test theme updates
- [ ] Test error scenarios

---

## Technical Details

### Seeding Service

**File:** `internal/templates/seeder.go`

```go
package templates

import (
    "context"
    "fmt"
    "log"
    
    "github.com/lenaxia/tinyrsvp/internal/models"
)

type Seeder struct {
    service Service
}

func NewSeeder(service Service) *Seeder {
    return &Seeder{service: service}
}

func (s *Seeder) SeedThemes(ctx context.Context) error {
    themes := s.getDefaultThemes()
    
    for _, theme := range themes {
        if err := s.seedTheme(ctx, theme); err != nil {
            log.Printf("Warning: Failed to seed theme %s: %v", theme.Name, err)
            // Continue with other themes
        }
    }
    
    return nil
}

func (s *Seeder) seedTheme(ctx context.Context, theme *models.Template) error {
    // Check if theme already exists by name and type
    existing, err := s.service.GetTemplateByName(ctx, theme.Name, theme.Type)
    if err == nil && existing != nil {
        // Theme exists, update if changed
        theme.ID = existing.ID
        return s.service.UpdateTemplate(ctx, theme)
    }
    
    // Theme doesn't exist, create it
    return s.service.CreateTemplate(ctx, theme)
}

func (s *Seeder) getDefaultThemes() []*models.Template {
    return []*models.Template{
        {
            Name:         "Simple & Clean",
            Type:         models.TemplateTypeRSVPPage,
            Category:     models.CategoryPlain,
            Description:  "Minimalist text-based invitation, perfect for accessibility and fast loading",
            HTMLContent:  s.loadTemplate("plain-text.html"),
            CSSContent:   stringPtr(s.loadCSS("plain-text.css")),
            ThumbnailURL: stringPtr("/static/images/themes/plain-text-thumb.jpg"),
            ImageURL:     nil,
            Tags:         []string{"accessible", "minimal", "text-only"},
            SortOrder:    0,
            IsDefault:    true,
        },
        {
            Name:         "Wedding Elegance",
            Type:         models.TemplateTypeRSVPPage,
            Category:     models.CategoryCard,
            Description:  "Elegant floral design perfect for weddings and formal celebrations",
            HTMLContent:  s.loadTemplate("wedding-elegance.html"),
            CSSContent:   stringPtr(s.loadCSS("wedding-elegance.css")),
            ThumbnailURL: stringPtr("/static/images/themes/wedding-elegance-thumb.jpg"),
            ImageURL:     stringPtr("/static/images/themes/wedding-elegance-header.jpg"),
            Tags:         []string{"wedding", "formal", "floral", "elegant"},
            SortOrder:    1,
            IsDefault:    false,
        },
        {
            Name:         "Birthday Celebration",
            Type:         models.TemplateTypeRSVPPage,
            Category:     models.CategoryCard,
            Description:  "Fun and colorful design for birthday parties and celebrations",
            HTMLContent:  s.loadTemplate("birthday-celebration.html"),
            CSSContent:   stringPtr(s.loadCSS("birthday-celebration.css")),
            ThumbnailURL: stringPtr("/static/images/themes/birthday-celebration-thumb.jpg"),
            ImageURL:     stringPtr("/static/images/themes/birthday-celebration-header.jpg"),
            Tags:         []string{"birthday", "celebration", "fun", "colorful"},
            SortOrder:    2,
            IsDefault:    false,
        },
        {
            Name:         "Corporate Professional",
            Type:         models.TemplateTypeRSVPPage,
            Category:     models.CategoryCard,
            Description:  "Clean and professional design for business events and meetings",
            HTMLContent:  s.loadTemplate("corporate-professional.html"),
            CSSContent:   stringPtr(s.loadCSS("corporate-professional.css")),
            ThumbnailURL: stringPtr("/static/images/themes/corporate-professional-thumb.jpg"),
            ImageURL:     stringPtr("/static/images/themes/corporate-professional-header.jpg"),
            Tags:         []string{"corporate", "professional", "business", "formal"},
            SortOrder:    3,
            IsDefault:    false,
        },
        {
            Name:         "Holiday Festive",
            Type:         models.TemplateTypeRSVPPage,
            Category:     models.CategoryCard,
            Description:  "Warm and festive design for holiday gatherings and seasonal events",
            HTMLContent:  s.loadTemplate("holiday-festive.html"),
            CSSContent:   stringPtr(s.loadCSS("holiday-festive.css")),
            ThumbnailURL: stringPtr("/static/images/themes/holiday-festive-thumb.jpg"),
            ImageURL:     stringPtr("/static/images/themes/holiday-festive-header.jpg"),
            Tags:         []string{"holiday", "festive", "seasonal", "warm"},
            SortOrder:    4,
            IsDefault:    false,
        },
        {
            Name:         "Garden Party",
            Type:         models.TemplateTypeRSVPPage,
            Category:     models.CategoryCard,
            Description:  "Fresh botanical design for outdoor events and garden parties",
            HTMLContent:  s.loadTemplate("garden-party.html"),
            CSSContent:   stringPtr(s.loadCSS("garden-party.css")),
            ThumbnailURL: stringPtr("/static/images/themes/garden-party-thumb.jpg"),
            ImageURL:     stringPtr("/static/images/themes/garden-party-header.jpg"),
            Tags:         []string{"garden", "nature", "outdoor", "botanical"},
            SortOrder:    5,
            IsDefault:    false,
        },
        {
            Name:         "Modern Minimalist",
            Type:         models.TemplateTypeRSVPPage,
            Category:     models.CategoryCard,
            Description:  "Contemporary minimal design with clean lines and bold typography",
            HTMLContent:  s.loadTemplate("modern-minimalist.html"),
            CSSContent:   stringPtr(s.loadCSS("modern-minimalist.css")),
            ThumbnailURL: stringPtr("/static/images/themes/modern-minimalist-thumb.jpg"),
            ImageURL:     stringPtr("/static/images/themes/modern-minimalist-header.jpg"),
            Tags:         []string{"modern", "minimal", "contemporary", "clean"},
            SortOrder:    6,
            IsDefault:    false,
        },
    }
}

func (s *Seeder) loadTemplate(filename string) string {
    content, err := os.ReadFile(filepath.Join("templates/web/rsvp_themes", filename))
    if err != nil {
        log.Printf("Warning: Failed to load template %s: %v", filename, err)
        return ""
    }
    return string(content)
}

func (s *Seeder) loadCSS(filename string) string {
    content, err := os.ReadFile(filepath.Join("static/css/themes", filename))
    if err != nil {
        log.Printf("Warning: Failed to load CSS %s: %v", filename, err)
        return ""
    }
    return string(content)
}

func stringPtr(s string) *string {
    if s == "" {
        return nil
    }
    return &s
}
```

### Startup Integration

**File:** `cmd/server/main.go`

```go
func main() {
    // ... existing setup ...
    
    // Initialize template service
    templateService := templates.NewService(templateRepo)
    
    // Seed default themes
    seeder := templates.NewSeeder(templateService)
    if err := seeder.SeedThemes(context.Background()); err != nil {
        log.Printf("Warning: Theme seeding failed: %v", err)
        // Don't fail startup, just log warning
    } else {
        log.Println("Themes seeded successfully")
    }
    
    // ... continue with server startup ...
}
```

---

## Tasks

### Seeder Implementation
- [ ] Create `seeder.go` file
- [ ] Implement `Seeder` struct
- [ ] Implement `SeedThemes()` method
- [ ] Implement `seedTheme()` helper
- [ ] Implement `getDefaultThemes()` method
- [ ] Implement file loading helpers
- [ ] Handle errors gracefully
- [ ] Write unit tests

### Template Service Extension
- [ ] Add `GetTemplateByName()` method
- [ ] Update service interface
- [ ] Implement in service
- [ ] Write service tests

### Startup Integration
- [ ] Update `main.go` to call seeder
- [ ] Add logging for seeding status
- [ ] Handle seeding errors
- [ ] Test startup with clean database
- [ ] Test startup with existing themes

### Idempotency Testing
- [ ] Test running seeder twice
- [ ] Verify no duplicates created
- [ ] Verify existing themes updated
- [ ] Test with partial existing themes
- [ ] Write idempotency tests

### File Loading
- [ ] Verify all template files exist
- [ ] Verify all CSS files exist
- [ ] Verify all image files exist
- [ ] Handle missing files gracefully
- [ ] Write file loading tests

---

## Definition of Done

- [ ] All acceptance criteria met
- [ ] Seeder implemented
- [ ] Integrated into startup
- [ ] All 7 themes seeded correctly
- [ ] Idempotency verified
- [ ] All unit tests passing
- [ ] All integration tests passing
- [ ] Error handling tested
- [ ] Changes committed to git

---

## Dependencies

**Depends on:**
- Story 11.01: Theme Model Extension
- Story 11.02: Theme Asset Creation

**Blocks:**
- Story 11.07: Theme Integration Testing

---

## References

- **Epic:** [11_EPIC_rsvp_themes.md](11_EPIC_rsvp_themes.md)
- **Analysis:** [11_ANALYSIS_rsvp_page_themes.md](11_ANALYSIS_rsvp_page_themes.md)
- **Template Service:** `internal/templates/service.go`
- **Main Entry:** `cmd/server/main.go`
- **Similar Pattern:** Email template seeding (if exists)

---

## Notes

- Seeding should be fast (<5 seconds)
- Log warnings for missing files but don't fail
- Consider adding CLI command for manual seeding (v1)
- Could add theme import/export functionality (v1)
- Monitor seeding performance

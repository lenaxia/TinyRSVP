# User Story 11.06: Theme Seeding System

**Epic:** [11_EPIC_rsvp_themes.md](11_EPIC_rsvp_themes.md)
**Priority:** High
**Status:** Complete
**Estimated Effort:** 1 day
**Owner:** LLM
**Completed:** 2026-01-11

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
- [x] Themes seeded on application startup
- [x] Seeding is idempotent (safe to run multiple times)
- [x] Existing themes not duplicated
- [x] Existing themes updated if changed
- [x] New themes added automatically
- [x] Seeding completes in <5 seconds

### Theme Data
- [x] All 7 themes seeded (1 plain + 6 card)
- [x] Theme metadata complete (name, description, category, tags)
- [x] Image URLs correct
- [x] Sort order set appropriately
- [x] One theme marked as default
- [x] Created_by is NULL (system themes)

### Validation
- [x] Seeded themes pass validation
- [x] Image URLs point to existing files
- [x] CSS files exist for all themes
- [x] HTML templates exist for all themes
- [x] No broken references

### Error Handling
- [x] Database errors logged and reported
- [x] Missing files logged as warnings
- [x] Seeding failure doesn't prevent startup (uses existing themes)
- [x] Partial seeding handled gracefully

### Testing
- [x] Unit tests for seeding logic
- [x] Integration tests with clean database
- [x] Integration tests with existing themes
- [x] Test idempotency (run twice)
- [x] Test theme updates
- [x] Test error scenarios

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
- [x] Create `seeder.go` file
- [x] Implement `Seeder` struct
- [x] Implement `SeedThemes()` method
- [x] Implement `seedTheme()` helper
- [x] Implement `getDefaultThemes()` method
- [x] Implement file loading helpers
- [x] Handle errors gracefully
- [x] Write unit tests

### Template Service Extension
- [x] Add `GetByNameAndType()` method to repository
- [x] Update repository interface
- [x] Implement in repository
- [x] Write repository tests

### Startup Integration
- [x] Update `main.go` to call seeder
- [x] Add logging for seeding status
- [x] Handle seeding errors
- [x] Test startup with clean database
- [x] Test startup with existing themes

### Idempotency Testing
- [x] Test running seeder twice
- [x] Verify no duplicates created
- [x] Verify existing themes updated
- [x] Test with partial existing themes
- [x] Write idempotency tests

### File Loading
- [x] Verify all template files exist
- [x] Verify all CSS files exist
- [x] Verify all image files exist
- [x] Handle missing files gracefully
- [x] Write file loading tests

---

## Definition of Done

- [x] All acceptance criteria met
- [x] Seeder implemented
- [x] Integrated into startup
- [x] All 7 themes seeded correctly
- [x] Idempotency verified
- [x] All unit tests passing
- [x] All integration tests passing
- [x] Error handling tested
- [x] Changes committed to git

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

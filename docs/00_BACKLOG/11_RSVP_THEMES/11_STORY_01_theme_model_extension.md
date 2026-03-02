# User Story 11.01: Theme Model Extension

**Epic:** [11_EPIC_rsvp_themes.md](11_EPIC_rsvp_themes.md)
**Priority:** High
**Status:** Complete
**Estimated Effort:** 2 days
**Owner:** LLM Assistant
**Completed:** 2026-01-11

---

## User Story

As a **system developer**,  
I want to **extend the template model to support theme metadata**,  
So that **we can store and manage RSVP page themes with categories, images, and descriptions**.

---

## Context

The current template model supports basic HTML/CSS content but lacks fields needed for a theme system:
- No category field (plain, card, modern, etc.)
- No description for theme picker display
- No thumbnail/image URLs for visual themes
- No tags for filtering/searching
- No sort order for gallery display

This story extends the model and database schema to support rich theme metadata.

---

## Acceptance Criteria

### Database Schema
- [x] Add `category` column to templates table (TEXT)
- [x] Add `description` column to templates table (TEXT) - already existed
- [x] Add `thumbnail_url` column to templates table (TEXT, nullable)
- [x] Add `image_url` column to templates table (TEXT, nullable)
- [x] Add `tags` column to templates table (TEXT, JSON array)
- [x] Add `sort_order` column to templates table (INTEGER, default 0)
- [x] Create index on `category` column
- [x] Create index on `sort_order` column
- [x] Migration includes both up and down scripts

### Model Updates
- [ ] Update `Template` struct with new fields
- [ ] Add JSON tags for API serialization
- [ ] Add validation for new fields
- [ ] Update template repository methods
- [ ] Add category constants (plain, card, modern, classic, fun)

### Repository Methods
- [x] `GetTemplatesByCategory(ctx, category)` returns themes by category
- [x] `ListThemes(ctx, type, category)` returns themes with filtering
- [x] Repository methods handle nullable image URLs
- [x] Repository methods parse tags JSON
- [x] Update Create() to handle new fields
- [x] Update GetByID() to scan new fields
- [x] Update GetByEventAndType() to scan new fields
- [x] Update GetDefaultByType() to scan new fields
- [x] Update List() to use scanTemplate helper
- [x] Update Update() to handle new fields

### Validation
- [x] Category must be valid enum value (with default to 'plain')
- [x] Description max 500 characters
- [x] Tags handled as JSON array (serialization/deserialization)
- [x] Sort order must be >= 0
- [x] Image URLs nullable (validation deferred to Story 11.02)

### Testing
- [x] Unit tests for model validation (template_category_test.go)
- [x] Unit tests for repository methods (template_repository_theme_test.go)
- [x] Integration tests for database operations
- [x] Test nullable fields handled correctly
- [x] Test JSON tags parsing
- [x] Test category filtering
- [x] Test sort order
- [x] Test backward compatibility with existing tests

---

## Technical Details

### Database Migration

**File:** `migrations/sqlite/000010_add_theme_fields.up.sql`

```sql
-- Add theme-specific fields to templates table
ALTER TABLE templates ADD COLUMN category TEXT DEFAULT 'plain';
ALTER TABLE templates ADD COLUMN description TEXT;
ALTER TABLE templates ADD COLUMN thumbnail_url TEXT;
ALTER TABLE templates ADD COLUMN image_url TEXT;
ALTER TABLE templates ADD COLUMN tags TEXT; -- JSON array
ALTER TABLE templates ADD COLUMN sort_order INTEGER DEFAULT 0;

-- Create indexes for theme queries
CREATE INDEX idx_templates_category ON templates(category);
CREATE INDEX idx_templates_sort_order ON templates(sort_order);

-- Update existing templates to have category
UPDATE templates SET category = 'plain' WHERE category IS NULL;
```

**File:** `migrations/sqlite/000010_add_theme_fields.down.sql`

```sql
-- Remove indexes
DROP INDEX IF EXISTS idx_templates_sort_order;
DROP INDEX IF EXISTS idx_templates_category;

-- Note: SQLite doesn't support DROP COLUMN easily
-- For rollback, would need to recreate table without new columns
-- This is acceptable for development, production rollbacks should be rare
```

### Model Extension

**File:** `internal/models/template.go`

```go
type TemplateCategory string

const (
    CategoryPlain   TemplateCategory = "plain"
    CategoryCard    TemplateCategory = "card"
    CategoryModern  TemplateCategory = "modern"
    CategoryClassic TemplateCategory = "classic"
    CategoryFun     TemplateCategory = "fun"
)

type Template struct {
    ID           int64            `json:"id"`
    Name         string           `json:"name"`
    Type         TemplateType     `json:"type"`
    HTMLContent  string           `json:"html_content"`
    TextContent  *string          `json:"text_content,omitempty"`
    CSSContent   *string          `json:"css_content,omitempty"`
    IsDefault    bool             `json:"is_default"`
    CreatedBy    *int64           `json:"created_by,omitempty"`
    CreatedAt    time.Time        `json:"created_at"`
    UpdatedAt    time.Time        `json:"updated_at"`
    
    // Theme-specific fields
    Category     TemplateCategory `json:"category"`
    Description  string           `json:"description"`
    ThumbnailURL *string          `json:"thumbnail_url,omitempty"`
    ImageURL     *string          `json:"image_url,omitempty"`
    Tags         []string         `json:"tags"`
    SortOrder    int              `json:"sort_order"`
}

func (t *Template) Validate() error {
    if t.Name == "" {
        return &ValidationError{Field: "name", Message: "Template name is required"}
    }
    if len(t.Name) > 100 {
        return &ValidationError{Field: "name", Message: "Template name cannot exceed 100 characters"}
    }
    if t.Type == "" {
        return &ValidationError{Field: "type", Message: "Template type is required"}
    }
    if t.HTMLContent == "" {
        return &ValidationError{Field: "html_content", Message: "HTML content is required"}
    }
    if t.Type == TemplateTypeInviteEmail && t.TextContent == nil {
        return &ValidationError{Field: "text_content", Message: "Text content required for email templates"}
    }
    
    // Validate category
    validCategories := map[TemplateCategory]bool{
        CategoryPlain: true, CategoryCard: true, CategoryModern: true,
        CategoryClassic: true, CategoryFun: true,
    }
    if !validCategories[t.Category] {
        return &ValidationError{Field: "category", Message: "Invalid template category"}
    }
    
    // Validate description
    if len(t.Description) > 500 {
        return &ValidationError{Field: "description", Message: "Description cannot exceed 500 characters"}
    }
    
    // Validate sort order
    if t.SortOrder < 0 {
        return &ValidationError{Field: "sort_order", Message: "Sort order must be >= 0"}
    }
    
    return nil
}
```

### Repository Methods

**File:** `internal/db/repositories/template_repository.go`

```go
func (r *templateRepository) GetTemplatesByCategory(ctx context.Context, category models.TemplateCategory) ([]*models.Template, error) {
    query := `
        SELECT id, name, type, html_content, text_content, css_content, 
               is_default, created_by, created_at, updated_at,
               category, description, thumbnail_url, image_url, tags, sort_order
        FROM templates
        WHERE category = ?
        ORDER BY sort_order ASC, name ASC
    `
    
    rows, err := r.db.QueryContext(ctx, query, category)
    if err != nil {
        return nil, fmt.Errorf("query templates by category: %w", err)
    }
    defer rows.Close()
    
    var templates []*models.Template
    for rows.Next() {
        t, err := r.scanTemplate(rows)
        if err != nil {
            return nil, err
        }
        templates = append(templates, t)
    }
    
    return templates, rows.Err()
}

func (r *templateRepository) ListThemes(ctx context.Context, templateType models.TemplateType, category *models.TemplateCategory) ([]*models.Template, error) {
    query := `
        SELECT id, name, type, html_content, text_content, css_content,
               is_default, created_by, created_at, updated_at,
               category, description, thumbnail_url, image_url, tags, sort_order
        FROM templates
        WHERE type = ?
    `
    
    args := []interface{}{templateType}
    
    if category != nil {
        query += " AND category = ?"
        args = append(args, *category)
    }
    
    query += " ORDER BY sort_order ASC, name ASC"
    
    rows, err := r.db.QueryContext(ctx, query, args...)
    if err != nil {
        return nil, fmt.Errorf("list themes: %w", err)
    }
    defer rows.Close()
    
    var templates []*models.Template
    for rows.Next() {
        t, err := r.scanTemplate(rows)
        if err != nil {
            return nil, err
        }
        templates = append(templates, t)
    }
    
    return templates, rows.Err()
}

func (r *templateRepository) scanTemplate(scanner interface{ Scan(...interface{}) error }) (*models.Template, error) {
    var t models.Template
    var tagsJSON *string
    
    err := scanner.Scan(
        &t.ID, &t.Name, &t.Type, &t.HTMLContent, &t.TextContent, &t.CSSContent,
        &t.IsDefault, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt,
        &t.Category, &t.Description, &t.ThumbnailURL, &t.ImageURL, &tagsJSON, &t.SortOrder,
    )
    if err != nil {
        return nil, err
    }
    
    // Parse tags JSON
    if tagsJSON != nil && *tagsJSON != "" {
        if err := json.Unmarshal([]byte(*tagsJSON), &t.Tags); err != nil {
            return nil, fmt.Errorf("parse tags JSON: %w", err)
        }
    }
    
    return &t, nil
}
```

---

## Tasks

### Database Migration
- [x] Create migration file `000010_add_theme_fields.up.sql`
- [x] Create rollback file `000010_add_theme_fields.down.sql`
- [x] Test migration on clean database (via integration tests)
- [x] Test migration on existing database with data (via integration tests)
- [x] Rollback migration created (drops indexes)

### Model Updates
- [x] Add `TemplateCategory` type and constants
- [x] Add new fields to `Template` struct
- [x] Update `Validate()` method
- [x] Add category validation with default
- [x] Add description length validation
- [x] Add sort order validation
- [x] Write unit tests for validation

### Repository Updates
- [x] Update `Create()` to handle new fields
- [x] Update `Update()` to handle new fields
- [x] Create `scanTemplate()` helper to scan new fields
- [x] Add `GetTemplatesByCategory()` method
- [x] Add `ListThemes()` method with filtering
- [x] Handle JSON tags parsing (serializeTags/deserializeTags)
- [x] Write unit tests for new methods
- [x] Write integration tests
- [x] Update all existing repository methods to use new fields

### Service Updates
- [x] Template service automatically works with new repository methods
- [x] Service tests pass with default category approach
- [x] Integration tests pass
- [x] Mock repositories updated with new interface methods

---

## Definition of Done

- [x] All acceptance criteria met
- [x] Database migration created and tested
- [x] Model extended with new fields
- [x] Repository methods implemented
- [x] Service methods updated (via interface)
- [x] All unit tests passing
- [x] All integration tests passing
- [x] Code reviewed (self-review via TDD)
- [x] Changes committed to git

---

## Dependencies

**Depends on:**
- ✅ Story 06.00: Template Struct (complete)
- ✅ Story 06.04: Template CRUD (complete)

**Blocks:**
- Story 11.02: Theme Asset Creation
- Story 11.03: Theme Picker UI
- All subsequent theme stories

---

## References

- **Epic:** [11_EPIC_rsvp_themes.md](11_EPIC_rsvp_themes.md)
- **Analysis:** [11_ANALYSIS_rsvp_page_themes.md](11_ANALYSIS_rsvp_page_themes.md)
- **HLD:** [docs/02_REVISED_HLD.md](../02_REVISED_HLD.md) Section 11
- **Template LLD:** [docs/lld/06_TEMPLATE_LLD.md](../lld/06_TEMPLATE_LLD.md)
- **Template Model:** `internal/models/template.go`
- **Template Repository:** `internal/db/repositories/template_repository.go`

---

## Notes

- Keep backward compatibility - existing templates get `category='plain'` by default
- Tags stored as JSON for flexibility
- Nullable image URLs allow plain text themes
- Sort order allows manual theme ordering in picker
- Category enum allows future expansion (modern, classic, fun)

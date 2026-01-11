# User Story 11.01: Theme Model Extension

**Epic:** [11_EPIC_rsvp_themes.md](11_EPIC_rsvp_themes.md)  
**Priority:** High  
**Status:** Not Started  
**Estimated Effort:** 2 days  
**Owner:** Unassigned

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
- [ ] Add `category` column to templates table (TEXT)
- [ ] Add `description` column to templates table (TEXT)
- [ ] Add `thumbnail_url` column to templates table (TEXT, nullable)
- [ ] Add `image_url` column to templates table (TEXT, nullable)
- [ ] Add `tags` column to templates table (TEXT, JSON array)
- [ ] Add `sort_order` column to templates table (INTEGER, default 0)
- [ ] Create index on `category` column
- [ ] Create index on `sort_order` column
- [ ] Migration includes both up and down scripts

### Model Updates
- [ ] Update `Template` struct with new fields
- [ ] Add JSON tags for API serialization
- [ ] Add validation for new fields
- [ ] Update template repository methods
- [ ] Add category constants (plain, card, modern, classic, fun)

### Repository Methods
- [ ] `GetTemplatesByCategory(ctx, category)` returns themes by category
- [ ] `GetTemplatesByType(ctx, type)` filters by template type
- [ ] `ListThemes(ctx, type, category)` returns themes with filtering
- [ ] Repository methods handle nullable image URLs
- [ ] Repository methods parse tags JSON

### Validation
- [ ] Category must be valid enum value
- [ ] Description max 500 characters
- [ ] Tags must be valid JSON array
- [ ] Sort order must be >= 0
- [ ] Image URLs validated if provided

### Testing
- [ ] Unit tests for model validation
- [ ] Unit tests for repository methods
- [ ] Integration tests for database operations
- [ ] Test nullable fields handled correctly
- [ ] Test JSON tags parsing
- [ ] Test category filtering
- [ ] Test sort order

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
- [ ] Create migration file `000010_add_theme_fields.up.sql`
- [ ] Create rollback file `000010_add_theme_fields.down.sql`
- [ ] Test migration on clean database
- [ ] Test migration on existing database with data
- [ ] Test rollback migration

### Model Updates
- [ ] Add `TemplateCategory` type and constants
- [ ] Add new fields to `Template` struct
- [ ] Update `Validate()` method
- [ ] Add category validation
- [ ] Add description length validation
- [ ] Add sort order validation
- [ ] Write unit tests for validation

### Repository Updates
- [ ] Update `Create()` to handle new fields
- [ ] Update `Update()` to handle new fields
- [ ] Update `scanTemplate()` to scan new fields
- [ ] Add `GetTemplatesByCategory()` method
- [ ] Add `ListThemes()` method with filtering
- [ ] Handle JSON tags parsing
- [ ] Write unit tests for new methods
- [ ] Write integration tests

### Service Updates
- [ ] Update template service to use new repository methods
- [ ] Add `ListThemesByCategory()` service method
- [ ] Update service tests
- [ ] Write integration tests

---

## Definition of Done

- [ ] All acceptance criteria met
- [ ] Database migration created and tested
- [ ] Model extended with new fields
- [ ] Repository methods implemented
- [ ] Service methods updated
- [ ] All unit tests passing
- [ ] All integration tests passing
- [ ] Code reviewed
- [ ] Changes committed to git

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

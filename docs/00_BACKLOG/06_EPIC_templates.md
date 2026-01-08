# Epic: Templates & Asset Management

**Priority:** Medium  
**Status:** Not Started  
**Target Version:** v0  
**Estimated Effort:** 1 week

---

## Overview

Implement template system for customizable email and web page rendering. Support image uploads with validation, storage provider abstraction (local FS in v0), and XSS prevention through Go html/template.

**Goal:** Enable event managers to customize invitation appearance while maintaining security and supporting asset uploads.

---

## Success Criteria

- [ ] Default templates provided for all types
- [ ] Event managers can customize templates
- [ ] Go html/template integration with auto-escaping
- [ ] Template validation prevents XSS
- [ ] Image upload with validation (type, size, dimensions)
- [ ] Local filesystem storage working
- [ ] Images accessible via public URLs
- [ ] Template variables properly documented
- [ ] CSS sanitization prevents malicious code

---

## User Stories

### Phase 1: Template System
- [ ] [`06_STORY_00_template_model.md`](06_STORY_template_model.md) - Template struct and repository
- [ ] [`06_STORY_01_template_rendering.md`](06_STORY_template_rendering.md) - Go html/template integration
- [ ] [`06_STORY_02_template_validation.md`](06_STORY_template_validation.md) - Template security validation
- [ ] [`06_STORY_03_default_templates.md`](06_STORY_default_templates.md) - System default templates

### Phase 2: Template Management
- [ ] [`06_STORY_04_template_crud.md`](06_STORY_template_crud.md) - Create/edit/delete templates
- [ ] [`06_STORY_05_template_variables.md`](06_STORY_template_variables.md) - Template variable system
- [ ] [`06_STORY_06_template_preview.md`](06_STORY_template_preview.md) - Preview before saving

### Phase 3: Asset Management
- [ ] [`06_STORY_07_image_upload.md`](06_STORY_image_upload.md) - Image upload with validation
- [ ] [`06_STORY_08_storage_provider.md`](06_STORY_storage_provider.md) - Storage provider interface
- [ ] [`06_STORY_09_local_storage.md`](06_STORY_local_storage.md) - Local filesystem implementation
- [ ] [`06_STORY_10_asset_serving.md`](06_STORY_asset_serving.md) - Serve uploaded assets

### Phase 4: Security
- [ ] [`06_STORY_11_xss_prevention.md`](06_STORY_xss_prevention.md) - XSS prevention in templates
- [ ] [`06_STORY_12_css_sanitization.md`](06_STORY_css_sanitization.md) - CSS sanitization
- [ ] [`06_STORY_013_file_validation.md`](06_STORY_file_validation.md) - File type validation

---

## Dependencies

**Depends on:** Epic 00 (Foundation), Epic 01 (Auth)  
**Blocks:** Epic 05 (Email - needs templates)

---

## Technical Overview

### Template Types

```
invite_email       → HTML + text email template
rsvp_page          → HTML web page template
confirmation_page  → HTML web page template
```

### Template Rendering Flow

```
Template + Data → Go html/template → Rendered HTML (auto-escaped)
```

### Storage Provider Interface

```go
type StorageProvider interface {
    PutObject(path string, data io.Reader, contentType string) error
    GetObject(path string) (io.ReadCloser, error)
    DeleteObject(path string) error
    GetPublicURL(path string) (string, error)
}
```

### Image Upload Flow

```
1. Validate file type (magic bytes)
2. Validate file size (<5MB)
3. Validate dimensions (<4096x4096)
4. Strip EXIF data
5. Generate unique filename
6. Store via provider
7. Return public URL
```

---

## Technical Decisions

### Template Engine: Go html/template
- Built into Go standard library
- Automatic HTML escaping (XSS-safe)
- No build step required
- Works for web and email

### Storage: Local FS (v0)
- Simple, no dependencies
- Files in mounted volume
- Served by application
- S3 support deferred to v1+

### Image Processing
- Re-encode to strip EXIF
- Validate by magic bytes (not extension)
- Enable automatic resizing
- Max 5MB, 4096x4096 pixels

### Template Versioning
- No versioning in v0
- Updates apply immediately
- Simplifies implementation
- Versioning in v1+ if needed
- design in a such a way that verisoning refactoring is possible

---

## Template Variables

### Available in All Templates
```
{{.Event.Title}}
{{.Event.Description}}
{{.Event.StartTime}}
{{.Event.EndTime}}
{{.Event.Timezone}}
{{.Event.Location}}
{{.Event.RSVPDeadline}}
```

### Invite Templates Only
```
{{.Invite.Name}}
{{.Invite.Email}}
{{.RSVPURL}}
{{.MaxPlusOnes}}
```

### RSVP Page Templates
```
{{.RSVP.Response}}
{{.RSVP.PlusOnes}}
{{.Questions}} (array)
```

### Template Functions
```
{{.StartTime.Format "Jan 2, 2006"}}  - Date formatting
{{.Title | upper}}                    - String operations
{{if .RSVPDeadline}}...{{end}}       - Conditionals
{{range .Questions}}...{{end}}        - Loops
```

---

## Security Measures

### XSS Prevention
- Go html/template auto-escapes
- No `template.HTML` type (disables escaping)
- User input always escaped
- Template validation on upload

### CSS Sanitization
- No `javascript:` URLs
- No `expression()` (IE)
- No `@import` external resources
- Whitelist safe CSS properties

### File Upload Security
- Magic byte validation
- File size limits
- Dimension limits
- EXIF stripping
- Unique filenames (prevent overwrite)

### Template Validation
- Parse template on upload
- Reject if parse fails
- Reject undefined variables
- Reject disallowed functions

---

## Validation Rules

### Template Name
- Required, 3-100 characters
- Alphanumeric + spaces + hyphens

### Template Type
- Must be: invite_email, rsvp_page, confirmation_page

### HTML Content
- Required
- Valid Go template syntax
- No disallowed functions
- No undefined variables

### Text Content
- Required for email templates
- Plain text only

### Image Upload
- File types: JPEG, PNG, GIF, WebP
- Max size: 5MB
- Max dimensions: 4096x4096
- Valid image (magic bytes)

---

## Storage Structure

### Local Filesystem
```
/data/uploads/
├── images/
│   ├── {event_id}/
│   │   └── {filename}
└── templates/
    └── {template_id}/
        └── {filename}
```

### Public URLs
```
/assets/images/{event_id}/{filename}
```

---

## References

- **HLD:** Section 11 (Templates), Section 12 (Asset Storage)
- **LLD:** [`lld/06_TEMPLATE_LLD.md`](../lld/06_TEMPLATE_LLD.md)
- **Database:** templates table
- **Go Docs:** https://pkg.go.dev/html/template

---

## Testing Strategy

### Unit Tests
- Template parsing
- Variable substitution
- XSS prevention
- CSS sanitization
- File validation

### Integration Tests
- Template rendering with real data
- Image upload flow
- Storage provider operations
- Template CRUD

### Security Tests
- XSS injection attempts
- CSS injection attempts
- File upload attacks
- Path traversal attempts

### Manual Tests
1. Create custom template
2. Upload image
3. Preview template
4. Send email with template
5. View RSVP page with template
6. Verify XSS protection

---

## Default Templates

### Invite Email Template
- Clean, professional design
- Event details prominently displayed
- Clear RSVP button
- Mobile-responsive

### RSVP Page Template
- Simple, focused form
- Event details at top
- Clear response options
- Mobile-first design

### Confirmation Page Template
- Thank you message
- RSVP summary
- Add to calendar button
- Update RSVP link

---

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| XSS vulnerabilities | Critical | Go html/template auto-escaping, validation |
| Malicious file uploads | High | Magic byte validation, EXIF stripping |
| Storage exhaustion | Medium | File size limits, quota monitoring |
| Template syntax errors | Low | Validation on upload, preview before save |
| CSS injection | Medium | CSS sanitization, whitelist approach |

---

## Performance Considerations

- Template compilation cached
- Static assets served with caching headers
- Image serving optimized
- No image resizing (keep simple)

---

## Future Enhancements (v1+)

- S3-compatible storage
- Image resizing/optimization
- Template versioning
- Template marketplace
- Custom CSS editor
- Drag-and-drop template builder

---

## Definition of Done

- [ ] All user stories complete
- [ ] Default templates created
- [ ] Template CRUD functional
- [ ] Go html/template integrated
- [ ] XSS prevention verified
- [ ] Image upload working
- [ ] Local storage implemented
- [ ] Asset serving functional
- [ ] All validation rules enforced
- [ ] Security review passed
- [ ] All tests passing
- [ ] Documentation updated

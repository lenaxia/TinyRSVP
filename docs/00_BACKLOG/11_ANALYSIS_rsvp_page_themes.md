# RSVP Page Theme System Analysis

**Date:** 2026-01-11  
**Author:** LLM Assistant  
**Status:** Analysis Complete  
**Related Stories:** 10.12 (Light/Dark Theme Switching)

---

## Executive Summary

This analysis addresses the need for a comprehensive **RSVP Page Theme System** that allows event managers to select from professionally designed themes when creating events. This is distinct from email templates - these themes control the visual appearance of the web pages guests see when they RSVP.

**Key Findings:**
1. **Current State**: Single plain RSVP page template with basic styling
2. **Infrastructure**: ~60% ready (CSS variables, template system exists)
3. **Recommendation**: Implement card-based theme system (like Evite) with extensibility
4. **Effort**: 2-3 weeks for Phase 1 (pre-designed themes)

**Scope Clarification:**
- ✅ RSVP page themes (web pages guests see)
- ✅ Card-based visual designs with images
- ✅ Integration with light/dark theme system
- ❌ Email templates (separate concern, already analyzed)
- ❌ Admin dashboard themes (uses system light/dark only)

---

## Table of Contents

1. [Current State Analysis](#current-state-analysis)
2. [Requirements & User Needs](#requirements--user-needs)
3. [Theme Architecture](#theme-architecture)
4. [Card-Based Design System](#card-based-design-system)
5. [Light/Dark Theme Integration](#lightdark-theme-integration)
6. [WYSIWYG Editor Analysis](#wysiwyg-editor-analysis)
7. [Implementation Phases](#implementation-phases)
8. [Technical Specifications](#technical-specifications)
9. [Recommendations](#recommendations)

---

## 1. Current State Analysis

### 1.1 What Exists ✅

**Template Infrastructure:**
- ✅ Template model with HTML/CSS content fields
- ✅ Template service with CRUD operations
- ✅ Template types: `invite_email`, `rsvp_page`, `confirmation_page`
- ✅ Go html/template rendering engine
- ✅ XSS prevention and security

**CSS Foundation:**
- ✅ CSS variables system with semantic tokens
- ✅ Basic light/dark theme support (6 variables via media query)
- ✅ Responsive grid system
- ✅ Component library (buttons, forms, cards)
- ✅ Mobile-first design

**Storage Infrastructure:**
- ✅ Storage provider interface
- ✅ Local filesystem implementation
- ✅ Image validation framework
- ✅ Asset serving capability

### 1.2 What's Missing ❌

**Theme System:**
- ❌ No theme selection UI for event creation
- ❌ No pre-designed RSVP page themes
- ❌ No theme preview functionality
- ❌ No theme-specific image assets
- ❌ No theme metadata (name, description, category)
- ❌ No theme-to-event association

**Visual Design:**
- ❌ No card-based invitation designs
- ❌ No theme-specific layouts
- ❌ No image placeholders for themes
- ❌ No theme gallery/picker UI

**Integration:**
- ❌ Light/dark theme system incomplete (only 6 variables)
- ❌ No theme inheritance (event theme + user light/dark preference)
- ❌ No theme customization options

---

## 2. Requirements & User Needs

### 2.1 User Personas

**Event Manager (Primary User):**
- Wants professional-looking RSVP pages without design skills
- Needs quick theme selection during event creation
- Values preview before committing to a theme
- May want to customize theme with own images later

**Guest (Secondary User):**
- Expects visually appealing invitation page
- Wants theme to match event type (wedding, birthday, corporate)
- Needs light/dark mode support for comfort
- Requires mobile-responsive design

**System Designer (Internal):**
- Needs tools to create new themes efficiently
- Wants consistent design system
- Requires theme versioning and updates

### 2.2 Core Requirements

**Functional Requirements:**
1. Event manager can select theme during event creation
2. Event manager can preview theme before selection
3. Event manager can change theme after event creation
4. Guest sees RSVP page styled with selected theme
5. Theme respects guest's light/dark mode preference
6. At least 1 plain text theme (no images)
7. At least 5-10 card-based image themes

**Non-Functional Requirements:**
1. Themes load in <2 seconds
2. Mobile-responsive (320px to 1920px)
3. Accessible (WCAG AA)
4. Works without JavaScript (progressive enhancement)
5. No layout shift during theme application

---

## 3. Theme Architecture

### 3.1 Theme Hierarchy

```
System Theme (Light/Dark)
    ↓ (User Preference)
Event Theme (Visual Design)
    ↓ (Event Manager Selection)
RSVP Page Rendering
```

**Separation of Concerns:**
- **System Theme**: Light/Dark mode (user preference, affects all pages)
- **Event Theme**: Visual design/layout (event-specific, affects RSVP page only)
- **Result**: Event theme adapts to system theme (e.g., "Wedding Card" in dark mode)

### 3.2 Theme Data Model

**Extend Existing Template Model:**
```go
type Template struct {
    ID          int64     `json:"id"`
    Name        string    `json:"name"`
    Type        string    `json:"type"` // "rsvp_page"
    HTMLContent string    `json:"html_content"`
    CSSContent  string    `json:"css_content"`
    IsDefault   bool      `json:"is_default"`
    CreatedBy   *int64    `json:"created_by,omitempty"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
    
    // NEW FIELDS FOR THEMES
    Category    string    `json:"category"`    // "plain", "card", "modern", "classic"
    Description string    `json:"description"` // "Elegant wedding invitation with floral design"
    ThumbnailURL *string  `json:"thumbnail_url,omitempty"` // Preview image
    ImageURL    *string   `json:"image_url,omitempty"`     // Theme header image
    Tags        []string  `json:"tags"`        // ["wedding", "formal", "floral"]
    SortOrder   int       `json:"sort_order"`  // Display order in picker
}
```

**Theme Categories:**
- `plain` - Text-only, no images (1 theme)
- `card` - Card-based designs with images (5-10 themes)
- `modern` - Future: Modern flat designs
- `classic` - Future: Traditional formal designs
- `fun` - Future: Playful/casual designs

### 3.3 Theme Selection Flow

```
Event Creation Form
    ↓
Theme Picker Step (NEW)
    ↓ (Select from gallery)
Theme Preview Modal
    ↓ (Confirm selection)
Event Details Form
    ↓
Event Created with Theme
```

**Alternative Flow (Simpler for v0):**
```
Event Creation Form
    ↓ (Theme dropdown in form)
Select Theme from Dropdown
    ↓ (Optional: Click preview)
Theme Preview Modal
    ↓
Complete Event Creation
```

---

## 4. Card-Based Design System

### 4.1 Card Theme Anatomy

**Inspired by Evite's card-based invitations:**

```
┌─────────────────────────────────────┐
│  ┌───────────────────────────────┐  │ ← Card Container
│  │                               │  │
│  │     [Theme Header Image]      │  │ ← Image Area (1200x400px)
│  │                               │  │
│  ├───────────────────────────────┤  │
│  │                               │  │
│  │   Event Title                 │  │ ← Content Area
│  │   Date & Time                 │  │
│  │   Location                    │  │
│  │   Description                 │  │
│  │                               │  │
│  │   [RSVP Form]                 │  │ ← Interactive Area
│  │                               │  │
│  └───────────────────────────────┘  │
└─────────────────────────────────────┘
```

**Key Characteristics:**
1. **Centered card** on page background
2. **Header image** sets theme tone
3. **Content area** with event details
4. **Form area** for RSVP submission
5. **Responsive** - card adapts to screen size

### 4.2 Theme Variations

**Plain Text Theme (Required):**
- No images
- Clean typography
- Minimal design
- Fast loading
- Accessibility-first
- Example: Simple centered text with subtle borders

**Card Themes (5-10 Initial):**

1. **Wedding Elegance**
   - Floral header image
   - Serif fonts
   - Soft colors (blush, gold)
   - Formal tone

2. **Birthday Celebration**
   - Balloons/confetti header
   - Playful fonts
   - Bright colors
   - Fun tone

3. **Corporate Professional**
   - Abstract geometric header
   - Sans-serif fonts
   - Blue/gray palette
   - Professional tone

4. **Holiday Festive**
   - Seasonal imagery
   - Decorative fonts
   - Red/green or seasonal colors
   - Warm tone

5. **Garden Party**
   - Nature/botanical header
   - Organic fonts
   - Green/earth tones
   - Casual tone

6. **Modern Minimalist**
   - Simple geometric header
   - Clean fonts
   - Monochrome with accent
   - Contemporary tone

7. **Beach/Summer**
   - Ocean/beach header
   - Relaxed fonts
   - Blue/sand colors
   - Casual tone

8. **Baby Shower**
   - Soft pastel header
   - Gentle fonts
   - Pink/blue pastels
   - Sweet tone

### 4.3 Image Specifications

**Header Images:**
- Dimensions: 1200x400px (3:1 ratio)
- Format: JPEG (optimized)
- Size: <150KB each
- Total: ~1.5MB for 10 themes
- Location: `/static/images/themes/`

**Thumbnail Images:**
- Dimensions: 300x200px
- Format: JPEG (optimized)
- Size: <30KB each
- Used in theme picker gallery

**Design Guidelines:**
- Images should work in both light and dark modes
- Avoid text in images (accessibility)
- Use subtle overlays if needed for text readability
- Provide alt text for all images

---

## 5. Light/Dark Theme Integration

### 5.1 Current Light/Dark System (Story 10.12)

**From Story 10.12:**
- Uses `[data-theme="light|dark"]` attribute on `<html>`
- JavaScript controller manages theme switching
- localStorage persistence
- System preference detection
- Affects ALL pages (admin + guest)

**Current Coverage:**
- Only 6 CSS variables have dark mode equivalents
- Needs expansion to ~50+ variables

### 5.2 Integration Strategy

**Two-Layer Theme System:**

```css
/* Layer 1: System Theme (Light/Dark) - User Preference */
:root {
    --color-background: #ffffff;
    --color-text: #111827;
    /* ... light mode defaults ... */
}

[data-theme="dark"] {
    --color-background: #0f172a;
    --color-text: #f1f5f9;
    /* ... dark mode overrides ... */
}

/* Layer 2: Event Theme (Visual Design) - Event Manager Selection */
[data-event-theme="wedding"] {
    --theme-primary: #f4c2c2;      /* Blush */
    --theme-secondary: #d4af37;    /* Gold */
    --theme-accent: #8b4789;       /* Plum */
    --theme-font-heading: 'Playfair Display', serif;
    --theme-font-body: 'Lato', sans-serif;
}

[data-event-theme="birthday"] {
    --theme-primary: #ff6b9d;      /* Pink */
    --theme-secondary: #ffd93d;    /* Yellow */
    --theme-accent: #6bcf7f;       /* Green */
    --theme-font-heading: 'Fredoka One', cursive;
    --theme-font-body: 'Open Sans', sans-serif;
}

/* Layer 3: Combined Application */
.rsvp-card {
    background: var(--color-surface);        /* From system theme */
    color: var(--color-text-primary);        /* From system theme */
    border-color: var(--theme-primary);      /* From event theme */
    font-family: var(--theme-font-body);     /* From event theme */
}

.rsvp-card h1 {
    color: var(--theme-primary);             /* From event theme */
    font-family: var(--theme-font-heading); /* From event theme */
}
```

**How It Works:**
1. Guest visits RSVP page with `?token=abc123`
2. Server renders page with `data-event-theme="wedding"` on `<html>`
3. Guest's browser applies their light/dark preference via JavaScript
4. Result: Wedding theme colors + guest's light/dark preference

**Example Combinations:**
- Wedding theme + Light mode = Blush/gold on white
- Wedding theme + Dark mode = Blush/gold on dark gray
- Birthday theme + Light mode = Bright colors on white
- Birthday theme + Dark mode = Bright colors on dark gray

### 5.3 Theme Variable Naming

**System Variables (Light/Dark):**
- `--color-background` - Page background
- `--color-surface` - Card/panel background
- `--color-text-primary` - Main text color
- `--color-border` - Border colors
- (All existing variables from Story 10.12)

**Event Theme Variables:**
- `--theme-primary` - Primary brand color
- `--theme-secondary` - Secondary brand color
- `--theme-accent` - Accent color
- `--theme-font-heading` - Heading font family
- `--theme-font-body` - Body font family
- `--theme-image-url` - Header image URL

**Benefits:**
- Clear separation of concerns
- Easy to add new event themes
- User preference respected
- No conflicts between layers

---

## 6. WYSIWYG Editor Analysis

### 6.1 Two Different Use Cases

**Use Case 1: System Designers Creating Themes**
- **Who:** Internal developers/designers
- **What:** Create new theme templates for the gallery
- **How:** Code-based (HTML/CSS files)
- **Frequency:** Infrequent (quarterly)
- **Complexity:** High (full control)

**Use Case 2: Event Managers Customizing Themes**
- **Who:** End users (event managers)
- **What:** Customize selected theme (change image, colors)
- **How:** UI-based customization options
- **Frequency:** Every event creation
- **Complexity:** Low (limited options)

### 6.2 WYSIWYG Editor for System Designers

**Question:** Do we need a WYSIWYG editor for creating themes?

**Answer:** **NO, not for v0-v1**

**Rationale:**
1. **Low Frequency**: New themes added rarely (maybe 2-3 per quarter)
2. **High Skill**: Theme creation requires HTML/CSS knowledge anyway
3. **Complexity**: WYSIWYG editors are complex to build and maintain
4. **Alternatives**: Code-based workflow is faster for skilled designers

**Recommended Workflow for System Designers:**
```
1. Create HTML template file
2. Create CSS theme file
3. Add theme images to /static/images/themes/
4. Seed theme in database via migration
5. Test in staging
6. Deploy to production
```

**Example Theme Creation:**
```html
<!-- templates/web/rsvp_themes/wedding_elegance.html -->
<div class="rsvp-card" data-event-theme="wedding-elegance">
    <div class="rsvp-card-header">
        <img src="/static/images/themes/wedding-elegance-header.jpg" 
             alt="Floral wedding design">
    </div>
    <div class="rsvp-card-content">
        <h1>{{.Event.Title}}</h1>
        <p class="event-date">{{.Event.StartTime}}</p>
        <!-- ... rest of template ... -->
    </div>
</div>
```

```css
/* static/css/themes/wedding-elegance.css */
[data-event-theme="wedding-elegance"] {
    --theme-primary: #f4c2c2;
    --theme-secondary: #d4af37;
    --theme-font-heading: 'Playfair Display', serif;
    /* ... rest of theme variables ... */
}
```

### 6.3 Customization UI for Event Managers

**Question:** Do event managers need a WYSIWYG editor?

**Answer:** **NO, but they need LIMITED customization options**

**Rationale:**
1. **Most users** want pre-designed themes (80% use case)
2. **Some users** want minor customization (15% use case)
3. **Few users** want full control (5% use case - defer to v2+)

**Recommended Customization Options (Phase 2):**

**Option 1: Image Upload**
- Replace theme header image with custom image
- Validation: size, format, dimensions
- Preview before saving

**Option 2: Color Picker (Future)**
- Override theme primary color
- Preview in real-time
- Limited to primary color only (simplicity)

**Option 3: Text Customization (Future)**
- Custom welcome message
- Custom RSVP button text
- Limited to text content, not layout

**UI Design:**
```
┌─────────────────────────────────────┐
│ Theme: Wedding Elegance             │
│                                     │
│ [Theme Preview]                     │
│                                     │
│ Customize (Optional):               │
│                                     │
│ ☐ Use custom header image           │
│   [Upload Image] (1200x400px)      │
│                                     │
│ ☐ Override primary color            │
│   [Color Picker] #f4c2c2            │
│                                     │
│ [Preview Changes] [Save]            │
└─────────────────────────────────────┘
```

**Implementation:**
- Simple form with conditional fields
- JavaScript for preview
- No WYSIWYG editor needed
- Saves customizations to event record

### 6.4 WYSIWYG Editor Decision Matrix

| Feature | System Designers | Event Managers |
|---------|-----------------|----------------|
| **Need WYSIWYG?** | ❌ No | ❌ No |
| **Alternative** | Code-based | Limited UI options |
| **Complexity** | High (full control) | Low (guided) |
| **Frequency** | Rare | Every event |
| **Priority** | v0 (code-based) | v0 (pre-designed), v1 (customization) |

**Conclusion:**
- **v0**: Pre-designed themes only, no customization
- **v1**: Add image upload and color picker
- **v2+**: Consider WYSIWYG if user demand justifies complexity

---

## 7. Implementation Phases

### Phase 1: Pre-Designed Theme Gallery (v0.1)

**Timeline:** 2-3 weeks  
**Priority:** High  
**Effort:** Medium

**Deliverables:**
1. 1 plain text theme
2. 5-10 card-based themes with images
3. Theme picker UI in event creation
4. Theme preview modal
5. Theme rendering on RSVP pages
6. Integration with light/dark system

**Tasks:**
- [ ] Design 10 theme concepts
- [ ] Create theme images (header + thumbnail)
- [ ] Build theme HTML templates
- [ ] Build theme CSS files
- [ ] Extend template model with theme fields
- [ ] Create theme picker UI component
- [ ] Create theme preview modal
- [ ] Update event creation form
- [ ] Update RSVP page rendering
- [ ] Seed themes in database
- [ ] Test all themes in light/dark modes
- [ ] Write integration tests

**Success Criteria:**
- Event manager can select theme during event creation
- Guest sees RSVP page with selected theme
- Theme adapts to guest's light/dark preference
- All themes work on mobile and desktop
- No performance degradation

### Phase 2: Custom Image Upload (v0.2)

**Timeline:** 1-2 weeks  
**Priority:** Medium  
**Effort:** Medium

**Deliverables:**
1. Image upload UI in theme customization
2. Image validation and processing
3. Image storage via storage provider
4. Image serving on RSVP pages
5. Preview with custom image

**Tasks:**
- [ ] Add image upload field to event form
- [ ] Implement image validation
- [ ] Integrate with storage provider
- [ ] Update RSVP rendering to use custom image
- [ ] Add preview functionality
- [ ] Test image upload flow
- [ ] Test image serving
- [ ] Security testing

**Success Criteria:**
- Event manager can upload custom header image
- Image is validated and stored securely
- RSVP page displays custom image
- Preview shows custom image before saving

### Phase 3: Color Customization (v1.0)

**Timeline:** 1 week  
**Priority:** Low  
**Effort:** Low

**Deliverables:**
1. Color picker UI
2. Real-time preview
3. Color override in rendering

**Tasks:**
- [ ] Add color picker to customization UI
- [ ] Implement real-time preview
- [ ] Store color override in event record
- [ ] Apply color override in RSVP rendering
- [ ] Test color combinations
- [ ] Accessibility testing (contrast)

**Success Criteria:**
- Event manager can override primary color
- Preview shows color change in real-time
- RSVP page uses custom color
- Color contrast meets WCAG AA

### Phase 4: WYSIWYG Editor (v2.0+)

**Timeline:** 3-4 weeks  
**Priority:** Very Low (defer based on demand)  
**Effort:** Very High

**Deliverables:**
1. Full WYSIWYG editor integration
2. Template variable system
3. Layout customization
4. Advanced preview

**Decision Point:**
- Only implement if Phase 1-3 show strong user demand
- Evaluate alternatives (e.g., Markdown editor)
- Consider maintenance burden

---

## 8. Technical Specifications

### 8.1 Database Schema Changes

**Add to `templates` table:**
```sql
ALTER TABLE templates ADD COLUMN category TEXT;
ALTER TABLE templates ADD COLUMN description TEXT;
ALTER TABLE templates ADD COLUMN thumbnail_url TEXT;
ALTER TABLE templates ADD COLUMN image_url TEXT;
ALTER TABLE templates ADD COLUMN tags TEXT; -- JSON array
ALTER TABLE templates ADD COLUMN sort_order INTEGER DEFAULT 0;

CREATE INDEX idx_templates_category ON templates(category);
CREATE INDEX idx_templates_sort_order ON templates(sort_order);
```

**Add to `events` table:**
```sql
ALTER TABLE events ADD COLUMN custom_theme_image_url TEXT;
ALTER TABLE events ADD COLUMN custom_theme_color TEXT;
```

### 8.2 File Structure

```
static/
├── images/
│   └── themes/
│       ├── plain-text-thumb.jpg
│       ├── wedding-elegance-header.jpg
│       ├── wedding-elegance-thumb.jpg
│       ├── birthday-celebration-header.jpg
│       ├── birthday-celebration-thumb.jpg
│       └── ... (more themes)
├── css/
│   └── themes/
│       ├── plain-text.css
│       ├── wedding-elegance.css
│       ├── birthday-celebration.css
│       └── ... (more themes)
└── js/
    └── theme_picker.js

templates/
└── web/
    ├── rsvp_themes/
    │   ├── plain-text.html
    │   ├── wedding-elegance.html
    │   ├── birthday-celebration.html
    │   └── ... (more themes)
    └── partials/
        └── theme_picker.html
```

### 8.3 Theme Rendering Logic

```go
// internal/handlers/rsvp.go
func (h *Handler) HandleRSVPPage(w http.ResponseWriter, r *http.Request) {
    // ... token validation ...
    
    // Get event with template
    event, err := h.eventService.GetEvent(ctx, invite.EventID)
    if err != nil {
        return err
    }
    
    // Get theme template
    var template *models.Template
    if event.TemplateID != nil {
        template, err = h.templateService.GetTemplate(ctx, *event.TemplateID)
    } else {
        template, err = h.templateService.GetDefaultTemplate(ctx, "rsvp_page")
    }
    
    // Prepare template data
    data := struct {
        Event           *models.Event
        Invite          *models.Invite
        Questions       []*models.PreferenceQuestion
        ThemeCategory   string
        ThemeImageURL   string
        ThemeColor      string
    }{
        Event:         event,
        Invite:        invite,
        Questions:     questions,
        ThemeCategory: template.Category,
        ThemeImageURL: event.CustomThemeImageURL, // or template.ImageURL
        ThemeColor:    event.CustomThemeColor,    // or nil
    }
    
    // Render with theme
    return h.renderer.Render(w, template.HTMLContent, data)
}
```

### 8.4 Theme Picker Component

```html
<!-- templates/web/partials/theme_picker.html -->
<div class="theme-picker">
    <h3>Select Theme</h3>
    
    <div class="theme-gallery">
        {{range .Themes}}
        <div class="theme-card" data-theme-id="{{.ID}}">
            <img src="{{.ThumbnailURL}}" alt="{{.Name}}">
            <h4>{{.Name}}</h4>
            <p>{{.Description}}</p>
            <button type="button" class="btn-preview" 
                    data-theme-id="{{.ID}}">
                Preview
            </button>
            <button type="button" class="btn-select" 
                    data-theme-id="{{.ID}}">
                Select
            </button>
        </div>
        {{end}}
    </div>
</div>

<!-- Theme Preview Modal -->
<div id="theme-preview-modal" class="modal" hidden>
    <div class="modal-content">
        <h3>Theme Preview</h3>
        <iframe id="theme-preview-frame" src=""></iframe>
        <button type="button" class="btn-close">Close</button>
    </div>
</div>
```

---

## 9. Recommendations

### 9.1 Immediate Actions (This Sprint)

1. **Approve Phase 1 Scope**
   - 1 plain text theme
   - 5-10 card-based themes
   - Theme picker UI
   - No customization yet

2. **Design Theme Concepts**
   - Sketch 10 theme designs
   - Get stakeholder feedback
   - Finalize 5-10 themes

3. **Create Theme Assets**
   - Design header images (1200x400px)
   - Create thumbnail images (300x200px)
   - Optimize for web (<150KB each)

4. **Extend Light/Dark System**
   - Complete Story 10.12 first
   - Ensure all ~50 CSS variables have dark mode
   - Test thoroughly

### 9.2 Phase 1 Implementation Order

1. **Week 1: Foundation**
   - Extend template model
   - Create database migration
   - Design and create theme images
   - Build theme HTML templates
   - Build theme CSS files

2. **Week 2: UI Components**
   - Build theme picker component
   - Build theme preview modal
   - Update event creation form
   - Update RSVP page rendering

3. **Week 3: Testing & Polish**
   - Seed themes in database
   - Integration testing
   - Visual testing (light/dark)
   - Mobile testing
   - Accessibility testing
   - Bug fixes and polish

### 9.3 Future Considerations (v1+)

1. **Phase 2: Custom Images**
   - Only implement if Phase 1 successful
   - Gauge user demand first

2. **Phase 3: Color Customization**
   - Low priority
   - Simple to add later

3. **Phase 4: WYSIWYG**
   - Defer to v2+ based on user feedback
   - Evaluate alternatives (Markdown, template builder)
   - Consider maintenance burden

4. **Additional Theme Categories**
   - Modern flat designs
   - Classic formal designs
   - Fun/playful designs
   - Seasonal themes

5. **Theme Marketplace (v3+)**
   - User-submitted themes
   - Theme ratings/reviews
   - Premium themes

---

## 10. Comparison with Competitors

### Evite
- ✅ Card-based designs
- ✅ Large theme gallery (100+)
- ✅ Image customization
- ✅ Text customization
- ❌ Self-hosted option

### Paperless Post
- ✅ Premium card designs
- ✅ Extensive customization
- ✅ Animation support
- ❌ Self-hosted option
- ❌ Expensive

### TinyRSVP (Proposed)
- ✅ Self-hosted
- ✅ Card-based designs (5-10 initially)
- ✅ Light/dark mode support
- ✅ Privacy-focused
- ✅ Free and open source
- ⏳ Image customization (Phase 2)
- ⏳ Color customization (Phase 3)
- ❌ Animation (not planned)

**Competitive Position:**
- Phase 1 gets us to 70% feature parity for core use cases
- Self-hosted + privacy is unique differentiator
- Light/dark mode support is unique feature
- Can expand theme library over time

---

## 11. Answers to Original Questions

### Q: Should we have a theme/template manager for event creation?

**A: YES - Theme picker during event creation**

Implement a theme selection step (or dropdown) in the event creation flow where event managers can:
1. Browse available themes in a gallery
2. Preview themes before selection
3. Select theme for their event
4. (Phase 2) Optionally customize with own image

This should be the **first visual decision** after basic event details (title, date, location).

### Q: Should themes be card-based like Evite?

**A: YES - Start with card-based, allow extensibility**

- **v0**: Card-based themes (like Evite)
- **v1+**: Add other styles (modern, classic, fun)
- **Architecture**: Category field allows future expansion

Card-based is proven, familiar, and works well for most event types.

### Q: Do we need a WYSIWYG editor?

**A: NO for v0-v1, MAYBE for v2+**

**For System Designers:** No, code-based workflow is better
**For Event Managers:** No, limited customization UI is sufficient

Only consider WYSIWYG if:
- Strong user demand after Phase 1-3
- Resources available for 3-4 week implementation
- Maintenance burden acceptable

### Q: How does this relate to light/dark theme switching?

**A: Two-layer system - they work together**

- **Layer 1**: System theme (light/dark) - user preference
- **Layer 2**: Event theme (visual design) - event manager selection
- **Result**: Event theme adapts to user's light/dark preference

Example: "Wedding Elegance" theme works in both light and dark modes, with colors adjusted automatically.

---

## 12. Infrastructure Assessment

### 12.1 Template System: 80% Ready ⚠️

**Complete:**
- ✅ Template model with CRUD operations
- ✅ Template types (invite_email, rsvp_page, confirmation_page)
- ✅ Go html/template rendering
- ✅ XSS prevention
- ✅ Default template system

**Needs:**
- ⚠️ Theme-specific fields (category, description, thumbnail, image, tags)
- ⚠️ Theme picker UI
- ⚠️ Theme preview functionality
- ⚠️ Multiple RSVP page templates (currently only 1)

### 12.2 CSS System: 60% Ready ⚠️

**Complete:**
- ✅ CSS variables system
- ✅ Basic light/dark support (6 variables)
- ✅ Responsive grid
- ✅ Component library

**Needs:**
- ⚠️ Complete light/dark palette (~50 variables) - Story 10.12
- ⚠️ Event theme variable layer
- ⚠️ Theme-specific CSS files
- ⚠️ Smooth transitions

### 12.3 Storage System: 100% Ready ✅

**Complete:**
- ✅ Storage provider interface
- ✅ Local filesystem implementation
- ✅ Image validation
- ✅ Asset serving
- ✅ Security (path validation, content type)

**Ready for:**
- ✅ Theme image storage
- ✅ Custom image uploads (Phase 2)

### 12.4 Overall Readiness: 70%

**Foundation is solid, needs:**
1. Complete Story 10.12 (light/dark system) first
2. Extend template model for themes
3. Create theme assets and templates
4. Build theme picker UI

---

## 13. Implementation Effort Breakdown

### Phase 1: Pre-Designed Theme Gallery

**Design Work (3-4 days):**
- Concept 10 theme designs
- Create header images (1200x400px)
- Create thumbnail images (300x200px)
- Optimize images for web
- Get stakeholder approval

**Backend Work (3-4 days):**
- Extend template model (category, description, etc.)
- Create database migration
- Update template service
- Create theme seeding system
- Write unit tests

**Frontend Work (4-5 days):**
- Create theme CSS files (10 themes)
- Create theme HTML templates (10 themes)
- Build theme picker component
- Build theme preview modal
- Update event creation form
- Write integration tests

**Integration & Testing (2-3 days):**
- Seed themes in database
- Test all themes in light/dark modes
- Mobile testing
- Accessibility testing
- Bug fixes

**Total: 12-16 days (2-3 weeks)**

### Phase 2: Custom Image Upload

**Backend Work (2-3 days):**
- Image upload handler
- Image validation
- Storage integration
- Update event model

**Frontend Work (2-3 days):**
- Image upload UI
- Preview functionality
- Form integration

**Testing (2 days):**
- Security testing
- Integration testing

**Total: 6-8 days (1-2 weeks)**

### Phase 3: Color Customization

**Backend Work (1 day):**
- Store color override in event

**Frontend Work (2-3 days):**
- Color picker UI
- Real-time preview
- CSS variable override

**Testing (1 day):**
- Contrast testing
- Integration testing

**Total: 4-5 days (1 week)**

---

## 14. Risk Assessment

### 14.1 Technical Risks

**Risk: Theme complexity grows unmanageable**
- **Likelihood:** Medium
- **Impact:** High
- **Mitigation:** Start with 5 themes, add gradually based on demand

**Risk: Light/dark integration issues**
- **Likelihood:** Medium
- **Impact:** Medium
- **Mitigation:** Complete Story 10.12 first, test thoroughly

**Risk: Performance degradation with many themes**
- **Likelihood:** Low
- **Impact:** Medium
- **Mitigation:** Lazy load theme CSS, optimize images

**Risk: Mobile rendering issues**
- **Likelihood:** Medium
- **Impact:** High
- **Mitigation:** Mobile-first design, extensive testing

### 14.2 Design Risks

**Risk: Themes don't match user expectations**
- **Likelihood:** Medium
- **Impact:** Medium
- **Mitigation:** Research Evite designs, get user feedback early

**Risk: Images don't work in dark mode**
- **Likelihood:** Medium
- **Impact:** Medium
- **Mitigation:** Design images for both modes, use overlays

**Risk: Too many themes overwhelm users**
- **Likelihood:** Low
- **Impact:** Low
- **Mitigation:** Start with 5-7 themes, categorize clearly

### 14.3 Scope Risks

**Risk: Feature creep (WYSIWYG, animations, etc.)**
- **Likelihood:** High
- **Impact:** High
- **Mitigation:** Strict phase boundaries, defer Phase 4

**Risk: Customization demands exceed simple UI**
- **Likelihood:** Medium
- **Impact:** Medium
- **Mitigation:** Phase 2-3 provide 80% of customization needs

---

## 15. Success Criteria

### Phase 1 Success Metrics

**Functional:**
- [ ] Event manager can select from 5-10 themes
- [ ] Event manager can preview themes before selection
- [ ] Guest sees RSVP page with selected theme
- [ ] Theme works in both light and dark modes
- [ ] All themes work on mobile and desktop
- [ ] Plain text theme available for accessibility

**Performance:**
- [ ] RSVP page loads in <2 seconds
- [ ] Theme images <150KB each
- [ ] No layout shift during load
- [ ] Smooth theme transitions

**Quality:**
- [ ] All themes pass accessibility audit (WCAG AA)
- [ ] All themes tested in light/dark modes
- [ ] All themes tested on mobile/tablet/desktop
- [ ] Integration tests pass
- [ ] Visual regression tests pass

### Phase 2 Success Metrics

**Functional:**
- [ ] Event manager can upload custom header image
- [ ] Image validation works (size, format, dimensions)
- [ ] Custom image displays on RSVP page
- [ ] Preview shows custom image

**Security:**
- [ ] Image validation prevents malicious uploads
- [ ] EXIF data stripped
- [ ] File size limits enforced
- [ ] Content type validation works

---

## 16. Open Questions

### 16.1 Design Questions

1. **How many themes for v0?**
   - Option A: 5 themes (faster to market)
   - Option B: 10 themes (more choice)
   - **Recommendation:** 7 themes (good balance)

2. **Should themes have animations?**
   - Option A: Static only (simpler)
   - Option B: Subtle animations (modern)
   - **Recommendation:** Static for v0, animations in v1+

3. **Should we support custom fonts?**
   - Option A: System fonts only (faster, privacy)
   - Option B: Google Fonts (more choice)
   - **Recommendation:** System fonts for v0, consider Google Fonts in v1

### 16.2 Technical Questions

1. **How to handle theme updates?**
   - Option A: Updates apply to all events using theme
   - Option B: Events snapshot theme at creation
   - **Recommendation:** Option A for v0 (simpler), Option B for v1 (safer)

2. **Should themes be versioned?**
   - Option A: No versioning (simpler)
   - Option B: Version themes (safer)
   - **Recommendation:** No versioning in v0, add in v1 if needed

3. **How to organize theme files?**
   - Option A: Single template with CSS classes
   - Option B: Separate template per theme
   - **Recommendation:** Option B (easier to maintain, clearer separation)

### 16.3 Scope Questions

1. **Should confirmation page use same theme?**
   - Option A: Yes, consistent experience
   - Option B: No, separate template
   - **Recommendation:** Yes, apply event theme to confirmation page too

2. **Should email invitations match theme?**
   - Option A: Yes, consistent branding
   - Option B: No, keep emails simple
   - **Recommendation:** No for v0 (email clients limited), consider v1+

---

## 17. Dependency Analysis

### 17.1 Prerequisites

**Must Complete First:**
- ✅ Story 10.12: Light/Dark Theme Switching
  - Provides foundation for two-layer theme system
  - Ensures all CSS variables have dark mode
  - Establishes theme switching patterns

**Should Complete First:**
- ✅ Story 06.03: Default Templates
  - Establishes template seeding patterns
  - Provides template CRUD foundation

### 17.2 Blocks

**This Epic Blocks:**
- Future customization features
- Theme marketplace (v3+)
- Advanced template features

**This Epic Enables:**
- Professional-looking RSVP pages
- Event-specific branding
- Competitive feature parity

---

## 18. Conclusion

### 18.1 Recommended Path Forward

**Immediate (This Sprint):**
1. ✅ Complete Story 10.12 (Light/Dark Theme Switching)
2. ✅ Create user stories for RSVP Page Theme System
3. ✅ Get stakeholder approval on theme concepts

**Phase 1 (Next 2-3 weeks):**
1. Design and create 7 theme concepts
2. Implement theme picker UI
3. Implement theme rendering
4. Test thoroughly
5. Deploy to production

**Phase 2 (Following 1-2 weeks):**
1. Add custom image upload
2. Test security thoroughly
3. Deploy to production

**Phase 3 (Future):**
1. Evaluate user demand
2. Add color customization if justified
3. Consider additional theme categories

**Phase 4 (v2+):**
1. Defer WYSIWYG based on user feedback
2. Focus on core value first

### 18.2 Key Decisions

**Architecture:**
- ✅ Two-layer theme system (system + event)
- ✅ Data attribute approach for themes
- ✅ Separate template per theme
- ✅ Category-based extensibility

**Scope:**
- ✅ 1 plain text + 5-10 card-based themes
- ✅ Theme picker during event creation
- ✅ Theme preview functionality
- ❌ No WYSIWYG editor for v0-v1
- ❌ No animations for v0

**Customization:**
- ✅ Phase 1: Pre-designed themes only
- ✅ Phase 2: Custom image upload
- ✅ Phase 3: Color picker
- ❌ Phase 4: WYSIWYG (defer)

### 18.3 Success Definition

**Phase 1 is successful if:**
1. Event managers can easily select themes
2. Themes look professional and polished
3. Themes work perfectly in light/dark modes
4. Mobile experience is excellent
5. No performance degradation
6. Users request more themes (indicates value)

**Proceed to Phase 2 if:**
1. Phase 1 successful
2. Users request customization
3. Resources available

**Proceed to Phase 4 (WYSIWYG) only if:**
1. Phase 1-3 successful
2. Strong user demand (>30% of users want it)
3. Simpler alternatives exhausted
4. Resources available for 3-4 week project

---

## 19. Next Steps

1. **Review this analysis** with stakeholder
2. **Create user stories** based on approved scope
3. **Complete Story 10.12** (light/dark theme) first
4. **Design theme concepts** (7 themes)
5. **Begin Phase 1 implementation**

---

**Status:** ✅ Analysis Complete  
**Next Action:** Create user stories for Epic 11 (RSVP Page Themes)

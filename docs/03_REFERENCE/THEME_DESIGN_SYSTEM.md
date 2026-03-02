# Theme Design System

**Version:** 1.0  
**Last Updated:** 2026-01-11  
**Status:** Complete

---

## Overview

TinyRSVP implements a two-layer theme system for RSVP pages:

1. **System Theme (Light/Dark)** - User preference, affects all pages
2. **Event Theme (Visual Design)** - Event manager selection, affects RSVP pages only

This document describes the 7 pre-designed themes available in v0.1.

---

## Theme Architecture

### Two-Layer CSS Variable System

```css
/* Layer 1: System Theme - User's light/dark preference */
:root {
    --color-background: #ffffff;
    --color-text-primary: #111827;
}

[data-theme="dark"] {
    --color-background: #0f172a;
    --color-text-primary: #f1f5f9;
}

/* Layer 2: Event Theme - Event manager's design choice */
[data-event-theme="wedding-elegance"] {
    --theme-primary: #f4c2c2;
    --theme-secondary: #d4af37;
    --theme-accent: #8b4789;
}

/* Layer 3: Combined Application */
.event-title {
    color: var(--theme-primary);      /* From event theme */
    background: var(--color-surface); /* From system theme */
}
```

### Data Attributes

- `data-theme="light|dark"` - Applied to `<html>` tag, controls system theme
- `data-event-theme="theme-name"` - Applied to `<html>` tag, controls event theme

---

## Available Themes

### 1. Plain Text Theme

**Name:** Simple & Clean  
**Category:** plain  
**Theme ID:** `plain-text`

**Description:** Minimalist text-based invitation, perfect for accessibility and fast loading.

**Colors:**
- Primary: Uses system primary color
- Secondary: Uses system gray
- Accent: Uses system primary

**Fonts:**
- Heading: System sans-serif
- Body: System sans-serif

**Features:**
- No header image
- Clean typography
- Minimal design
- Fast loading (<1 second)
- Accessibility-first

**Files:**
- Template: `templates/web/rsvp_themes/plain-text.html`
- CSS: `static/css/themes/plain-text.css`
- Thumbnail: `static/images/themes/plain-text-thumb.svg`

---

### 2. Wedding Elegance

**Name:** Wedding Elegance  
**Category:** card  
**Theme ID:** `wedding-elegance`

**Description:** Elegant floral design perfect for weddings and formal celebrations.

**Colors:**
- Primary: `#f4c2c2` (Blush)
- Secondary: `#d4af37` (Gold)
- Accent: `#8b4789` (Plum)

**Dark Mode Adjustments:**
- Primary: `#f4c2c2` (unchanged)
- Secondary: `#e6c84e` (brighter gold)
- Accent: `#a855b5` (brighter plum)

**Fonts:**
- Heading: Georgia, Times New Roman, serif
- Body: System sans-serif

**Features:**
- Floral header image
- Elegant serif typography
- Soft romantic colors
- Formal tone

**Files:**
- Template: `templates/web/rsvp_themes/wedding-elegance.html`
- CSS: `static/css/themes/wedding-elegance.css`
- Header: `static/images/themes/wedding-elegance-header.svg` (1200x400px)
- Thumbnail: `static/images/themes/wedding-elegance-thumb.svg` (300x200px)

---

### 3. Birthday Celebration

**Name:** Birthday Celebration  
**Category:** card  
**Theme ID:** `birthday-celebration`

**Description:** Fun and colorful design for birthday parties and celebrations.

**Colors:**
- Primary: `#ff6b9d` (Pink)
- Secondary: `#ffd93d` (Yellow)
- Accent: `#6bcf7f` (Green)

**Dark Mode Adjustments:**
- Colors remain vibrant in dark mode (no adjustment needed)

**Fonts:**
- Heading: System sans-serif (rounded)
- Body: System sans-serif

**Features:**
- Balloons and confetti header
- Playful bright colors
- Fun energetic tone
- Gradient button

**Files:**
- Template: `templates/web/rsvp_themes/birthday-celebration.html`
- CSS: `static/css/themes/birthday-celebration.css`
- Header: `static/images/themes/birthday-celebration-header.svg` (1200x400px)
- Thumbnail: `static/images/themes/birthday-celebration-thumb.svg` (300x200px)

---

### 4. Corporate Professional

**Name:** Corporate Professional  
**Category:** card  
**Theme ID:** `corporate-professional`

**Description:** Clean and professional design for business events and meetings.

**Colors:**
- Primary: `#2563eb` (Blue)
- Secondary: `#64748b` (Gray)
- Accent: `#0ea5e9` (Light Blue)

**Dark Mode Adjustments:**
- Primary: `#3b82f6` (brighter blue)
- Secondary: `#94a3b8` (lighter gray)
- Accent: `#06b6d4` (brighter cyan)

**Fonts:**
- Heading: System sans-serif
- Body: System sans-serif

**Features:**
- Abstract geometric header
- Professional color palette
- Uppercase button text
- Clean modern layout

**Files:**
- Template: `templates/web/rsvp_themes/corporate-professional.html`
- CSS: `static/css/themes/corporate-professional.css`
- Header: `static/images/themes/corporate-professional-header.svg` (1200x400px)
- Thumbnail: `static/images/themes/corporate-professional-thumb.svg` (300x200px)

---

### 5. Holiday Festive

**Name:** Holiday Festive  
**Category:** card  
**Theme ID:** `holiday-festive`

**Description:** Warm and festive design for holiday gatherings and seasonal events.

**Colors:**
- Primary: `#dc2626` (Red)
- Secondary: `#16a34a` (Green)
- Accent: `#d4af37` (Gold)

**Dark Mode Adjustments:**
- Primary: `#ef4444` (brighter red)
- Secondary: `#22c55e` (brighter green)
- Accent: `#fbbf24` (brighter gold)

**Fonts:**
- Heading: Georgia, Times New Roman, serif
- Body: System sans-serif

**Features:**
- Snowflakes and ornaments header
- Traditional holiday colors
- Warm festive tone
- Decorative borders

**Files:**
- Template: `templates/web/rsvp_themes/holiday-festive.html`
- CSS: `static/css/themes/holiday-festive.css`
- Header: `static/images/themes/holiday-festive-header.svg` (1200x400px)
- Thumbnail: `static/images/themes/holiday-festive-thumb.svg` (300x200px)

---

### 6. Garden Party

**Name:** Garden Party  
**Category:** card  
**Theme ID:** `garden-party`

**Description:** Fresh botanical design for outdoor events and garden parties.

**Colors:**
- Primary: `#16a34a` (Green)
- Secondary: `#84cc16` (Lime)
- Accent: `#fbbf24` (Amber)

**Dark Mode Adjustments:**
- Primary: `#22c55e` (brighter green)
- Secondary: `#a3e635` (brighter lime)
- Accent: `#fcd34d` (brighter amber)

**Fonts:**
- Heading: Georgia, Times New Roman, serif
- Body: System sans-serif

**Features:**
- Botanical leaves and flowers header
- Natural earth tones
- Organic casual tone
- Nature-inspired design

**Files:**
- Template: `templates/web/rsvp_themes/garden-party.html`
- CSS: `static/css/themes/garden-party.css`
- Header: `static/images/themes/garden-party-header.svg` (1200x400px)
- Thumbnail: `static/images/themes/garden-party-thumb.svg` (300x200px)

---

### 7. Modern Minimalist

**Name:** Modern Minimalist  
**Category:** card  
**Theme ID:** `modern-minimalist`

**Description:** Contemporary minimal design with clean lines and bold typography.

**Colors:**
- Primary: `#0f172a` (Dark Blue)
- Secondary: `#64748b` (Gray)
- Accent: `#06b6d4` (Cyan)

**Dark Mode Adjustments:**
- Primary: `#f1f5f9` (inverted to light)
- Secondary: `#cbd5e1` (lighter gray)
- Accent: `#22d3ee` (brighter cyan)

**Fonts:**
- Heading: System sans-serif
- Body: System sans-serif

**Features:**
- Simple geometric header
- Minimal clean design
- Contemporary tone
- Uppercase labels

**Files:**
- Template: `templates/web/rsvp_themes/modern-minimalist.html`
- CSS: `static/css/themes/modern-minimalist.css`
- Header: `static/images/themes/modern-minimalist-header.svg` (1200x400px)
- Thumbnail: `static/images/themes/modern-minimalist-thumb.svg` (300x200px)

---

## Theme Variables

Each theme defines the following CSS variables:

### Required Variables

```css
[data-event-theme="theme-name"] {
    --theme-primary: #hexcolor;      /* Primary brand color */
    --theme-secondary: #hexcolor;    /* Secondary brand color */
    --theme-accent: #hexcolor;       /* Accent color */
    --theme-font-heading: font-stack; /* Heading font family */
    --theme-font-body: font-stack;    /* Body font family */
}
```

### Dark Mode Overrides

```css
[data-theme="dark"][data-event-theme="theme-name"] {
    --theme-primary: #hexcolor;      /* Adjusted for dark mode */
    --theme-secondary: #hexcolor;    /* Adjusted for dark mode */
    --theme-accent: #hexcolor;       /* Adjusted for dark mode */
}
```

---

## Image Specifications

### Header Images

**Dimensions:** 1200x400px (3:1 aspect ratio)  
**Format:** SVG (scalable, small file size)  
**Max Size:** 50KB per file  
**Location:** `/static/images/themes/`

**Requirements:**
- Must include `viewBox="0 0 1200 400"` for responsiveness
- Should work in both light and dark modes
- No text in images (accessibility)
- Appropriate for theme category

### Thumbnail Images

**Dimensions:** 300x200px (3:2 aspect ratio)  
**Format:** SVG (scalable, small file size)  
**Max Size:** 30KB per file  
**Location:** `/static/images/themes/`

**Requirements:**
- Must include `viewBox="0 0 300 200"` for responsiveness
- Should represent the theme visually
- Used in theme picker UI

---

## HTML Template Structure

### Standard Structure

All theme templates follow this structure:

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
                <h1 class="event-title">{{.Event.Title}}</h1>
                
                <div class="event-details">
                    <!-- Event information -->
                </div>
                
                <form method="POST" action="/api/rsvp/{{.Token}}" class="rsvp-form">
                    <!-- RSVP form fields -->
                </form>
            </div>
        </div>
    </div>
</body>
</html>
```

### Go Template Variables

All templates use these Go template variables:

- `{{.Event.Title}}` - Event title
- `{{.Event.StartTime}}` - Event start time
- `{{.Event.EndTime}}` - Event end time (optional)
- `{{.Event.Location}}` - Event location (optional)
- `{{.Event.Timezone}}` - Event timezone (optional)
- `{{.Event.Description}}` - Event description (optional)
- `{{.Event.RSVPDeadline}}` - RSVP deadline (optional)
- `{{.Event.CustomThemeImageURL}}` - Custom header image (optional)
- `{{.Token}}` - Guest invite token
- `{{.CSRFToken}}` - CSRF protection token
- `{{.MaxPlusOnes}}` - Maximum additional guests
- `{{.Questions}}` - Preference questions array

### Template Functions

- `{{formatDateTime .Event.StartTime}}` - Format date and time
- `{{formatTime .Event.EndTime}}` - Format time only
- `{{formatDate .Event.RSVPDeadline "Monday, January 2, 2006"}}` - Format date with layout
- `{{if gt .MaxPlusOnes 0}}` - Conditional rendering

---

## CSS Class Structure

### Container Classes

- `.rsvp-container` - Outer container, centers card
- `.rsvp-card` - Main card element
- `.rsvp-card-header` - Header image section (card-based themes only)
- `.rsvp-card-content` - Content area with padding

### Content Classes

- `.event-title` - Event title (h1)
- `.event-details` - Event information section
- `.event-date` - Date/time information
- `.event-location` - Location information
- `.event-description` - Event description

### Form Classes

- `.rsvp-form` - Form element
- `.form-group` - Form field wrapper
- `.radio-group` - Radio button group
- `.btn-primary` - Primary submit button
- `.required` - Required field indicator
- `.help-text` - Help text for form fields

---

## Responsive Behavior

All themes are mobile-responsive with breakpoints:

### Mobile (320px - 767px)

- Single column layout
- Reduced padding
- Smaller header images (200px height)
- Smaller typography
- Full-width buttons

### Tablet (768px - 1023px)

- Standard layout
- Normal padding
- Full header images (300px height)
- Standard typography

### Desktop (1024px+)

- Standard layout
- Maximum width constraints (800px)
- Full header images (300px height)
- Standard typography

---

## Creating New Themes

### Step 1: Design Concept

1. Choose theme category (plain, card, modern, etc.)
2. Select color palette (primary, secondary, accent)
3. Choose typography (heading and body fonts)
4. Design header image concept

### Step 2: Create Assets

**Header Image (Card-based themes only):**
```svg
<svg viewBox="0 0 1200 400" xmlns="http://www.w3.org/2000/svg">
    <!-- SVG content with theme colors -->
</svg>
```

**Thumbnail Image:**
```svg
<svg viewBox="0 0 300 200" xmlns="http://www.w3.org/2000/svg">
    <!-- SVG content representing theme -->
</svg>
```

### Step 3: Create CSS File

```css
/* static/css/themes/new-theme.css */

[data-event-theme="new-theme"] {
    --theme-primary: #hexcolor;
    --theme-secondary: #hexcolor;
    --theme-accent: #hexcolor;
    --theme-font-heading: font-stack;
    --theme-font-body: font-stack;
}

[data-theme="dark"][data-event-theme="new-theme"] {
    --theme-primary: #hexcolor;
    --theme-secondary: #hexcolor;
    --theme-accent: #hexcolor;
}

/* Theme-specific styles */
[data-event-theme="new-theme"] .rsvp-card {
    /* Custom card styling */
}

/* Responsive adjustments */
@media (max-width: 767px) {
    /* Mobile overrides */
}
```

### Step 4: Create HTML Template

Copy an existing theme template and update:
1. `data-event-theme` attribute
2. CSS file link
3. Header image path
4. Alt text

### Step 5: Add Tests

Update test files to include new theme:
- `static/images/themes/theme_assets_test.go`
- `static/css/themes/theme_css_test.go`
- `templates/web/rsvp_themes/theme_templates_test.go`

### Step 6: Run Tests

```bash
go test -timeout 30s ./static/images/themes/...
go test -timeout 30s ./static/css/themes/...
go test -timeout 30s ./templates/web/rsvp_themes/...
```

---

## Design Guidelines

### Color Selection

**Primary Color:**
- Main brand color for the theme
- Used for titles, borders, key elements
- Should have good contrast with backgrounds

**Secondary Color:**
- Supporting color
- Used for accents, hover states
- Should complement primary

**Accent Color:**
- Highlight color
- Used for labels, required indicators
- Should stand out but not clash

### Dark Mode Strategy

**Approach 1: Keep Colors (Vibrant Themes)**
- Birthday, Garden Party themes
- Colors work well in both modes
- No adjustment needed

**Approach 2: Brighten Colors (Subtle Themes)**
- Wedding, Holiday themes
- Increase brightness for dark mode
- Maintain color identity

**Approach 3: Invert Colors (High Contrast Themes)**
- Modern Minimalist theme
- Invert dark/light for contrast
- Maintain readability

### Typography

**System Font Stacks:**
```css
/* Sans-serif (default) */
--font-family-sans: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;

/* Serif (formal themes) */
--font-family-serif: Georgia, "Times New Roman", serif;
```

**Font Pairing:**
- Serif heading + Sans body = Formal/elegant
- Sans heading + Sans body = Modern/professional
- Rounded heading + Sans body = Fun/playful

### Image Design

**Header Images (1200x400px):**
- Use theme colors
- Abstract or decorative elements
- No text (accessibility)
- Works in light and dark modes
- Subtle, not overwhelming

**Thumbnail Images (300x200px):**
- Simplified version of header
- Shows theme identity clearly
- Includes theme name text
- Card representation

---

## Accessibility

### WCAG AA Compliance

All themes must meet:
- Color contrast ratio ≥ 4.5:1 for normal text
- Color contrast ratio ≥ 3:1 for large text
- Keyboard navigation support
- Screen reader friendly
- Semantic HTML structure

### Best Practices

- Alt text for all images
- Proper heading hierarchy (h1, h2, h3)
- Form labels associated with inputs
- Focus indicators visible
- No text in images

---

## Performance

### Target Metrics

- First Contentful Paint: <1s
- Time to Interactive: <2s
- Total page weight: <100KB

### Optimization

**SVG Images:**
- Scalable without quality loss
- Small file sizes (<50KB)
- No HTTP requests for data URIs
- CSS-styleable

**CSS:**
- Minimal theme-specific styles
- Leverage CSS variables
- No redundant rules

**HTML:**
- Semantic structure
- Minimal DOM nodes
- Progressive enhancement

---

## Testing

### Asset Tests

```bash
# Test images exist and meet specifications
go test -timeout 30s ./static/images/themes/...

# Test CSS files exist and have required variables
go test -timeout 30s ./static/css/themes/...

# Test HTML templates exist and have correct structure
go test -timeout 30s ./templates/web/rsvp_themes/...
```

### Visual Testing

Test each theme in:
- Light mode
- Dark mode
- Mobile (320px, 375px, 414px)
- Tablet (768px, 1024px)
- Desktop (1280px, 1920px)

### Browser Testing

Test in:
- Chrome/Edge (Chromium)
- Firefox
- Safari
- Mobile browsers

---

## File Organization

```
static/
├── images/
│   └── themes/
│       ├── plain-text-thumb.svg
│       ├── wedding-elegance-header.svg
│       ├── wedding-elegance-thumb.svg
│       ├── birthday-celebration-header.svg
│       ├── birthday-celebration-thumb.svg
│       ├── corporate-professional-header.svg
│       ├── corporate-professional-thumb.svg
│       ├── holiday-festive-header.svg
│       ├── holiday-festive-thumb.svg
│       ├── garden-party-header.svg
│       ├── garden-party-thumb.svg
│       ├── modern-minimalist-header.svg
│       ├── modern-minimalist-thumb.svg
│       └── theme_assets_test.go
├── css/
│   └── themes/
│       ├── plain-text.css
│       ├── wedding-elegance.css
│       ├── birthday-celebration.css
│       ├── corporate-professional.css
│       ├── holiday-festive.css
│       ├── garden-party.css
│       ├── modern-minimalist.css
│       └── theme_css_test.go

templates/
└── web/
    └── rsvp_themes/
        ├── plain-text.html
        ├── wedding-elegance.html
        ├── birthday-celebration.html
        ├── corporate-professional.html
        ├── holiday-festive.html
        ├── garden-party.html
        ├── modern-minimalist.html
        └── theme_templates_test.go
```

---

## Integration with Template System

### Database Schema

Themes are stored in the `templates` table with these fields:

```sql
CREATE TABLE templates (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    html_content TEXT NOT NULL,
    css_content TEXT,
    is_default BOOLEAN DEFAULT FALSE,
    category TEXT NOT NULL DEFAULT 'plain',
    description TEXT,
    thumbnail_url TEXT,
    image_url TEXT,
    tags TEXT NOT NULL DEFAULT '[]',
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_by INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### Theme Selection

Event managers select themes during event creation:
1. Browse theme gallery (thumbnails)
2. Preview theme (optional)
3. Select theme
4. Theme ID stored with event

### Theme Rendering

When guest visits RSVP page:
1. Server loads event with template ID
2. Server renders HTML template with event data
3. Browser applies guest's light/dark preference
4. Result: Event theme + user's system theme

---

## Future Enhancements

### Phase 2: Custom Image Upload (v0.2)

- Event managers can upload custom header images
- Image validation and processing
- Storage via storage provider
- Preview with custom image

### Phase 3: Color Customization (v1.0)

- Color picker UI
- Override theme primary color
- Real-time preview
- Contrast validation

### Phase 4: Additional Themes

Potential future themes:
- Beach/Summer
- Baby Shower
- Graduation
- Retirement
- Anniversary
- Seasonal variants

---

## Troubleshooting

### Theme Not Applying

**Check:**
1. `data-event-theme` attribute on `<html>` tag
2. CSS file linked in `<head>`
3. CSS file exists at correct path
4. Theme variables defined in CSS

### Colors Wrong in Dark Mode

**Check:**
1. Dark mode overrides defined
2. `[data-theme="dark"]` selector present
3. Colors adjusted for dark backgrounds
4. Contrast ratios maintained

### Images Not Loading

**Check:**
1. Image files exist at correct path
2. Image paths in HTML template correct
3. SVG files have valid XML
4. `viewBox` attribute present

### Mobile Layout Issues

**Check:**
1. Responsive media queries defined
2. Mobile-first approach used
3. Viewport meta tag present
4. Touch-friendly tap targets (44px minimum)

---

## References

- **Story:** [11_STORY_02_theme_asset_creation.md](../../docs/00_BACKLOG/11_STORY_02_theme_asset_creation.md)
- **Epic:** [11_EPIC_rsvp_themes.md](../../docs/00_BACKLOG/11_EPIC_rsvp_themes.md)
- **Analysis:** [11_ANALYSIS_rsvp_page_themes.md](../../docs/00_BACKLOG/11_ANALYSIS_rsvp_page_themes.md)
- **CSS Variables:** [variables.css](../css/variables.css)
- **Theme Controller:** [theme_controller.js](../js/theme_controller.js)

---

**Status:** ✅ Complete  
**Version:** 1.0  
**Last Updated:** 2026-01-11

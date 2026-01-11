# User Story 11.02: Theme Asset Creation

**Epic:** [11_EPIC_rsvp_themes.md](11_EPIC_rsvp_themes.md)  
**Priority:** High  
**Status:** Not Started  
**Estimated Effort:** 3-4 days  
**Owner:** Unassigned

---

## User Story

As a **system designer**,  
I want to **create professional theme assets (images, CSS, HTML templates)**,  
So that **event managers have 7 beautiful pre-designed themes to choose from**.

---

## Context

This story involves creating the actual visual assets and code for 7 RSVP page themes:
- 1 plain text theme (no images, accessibility-first)
- 6 card-based themes with header images (wedding, birthday, corporate, holiday, garden, modern)

Each theme includes:
- Header image (1200x400px, <150KB)
- Thumbnail image (300x200px, <30KB)
- HTML template with theme structure
- CSS file with theme-specific variables
- Theme metadata (name, description, tags)

---

## Acceptance Criteria

### Plain Text Theme
- [ ] HTML template created (`templates/web/rsvp_themes/plain-text.html`)
- [ ] CSS file created (`static/css/themes/plain-text.css`)
- [ ] Thumbnail image created (300x200px, shows text layout)
- [ ] No header image (theme works without images)
- [ ] Clean typography, minimal design
- [ ] Fast loading (<1 second)
- [ ] Accessibility-first (high contrast, clear hierarchy)

### Card-Based Themes (6 themes)
- [ ] Wedding Elegance theme complete
- [ ] Birthday Celebration theme complete
- [ ] Corporate Professional theme complete
- [ ] Holiday Festive theme complete
- [ ] Garden Party theme complete
- [ ] Modern Minimalist theme complete

### Per Theme Requirements
- [ ] Header image (1200x400px, JPEG, <150KB)
- [ ] Thumbnail image (300x200px, JPEG, <30KB)
- [ ] HTML template file
- [ ] CSS theme file with variables
- [ ] Works in light and dark modes
- [ ] Mobile-responsive (320px to 1920px)
- [ ] Images optimized for web
- [ ] Alt text provided for images

### Image Quality Standards
- [ ] Images professionally designed or sourced
- [ ] Images work in both light and dark modes
- [ ] No text in images (accessibility)
- [ ] Appropriate for theme category
- [ ] Optimized file sizes
- [ ] Proper aspect ratios maintained

### CSS Theme Variables
- [ ] Each theme defines `--theme-primary`
- [ ] Each theme defines `--theme-secondary`
- [ ] Each theme defines `--theme-accent`
- [ ] Each theme defines `--theme-font-heading`
- [ ] Each theme defines `--theme-font-body`
- [ ] Variables work with system light/dark theme
- [ ] Smooth transitions defined

### HTML Template Structure
- [ ] Consistent structure across all themes
- [ ] Uses Go template variables ({{.Event.Title}}, etc.)
- [ ] Includes RSVP form
- [ ] Includes preference questions section
- [ ] Mobile-responsive layout
- [ ] Semantic HTML
- [ ] Accessibility attributes

---

## Technical Details

### Theme Specifications

#### 1. Plain Text Theme
**Name:** "Simple & Clean"  
**Category:** plain  
**Description:** "Minimalist text-based invitation, perfect for accessibility and fast loading"  
**Tags:** ["accessible", "minimal", "text-only"]  
**Colors:**
- Primary: Uses system colors only
- No theme-specific colors
**Fonts:**
- System font stack

#### 2. Wedding Elegance
**Name:** "Wedding Elegance"  
**Category:** card  
**Description:** "Elegant floral design perfect for weddings and formal celebrations"  
**Tags:** ["wedding", "formal", "floral", "elegant"]  
**Colors:**
- Primary: #f4c2c2 (Blush)
- Secondary: #d4af37 (Gold)
- Accent: #8b4789 (Plum)
**Fonts:**
- Heading: 'Playfair Display', serif (or similar system serif)
- Body: 'Lato', sans-serif (or system sans-serif)
**Image:** Floral border or watercolor flowers

#### 3. Birthday Celebration
**Name:** "Birthday Celebration"  
**Category:** card  
**Description:** "Fun and colorful design for birthday parties and celebrations"  
**Tags:** ["birthday", "celebration", "fun", "colorful"]  
**Colors:**
- Primary: #ff6b9d (Pink)
- Secondary: #ffd93d (Yellow)
- Accent: #6bcf7f (Green)
**Fonts:**
- Heading: 'Fredoka One', cursive (or system rounded)
- Body: 'Open Sans', sans-serif (or system sans-serif)
**Image:** Balloons, confetti, or party elements

#### 4. Corporate Professional
**Name:** "Corporate Professional"  
**Category:** card  
**Description:** "Clean and professional design for business events and meetings"  
**Tags:** ["corporate", "professional", "business", "formal"]  
**Colors:**
- Primary: #2563eb (Blue)
- Secondary: #64748b (Gray)
- Accent: #0ea5e9 (Light Blue)
**Fonts:**
- Heading: System sans-serif
- Body: System sans-serif
**Image:** Abstract geometric shapes or skyline

#### 5. Holiday Festive
**Name:** "Holiday Festive"  
**Category:** card  
**Description:** "Warm and festive design for holiday gatherings and seasonal events"  
**Tags:** ["holiday", "festive", "seasonal", "warm"]  
**Colors:**
- Primary: #dc2626 (Red)
- Secondary: #16a34a (Green)
- Accent: #d4af37 (Gold)
**Fonts:**
- Heading: System serif
- Body: System sans-serif
**Image:** Winter/holiday themed (snowflakes, ornaments, etc.)

#### 6. Garden Party
**Name:** "Garden Party"  
**Category:** card  
**Description:** "Fresh botanical design for outdoor events and garden parties"  
**Tags:** ["garden", "nature", "outdoor", "botanical"]  
**Colors:**
- Primary: #16a34a (Green)
- Secondary: #84cc16 (Lime)
- Accent: #fbbf24 (Amber)
**Fonts:**
- Heading: System serif
- Body: System sans-serif
**Image:** Botanical elements, leaves, or garden scene

#### 7. Modern Minimalist
**Name:** "Modern Minimalist"  
**Category:** card  
**Description:** "Contemporary minimal design with clean lines and bold typography"  
**Tags:** ["modern", "minimal", "contemporary", "clean"]  
**Colors:**
- Primary: #0f172a (Dark Blue)
- Secondary: #64748b (Gray)
- Accent: #06b6d4 (Cyan)
**Fonts:**
- Heading: System sans-serif
- Body: System sans-serif
**Image:** Simple geometric shapes or abstract minimal design

### File Structure

```
static/
├── images/
│   └── themes/
│       ├── plain-text-thumb.jpg
│       ├── wedding-elegance-header.jpg
│       ├── wedding-elegance-thumb.jpg
│       ├── birthday-celebration-header.jpg
│       ├── birthday-celebration-thumb.jpg
│       ├── corporate-professional-header.jpg
│       ├── corporate-professional-thumb.jpg
│       ├── holiday-festive-header.jpg
│       ├── holiday-festive-thumb.jpg
│       ├── garden-party-header.jpg
│       ├── garden-party-thumb.jpg
│       ├── modern-minimalist-header.jpg
│       └── modern-minimalist-thumb.jpg
├── css/
│   └── themes/
│       ├── plain-text.css
│       ├── wedding-elegance.css
│       ├── birthday-celebration.css
│       ├── corporate-professional.css
│       ├── holiday-festive.css
│       ├── garden-party.css
│       └── modern-minimalist.css

templates/
└── web/
    └── rsvp_themes/
        ├── plain-text.html
        ├── wedding-elegance.html
        ├── birthday-celebration.html
        ├── corporate-professional.html
        ├── holiday-festive.html
        ├── garden-party.html
        └── modern-minimalist.html
```

### HTML Template Example

**File:** `templates/web/rsvp_themes/wedding-elegance.html`

```html
<!DOCTYPE html>
<html lang="en" data-event-theme="wedding-elegance">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>RSVP: {{.Event.Title}}</title>
    <link rel="stylesheet" href="/static/css/variables.css">
    <link rel="stylesheet" href="/static/css/themes/wedding-elegance.css">
    <script src="/static/js/theme_controller.js" defer></script>
</head>
<body>
    <div class="rsvp-container">
        <div class="rsvp-card">
            <!-- Header Image -->
            <div class="rsvp-card-header">
                <img src="{{if .Event.CustomThemeImageURL}}{{.Event.CustomThemeImageURL}}{{else}}/static/images/themes/wedding-elegance-header.jpg{{end}}" 
                     alt="Wedding invitation design"
                     class="theme-header-image">
            </div>
            
            <!-- Content Area -->
            <div class="rsvp-card-content">
                <h1 class="event-title">{{.Event.Title}}</h1>
                
                <div class="event-details">
                    <p class="event-date">
                        <strong>When:</strong> {{.Event.StartTime.Format "Monday, January 2, 2006 at 3:04 PM"}}
                    </p>
                    {{if .Event.Location}}
                    <p class="event-location">
                        <strong>Where:</strong> {{.Event.Location}}
                    </p>
                    {{end}}
                    {{if .Event.Description}}
                    <p class="event-description">{{.Event.Description}}</p>
                    {{end}}
                </div>
                
                <!-- RSVP Form -->
                <form method="POST" action="/rsvp/{{.Token}}" class="rsvp-form">
                    <input type="hidden" name="csrf_token" value="{{.CSRFToken}}">
                    
                    <div class="form-group">
                        <label for="response">Will you attend?</label>
                        <select id="response" name="response" required>
                            <option value="">Please select...</option>
                            <option value="yes">Yes, I'll be there!</option>
                            <option value="no">Sorry, can't make it</option>
                            <option value="maybe">Maybe</option>
                        </select>
                    </div>
                    
                    {{if gt .Event.MaxPlusOnes 0}}
                    <div class="form-group">
                        <label for="plus_ones">Number of guests (including you)</label>
                        <input type="number" id="plus_ones" name="plus_ones" 
                               min="1" max="{{add .Event.MaxPlusOnes 1}}" value="1">
                    </div>
                    {{end}}
                    
                    {{range .Questions}}
                    <div class="form-group">
                        <label for="question_{{.ID}}">
                            {{.QuestionText}}
                            {{if .Required}}<span class="required">*</span>{{end}}
                        </label>
                        {{if eq .QuestionType "text"}}
                        <input type="text" id="question_{{.ID}}" name="answer_{{.ID}}" 
                               {{if .Required}}required{{end}}>
                        {{else if eq .QuestionType "select"}}
                        <select id="question_{{.ID}}" name="answer_{{.ID}}" 
                                {{if .Required}}required{{end}}>
                            <option value="">Please select...</option>
                            {{range .Options}}
                            <option value="{{.Value}}">{{.Label}}</option>
                            {{end}}
                        </select>
                        {{else if eq .QuestionType "boolean"}}
                        <div class="radio-group">
                            <label>
                                <input type="radio" name="answer_{{.ID}}" value="true" 
                                       {{if .Required}}required{{end}}> Yes
                            </label>
                            <label>
                                <input type="radio" name="answer_{{.ID}}" value="false"> No
                            </label>
                        </div>
                        {{end}}
                    </div>
                    {{end}}
                    
                    <button type="submit" class="btn btn-primary">Submit RSVP</button>
                </form>
            </div>
        </div>
    </div>
</body>
</html>
```

### CSS Theme Example

**File:** `static/css/themes/wedding-elegance.css`

```css
/* Wedding Elegance Theme */
[data-event-theme="wedding-elegance"] {
    /* Theme Colors */
    --theme-primary: #f4c2c2;      /* Blush */
    --theme-secondary: #d4af37;    /* Gold */
    --theme-accent: #8b4789;       /* Plum */
    
    /* Theme Fonts */
    --theme-font-heading: 'Georgia', 'Times New Roman', serif;
    --theme-font-body: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
}

/* Dark mode adjustments for Wedding theme */
[data-theme="dark"][data-event-theme="wedding-elegance"] {
    --theme-primary: #f4c2c2;      /* Keep blush visible */
    --theme-secondary: #e6c84e;    /* Brighter gold for dark mode */
    --theme-accent: #a855b5;       /* Brighter plum */
}

/* Theme-specific styles */
[data-event-theme="wedding-elegance"] .rsvp-card {
    border: 2px solid var(--theme-primary);
    border-radius: 12px;
    overflow: hidden;
}

[data-event-theme="wedding-elegance"] .rsvp-card-header {
    position: relative;
    height: 300px;
    overflow: hidden;
}

[data-event-theme="wedding-elegance"] .theme-header-image {
    width: 100%;
    height: 100%;
    object-fit: cover;
}

[data-event-theme="wedding-elegance"] .event-title {
    font-family: var(--theme-font-heading);
    color: var(--theme-primary);
    font-size: 2.5rem;
    text-align: center;
    margin: var(--spacing-6) 0 var(--spacing-4);
}

[data-event-theme="wedding-elegance"] .event-details {
    font-family: var(--theme-font-body);
    text-align: center;
    margin-bottom: var(--spacing-6);
}

[data-event-theme="wedding-elegance"] .btn-primary {
    background: var(--theme-primary);
    border-color: var(--theme-primary);
    color: var(--color-text-primary);
}

[data-event-theme="wedding-elegance"] .btn-primary:hover {
    background: var(--theme-secondary);
    border-color: var(--theme-secondary);
}

/* Responsive adjustments */
@media (max-width: 767px) {
    [data-event-theme="wedding-elegance"] .rsvp-card-header {
        height: 200px;
    }
    
    [data-event-theme="wedding-elegance"] .event-title {
        font-size: 2rem;
    }
}
```

---

## Tasks

### Design Phase
- [ ] Research Evite card designs for inspiration
- [ ] Sketch 7 theme concepts
- [ ] Get stakeholder feedback on concepts
- [ ] Finalize theme designs
- [ ] Create design specifications document

### Image Creation
- [ ] Source or create Wedding Elegance images
- [ ] Source or create Birthday Celebration images
- [ ] Source or create Corporate Professional images
- [ ] Source or create Holiday Festive images
- [ ] Source or create Garden Party images
- [ ] Source or create Modern Minimalist images
- [ ] Create Plain Text thumbnail
- [ ] Optimize all images for web
- [ ] Verify images work in light/dark modes
- [ ] Add images to `/static/images/themes/`

### HTML Templates
- [ ] Create plain-text.html template
- [ ] Create wedding-elegance.html template
- [ ] Create birthday-celebration.html template
- [ ] Create corporate-professional.html template
- [ ] Create holiday-festive.html template
- [ ] Create garden-party.html template
- [ ] Create modern-minimalist.html template
- [ ] Ensure consistent structure across templates
- [ ] Test template variables render correctly
- [ ] Validate HTML syntax

### CSS Theme Files
- [ ] Create plain-text.css
- [ ] Create wedding-elegance.css
- [ ] Create birthday-celebration.css
- [ ] Create corporate-professional.css
- [ ] Create holiday-festive.css
- [ ] Create garden-party.css
- [ ] Create modern-minimalist.css
- [ ] Define theme variables for each
- [ ] Add dark mode adjustments for each
- [ ] Test responsive behavior
- [ ] Validate CSS syntax

### Testing
- [ ] Visual test each theme in light mode
- [ ] Visual test each theme in dark mode
- [ ] Test on mobile (320px, 375px, 414px)
- [ ] Test on tablet (768px, 1024px)
- [ ] Test on desktop (1280px, 1920px)
- [ ] Test with long event titles
- [ ] Test with long descriptions
- [ ] Test with many preference questions
- [ ] Test without JavaScript
- [ ] Accessibility audit (WCAG AA)
- [ ] Performance testing (load times)

### Documentation
- [ ] Document theme creation process
- [ ] Document theme specifications
- [ ] Document image requirements
- [ ] Add theme examples to README
- [ ] Create theme design guidelines

---

## Definition of Done

- [ ] All 7 themes created (1 plain + 6 card)
- [ ] All images optimized and added to repository
- [ ] All HTML templates created and tested
- [ ] All CSS files created and tested
- [ ] Themes work in light and dark modes
- [ ] Themes mobile-responsive
- [ ] Accessibility audit passed
- [ ] Performance targets met (<2s load)
- [ ] Documentation complete
- [ ] Changes committed to git

---

## Dependencies

**Depends on:**
- Story 11.01: Theme Model Extension
- Story 10.12: Light/Dark Theme Switching (should complete first)

**Blocks:**
- Story 11.03: Theme Picker UI
- Story 11.04: Theme Preview Modal
- Story 11.05: Theme Rendering Engine

---

## References

- **Epic:** [11_EPIC_rsvp_themes.md](11_EPIC_rsvp_themes.md)
- **Analysis:** [11_ANALYSIS_rsvp_page_themes.md](11_ANALYSIS_rsvp_page_themes.md)
- **Light/Dark Theme:** [10_STORY_12_theme_switching.md](10_STORY_12_theme_switching.md)
- **Current RSVP Page:** `templates/web/rsvp_page.html`
- **CSS Variables:** `static/css/variables.css`

---

## Notes

### Image Sourcing Options
1. **Create custom:** Use design tools (Figma, Canva)
2. **Stock photos:** Unsplash, Pexels (free, attribution)
3. **AI generation:** Stable Diffusion, DALL-E (verify licensing)
4. **Placeholder:** Use colored gradients initially, replace later

### Font Considerations
- Use system font stack for v0 (privacy, performance)
- Consider Google Fonts in v1 if needed
- Ensure fallbacks for all fonts
- Test font rendering across platforms

### Dark Mode Image Strategy
- Design images to work in both modes
- Use subtle overlays if needed
- Test image visibility in dark mode
- Consider providing dark mode variants (v1+)

### Accessibility Checklist
- [ ] Alt text for all images
- [ ] Semantic HTML structure
- [ ] Sufficient color contrast
- [ ] Keyboard navigation works
- [ ] Screen reader friendly
- [ ] No text in images

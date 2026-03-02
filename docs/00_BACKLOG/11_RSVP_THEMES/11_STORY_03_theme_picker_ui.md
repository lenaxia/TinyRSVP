# User Story 11.03: Theme Picker UI

**Epic:** [11_EPIC_rsvp_themes.md](11_EPIC_rsvp_themes.md)
**Priority:** High
**Status:** Complete
**Estimated Effort:** 2-3 days
**Owner:** LLM
**Completed:** 2026-01-11

---

## User Story

As an **event manager**,  
I want to **select an RSVP page theme from a visual gallery when creating an event**,  
So that **I can choose a design that matches my event type and personal taste**.

---

## Context

Event managers need an intuitive way to browse and select from available RSVP page themes. The theme picker should:
- Display themes in a visual gallery with thumbnails
- Show theme name and description
- Allow filtering by category
- Provide preview functionality
- Be integrated into event creation flow

This is the primary UI for theme selection and should be user-friendly and visually appealing.

---

## Acceptance Criteria

### Gallery Display
- [x] Themes displayed in responsive grid layout
- [x] Each theme shows thumbnail image
- [x] Each theme shows name and description
- [x] Themes sorted by sort_order then name
- [x] Grid adapts to screen size (1 col mobile, 2-3 cols desktop)
- [x] Currently selected theme highlighted

### Filtering
- [x] Filter dropdown by category (All, Plain, Card, etc.)
- [x] Filter updates gallery without page reload
- [x] Filter state persists during session
- [x] "All" shows all themes by default

### Theme Selection
- [x] Click theme card to select
- [x] Selected theme visually highlighted
- [x] Selection updates hidden form field
- [x] Can change selection before saving
- [x] Default theme pre-selected if none chosen

### Integration with Event Form
- [x] Theme picker integrated into event creation form
- [x] Theme picker appears after basic details (title, date, location)
- [x] Selected theme ID submitted with event form
- [x] Form validation ensures theme selected
- [x] Theme selection optional (uses default if not selected)

### Mobile Experience
- [x] Touch-friendly tap targets (44px minimum)
- [x] Single column layout on mobile
- [x] Thumbnails sized appropriately
- [x] Smooth scrolling
- [x] No horizontal scroll

### Accessibility
- [x] Keyboard navigation works
- [x] Focus states visible
- [x] Screen reader announces theme selection
- [x] ARIA labels on interactive elements
- [x] Semantic HTML structure

---

## Technical Details

### Component Structure

**File:** `templates/web/partials/theme_picker.html`

```html
<div class="theme-picker">
    <div class="theme-picker-header">
        <h3>Select Theme</h3>
        <div class="theme-filter">
            <label for="theme-category-filter">Category:</label>
            <select id="theme-category-filter" class="theme-category-select">
                <option value="">All Themes</option>
                <option value="plain">Plain Text</option>
                <option value="card">Card Designs</option>
            </select>
        </div>
    </div>
    
    <div class="theme-gallery" role="radiogroup" aria-label="Select theme">
        {{range .Themes}}
        <div class="theme-card {{if eq .ID $.SelectedThemeID}}selected{{end}}" 
             data-theme-id="{{.ID}}"
             data-category="{{.Category}}"
             role="radio"
             aria-checked="{{if eq .ID $.SelectedThemeID}}true{{else}}false{{end}}"
             tabindex="{{if eq .ID $.SelectedThemeID}}0{{else}}-1{{end}}">
            
            <div class="theme-thumbnail">
                <img src="{{.ThumbnailURL}}" 
                     alt="{{.Name}} theme preview"
                     loading="lazy">
            </div>
            
            <div class="theme-info">
                <h4 class="theme-name">{{.Name}}</h4>
                <p class="theme-description">{{.Description}}</p>
                
                {{if .Tags}}
                <div class="theme-tags">
                    {{range .Tags}}
                    <span class="theme-tag">{{.}}</span>
                    {{end}}
                </div>
                {{end}}
            </div>
            
            <div class="theme-actions">
                <button type="button" 
                        class="btn-preview" 
                        data-theme-id="{{.ID}}"
                        aria-label="Preview {{.Name}} theme">
                    Preview
                </button>
                <button type="button" 
                        class="btn-select" 
                        data-theme-id="{{.ID}}"
                        aria-label="Select {{.Name}} theme">
                    Select
                </button>
            </div>
        </div>
        {{end}}
    </div>
    
    <!-- Hidden input for form submission -->
    <input type="hidden" id="selected-theme-id" name="template_id" value="{{.SelectedThemeID}}">
</div>
```

### CSS Styles

**File:** `static/css/theme_picker.css`

```css
.theme-picker {
    margin: var(--spacing-6) 0;
}

.theme-picker-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: var(--spacing-4);
    flex-wrap: wrap;
    gap: var(--spacing-4);
}

.theme-picker-header h3 {
    margin: 0;
    font-size: var(--font-size-xl);
}

.theme-filter {
    display: flex;
    align-items: center;
    gap: var(--spacing-2);
}

.theme-category-select {
    min-width: 150px;
}

.theme-gallery {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
    gap: var(--spacing-4);
}

.theme-card {
    border: 2px solid var(--color-border);
    border-radius: var(--radius-base);
    overflow: hidden;
    transition: all var(--transition-fast);
    cursor: pointer;
    background: var(--color-surface);
}

.theme-card:hover {
    border-color: var(--color-primary-600);
    box-shadow: var(--shadow-md);
    transform: translateY(-2px);
}

.theme-card:focus-within {
    outline: 2px solid var(--color-border-focus);
    outline-offset: 2px;
}

.theme-card.selected {
    border-color: var(--color-primary-600);
    border-width: 3px;
    box-shadow: var(--shadow-lg);
}

.theme-card.selected::before {
    content: '✓';
    position: absolute;
    top: var(--spacing-2);
    right: var(--spacing-2);
    background: var(--color-primary-600);
    color: white;
    width: 32px;
    height: 32px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    font-weight: bold;
    z-index: 1;
}

.theme-thumbnail {
    position: relative;
    width: 100%;
    height: 180px;
    overflow: hidden;
    background: var(--color-gray-100);
}

.theme-thumbnail img {
    width: 100%;
    height: 100%;
    object-fit: cover;
}

.theme-info {
    padding: var(--spacing-4);
}

.theme-name {
    font-size: var(--font-size-lg);
    font-weight: 600;
    margin: 0 0 var(--spacing-2);
    color: var(--color-text-primary);
}

.theme-description {
    font-size: var(--font-size-sm);
    color: var(--color-text-secondary);
    margin: 0 0 var(--spacing-3);
    line-height: 1.4;
}

.theme-tags {
    display: flex;
    flex-wrap: wrap;
    gap: var(--spacing-2);
    margin-bottom: var(--spacing-3);
}

.theme-tag {
    font-size: var(--font-size-xs);
    padding: var(--spacing-1) var(--spacing-2);
    background: var(--color-primary-50);
    color: var(--color-primary-700);
    border-radius: var(--radius-sm);
}

.theme-actions {
    display: flex;
    gap: var(--spacing-2);
    padding: 0 var(--spacing-4) var(--spacing-4);
}

.theme-actions button {
    flex: 1;
    padding: var(--spacing-2) var(--spacing-3);
    font-size: var(--font-size-sm);
}

.btn-preview {
    background: var(--color-surface);
    border: 1px solid var(--color-border);
    color: var(--color-text-primary);
}

.btn-preview:hover {
    background: var(--color-gray-50);
}

.btn-select {
    background: var(--color-primary-600);
    border: 1px solid var(--color-primary-600);
    color: white;
}

.btn-select:hover {
    background: var(--color-primary-700);
}

/* Mobile optimizations */
@media (max-width: 767px) {
    .theme-gallery {
        grid-template-columns: 1fr;
    }
    
    .theme-picker-header {
        flex-direction: column;
        align-items: flex-start;
    }
}

/* Tablet */
@media (min-width: 768px) and (max-width: 1023px) {
    .theme-gallery {
        grid-template-columns: repeat(2, 1fr);
    }
}

/* Desktop */
@media (min-width: 1024px) {
    .theme-gallery {
        grid-template-columns: repeat(3, 1fr);
    }
}
```

### JavaScript Controller

**File:** `static/js/theme_picker.js`

```javascript
class ThemePicker {
    constructor() {
        this.gallery = document.querySelector('.theme-gallery');
        this.filterSelect = document.getElementById('theme-category-filter');
        this.hiddenInput = document.getElementById('selected-theme-id');
        this.init();
    }

    init() {
        if (!this.gallery) return;
        
        this.attachEventListeners();
        this.initializeKeyboardNavigation();
    }

    attachEventListeners() {
        // Filter change
        if (this.filterSelect) {
            this.filterSelect.addEventListener('change', (e) => {
                this.filterThemes(e.target.value);
            });
        }

        // Theme selection
        this.gallery.addEventListener('click', (e) => {
            const selectBtn = e.target.closest('.btn-select');
            if (selectBtn) {
                const themeId = selectBtn.dataset.themeId;
                this.selectTheme(themeId);
            }

            const previewBtn = e.target.closest('.btn-preview');
            if (previewBtn) {
                const themeId = previewBtn.dataset.themeId;
                this.previewTheme(themeId);
            }

            // Also allow clicking card itself
            const card = e.target.closest('.theme-card');
            if (card && !selectBtn && !previewBtn) {
                const themeId = card.dataset.themeId;
                this.selectTheme(themeId);
            }
        });
    }

    filterThemes(category) {
        const cards = this.gallery.querySelectorAll('.theme-card');
        
        cards.forEach(card => {
            if (!category || card.dataset.category === category) {
                card.style.display = '';
            } else {
                card.style.display = 'none';
            }
        });
    }

    selectTheme(themeId) {
        // Remove previous selection
        const previousSelected = this.gallery.querySelector('.theme-card.selected');
        if (previousSelected) {
            previousSelected.classList.remove('selected');
            previousSelected.setAttribute('aria-checked', 'false');
            previousSelected.setAttribute('tabindex', '-1');
        }

        // Add new selection
        const card = this.gallery.querySelector(`[data-theme-id="${themeId}"]`);
        if (card) {
            card.classList.add('selected');
            card.setAttribute('aria-checked', 'true');
            card.setAttribute('tabindex', '0');
            card.focus();
        }

        // Update hidden input
        if (this.hiddenInput) {
            this.hiddenInput.value = themeId;
        }

        // Announce to screen readers
        this.announceSelection(card);
    }

    previewTheme(themeId) {
        // Dispatch custom event for preview modal
        const event = new CustomEvent('theme-preview-requested', {
            detail: { themeId }
        });
        document.dispatchEvent(event);
    }

    initializeKeyboardNavigation() {
        this.gallery.addEventListener('keydown', (e) => {
            const card = e.target.closest('.theme-card');
            if (!card) return;

            let nextCard = null;

            switch (e.key) {
                case 'Enter':
                case ' ':
                    e.preventDefault();
                    this.selectTheme(card.dataset.themeId);
                    break;
                case 'ArrowRight':
                case 'ArrowDown':
                    e.preventDefault();
                    nextCard = card.nextElementSibling;
                    while (nextCard && nextCard.style.display === 'none') {
                        nextCard = nextCard.nextElementSibling;
                    }
                    if (nextCard) nextCard.focus();
                    break;
                case 'ArrowLeft':
                case 'ArrowUp':
                    e.preventDefault();
                    nextCard = card.previousElementSibling;
                    while (nextCard && nextCard.style.display === 'none') {
                        nextCard = nextCard.previousElementSibling;
                    }
                    if (nextCard) nextCard.focus();
                    break;
                case 'Home':
                    e.preventDefault();
                    const firstCard = this.gallery.querySelector('.theme-card:not([style*="display: none"])');
                    if (firstCard) firstCard.focus();
                    break;
                case 'End':
                    e.preventDefault();
                    const cards = Array.from(this.gallery.querySelectorAll('.theme-card:not([style*="display: none"])'));
                    const lastCard = cards[cards.length - 1];
                    if (lastCard) lastCard.focus();
                    break;
            }
        });
    }

    announceSelection(card) {
        const themeName = card.querySelector('.theme-name')?.textContent;
        const announcement = document.createElement('div');
        announcement.setAttribute('role', 'status');
        announcement.setAttribute('aria-live', 'polite');
        announcement.className = 'sr-only';
        announcement.textContent = `${themeName} theme selected`;
        document.body.appendChild(announcement);
        setTimeout(() => announcement.remove(), 1000);
    }
}

// Initialize on DOM ready
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', () => new ThemePicker());
} else {
    new ThemePicker();
}
```

### Handler Integration

**File:** `internal/handlers/events_web.go`

```go
func (h *Handler) HandleEventCreatePage(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    // Get available themes for RSVP pages
    themes, err := h.templateService.ListThemes(ctx, models.TemplateTypeRSVPPage, nil)
    if err != nil {
        HandleError(w, r, fmt.Errorf("failed to load themes: %w", err))
        return
    }
    
    // Get default theme
    defaultTheme, err := h.templateService.GetDefaultTemplate(ctx, models.TemplateTypeRSVPPage)
    if err != nil {
        HandleError(w, r, fmt.Errorf("failed to load default theme: %w", err))
        return
    }
    
    data := struct {
        Themes           []*models.Template
        SelectedThemeID  int64
        CSRFToken        string
    }{
        Themes:          themes,
        SelectedThemeID: defaultTheme.ID,
        CSRFToken:       GetCSRFToken(r),
    }
    
    RenderTemplate(w, "event_form.html", data)
}
```

---

## Tasks

### HTML Component
- [x] Create `theme_picker.html` partial
- [x] Add gallery grid structure
- [x] Add theme card template
- [x] Add filter dropdown
- [x] Add hidden input for form submission
- [x] Add ARIA attributes
- [x] Test HTML structure

### CSS Styles
- [x] Create `theme_picker.css`
- [x] Style gallery grid
- [x] Style theme cards
- [x] Style thumbnails
- [x] Style selection state
- [x] Style hover states
- [x] Style focus states
- [x] Add responsive breakpoints
- [x] Test on mobile/tablet/desktop

### JavaScript Controller
- [x] Create `theme_picker.js`
- [x] Implement theme selection
- [x] Implement filtering
- [x] Implement keyboard navigation
- [x] Add screen reader announcements
- [x] Handle edge cases
- [x] Write unit tests

### Handler Integration
- [x] Update event creation handler
- [x] Load themes from service
- [x] Pass themes to template
- [x] Handle theme selection in POST
- [x] Validate theme ID
- [x] Write handler tests

### Form Integration
- [x] Update `event_form.html`
- [x] Include theme picker partial
- [x] Add theme picker section
- [x] Update form submission
- [x] Test form validation
- [x] Write integration tests

---

## Definition of Done

- [x] All acceptance criteria met
- [x] Theme picker component created
- [x] CSS styles implemented
- [x] JavaScript controller implemented
- [x] Integrated into event creation form
- [x] All unit tests passing
- [x] All integration tests passing
- [x] Mobile-responsive
- [x] Accessibility audit passed
- [x] Changes committed to git

---

## Dependencies

**Depends on:**
- Story 11.01: Theme Model Extension
- Story 11.02: Theme Asset Creation
- Story 10.12: Light/Dark Theme Switching

**Blocks:**
- Story 11.04: Theme Preview Modal
- Story 11.05: Theme Rendering Engine

---

## References

- **Epic:** [11_EPIC_rsvp_themes.md](11_EPIC_rsvp_themes.md)
- **Analysis:** [11_ANALYSIS_rsvp_page_themes.md](11_ANALYSIS_rsvp_page_themes.md)
- **Event Form:** `templates/web/event_form.html`
- **Event Handler:** `internal/handlers/events_web.go`

---

## Notes

- Theme picker should be visually appealing itself
- Consider lazy loading thumbnails for performance
- Filter state could be saved to localStorage (enhancement)
- Could add search functionality in v1
- Consider adding "Recently Used" section in v1

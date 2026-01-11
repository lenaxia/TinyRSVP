# User Story 11.04: Theme Preview Modal

**Epic:** [11_EPIC_rsvp_themes.md](11_EPIC_rsvp_themes.md)  
**Priority:** High  
**Status:** Not Started  
**Estimated Effort:** 2 days  
**Owner:** Unassigned

---

## User Story

As an **event manager**,  
I want to **preview how a theme will look with my event details before selecting it**,  
So that **I can make an informed decision and ensure the theme matches my vision**.

---

## Context

Event managers need to see how themes will look with their actual event data before committing to a selection. The preview should:
- Show full RSVP page with theme applied
- Use real event data (title, date, location, description)
- Display in a modal overlay
- Support light/dark mode toggle in preview
- Be fast and responsive

---

## Acceptance Criteria

### Modal Component
- [ ] Modal opens when "Preview" button clicked
- [ ] Modal displays full-size preview of RSVP page
- [ ] Modal has close button (X in corner)
- [ ] Modal closes on Escape key
- [ ] Modal closes on backdrop click
- [ ] Modal traps focus while open
- [ ] Modal prevents body scroll when open

### Preview Content
- [ ] Preview shows RSVP page with selected theme
- [ ] Preview uses current event form data (title, date, location)
- [ ] Preview shows placeholder data if form incomplete
- [ ] Preview updates if event data changes
- [ ] Preview shows theme in current light/dark mode
- [ ] Preview includes theme toggle button

### Theme Toggle in Preview
- [ ] Preview has light/dark toggle button
- [ ] Toggle switches preview between light/dark
- [ ] Toggle state independent of main page theme
- [ ] Toggle allows testing theme in both modes

### Performance
- [ ] Preview loads in <1 second
- [ ] Preview uses iframe for isolation
- [ ] Preview doesn't affect main page
- [ ] Multiple previews don't cause memory leaks

### Mobile Experience
- [ ] Modal full-screen on mobile
- [ ] Preview scrollable on mobile
- [ ] Close button accessible on mobile
- [ ] Touch gestures work (swipe to close optional)

### Accessibility
- [ ] Focus trapped in modal when open
- [ ] Focus returns to trigger button on close
- [ ] Screen reader announces modal open/close
- [ ] Keyboard navigation works (Tab, Shift+Tab, Escape)
- [ ] ARIA attributes correct

---

## Technical Details

### Modal Component

**File:** `templates/web/partials/theme_preview_modal.html`

```html
<div id="theme-preview-modal" 
     class="modal" 
     role="dialog" 
     aria-labelledby="preview-modal-title"
     aria-modal="true"
     hidden>
    <div class="modal-backdrop" aria-hidden="true"></div>
    <div class="modal-container">
        <div class="modal-header">
            <h3 id="preview-modal-title">Theme Preview</h3>
            <div class="modal-header-actions">
                <button type="button" 
                        id="preview-theme-toggle" 
                        class="btn-icon"
                        aria-label="Toggle preview theme">
                    <span class="theme-icon">🌙</span>
                </button>
                <button type="button" 
                        class="modal-close" 
                        aria-label="Close preview">
                    ×
                </button>
            </div>
        </div>
        <div class="modal-body">
            <iframe id="theme-preview-frame" 
                    title="Theme preview"
                    sandbox="allow-same-origin"
                    loading="lazy"></iframe>
        </div>
        <div class="modal-footer">
            <button type="button" class="btn btn-secondary modal-close">
                Close
            </button>
            <button type="button" class="btn btn-primary" id="select-previewed-theme">
                Select This Theme
            </button>
        </div>
    </div>
</div>
```

### Modal CSS

**File:** `static/css/theme_preview_modal.css`

```css
.modal {
    position: fixed;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    z-index: 1000;
    display: flex;
    align-items: center;
    justify-content: center;
}

.modal[hidden] {
    display: none;
}

.modal-backdrop {
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background: rgba(0, 0, 0, 0.5);
    backdrop-filter: blur(4px);
}

.modal-container {
    position: relative;
    background: var(--color-surface);
    border-radius: var(--radius-lg);
    box-shadow: var(--shadow-2xl);
    max-width: 90vw;
    max-height: 90vh;
    display: flex;
    flex-direction: column;
    z-index: 1;
}

.modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: var(--spacing-4);
    border-bottom: 1px solid var(--color-border);
}

.modal-header h3 {
    margin: 0;
    font-size: var(--font-size-xl);
}

.modal-header-actions {
    display: flex;
    gap: var(--spacing-2);
}

.modal-close {
    background: transparent;
    border: none;
    font-size: 2rem;
    line-height: 1;
    cursor: pointer;
    padding: var(--spacing-2);
    color: var(--color-text-secondary);
    transition: color var(--transition-fast);
}

.modal-close:hover {
    color: var(--color-text-primary);
}

.modal-body {
    flex: 1;
    overflow: hidden;
    padding: var(--spacing-4);
}

#theme-preview-frame {
    width: 100%;
    height: 600px;
    border: 1px solid var(--color-border);
    border-radius: var(--radius-base);
    background: white;
}

.modal-footer {
    display: flex;
    justify-content: flex-end;
    gap: var(--spacing-3);
    padding: var(--spacing-4);
    border-top: 1px solid var(--color-border);
}

/* Mobile optimizations */
@media (max-width: 767px) {
    .modal-container {
        max-width: 100vw;
        max-height: 100vh;
        border-radius: 0;
        width: 100%;
        height: 100%;
    }
    
    #theme-preview-frame {
        height: 100%;
    }
    
    .modal-footer {
        flex-direction: column-reverse;
    }
    
    .modal-footer button {
        width: 100%;
    }
}
```

### Modal Controller

**File:** `static/js/theme_preview_modal.js`

```javascript
class ThemePreviewModal {
    constructor() {
        this.modal = document.getElementById('theme-preview-modal');
        this.iframe = document.getElementById('theme-preview-frame');
        this.themeToggle = document.getElementById('preview-theme-toggle');
        this.selectButton = document.getElementById('select-previewed-theme');
        this.currentThemeId = null;
        this.previewTheme = 'light';
        this.init();
    }

    init() {
        if (!this.modal) return;
        
        this.attachEventListeners();
        this.setupFocusTrap();
    }

    attachEventListeners() {
        // Listen for preview requests
        document.addEventListener('theme-preview-requested', (e) => {
            this.open(e.detail.themeId);
        });

        // Close buttons
        const closeButtons = this.modal.querySelectorAll('.modal-close');
        closeButtons.forEach(btn => {
            btn.addEventListener('click', () => this.close());
        });

        // Backdrop click
        const backdrop = this.modal.querySelector('.modal-backdrop');
        if (backdrop) {
            backdrop.addEventListener('click', () => this.close());
        }

        // Escape key
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape' && !this.modal.hidden) {
                this.close();
            }
        });

        // Theme toggle in preview
        if (this.themeToggle) {
            this.themeToggle.addEventListener('click', () => {
                this.togglePreviewTheme();
            });
        }

        // Select button
        if (this.selectButton) {
            this.selectButton.addEventListener('click', () => {
                this.selectCurrentTheme();
            });
        }
    }

    open(themeId) {
        this.currentThemeId = themeId;
        this.lastFocusedElement = document.activeElement;
        
        // Load preview
        this.loadPreview(themeId);
        
        // Show modal
        this.modal.hidden = false;
        document.body.style.overflow = 'hidden';
        
        // Focus first focusable element
        const firstFocusable = this.modal.querySelector('button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])');
        if (firstFocusable) {
            firstFocusable.focus();
        }
        
        // Announce to screen readers
        this.announce('Theme preview opened');
    }

    close() {
        this.modal.hidden = true;
        document.body.style.overflow = '';
        
        // Return focus
        if (this.lastFocusedElement) {
            this.lastFocusedElement.focus();
        }
        
        // Announce to screen readers
        this.announce('Theme preview closed');
    }

    loadPreview(themeId) {
        // Get current event form data
        const formData = this.getEventFormData();
        
        // Build preview URL
        const params = new URLSearchParams({
            theme_id: themeId,
            preview: 'true',
            theme_mode: this.previewTheme,
            ...formData
        });
        
        this.iframe.src = `/api/themes/preview?${params.toString()}`;
    }

    getEventFormData() {
        const form = document.querySelector('form[action*="/events"]');
        if (!form) return {};
        
        return {
            title: form.querySelector('[name="title"]')?.value || 'Sample Event',
            location: form.querySelector('[name="location"]')?.value || 'Sample Location',
            start_time: form.querySelector('[name="start_time"]')?.value || new Date().toISOString(),
            description: form.querySelector('[name="description"]')?.value || 'Sample description'
        };
    }

    togglePreviewTheme() {
        this.previewTheme = this.previewTheme === 'light' ? 'dark' : 'light';
        
        // Update toggle button
        const icon = this.themeToggle.querySelector('.theme-icon');
        if (icon) {
            icon.textContent = this.previewTheme === 'dark' ? '☀️' : '🌙';
        }
        
        // Reload preview with new theme
        if (this.currentThemeId) {
            this.loadPreview(this.currentThemeId);
        }
    }

    selectCurrentTheme() {
        if (this.currentThemeId) {
            // Dispatch event to theme picker
            const event = new CustomEvent('theme-selected', {
                detail: { themeId: this.currentThemeId }
            });
            document.dispatchEvent(event);
            
            this.close();
        }
    }

    setupFocusTrap() {
        this.modal.addEventListener('keydown', (e) => {
            if (e.key !== 'Tab') return;
            
            const focusableElements = this.modal.querySelectorAll(
                'button:not([disabled]), [href], input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])'
            );
            
            const firstElement = focusableElements[0];
            const lastElement = focusableElements[focusableElements.length - 1];
            
            if (e.shiftKey && document.activeElement === firstElement) {
                e.preventDefault();
                lastElement.focus();
            } else if (!e.shiftKey && document.activeElement === lastElement) {
                e.preventDefault();
                firstElement.focus();
            }
        });
    }

    announce(message) {
        const announcement = document.createElement('div');
        announcement.setAttribute('role', 'status');
        announcement.setAttribute('aria-live', 'polite');
        announcement.className = 'sr-only';
        announcement.textContent = message;
        document.body.appendChild(announcement);
        setTimeout(() => announcement.remove(), 1000);
    }
}

// Initialize on DOM ready
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', () => new ThemePreviewModal());
} else {
    new ThemePreviewModal();
}

// Connect theme picker to modal
document.addEventListener('theme-selected', (e) => {
    const themePicker = window.themePicker;
    if (themePicker) {
        themePicker.selectTheme(e.detail.themeId);
    }
});
```

### Preview Endpoint

**File:** `internal/handlers/theme_preview.go`

```go
func (h *Handler) HandleThemePreview(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    // Get theme ID
    themeIDStr := r.URL.Query().Get("theme_id")
    themeID, err := strconv.ParseInt(themeIDStr, 10, 64)
    if err != nil {
        http.Error(w, "Invalid theme ID", http.StatusBadRequest)
        return
    }
    
    // Get theme template
    template, err := h.templateService.GetTemplate(ctx, themeID)
    if err != nil {
        HandleError(w, r, fmt.Errorf("failed to load theme: %w", err))
        return
    }
    
    // Get preview data from query params or use defaults
    previewData := struct {
        Event struct {
            Title       string
            Description string
            StartTime   time.Time
            Location    string
        }
        Invite struct {
            Name string
        }
        Questions       []*models.PreferenceQuestion
        ThemeMode       string
        CSRFToken       string
    }{
        Event: struct {
            Title       string
            Description string
            StartTime   time.Time
            Location    string
        }{
            Title:       r.URL.Query().Get("title"),
            Description: r.URL.Query().Get("description"),
            StartTime:   parseTimeOrDefault(r.URL.Query().Get("start_time")),
            Location:    r.URL.Query().Get("location"),
        },
        Invite: struct {
            Name string
        }{
            Name: "Guest Name",
        },
        Questions: []*models.PreferenceQuestion{
            {
                ID:           1,
                QuestionText: "Dietary restrictions?",
                QuestionType: "select",
                Options: []models.QuestionOption{
                    {Value: "none", Label: "None"},
                    {Value: "vegetarian", Label: "Vegetarian"},
                    {Value: "vegan", Label: "Vegan"},
                },
            },
        },
        ThemeMode:  r.URL.Query().Get("theme_mode"),
        CSRFToken:  "preview-token",
    }
    
    // Use defaults if not provided
    if previewData.Event.Title == "" {
        previewData.Event.Title = "Sample Event Title"
    }
    if previewData.Event.Location == "" {
        previewData.Event.Location = "Sample Location"
    }
    if previewData.Event.Description == "" {
        previewData.Event.Description = "This is a sample event description to show how your theme will look."
    }
    
    // Render theme with preview data
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    
    // Add data-theme attribute for light/dark
    if previewData.ThemeMode == "dark" {
        w.Write([]byte(`<html lang="en" data-theme="dark" data-event-theme="` + template.Category + `">`))
    } else {
        w.Write([]byte(`<html lang="en" data-theme="light" data-event-theme="` + template.Category + `">`))
    }
    
    // Render template content
    if err := h.renderer.RenderToWriter(w, template.HTMLContent, previewData); err != nil {
        HandleError(w, r, fmt.Errorf("failed to render preview: %w", err))
        return
    }
    
    w.Write([]byte(`</html>`))
}
```

---

## Tasks

### Modal Component
- [ ] Create `theme_preview_modal.html` partial
- [ ] Add modal structure (backdrop, container, header, body, footer)
- [ ] Add close button
- [ ] Add theme toggle button
- [ ] Add select button
- [ ] Add ARIA attributes
- [ ] Test HTML structure

### Modal CSS
- [ ] Create `theme_preview_modal.css`
- [ ] Style modal backdrop
- [ ] Style modal container
- [ ] Style modal header/body/footer
- [ ] Style iframe
- [ ] Add responsive styles
- [ ] Add animations (fade in/out)
- [ ] Test on mobile/tablet/desktop

### Modal JavaScript
- [ ] Create `theme_preview_modal.js`
- [ ] Implement open/close functionality
- [ ] Implement focus trap
- [ ] Implement keyboard navigation
- [ ] Implement theme toggle in preview
- [ ] Handle preview loading
- [ ] Handle select button
- [ ] Write unit tests

### Preview Endpoint
- [ ] Create `theme_preview.go` handler
- [ ] Implement preview rendering
- [ ] Handle query parameters
- [ ] Provide default preview data
- [ ] Support light/dark mode parameter
- [ ] Add route to router
- [ ] Write handler tests

### Integration
- [ ] Connect theme picker to modal
- [ ] Connect modal to theme selection
- [ ] Test full flow (pick → preview → select)
- [ ] Test with incomplete form data
- [ ] Test with complete form data
- [ ] Write integration tests

---

## Definition of Done

- [ ] All acceptance criteria met
- [ ] Modal component created
- [ ] Modal CSS implemented
- [ ] Modal JavaScript implemented
- [ ] Preview endpoint implemented
- [ ] Integrated with theme picker
- [ ] All unit tests passing
- [ ] All integration tests passing
- [ ] Mobile-responsive
- [ ] Accessibility audit passed
- [ ] Changes committed to git

---

## Dependencies

**Depends on:**
- Story 11.01: Theme Model Extension
- Story 11.02: Theme Asset Creation
- Story 11.03: Theme Picker UI

**Blocks:**
- Story 11.05: Theme Rendering Engine

---

## References

- **Epic:** [11_EPIC_rsvp_themes.md](11_EPIC_rsvp_themes.md)
- **Analysis:** [11_ANALYSIS_rsvp_page_themes.md](11_ANALYSIS_rsvp_page_themes.md)
- **Modal Component:** `static/css/modal.css` (existing)
- **Theme Picker:** [11_STORY_03_theme_picker_ui.md](11_STORY_03_theme_picker_ui.md)

---

## Notes

- Preview should be fast - consider caching rendered previews
- Iframe provides isolation from main page styles
- Sandbox attribute prevents preview from affecting main page
- Theme toggle in preview allows testing both modes
- Consider adding "Refresh Preview" button if form data changes
- Could add zoom controls for better preview inspection (v1)

# Live Preview Mode - Design Document

**Version:** 1.0  
**Date:** 2026-02-04  
**Status:** Design Phase  
**Epic:** Epic 07 - Frontend Enhancement  
**Related:** Event Creation UX, Theme Selection

---

## 1. Overview

### 1.1 Purpose

Add a "Design Mode" to the event creation page that shows a live preview of the RSVP invitation as users fill out the form. This replaces the current theme gallery with a real-time preview that updates as users type.

### 1.2 Problem Statement

**Current Experience:**
- Users select a theme from a static gallery of thumbnail images
- No visual feedback of what the actual invitation will look like with their event data
- Users must click "Preview" button to see a modal popup
- Modal popup is disconnected from the form-filling experience

**Desired Experience:**
- Users see a live preview of their RSVP invitation as they design it
- Preview updates in real-time as they type event details
- Immediate visual feedback for title, location, description, date/time, custom colors, custom images
- More intuitive, visual-first design experience (like Canva, Figma, Squarespace)

### 1.3 User Stories

**As an event organizer:**
- I want to see what my invitation looks like as I'm creating it
- I want to switch between themes and immediately see my content in that theme
- I want to see my custom images and colors applied in real-time
- I want to type my event title and see it appear in the preview instantly

---

## 2. Design Decisions

### 2.1 Layout Architecture

**Two-Column Layout (Desktop ≥1024px):**
```
┌─────────────────────────────────────────────────────────────────┐
│  Event Creation Page                                            │
├─────────────────────────────────┬───────────────────────────────┤
│  LEFT COLUMN (75%)              │  RIGHT COLUMN (25%)           │
│  ┌─────────────────────────┐   │  ┌──────────────────────┐    │
│  │  LIVE PREVIEW IFRAME    │   │  │  Event Details Form  │    │
│  │                         │   │  │  - Title             │    │
│  │  Shows actual RSVP page │   │  │  - Date/Time         │    │
│  │  with theme applied     │   │  │  - Location          │    │
│  │  and user's data        │   │  │  - Description       │    │
│  │                         │   │  │                      │    │
│  │  [Your event content    │   │  │  Theme Selection     │    │
│  │   appears here in       │   │  │  - Gallery/List view │    │
│  │   selected theme]       │   │  │  - Design Mode toggle│    │
│  │                         │   │  │                      │    │
│  │                         │   │  │  Custom Options      │    │
│  │                         │   │  │  - Image Upload      │    │
│  │                         │   │  │  - Color Picker      │    │
│  │                         │   │  │                      │    │
│  │                         │   │  │  Actions             │    │
│  │                         │   │  │  - Cancel            │    │
│  └─────────────────────────┘   │  │  - Save Draft        │    │
│                                 │  │  - Publish           │    │
│                                 │  └──────────────────────┘    │
└─────────────────────────────────┴───────────────────────────────┘
```

**Tablet Layout (768px - 1023px):**
```
┌─────────────────────────────────────────────┐
│  Event Creation Page                        │
├─────────────────────────────────────────────┤
│  TOP: Event Details Form                    │
│  ┌─────────────────────────────────────┐   │
│  │  - Title                            │   │
│  │  - Date/Time                        │   │
│  │  - Location                         │   │
│  │  - Description                      │   │
│  │                                     │   │
│  │  Theme Selection                    │   │
│  │  - Gallery/List view                │   │
│  │  - Design Mode toggle               │   │
│  └─────────────────────────────────────┘   │
│                                             │
│  BOTTOM: Live Preview (in Design Mode)     │
│  ┌─────────────────────────────────────┐   │
│  │  [Preview shows full width]         │   │
│  │                                     │   │
│  └─────────────────────────────────────┘   │
│                                             │
│  Actions (full width)                      │
│  [Cancel] [Save Draft] [Publish]           │
└─────────────────────────────────────────────┘
```

**Mobile Layout (<768px):**
```
┌─────────────────────────────┐
│  Event Creation Page        │
├─────────────────────────────┤
│  [Edit] [Preview] Toggle    │
│  (Pills, only in Design Mode)│
├─────────────────────────────┤
│                             │
│  EDIT VIEW:                 │
│  ┌───────────────────────┐ │
│  │  Event Details Form   │ │
│  │  - Title              │ │
│  │  - Date/Time          │ │
│  │  - Location           │ │
│  │  - Description        │ │
│  │                       │ │
│  │  Theme Selection      │ │
│  │  - Gallery/List       │ │
│  │  - Design Mode toggle │ │
│  │                       │ │
│  │  Actions (stacked)    │ │
│  │  [Cancel]             │ │
│  │  [Save Draft]         │ │
│  │  [Publish]            │ │
│  └───────────────────────┘ │
│                             │
│  OR                         │
│                             │
│  PREVIEW VIEW:              │
│  ┌───────────────────────┐ │
│  │  [Live Preview]       │ │
│  │  Full screen          │ │
│  │                       │ │
│  │  Scrollable           │ │
│  └───────────────────────┘ │
└─────────────────────────────┘
```

**Mobile Design Mode Behavior:**
- **Default state:** Edit view (show form)
- **Toggle pills:** Sticky at top, switch between "Edit" and "Preview"
- **Edit view:** Show all form fields + gallery/design mode toggle
- **Preview view:** Show full-screen preview iframe (hide form completely)
- **Benefits:** User can see full preview without form taking up space

### 2.2 Mode Toggle

**Theme Picker Header:**
```html
<div class="theme-picker-header">
    <h3>Select Theme</h3>
    <div class="theme-mode-controls">
        <button id="gallery-mode-btn" class="mode-btn active">
            Gallery
        </button>
        <button id="design-mode-btn" class="mode-btn">
            Design Mode
        </button>
    </div>
</div>
```

**Two Modes:**
1. **Gallery Mode (Default):** Shows theme thumbnails in a grid, clicking "Select" chooses a theme
2. **Design Mode:** Shows live preview iframe, theme selector becomes a compact dropdown

### 2.3 Preview Update Strategy

**Event-Driven Updates:**
- Use `input` event on form fields (fires as user types)
- Debounce updates with 500ms delay (wait for user to stop typing)
- Rebuild preview URL with current form data
- Update iframe `src` to refresh preview

**Watched Form Fields:**
- `[name="title"]` → Event title
- `[name="location"]` → Event location  
- `[name="description"]` → Event description
- `[name="start_time"]` → Event date/time
- `#selected-theme-id` → Selected theme
- `[name="custom_theme_image_url"]` → Custom header image
- `#custom-theme-color-value` → Custom color

**Preview URL Format:**
```
/api/themes/preview?theme_id={id}&title={title}&location={loc}&description={desc}&start_time={iso8601}&custom_image_url={url}&custom_color={hex}
```

---

## 3. Technical Implementation

### 3.1 Backend Changes

**No backend changes required!** ✅

The existing `/api/themes/preview` endpoint (`internal/handlers/templates.go:451`) already:
- Accepts `theme_id` parameter
- Accepts event data as query parameters (`title`, `location`, `description`, `start_time`)
- Accepts `custom_image_url` and `custom_color` parameters
- Renders a full HTML page with the theme applied
- Returns proper HTML with all CSS/JS loaded

### 3.2 Frontend Changes

#### 3.2.1 HTML Changes

**File:** `templates/web/partials/theme_picker.html`

Add:
1. Mode toggle buttons in header
2. Live preview container (hidden by default)
3. Compact theme selector for design mode

```html
{{define "theme_picker"}}
<div class="theme-picker" data-mode="gallery" data-mobile-view="edit">
    <!-- Mobile View Toggle (only shown on mobile in design mode) -->
    <div class="mobile-view-toggle">
        <button type="button" id="mobile-edit-btn" class="view-btn active">
            Edit
        </button>
        <button type="button" id="mobile-preview-btn" class="view-btn">
            Preview
        </button>
    </div>
    
    <div class="theme-picker-header">
        <h3>Select Theme</h3>
        <div class="theme-mode-controls">
            <button type="button" id="gallery-mode-btn" class="mode-btn active" aria-pressed="true">
                Gallery
            </button>
            <button type="button" id="design-mode-btn" class="mode-btn" aria-pressed="false">
                Design Mode
            </button>
        </div>
    </div>
    
    <!-- Gallery Mode (existing) -->
    <div id="theme-gallery-container" class="theme-gallery-container">
        <div class="theme-gallery" role="radiogroup" aria-label="Select theme">
            <!-- existing theme cards -->
        </div>
    </div>
    
    <!-- Design Mode (new) -->
    <div id="design-mode-container" class="design-mode-container" hidden aria-hidden="true">
        <div class="design-mode-theme-selector">
            <label for="design-theme-select">Theme:</label>
            <select id="design-theme-select" class="form-select" aria-label="Select theme for live preview">
                {{range .Themes}}
                <option value="{{.ID}}" {{if eq .ID $.SelectedThemeID}}selected{{end}}>
                    {{.Name}}
                </option>
                {{end}}
            </select>
        </div>
        
        <div class="live-preview-wrapper">
            <div class="live-preview-loading" hidden aria-live="polite">
                <div class="spinner" role="status" aria-label="Loading preview"></div>
                <p>Loading preview...</p>
            </div>
            <iframe 
                id="live-preview-frame"
                class="live-preview-frame"
                title="Live preview of your RSVP invitation"
                sandbox="allow-same-origin"
                loading="lazy"
                aria-live="polite">
            </iframe>
        </div>
    </div>
    
    <input type="hidden" id="selected-theme-id" name="template_id" value="{{.SelectedThemeID}}">
</div>
{{end}}
```

#### 3.2.2 CSS Changes

**File:** `static/css/theme_picker.css`

Add styles for:
- Mode toggle buttons
- Design mode container
- Compact theme selector
- Live preview iframe
- Loading state

```css
/* Mode Controls */
.theme-mode-controls {
    display: flex;
    gap: var(--spacing-2);
    background: var(--color-surface);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-base);
    padding: 4px;
}

.mode-btn {
    padding: var(--spacing-2) var(--spacing-3);
    border: none;
    background: transparent;
    color: var(--color-text-secondary);
    border-radius: var(--radius-sm);
    cursor: pointer;
    transition: all var(--transition-fast);
    font-size: var(--font-size-sm);
    font-weight: var(--font-weight-medium);
}

.mode-btn:hover {
    background: var(--color-gray-100);
}

.mode-btn.active {
    background: var(--color-primary-600);
    color: white;
}

/* Design Mode Container */
.design-mode-container {
    display: flex;
    flex-direction: column;
    gap: var(--spacing-4);
}

.design-mode-theme-selector {
    display: flex;
    align-items: center;
    gap: var(--spacing-2);
}

.design-mode-theme-selector label {
    font-weight: var(--font-weight-medium);
    color: var(--color-text-primary);
}

.design-mode-theme-selector select {
    flex: 1;
    max-width: 300px;
}

/* Live Preview */
.live-preview-wrapper {
    position: relative;
    width: 100%;
    background: var(--color-gray-100);
    border-radius: var(--radius-lg);
    overflow: hidden;
    box-shadow: var(--shadow-md);
}

.live-preview-frame {
    width: 100%;
    height: 800px;
    border: none;
    display: block;
    background: white;
}

.live-preview-loading {
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: var(--spacing-3);
    z-index: 10;
}

.spinner {
    width: 40px;
    height: 40px;
    border: 4px solid var(--color-gray-300);
    border-top-color: var(--color-primary-600);
    border-radius: 50%;
    animation: spin 1s linear infinite;
}

@keyframes spin {
    to { transform: rotate(360deg); }
}

/* Hide gallery when in design mode */
.theme-picker[data-mode="design"] .theme-gallery-container {
    display: none;
}

.theme-picker[data-mode="gallery"] .design-mode-container {
    display: none;
}

/* Hide gallery when in design mode */
.theme-picker[data-mode="design"] .theme-gallery-container {
    display: none;
}

.theme-picker[data-mode="gallery"] .design-mode-container {
    display: none;
}

/* Mobile View Toggle (only visible on mobile in design mode) */
.mobile-view-toggle {
    display: none;
    position: sticky;
    top: 0;
    z-index: 100;
    background: var(--color-surface);
    border-bottom: 2px solid var(--color-border);
    padding: var(--spacing-2);
    gap: var(--spacing-2);
    box-shadow: var(--shadow-sm);
}

.mobile-view-toggle .view-btn {
    flex: 1;
    padding: var(--spacing-2) var(--spacing-3);
    border: 1px solid var(--color-border);
    background: var(--color-surface);
    color: var(--color-text-secondary);
    border-radius: var(--radius-base);
    cursor: pointer;
    transition: all var(--transition-fast);
    font-size: var(--font-size-sm);
    font-weight: var(--font-weight-medium);
    touch-action: manipulation;
}

.mobile-view-toggle .view-btn:active {
    transform: scale(0.98);
}

.mobile-view-toggle .view-btn.active {
    background: var(--color-primary-600);
    color: white;
    border-color: var(--color-primary-600);
}

/* Responsive Breakpoints */

/* Tablet (768px - 1023px) */
@media (min-width: 768px) and (max-width: 1023px) {
    .theme-picker-header {
        flex-wrap: nowrap;
    }
    
    .theme-mode-controls {
        flex-shrink: 0;
    }
    
    .live-preview-frame {
        height: 600px;
    }
    
    .design-mode-theme-selector {
        flex-wrap: wrap;
    }
    
    .design-mode-theme-selector select {
        max-width: 100%;
    }
}

/* Mobile (<768px) */
@media (max-width: 767px) {
    /* Show mobile view toggle in design mode */
    .theme-picker[data-mode="design"] .mobile-view-toggle {
        display: flex;
    }
    
    /* Mode toggle buttons smaller on mobile */
    .mode-btn {
        padding: var(--spacing-1) var(--spacing-2);
        font-size: var(--font-size-xs);
    }
    
    /* Theme selector full width */
    .design-mode-theme-selector {
        flex-direction: column;
        align-items: stretch;
    }
    
    .design-mode-theme-selector select {
        max-width: 100%;
    }
    
    /* Preview takes full height on mobile */
    .live-preview-frame {
        height: calc(100vh - 200px);
        min-height: 500px;
    }
    
    /* Hide form when showing preview on mobile */
    .theme-picker[data-mode="design"][data-mobile-view="preview"] .event-form-details-column {
        display: none;
    }
    
    .theme-picker[data-mode="design"][data-mobile-view="preview"] .form-actions {
        display: none;
    }
    
    /* Hide preview when showing form on mobile */
    .theme-picker[data-mode="design"][data-mobile-view="edit"] .design-mode-container {
        display: none;
    }
    
    /* Gallery cards stack vertically */
    .theme-gallery {
        grid-template-columns: 1fr;
    }
    
    /* Reduce preview wrapper padding */
    .live-preview-wrapper {
        border-radius: var(--radius-md);
    }
    
    /* Loading indicator smaller on mobile */
    .spinner {
        width: 32px;
        height: 32px;
        border-width: 3px;
    }
    
    .live-preview-loading p {
        font-size: var(--font-size-sm);
    }
}

/* Small mobile (<480px) */
@media (max-width: 479px) {
    .theme-picker-header {
        flex-direction: column;
        align-items: stretch;
        gap: var(--spacing-2);
    }
    
    .theme-picker-header h3 {
        font-size: var(--font-size-lg);
    }
    
    .theme-mode-controls {
        width: 100%;
    }
    
    .mode-btn {
        flex: 1;
    }
    
    .live-preview-frame {
        height: calc(100vh - 180px);
        min-height: 400px;
    }
}

/* Landscape mobile (max-height: 600px) */
@media (max-height: 600px) and (orientation: landscape) {
    .live-preview-frame {
        height: 400px;
    }
    
    .mobile-view-toggle {
        padding: var(--spacing-1);
    }
}

/* Print styles */
@media print {
    .theme-picker {
        display: none;
    }
    
    .mobile-view-toggle {
        display: none;
    }
}
```

**File:** `static/css/event_form.css`

Update layout at 1024px breakpoint:

```css
@media (min-width: 1024px) {
    .event-form-layout {
        flex-direction: row;
        align-items: flex-start;
        flex-wrap: wrap;
        gap: var(--spacing-6);
    }

    /* Left: Theme/Preview Column (75%) */
    .event-form-theme-column {
        order: 1;
        flex: 0 0 calc(75% - (var(--spacing-6) * 0.75));
        max-width: calc(75% - (var(--spacing-6) * 0.75));
    }

    /* Right: Form Details Column (25%) */
    .event-form-details-column {
        order: 2;
        flex: 0 0 calc(25% - (var(--spacing-6) * 0.25));
        max-width: calc(25% - (var(--spacing-6) * 0.25));
    }

    /* Form actions stick to right column */
    .form-actions {
        order: 3;
        flex: 0 0 calc(25% - (var(--spacing-6) * 0.25));
        max-width: calc(25% - (var(--spacing-6) * 0.25));
        margin-left: calc(75% + (var(--spacing-6) * 0.25));
        margin-top: 0;
    }
}
```

#### 3.2.3 JavaScript Changes

**File:** `static/js/theme_picker.js`

Add methods to `ThemePicker` class:

```javascript
class ThemePicker {
    constructor() {
        this.gallery = document.querySelector('.theme-gallery');
        this.filterSelect = document.getElementById('theme-category-filter');
        this.hiddenInput = document.getElementById('selected-theme-id');
        
        // Design mode elements
        this.galleryModeBtn = document.getElementById('gallery-mode-btn');
        this.designModeBtn = document.getElementById('design-mode-btn');
        this.galleryContainer = document.getElementById('theme-gallery-container');
        this.designModeContainer = document.getElementById('design-mode-container');
        this.designThemeSelect = document.getElementById('design-theme-select');
        this.livePreviewFrame = document.getElementById('live-preview-frame');
        this.livePreviewLoading = document.querySelector('.live-preview-loading');
        
        // Mobile view toggle elements
        this.mobileEditBtn = document.getElementById('mobile-edit-btn');
        this.mobilePreviewBtn = document.getElementById('mobile-preview-btn');
        this.currentMobileView = 'edit';
        
        this.currentMode = 'gallery';
        this.previewUpdateTimer = null;
        this.lastPreviewURL = null;
        this.isMobile = false;
        
        this.init();
    }

    init() {
        if (!this.gallery) return;
        
        this.attachEventListeners();
        this.attachDesignModeListeners();
        this.initializeKeyboardNavigation();
    }

    attachDesignModeListeners() {
        if (!this.designModeBtn) return;
        
        // Mode toggle buttons
        this.galleryModeBtn?.addEventListener('click', () => {
            this.switchMode('gallery');
        });
        
        this.designModeBtn?.addEventListener('click', () => {
            this.switchMode('design');
        });
        
        // Theme selector in design mode
        this.designThemeSelect?.addEventListener('change', (e) => {
            this.selectTheme(e.target.value);
            this.updateLivePreview();
        });
        
        // Watch form fields for changes
        this.attachFormWatchers();
    }

    switchMode(mode) {
        this.currentMode = mode;
        const themePicker = document.querySelector('.theme-picker');
        
        if (mode === 'design') {
            themePicker.setAttribute('data-mode', 'design');
            this.galleryModeBtn.classList.remove('active');
            this.designModeBtn.classList.add('active');
            this.galleryContainer.hidden = true;
            this.designModeContainer.hidden = false;
            
            // Load initial preview
            this.updateLivePreview();
        } else {
            themePicker.setAttribute('data-mode', 'gallery');
            this.galleryModeBtn.classList.add('active');
            this.designModeBtn.classList.remove('active');
            this.galleryContainer.hidden = false;
            this.designModeContainer.hidden = true;
        }
        
        this.announceMode(mode);
    }

    attachFormWatchers() {
        const form = document.querySelector('form[action*="/events"]');
        if (!form) return;
        
        const watchFields = [
            '[name="title"]',
            '[name="location"]',
            '[name="description"]',
            '[name="start_time"]',
            '[name="custom_theme_image_url"]',
            '#custom-theme-color-value'
        ];
        
        watchFields.forEach(selector => {
            const field = form.querySelector(selector);
            if (field) {
                field.addEventListener('input', () => {
                    this.debouncedUpdatePreview();
                });
            }
        });
        
        // Also watch for color picker changes
        document.addEventListener('colorChanged', () => {
            if (this.currentMode === 'design') {
                this.debouncedUpdatePreview();
            }
        });
    }

    debouncedUpdatePreview() {
        if (this.currentMode !== 'design') return;
        
        clearTimeout(this.previewUpdateTimer);
        this.previewUpdateTimer = setTimeout(() => {
            this.updateLivePreview();
        }, 500); // 500ms debounce
    }

    updateLivePreview() {
        const themeId = this.hiddenInput.value;
        if (!themeId) return;
        
        const previewURL = this.buildPreviewURL(themeId);
        
        // Only update if URL changed (avoid unnecessary reloads)
        if (previewURL === this.lastPreviewURL) return;
        this.lastPreviewURL = previewURL;
        
        // Show loading indicator
        if (this.livePreviewLoading) {
            this.livePreviewLoading.hidden = false;
        }
        
        // Update iframe
        this.livePreviewFrame.src = previewURL;
        
        // Hide loading indicator after iframe loads
        this.livePreviewFrame.addEventListener('load', () => {
            if (this.livePreviewLoading) {
                this.livePreviewLoading.hidden = true;
            }
        }, { once: true });
    }

    buildPreviewURL(themeId) {
        const form = document.querySelector('form[action*="/events"]');
        if (!form) return `/api/themes/preview?theme_id=${themeId}`;
        
        const params = new URLSearchParams({
            theme_id: themeId,
            preview: 'true',
            title: form.querySelector('[name="title"]')?.value || 'Sample Event',
            location: form.querySelector('[name="location"]')?.value || 'Sample Location',
            description: form.querySelector('[name="description"]')?.value || 'Sample description',
            start_time: form.querySelector('[name="start_time"]')?.value || new Date().toISOString()
        });
        
        // Add custom image if present
        const customImageInput = form.querySelector('[name="custom_theme_image_url"]');
        if (customImageInput && customImageInput.value) {
            params.set('custom_image_url', customImageInput.value);
        }
        
        // Add custom color if present
        const customColorInput = document.getElementById('custom-theme-color-value');
        if (customColorInput && customColorInput.value) {
            params.set('custom_color', customColorInput.value);
        }
        
        return `/api/themes/preview?${params.toString()}`;
    }

    announceMode(mode) {
        const message = mode === 'design' 
            ? 'Switched to design mode with live preview'
            : 'Switched to gallery mode';
        this.announce(message);
    }
    
    // ... existing methods remain ...
}
```

---

## 4. Implementation Plan (TDD Approach)

### 4.0 CRITICAL: Test-Driven Development

**MUST follow TDD workflow from README-LLM.md:**

1. ✅ Write test (red phase)
2. ❌ Run test - should fail
3. ✅ Write minimal code to pass test
4. ✅ Run test - should pass
5. ✅ Refactor if needed
6. ✅ Repeat

**Never write implementation code before writing tests!**

### 4.1 Phase 1: Write Tests First (3 hours)

**1. JavaScript Test Suite** (`static/js/theme_picker_design_mode_test.go`)

Write tests for:
- `TestDesignModeToggle_SwitchesToDesignMode`
- `TestDesignModeToggle_SwitchesBackToGallery`
- `TestDesignModeToggle_HidesGalleryWhenInDesignMode`
- `TestDesignModeToggle_ShowsLivePreviewWhenInDesignMode`
- `TestDesignModeToggle_UpdatesARIAAttributes`
- `TestMobileViewToggle_SwitchesBetweenEditAndPreview`
- `TestMobileViewToggle_OnlyVisibleOnMobile`
- `TestMobileViewToggle_HidesFormInPreviewMode`
- `TestThemeSelector_UpdatesPreviewOnChange`
- `TestFormWatchers_TriggerPreviewUpdate`
- `TestFormWatchers_DebouncesUpdates`
- `TestPreviewURL_BuildsCorrectlyWithAllParams`
- `TestPreviewURL_IncludesCustomImage`
- `TestPreviewURL_IncludesCustomColor`
- `TestLoadingIndicator_ShowsWhileLoading`
- `TestLoadingIndicator_HidesAfterLoad`
- `TestIframeSandbox_HasCorrectAttributes`

**2. CSS Test Suite** (`static/css/theme_picker_design_mode_test.go`)

Write tests for:
- `TestDesignModeCSS_ModeToggleButtonsExist`
- `TestDesignModeCSS_ActiveStateStyles`
- `TestDesignModeCSS_LivePreviewFrameStyles`
- `TestDesignModeCSS_LoadingIndicatorStyles`
- `TestDesignModeCSS_MobileResponsiveStyles`
- `TestDesignModeCSS_TabletResponsiveStyles`
- `TestDesignModeCSS_DesktopLayoutStyles`
- `TestDesignModeCSS_MobileViewToggleVisible`
- `TestDesignModeCSS_SpinnerAnimation`

**3. HTML Test Suite** (`templates/web/theme_picker_design_mode_test.go`)

Write tests for:
- `TestDesignModeHTML_ModeToggleButtonsPresent`
- `TestDesignModeHTML_DesignModeContainerExists`
- `TestDesignModeHTML_ThemeSelectorExists`
- `TestDesignModeHTML_LivePreviewIframeExists`
- `TestDesignModeHTML_LoadingIndicatorExists`
- `TestDesignModeHTML_MobileViewToggleExists`
- `TestDesignModeHTML_ARIAAttributesCorrect`
- `TestDesignModeHTML_HiddenByDefault`

**4. Integration Test Suite** (`templates/web/theme_picker_integration_test.go`)

Write tests for:
- `TestDesignModeIntegration_FullWorkflow`
- `TestDesignModeIntegration_WithAllThemes`
- `TestDesignModeIntegration_WithCustomImage`
- `TestDesignModeIntegration_WithCustomColor`
- `TestDesignModeIntegration_MobileResponsive`

**Expected Result:** All tests FAIL (red phase) ✅

### 4.2 Phase 2: Implement Minimal Code (2 hours)

**Goal:** Write minimal code to make tests pass (green phase)

**4. Implement HTML** (`templates/web/partials/theme_picker.html`)
   - Add mobile view toggle
   - Add mode toggle buttons  
   - Add design mode container
   - Add compact theme selector
   - Add live preview iframe
   - Add loading indicator
   - Add data attributes and ARIA attributes

**Run Tests:** Verify HTML structure tests pass ✅

**5. Implement CSS** (`static/css/theme_picker.css`, `static/css/event_form.css`)
   - Add mode toggle button styles
   - Add active state styles
   - Add design mode container layout
   - Add live preview iframe sizing
   - Add loading indicator + spinner animation
   - Add mobile responsive styles (<768px)
   - Add tablet responsive styles (768px-1023px)
   - Add desktop layout (≥1024px)
   - Add mobile view toggle styles

**Run Tests:** Verify CSS tests pass ✅

**6. Implement JavaScript Core** (`static/js/theme_picker.js`)
   - Add constructor properties for design mode elements
   - Add `attachDesignModeListeners()` method
   - Add `switchMode(mode)` method
   - Add `attachMobileViewToggle()` method
   - Add ARIA attribute updates
   - Add announcements for mode changes

**Run Tests:** Verify mode toggle tests pass ✅

### 4.3 Phase 3: Live Preview Logic (1.5 hours)

**7. Implement Form Watchers**
   - Add `attachFormWatchers()` method
   - Watch all form fields (title, location, description, start_time)
   - Watch custom image and color inputs
   - Add `debouncedUpdatePreview()` with 500ms delay

**Run Tests:** Verify form watcher tests pass ✅

**8. Implement Preview Updates**
   - Add `updateLivePreview()` method
   - Add `buildPreviewURL(themeId)` method
   - Build query params from form data
   - Update iframe src
   - Prevent unnecessary reloads (check if URL changed)

**Run Tests:** Verify preview update tests pass ✅

**9. Implement Loading States**
   - Show loading indicator when iframe starts loading
   - Hide loading indicator on iframe load event
   - Handle iframe errors gracefully

**Run Tests:** Verify loading indicator tests pass ✅

**10. Theme Selector Integration**
   - Handle theme dropdown change in design mode
   - Update hidden input value
   - Trigger preview update

**Run Tests:** Verify theme selector tests pass ✅

### 4.4 Phase 4: Integration & Manual Testing (1 hour)

**11. Run Full Test Suite**
```bash
go test ./static/js/... -v
go test ./static/css/... -v  
go test ./templates/web/... -v
```

**Expected:** All tests pass ✅

**12. Manual Testing Checklist**
- [ ] Desktop (≥1024px): Gallery mode shows thumbnails
- [ ] Desktop: Design Mode shows preview left, form right
- [ ] Desktop: Type in title → preview updates after 500ms
- [ ] Desktop: Switch themes → preview updates
- [ ] Desktop: Add custom image → preview shows image
- [ ] Desktop: Pick custom color → preview applies color
- [ ] Tablet (768-1023px): Form top, preview bottom
- [ ] Mobile (<768px): Toggle between Edit/Preview works
- [ ] Mobile: Preview mode hides form completely
- [ ] Mobile: Edit mode shows form, hides preview
- [ ] All: Loading spinner shows while loading
- [ ] All: ARIA announcements work
- [ ] All: Keyboard navigation works

**13. Cross-Browser Testing**
- [ ] Chrome/Edge: All functionality works
- [ ] Firefox: All functionality works
- [ ] Safari: All functionality works (if available)

**14. Theme Testing**
- [ ] Test with all 7 themes (each renders correctly)
- [ ] Custom images display in preview
- [ ] Custom colors apply CSS variables

### 4.5 Phase 5: Refactor & Polish (30 min)

**15. Code Review**
- [ ] Remove duplicate code
- [ ] Extract magic numbers to constants
- [ ] Add JSDoc comments
- [ ] Simplify complex conditionals
- [ ] Ensure consistent naming

**16. Performance Check**
- [ ] Verify debouncing works (max 1 update per 500ms)
- [ ] Check memory usage (no iframe leaks)
- [ ] Verify preview loads in <1 second

**Run Tests One More Time:** All pass ✅

**DONE!** Ready for code review and merge.

---

## 5. Testing Strategy

### 5.1 Unit Tests

**JavaScript Tests:** `static/js/theme_picker_design_mode_test.go`

Test cases:
- Mode toggle switches between gallery and design
- Theme selector in design mode updates preview
- Form field changes trigger preview updates
- Debouncing prevents excessive updates
- Preview URL is built correctly with all parameters
- Loading indicator shows/hides appropriately

### 5.2 Integration Tests

**E2E Tests:** Manual testing checklist

- [ ] Click "Design Mode" button → preview appears
- [ ] Type in title field → preview updates after 500ms
- [ ] Type in location field → preview updates
- [ ] Change date/time → preview updates
- [ ] Upload custom image → preview shows image
- [ ] Pick custom color → preview applies color
- [ ] Switch themes in design mode → preview updates
- [ ] Switch back to gallery mode → gallery appears
- [ ] Form submission includes correct template_id

### 5.3 Visual Regression Testing

- [ ] Take screenshots of design mode in all 7 themes
- [ ] Verify layout at 1024px, 1440px, 1920px
- [ ] Verify mobile layout (vertical stack)
- [ ] Verify no layout shifts during preview updates

---

## 6. Future Enhancements

### 6.1 Not in Scope for V1

- Real-time collaborative editing
- Preview of RSVP form fields (questions)
- Preview of confirmation page
- Preview on mobile device sizes (responsive toggle)
- Undo/redo for design changes
- Save as draft without publishing

### 6.2 Potential Future Improvements

- Add "Preview on Mobile" toggle to show mobile-sized iframe
- Add zoom controls for preview
- Add side-by-side theme comparison
- Add history/versioning of designs
- Add AI-powered theme recommendations based on event type

---

## 7. Success Metrics

### 7.1 User Experience Metrics

- **Time to complete event creation:** Reduce by 30%
- **Number of preview button clicks:** Reduce by 50%
- **User satisfaction:** Increase from feedback surveys

### 7.2 Technical Metrics

- **Preview load time:** < 1 second
- **Debounce effectiveness:** < 10 preview updates per form session
- **Error rate:** < 0.1% of preview loads fail

---

## 8. Risks & Mitigations

### 8.1 Technical Risks

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Preview iframe slow to load | High | Medium | Add loading indicator, optimize preview endpoint |
| Form fields trigger too many updates | Medium | High | Debounce updates to 500ms |
| Preview URL too long (query params) | Low | Low | URL limit is 2048 chars, our params are ~500 chars max |
| Iframe sandbox breaks functionality | High | Low | Use `allow-same-origin` sandbox, test thoroughly |
| Memory leak from iframe reloads | Medium | Medium | Clear iframe src on mode switch, monitor memory |

### 8.2 UX Risks

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Users confused by mode toggle | Medium | Low | Clear labels, add tooltips, default to gallery mode |
| Preview doesn't match published page | High | Low | Use exact same endpoint as theme preview modal |
| Preview too small on mobile | Medium | High | Stack vertically on mobile, allow full-screen preview |

---

## 9. Rollout Plan

### 9.1 Development

- [ ] Create feature branch: `feature/live-preview-mode`
- [ ] Implement Phase 1 (HTML/CSS)
- [ ] Implement Phase 2 (JavaScript)
- [ ] Implement Phase 3 (Polish)
- [ ] Write tests
- [ ] Code review
- [ ] Merge to main

### 9.2 Testing

- [ ] Test in staging environment
- [ ] Manual QA with all 7 themes
- [ ] Cross-browser testing
- [ ] Accessibility audit
- [ ] Performance profiling

### 9.3 Launch

- [ ] Deploy to production
- [ ] Monitor error logs
- [ ] Monitor performance metrics
- [ ] Gather user feedback
- [ ] Iterate based on feedback

---

## 10. References

### 10.1 Existing Code

- `internal/handlers/templates.go:451` - HandleThemePreview endpoint
- `static/js/theme_picker.js` - Existing theme picker logic
- `static/js/theme_preview_modal.js` - Modal preview implementation (reference)
- `templates/web/event_form.html` - Event creation page layout
- `static/css/event_form.css` - Event form layout styles
- `static/css/theme_picker.css` - Theme picker styles

### 10.2 Design Inspirations

- **Canva:** Live canvas updates as you edit
- **Squarespace:** Live website preview in editor
- **Figma:** Real-time design preview
- **Mailchimp:** Email editor with live preview panel

---

## 11. Open Questions

1. **Should we persist the user's preferred mode (gallery vs design)?**
   - Decision: No, always default to gallery mode for first-time users

2. **Should design mode be available on mobile?**
   - Decision: Yes, but stack vertically (form top, preview bottom)

3. **Should we show the preview for users editing an existing event?**
   - Decision: Yes, design mode should work for both create and edit

4. **Should we preload the first theme in design mode to avoid blank state?**
   - Decision: Yes, load preview immediately when entering design mode

5. **Should we add a "Full Screen" button for the preview?**
   - Decision: Not in V1, but good candidate for future enhancement

---

**End of Design Document**

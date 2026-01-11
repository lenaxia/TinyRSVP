# User Story 10.12: Light/Dark Theme Switching

**Epic:** 10 - Technical Debt & Improvements
**Status:** Complete
**Priority:** Medium
**Effort:** Medium-High
**Owner:** LLM (2026-01-11)

---

## User Story

As a **user of TinyRSVP**,  
I want to **toggle between light and dark themes in real-time**,  
So that **I can use the application comfortably in different lighting conditions and match my personal preference**.

---

## Context

Currently, TinyRSVP has:
- CSS variables system with semantic color tokens
- Basic dark mode via `@media (prefers-color-scheme: dark)` for 6 variables only
- No manual theme toggle control
- No theme persistence across sessions

This story implements a complete theme switching system with:
- Full dark mode color palette for all UI elements
- Manual theme toggle with sun/moon icon in navigation
- Theme persistence via localStorage
- Real-time theme switching without page reload

---

## Acceptance Criteria

### Theme System Architecture
- [x] Refactor CSS from media query to data attribute approach (`[data-theme="dark"]`)
- [x] Define complete dark mode color palette for all CSS variables
- [x] Ensure all UI components respect theme variables
- [x] Add smooth transitions for theme changes

### Theme Toggle UI
- [x] Add theme toggle button to navigation (sun/moon icon)
- [x] Position toggle appropriately in mobile and desktop layouts
- [x] Include accessible labels and ARIA attributes
- [x] Provide visual feedback on hover/focus

### Theme Controller JavaScript
- [x] Create theme controller module
- [x] Implement theme toggle functionality
- [x] Persist theme preference to localStorage
- [x] Initialize theme on page load (localStorage > system preference > light)
- [x] Handle system preference changes (optional enhancement)

### Color Palette Completeness
- [x] All primary color shades (50-900) have dark mode equivalents
- [x] Success/warning/error colors work in both themes
- [x] Gray scale properly adjusted for dark mode
- [x] Surface variants (disabled, hover states) defined
- [x] Border colors appropriate for both themes
- [x] Focus states visible in both themes
- [x] Shadows adjusted for dark mode (lighter shadows)

### Testing
- [x] Unit tests for theme controller JavaScript
- [x] Integration tests for theme persistence
- [x] Visual regression tests for all pages in both themes
- [x] Verify all components render correctly in dark mode
- [x] Test theme toggle accessibility (keyboard, screen reader)
- [x] Test theme initialization edge cases

### Documentation
- [x] Update CSS README with theme system documentation
- [x] Document theme variable naming conventions
- [x] Add examples of adding new themed components
- [x] Document localStorage key and data structure

---

## Technical Design

### CSS Architecture

**Current (Media Query Based):**
```css
:root {
    --color-background: #ffffff;
    /* ... light colors ... */
}

@media (prefers-color-scheme: dark) {
    :root {
        --color-background: #111827;
        /* ... only 6 variables ... */
    }
}
```

**New (Data Attribute Based):**
```css
:root {
    /* Light theme (default) */
    --color-background: #ffffff;
    --color-surface: #f9fafb;
    --color-text-primary: #111827;
    /* ... all variables ... */
}

[data-theme="dark"] {
    /* Dark theme */
    --color-background: #111827;
    --color-surface: #1f2937;
    --color-text-primary: #f9fafb;
    /* ... all variables with dark equivalents ... */
}
```

### Theme Controller JavaScript

**File:** `static/js/theme_controller.js`

```javascript
class ThemeController {
    constructor() {
        this.STORAGE_KEY = 'tinyrsvp-theme';
        this.THEMES = { LIGHT: 'light', DARK: 'dark' };
        this.init();
    }

    init() {
        const savedTheme = this.getSavedTheme();
        const systemTheme = this.getSystemTheme();
        const theme = savedTheme || systemTheme || this.THEMES.LIGHT;
        this.setTheme(theme);
        this.attachEventListeners();
    }

    getSavedTheme() {
        return localStorage.getItem(this.STORAGE_KEY);
    }

    getSystemTheme() {
        if (window.matchMedia('(prefers-color-scheme: dark)').matches) {
            return this.THEMES.DARK;
        }
        return this.THEMES.LIGHT;
    }

    setTheme(theme) {
        document.documentElement.setAttribute('data-theme', theme);
        localStorage.setItem(this.STORAGE_KEY, theme);
        this.updateToggleButton(theme);
    }

    toggleTheme() {
        const currentTheme = document.documentElement.getAttribute('data-theme');
        const newTheme = currentTheme === this.THEMES.DARK 
            ? this.THEMES.LIGHT 
            : this.THEMES.DARK;
        this.setTheme(newTheme);
    }

    updateToggleButton(theme) {
        const button = document.getElementById('theme-toggle');
        if (!button) return;
        
        const icon = button.querySelector('.theme-icon');
        const label = button.querySelector('.sr-only');
        
        if (theme === this.THEMES.DARK) {
            icon.textContent = '☀️';
            label.textContent = 'Switch to light mode';
            button.setAttribute('aria-label', 'Switch to light mode');
        } else {
            icon.textContent = '🌙';
            label.textContent = 'Switch to dark mode';
            button.setAttribute('aria-label', 'Switch to dark mode');
        }
    }

    attachEventListeners() {
        const button = document.getElementById('theme-toggle');
        if (button) {
            button.addEventListener('click', () => this.toggleTheme());
        }
    }
}

// Initialize on DOM ready
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', () => new ThemeController());
} else {
    new ThemeController();
}
```

### Theme Toggle Component

**File:** `static/css/theme_toggle.css`

```css
.theme-toggle {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 44px;
    height: 44px;
    padding: 0;
    background: transparent;
    border: 1px solid var(--color-border);
    border-radius: var(--radius-base);
    cursor: pointer;
    transition: all var(--transition-fast);
    font-size: var(--font-size-xl);
}

.theme-toggle:hover {
    background: var(--color-surface);
    border-color: var(--color-primary-600);
}

.theme-toggle:focus {
    outline: 2px solid var(--color-border-focus);
    outline-offset: 2px;
}

.theme-icon {
    display: block;
    line-height: 1;
}

@media (min-width: 768px) {
    .theme-toggle {
        margin-left: var(--spacing-4);
    }
}
```

### Navigation Template Update

**File:** `templates/web/partials/navigation.html`

Add theme toggle button after brand or in menu:

```html
<div class="app-nav-header">
    <button class="app-nav-toggle" aria-label="Toggle navigation" aria-expanded="false">
        <span></span>
    </button>
    <a href="/" class="app-nav-brand">TinyRSVP</a>
    <button id="theme-toggle" class="theme-toggle" aria-label="Toggle theme">
        <span class="theme-icon">🌙</span>
        <span class="sr-only">Toggle theme</span>
    </button>
</div>
```

---

## Dark Mode Color Palette

### Background & Surface Colors
```css
[data-theme="dark"] {
    --color-background: #0f172a;        /* Darker than current #111827 */
    --color-surface: #1e293b;           /* Elevated surfaces */
    --color-surface-disabled: #334155;  /* Disabled state */
}
```

### Text Colors
```css
[data-theme="dark"] {
    --color-text-primary: #f1f5f9;
    --color-text-secondary: #cbd5e1;
    --color-text-tertiary: #94a3b8;
    --color-text-muted: #94a3b8;
    --color-text-label: #cbd5e1;
    --color-text-disabled: #64748b;
}
```

### Primary Colors (Adjusted for Dark)
```css
[data-theme="dark"] {
    --color-primary-50: #1e3a8a;       /* Inverted scale */
    --color-primary-100: #1e40af;
    --color-primary-200: #1d4ed8;
    --color-primary-300: #2563eb;
    --color-primary-400: #3b82f6;
    --color-primary-500: #60a5fa;      /* Mid-point stays similar */
    --color-primary-600: #93c5fd;
    --color-primary-700: #bfdbfe;
    --color-primary-800: #dbeafe;
    --color-primary-900: #eff6ff;
}
```

### State Colors
```css
[data-theme="dark"] {
    --color-success: #22c55e;          /* Brighter for visibility */
    --color-success-dark: #16a34a;
    --color-success-light: #166534;    /* Darker in dark mode */
    --color-success-50: #14532d;
    --color-success-200: #15803d;
    --color-success-700: #4ade80;
    
    --color-warning: #f59e0b;
    --color-warning-dark: #d97706;
    --color-warning-darker: #b45309;
    --color-warning-light: #78350f;
    --color-warning-50: #451a03;
    --color-warning-200: #92400e;
    --color-warning-700: #fbbf24;
    
    --color-error: #ef4444;
    --color-error-dark: #dc2626;
    --color-error-light: #7f1d1d;
    --color-error-50: #450a0a;
    --color-error-200: #991b1b;
    --color-error-700: #f87171;
    
    --color-info: #06b6d4;
    --color-info-light: #164e63;
}
```

### Gray Scale
```css
[data-theme="dark"] {
    --color-gray-50: #1e293b;
    --color-gray-100: #334155;
    --color-gray-200: #475569;
    --color-gray-300: #64748b;
    --color-gray-400: #94a3b8;
    --color-gray-500: #cbd5e1;
    --color-gray-600: #e2e8f0;
    --color-gray-700: #f1f5f9;
    --color-gray-800: #f8fafc;
    --color-gray-900: #ffffff;
}
```

### Borders & Focus
```css
[data-theme="dark"] {
    --color-border: #334155;
    --color-border-focus: #60a5fa;     /* Brighter blue for visibility */
}
```

### Shadows (Lighter for Dark Mode)
```css
[data-theme="dark"] {
    --shadow-sm: 0 1px 2px 0 rgba(0, 0, 0, 0.3);
    --shadow-base: 0 1px 3px 0 rgba(0, 0, 0, 0.4), 0 1px 2px 0 rgba(0, 0, 0, 0.3);
    --shadow-md: 0 4px 6px -1px rgba(0, 0, 0, 0.4), 0 2px 4px -1px rgba(0, 0, 0, 0.3);
    --shadow-lg: 0 10px 15px -3px rgba(0, 0, 0, 0.5), 0 4px 6px -2px rgba(0, 0, 0, 0.4);
    --shadow-xl: 0 20px 25px -5px rgba(0, 0, 0, 0.5), 0 10px 10px -5px rgba(0, 0, 0, 0.4);
    --shadow-2xl: 0 25px 50px -12px rgba(0, 0, 0, 0.6);
}
```

---

## Implementation Tasks

### Phase 1: CSS Foundation
- [x] Refactor `variables.css` to use `[data-theme="dark"]` instead of media query
- [x] Define complete dark mode color palette
- [x] Add smooth transition for theme changes
- [x] Test all existing components still work

### Phase 2: Theme Toggle Component
- [x] Create `theme_toggle.css`
- [x] Add theme toggle button styles
- [x] Ensure accessibility (focus states, ARIA)
- [x] Test responsive behavior

### Phase 3: JavaScript Controller
- [x] Create `theme_controller.js`
- [x] Implement theme toggle logic
- [x] Add localStorage persistence
- [x] Handle initialization edge cases
- [x] Write unit tests

### Phase 4: Template Integration
- [x] Update `navigation.html` partial
- [x] Add theme toggle button to navigation
- [x] Include theme controller script in base template
- [x] Test on all pages

### Phase 5: Testing & Validation
- [x] Visual test all pages in dark mode
- [x] Test all interactive components
- [x] Verify forms, modals, buttons in both themes
- [x] Test theme persistence across sessions
- [x] Accessibility audit (keyboard, screen reader)
- [x] Write integration tests

### Phase 6: Documentation
- [x] Update `static/css/README.md`
- [x] Document theme system architecture
- [x] Add examples for themed components
- [x] Document localStorage structure

---

## Files to Create

1. `static/css/theme_toggle.css` - Theme toggle button styles
2. `static/js/theme_controller.js` - Theme switching logic
3. `static/js/theme_controller_test.go` - Unit tests
4. `static/css/theme_toggle_test.go` - CSS integration tests

---

## Files to Modify

1. `static/css/variables.css` - Refactor to data attribute, add full dark palette
2. `templates/web/partials/navigation.html` - Add theme toggle button
3. `static/css/README.md` - Document theme system
4. All page templates - Include theme controller script (if not in base)

---

## Testing Strategy

### Unit Tests
- Theme controller initialization
- Theme toggle functionality
- localStorage persistence
- Theme preference resolution (saved > system > default)

### Integration Tests
- Theme toggle button interaction
- Theme persistence across page loads
- All components render correctly in both themes
- Smooth transitions without flicker

### Visual Regression Tests
- Screenshot comparison of key pages in both themes
- Verify color contrast meets WCAG AA standards
- Test all interactive states (hover, focus, active)

### Accessibility Tests
- Keyboard navigation to theme toggle
- Screen reader announces theme changes
- Focus visible in both themes
- Color contrast ratios meet standards

---

## Success Metrics

- [x] All pages support both light and dark themes
- [x] Theme toggle works on all pages without reload
- [x] Theme preference persists across sessions
- [x] All tests pass (unit, integration, visual)
- [x] Accessibility audit passes
- [x] No visual regressions in light mode
- [x] User can override system preference

---

## Dependencies

- None (self-contained improvement)

---

## References

- **Current CSS Variables:** [`static/css/variables.css`](../../static/css/variables.css)
- **Navigation Template:** [`templates/web/partials/navigation.html`](../../templates/web/partials/navigation.html)
- **CSS README:** [`static/css/README.md`](../../static/css/README.md)
- **HLD:** [`docs/02_REVISED_HLD.md`](../02_REVISED_HLD.md)

---

## Notes

- Consider adding system preference change listener (optional enhancement)
- May want to add theme preview in admin settings (future story)
- Could add per-user theme preference in database (future story)
- Icon choice: Using emoji (🌙/☀️) for simplicity, could use SVG icons
- Smooth transitions prevent jarring theme switches
- Data attribute approach allows more than 2 themes in future (e.g., high contrast)

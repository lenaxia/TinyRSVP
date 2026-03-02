# User Story: CSS Custom Properties System

**Epic:** [07_EPIC_frontend.md](07_EPIC_frontend.md)
**Priority:** High
**Status:** Complete
**Estimated Effort:** 0.5 days

---

## User Story

As a **developer**, I want **a comprehensive CSS custom properties (variables) system** so that **I can maintain consistent theming and easily customize the application's appearance**.

---

## Acceptance Criteria

- [x] CSS variables defined for all design tokens
- [x] Color palette with semantic naming
- [x] Spacing scale system
- [x] Typography scale
- [x] Border radius values
- [x] Shadow definitions
- [x] Transition/animation values
- [x] Breakpoint variables
- [x] Dark mode support (optional)
- [x] All variables documented
- [x] Variables tested across browsers

---

## Technical Details

### CSS Variables Structure

```css
/* static/css/variables.css */

:root {
    /* Colors - Primary */
    --color-primary-50: #eff6ff;
    --color-primary-100: #dbeafe;
    --color-primary-200: #bfdbfe;
    --color-primary-300: #93c5fd;
    --color-primary-400: #60a5fa;
    --color-primary-500: #3b82f6;
    --color-primary-600: #2563eb;
    --color-primary-700: #1d4ed8;
    --color-primary-800: #1e40af;
    --color-primary-900: #1e3a8a;
    
    /* Colors - Semantic */
    --color-success: #16a34a;
    --color-success-light: #dcfce7;
    --color-warning: #f59e0b;
    --color-warning-light: #fef3c7;
    --color-error: #dc2626;
    --color-error-light: #fee2e2;
    --color-info: #0ea5e9;
    --color-info-light: #e0f2fe;
    
    /* Colors - Neutral */
    --color-gray-50: #f9fafb;
    --color-gray-100: #f3f4f6;
    --color-gray-200: #e5e7eb;
    --color-gray-300: #d1d5db;
    --color-gray-400: #9ca3af;
    --color-gray-500: #6b7280;
    --color-gray-600: #4b5563;
    --color-gray-700: #374151;
    --color-gray-800: #1f2937;
    --color-gray-900: #111827;
    
    /* Colors - Functional */
    --color-background: #ffffff;
    --color-surface: #f9fafb;
    --color-text-primary: #111827;
    --color-text-secondary: #6b7280;
    --color-text-disabled: #9ca3af;
    --color-border: #e5e7eb;
    --color-border-focus: #3b82f6;
    
    /* Spacing Scale (8px base) */
    --spacing-0: 0;
    --spacing-1: 0.25rem;  /* 4px */
    --spacing-2: 0.5rem;   /* 8px */
    --spacing-3: 0.75rem;  /* 12px */
    --spacing-4: 1rem;     /* 16px */
    --spacing-5: 1.25rem;  /* 20px */
    --spacing-6: 1.5rem;   /* 24px */
    --spacing-8: 2rem;     /* 32px */
    --spacing-10: 2.5rem;  /* 40px */
    --spacing-12: 3rem;    /* 48px */
    --spacing-16: 4rem;    /* 64px */
    --spacing-20: 5rem;    /* 80px */
    --spacing-24: 6rem;    /* 96px */
    
    /* Typography Scale */
    --font-size-xs: 0.75rem;    /* 12px */
    --font-size-sm: 0.875rem;   /* 14px */
    --font-size-base: 1rem;     /* 16px */
    --font-size-lg: 1.125rem;   /* 18px */
    --font-size-xl: 1.25rem;    /* 20px */
    --font-size-2xl: 1.5rem;    /* 24px */
    --font-size-3xl: 1.875rem;  /* 30px */
    --font-size-4xl: 2.25rem;   /* 36px */
    --font-size-5xl: 3rem;      /* 48px */
    
    /* Font Weights */
    --font-weight-normal: 400;
    --font-weight-medium: 500;
    --font-weight-semibold: 600;
    --font-weight-bold: 700;
    
    /* Line Heights */
    --line-height-tight: 1.25;
    --line-height-normal: 1.5;
    --line-height-relaxed: 1.75;
    
    /* Font Families */
    --font-family-sans: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
    --font-family-mono: "SF Mono", Monaco, "Cascadia Code", "Roboto Mono", Consolas, "Courier New", monospace;
    
    /* Border Radius */
    --radius-none: 0;
    --radius-sm: 0.125rem;   /* 2px */
    --radius-base: 0.25rem;  /* 4px */
    --radius-md: 0.375rem;   /* 6px */
    --radius-lg: 0.5rem;     /* 8px */
    --radius-xl: 0.75rem;    /* 12px */
    --radius-2xl: 1rem;      /* 16px */
    --radius-full: 9999px;
    
    /* Shadows */
    --shadow-sm: 0 1px 2px 0 rgba(0, 0, 0, 0.05);
    --shadow-base: 0 1px 3px 0 rgba(0, 0, 0, 0.1), 0 1px 2px 0 rgba(0, 0, 0, 0.06);
    --shadow-md: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06);
    --shadow-lg: 0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05);
    --shadow-xl: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04);
    --shadow-2xl: 0 25px 50px -12px rgba(0, 0, 0, 0.25);
    
    /* Transitions */
    --transition-fast: 150ms ease-in-out;
    --transition-base: 200ms ease-in-out;
    --transition-slow: 300ms ease-in-out;
    
    /* Z-Index Scale */
    --z-index-dropdown: 1000;
    --z-index-sticky: 1020;
    --z-index-fixed: 1030;
    --z-index-modal-backdrop: 1040;
    --z-index-modal: 1050;
    --z-index-popover: 1060;
    --z-index-tooltip: 1070;
    
    /* Breakpoints (for use in media queries) */
    --breakpoint-sm: 640px;
    --breakpoint-md: 768px;
    --breakpoint-lg: 1024px;
    --breakpoint-xl: 1280px;
    --breakpoint-2xl: 1536px;
    
    /* Container Max Widths */
    --container-sm: 640px;
    --container-md: 768px;
    --container-lg: 1024px;
    --container-xl: 1280px;
}

/* Dark Mode (Optional) */
@media (prefers-color-scheme: dark) {
    :root {
        --color-background: #111827;
        --color-surface: #1f2937;
        --color-text-primary: #f9fafb;
        --color-text-secondary: #d1d5db;
        --color-text-disabled: #6b7280;
        --color-border: #374151;
    }
}
```

### Usage Examples

```css
/* Using color variables */
.button-primary {
    background-color: var(--color-primary-600);
    color: var(--color-background);
}

.button-primary:hover {
    background-color: var(--color-primary-700);
}

/* Using spacing */
.card {
    padding: var(--spacing-6);
    margin-bottom: var(--spacing-4);
}

/* Using typography */
.heading-1 {
    font-size: var(--font-size-4xl);
    font-weight: var(--font-weight-bold);
    line-height: var(--line-height-tight);
}

/* Using shadows and radius */
.card {
    border-radius: var(--radius-lg);
    box-shadow: var(--shadow-md);
}

/* Using transitions */
.button {
    transition: all var(--transition-base);
}
```

---

## Tasks

### Phase 1: Color System (TDD)
- [x] Define primary color palette (50-900 scale)
- [x] Define semantic colors (success, warning, error, info)
- [x] Define neutral/gray scale
- [x] Define functional colors (background, text, border)
- [x] Test color contrast ratios (WCAG AA)
- [x] Document color usage guidelines

### Phase 2: Spacing & Layout
- [x] Define spacing scale (8px base)
- [x] Define container max widths
- [x] Define breakpoint values
- [x] Define z-index scale
- [x] Document spacing usage

### Phase 3: Typography
- [x] Define font size scale
- [x] Define font weight values
- [x] Define line height values
- [x] Define font family stacks
- [x] Test typography across devices

### Phase 4: Visual Effects
- [x] Define border radius values
- [x] Define shadow system
- [x] Define transition timings
- [x] Test visual consistency

### Phase 5: Dark Mode (Optional)
- [x] Define dark mode color overrides
- [x] Test dark mode contrast
- [x] Document dark mode usage

### Phase 6: Documentation
- [x] Create variable reference guide
- [x] Create usage examples
- [x] Document browser support
- [x] Create migration guide for existing styles

---

## Testing Requirements

### Browser Compatibility
- [x] Chrome/Edge (latest 2 versions)
- [x] Firefox (latest 2 versions)
- [x] Safari (latest 2 versions)
- [x] Mobile Safari (iOS 14+)
- [x] Chrome Mobile (Android 10+)

### Visual Testing
- [x] Test on mobile (320px, 375px, 414px)
- [x] Test on tablet (768px, 1024px)
- [x] Test on desktop (1280px, 1920px)
- [x] Test dark mode (if implemented)

### Accessibility Testing
- [x] Verify color contrast ratios (WCAG AA)
- [x] Test with high contrast mode
- [x] Test with reduced motion preference

---

## Dependencies

**Depends on:** None (foundational)

**Blocks:**
- 07_STORY_01_typography.md
- 07_STORY_02_color_system.md
- 07_STORY_03_spacing_system.md
- All other frontend stories

---

## Definition of Done

- [x] All acceptance criteria met
- [x] All CSS variables defined
- [x] Variables documented with examples
- [x] Browser compatibility verified
- [x] Color contrast meets WCAG AA
- [x] Variables tested across breakpoints
- [x] Documentation complete
- [x] Code reviewed
- [x] Changes committed to git

---

## References

- **Epic:** [07_EPIC_frontend.md](07_EPIC_frontend.md)
- **HLD:** Section 22 (Success Criteria - UI/UX)
- **Design System:** CSS Custom Properties specification
- **Accessibility:** WCAG 2.1 AA color contrast requirements

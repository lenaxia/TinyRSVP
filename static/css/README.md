# CSS Variables System

This directory contains the CSS custom properties (variables) system for TinyRSVP.

## Purpose

Provides a comprehensive design token system using CSS custom properties for consistent theming and easy customization across the application.

## Files

- [`variables.css`](variables.css) - Core CSS custom properties defining the design system
- [`variables_test.go`](variables_test.go) - Test suite validating CSS variables

## CSS Variables Reference

### Color System

#### Primary Colors (Blue Scale)
- `--color-primary-50` through `--color-primary-900` - Primary brand color palette (50 lightest, 900 darkest)

#### Semantic Colors
- `--color-success` / `--color-success-light` - Success states (green)
- `--color-warning` / `--color-warning-light` - Warning states (amber)
- `--color-error` / `--color-error-light` - Error states (red)
- `--color-info` / `--color-info-light` - Informational states (cyan)

#### Neutral/Gray Scale
- `--color-gray-50` through `--color-gray-900` - Neutral gray palette

#### Functional Colors
- `--color-background` - Page background color
- `--color-surface` - Card/surface background color
- `--color-text-primary` - Primary text color
- `--color-text-secondary` - Secondary/muted text color
- `--color-text-disabled` - Disabled text color
- `--color-border` - Default border color
- `--color-border-focus` - Focus state border color

### Spacing Scale (8px base)

- `--spacing-0` - 0
- `--spacing-1` - 4px (0.25rem)
- `--spacing-2` - 8px (0.5rem)
- `--spacing-3` - 12px (0.75rem)
- `--spacing-4` - 16px (1rem)
- `--spacing-5` - 20px (1.25rem)
- `--spacing-6` - 24px (1.5rem)
- `--spacing-8` - 32px (2rem)
- `--spacing-10` - 40px (2.5rem)
- `--spacing-12` - 48px (3rem)
- `--spacing-16` - 64px (4rem)
- `--spacing-20` - 80px (5rem)
- `--spacing-24` - 96px (6rem)

### Typography

#### Font Sizes
- `--font-size-xs` - 12px (0.75rem)
- `--font-size-sm` - 14px (0.875rem)
- `--font-size-base` - 16px (1rem)
- `--font-size-lg` - 18px (1.125rem)
- `--font-size-xl` - 20px (1.25rem)
- `--font-size-2xl` - 24px (1.5rem)
- `--font-size-3xl` - 30px (1.875rem)
- `--font-size-4xl` - 36px (2.25rem)
- `--font-size-5xl` - 48px (3rem)

#### Font Weights
- `--font-weight-normal` - 400
- `--font-weight-medium` - 500
- `--font-weight-semibold` - 600
- `--font-weight-bold` - 700

#### Line Heights
- `--line-height-tight` - 1.25
- `--line-height-normal` - 1.5
- `--line-height-relaxed` - 1.75

#### Font Families
- `--font-family-sans` - System sans-serif stack
- `--font-family-mono` - Monospace font stack

### Border Radius

- `--radius-none` - 0
- `--radius-sm` - 2px (0.125rem)
- `--radius-base` - 4px (0.25rem)
- `--radius-md` - 6px (0.375rem)
- `--radius-lg` - 8px (0.5rem)
- `--radius-xl` - 12px (0.75rem)
- `--radius-2xl` - 16px (1rem)
- `--radius-full` - 9999px (fully rounded)

### Shadows

- `--shadow-sm` - Small shadow
- `--shadow-base` - Base shadow
- `--shadow-md` - Medium shadow
- `--shadow-lg` - Large shadow
- `--shadow-xl` - Extra large shadow
- `--shadow-2xl` - 2X large shadow

### Transitions

- `--transition-fast` - 150ms ease-in-out
- `--transition-base` - 200ms ease-in-out
- `--transition-slow` - 300ms ease-in-out

### Z-Index Scale

- `--z-index-dropdown` - 1000
- `--z-index-sticky` - 1020
- `--z-index-fixed` - 1030
- `--z-index-modal-backdrop` - 1040
- `--z-index-modal` - 1050
- `--z-index-popover` - 1060
- `--z-index-tooltip` - 1070

### Breakpoints

- `--breakpoint-sm` - 640px
- `--breakpoint-md` - 768px
- `--breakpoint-lg` - 1024px
- `--breakpoint-xl` - 1280px
- `--breakpoint-2xl` - 1536px

### Container Max Widths

- `--container-sm` - 640px
- `--container-md` - 768px
- `--container-lg` - 1024px
- `--container-xl` - 1280px

## Usage Examples

### Using Color Variables

```css
.button-primary {
    background-color: var(--color-primary-600);
    color: var(--color-background);
}

.button-primary:hover {
    background-color: var(--color-primary-700);
}
```

### Using Spacing

```css
.card {
    padding: var(--spacing-6);
    margin-bottom: var(--spacing-4);
}
```

### Using Typography

```css
.heading-1 {
    font-size: var(--font-size-4xl);
    font-weight: var(--font-weight-bold);
    line-height: var(--line-height-tight);
}
```

### Using Shadows and Radius

```css
.card {
    border-radius: var(--radius-lg);
    box-shadow: var(--shadow-md);
}
```

### Using Transitions

```css
.button {
    transition: all var(--transition-base);
}
```

## Dark Mode Support

The variables include automatic dark mode support via `prefers-color-scheme` media query. The following functional colors are overridden in dark mode:

- `--color-background`
- `--color-surface`
- `--color-text-primary`
- `--color-text-secondary`
- `--color-text-disabled`
- `--color-border`

## Accessibility

All color combinations meet WCAG 2.1 AA contrast requirements:
- Text on background: 4.5:1 minimum
- Large text on background: 3:1 minimum

## Browser Support

CSS custom properties are supported in:
- Chrome/Edge 49+
- Firefox 31+
- Safari 9.1+
- iOS Safari 9.3+
- Chrome for Android 49+

## Testing

Run tests with:
```bash
cd static/css && go test -timeout 30s -v
```

Tests validate:
- All required variables are defined
- Syntax is valid
- Dark mode support is present
- Color contrast meets WCAG AA standards

## Integration

To use these variables in your HTML templates:

```html
<link rel="stylesheet" href="/static/css/variables.css">
```

Then reference variables in your component CSS files.

## Related Stories

- **Epic:** [07_EPIC_frontend.md](../../docs/00_BACKLOG/07_EPIC_frontend.md)
- **Story:** [07_STORY_00_css_variables.md](../../docs/00_BACKLOG/07_STORY_00_css_variables.md)
- **Blocks:** All other frontend stories (01-21)

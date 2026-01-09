# CSS Design System

This directory contains the CSS design system for TinyRSVP, including custom properties (variables), typography system, and color utilities.

## Purpose

Provides a comprehensive design token system using CSS custom properties for consistent theming, typography, colors, and easy customization across the application.

## Files

- [`variables.css`](variables.css) - Core CSS custom properties defining the design system
- [`variables_test.go`](variables_test.go) - Test suite validating CSS variables
- [`typography.css`](typography.css) - Typography system with heading hierarchy and text utilities
- [`typography_test.go`](typography_test.go) - Test suite validating typography styles
- [`typography_integration_test.go`](typography_integration_test.go) - Integration tests for typography system
- [`colors.css`](colors.css) - Color utility classes for backgrounds, text, and borders
- [`colors_test.go`](colors_test.go) - Test suite validating color utilities
- [`colors_integration_test.go`](colors_integration_test.go) - Integration tests for color system
- [`spacing.css`](spacing.css) - Spacing utility classes for margins, padding, and gaps
- [`spacing_test.go`](spacing_test.go) - Test suite validating spacing utilities
- [`spacing_integration_test.go`](spacing_integration_test.go) - Integration tests for spacing system
- [`grid.css`](grid.css) - Grid and flexbox layout utilities with responsive variants
- [`grid_integration_test.go`](grid_integration_test.go) - Integration tests for grid system

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

## Typography System

The typography system provides consistent text styling across the application with a mobile-first, responsive approach.

### Features

- **Heading Hierarchy:** h1-h6 with semantic HTML and utility classes (.h1-.h6)
- **Responsive Scaling:** Headings scale up on tablet+ screens (768px+)
- **Optimal Readability:** Paragraphs limited to 65ch for optimal line length
- **Text Utilities:** Font sizes, weights, colors, and alignment classes
- **Accessibility:** Font smoothing, focus indicators, and WCAG AA compliance
- **Link Styles:** Underlined with hover/focus states
- **Code Styling:** Inline code and code blocks with monospace font

### Typography Utility Classes

```css
/* Font Sizes */
.text-large, .text-small, .text-xs

/* Font Weights */
.text-bold, .text-semibold, .text-medium, .text-normal

/* Text Colors */
.text-primary, .text-secondary, .text-disabled
.text-success, .text-error, .text-warning

/* Text Alignment */
.text-left, .text-center, .text-right
```

### Usage Example

```html
<link rel="stylesheet" href="/static/css/variables.css">
<link rel="stylesheet" href="/static/css/typography.css">

<h1>Main Heading</h1>
<p class="text-large text-secondary">Introduction paragraph</p>
<p>Regular body text with optimal line length.</p>
```

## Color System

The color system provides comprehensive utility classes for backgrounds, text, and borders with semantic naming and WCAG AA compliance.

### Features

- **Primary Scale:** Full 50-900 color scale for brand colors
- **Gray Scale:** Complete neutral palette (50-900)
- **Semantic Colors:** Success, warning, error, info with variants
- **Hover States:** Interactive feedback for buttons and links
- **Light/Dark Variants:** Subtle backgrounds and emphasized text
- **WCAG AA Compliance:** All colors meet 4.5:1 contrast ratio

### Color Utility Classes

```css
/* Background Colors */
.bg-primary, .bg-primary-50 through .bg-primary-900
.bg-gray-50 through .bg-gray-900
.bg-success, .bg-warning, .bg-error, .bg-info
.bg-success-light, .bg-warning-light, .bg-error-light, .bg-info-light
.bg-white, .bg-transparent, .bg-surface, .bg-background

/* Text Colors */
.text-primary-600, .text-primary-700
.text-success, .text-warning, .text-error, .text-info
.text-success-dark, .text-warning-dark, .text-error-dark
.text-gray-600, .text-gray-700, .text-gray-800, .text-gray-900

/* Border Colors */
.border-primary, .border-success, .border-warning, .border-error, .border-info
.border-gray-200, .border-gray-300

/* Hover States */
.bg-primary:hover, .bg-success:hover, .bg-error:hover
```

### Usage Example

```html
<link rel="stylesheet" href="/static/css/variables.css">
<link rel="stylesheet" href="/static/css/typography.css">
<link rel="stylesheet" href="/static/css/colors.css">

<div class="bg-success-light border-success">
    <p class="text-success-dark">Success message</p>
</div>

<button class="bg-primary text-white">Primary Button</button>
```

## Spacing System

The spacing system provides consistent margin, padding, and gap utilities based on an 8px scale for visual rhythm and layout consistency.

### Features

- **8px Base Scale:** Consistent spacing increments (0, 4px, 8px, 12px, 16px, 20px, 24px, 32px, 40px, 48px, 64px, 80px, 96px)
- **Margin Utilities:** All directions (m, mt, mr, mb, ml, mx, my)
- **Padding Utilities:** All directions (p, pt, pr, pb, pl, px, py)
- **Gap Utilities:** For flexbox/grid layouts (gap, gap-x, gap-y)
- **Negative Margins:** For overlapping layouts (-m-1 through -m-12)
- **Auto Margins:** For centering (m-auto, mx-auto, my-auto, etc.)
- **Responsive Variants:** Tablet (md:) and desktop (lg:) breakpoints

### Spacing Utility Classes

```css
/* Margin - All Sides */
.m-0, .m-1, .m-2, .m-3, .m-4, .m-5, .m-6, .m-8, .m-10, .m-12, .m-16, .m-20, .m-24

/* Margin - Directional */
.mt-4  /* margin-top */
.mr-4  /* margin-right */
.mb-4  /* margin-bottom */
.ml-4  /* margin-left */
.mx-4  /* margin-left + margin-right */
.my-4  /* margin-top + margin-bottom */

/* Padding - All Sides */
.p-0, .p-1, .p-2, .p-3, .p-4, .p-5, .p-6, .p-8, .p-10, .p-12, .p-16, .p-20, .p-24

/* Padding - Directional */
.pt-4  /* padding-top */
.pr-4  /* padding-right */
.pb-4  /* padding-bottom */
.pl-4  /* padding-left */
.px-4  /* padding-left + padding-right */
.py-4  /* padding-top + padding-bottom */

/* Gap - Flexbox/Grid */
.gap-4    /* gap (both axes) */
.gap-x-4  /* column-gap */
.gap-y-4  /* row-gap */

/* Negative Margins */
.-m-1, .-m-2, .-m-3, .-m-4, .-m-5, .-m-6, .-m-8, .-m-10, .-m-12

/* Auto Margins */
.m-auto, .mx-auto, .my-auto, .mt-auto, .mr-auto, .mb-auto, .ml-auto

/* Responsive Variants */
.md\:m-4   /* margin at tablet+ */
.md\:p-4   /* padding at tablet+ */
.md\:gap-4 /* gap at tablet+ */
.lg\:m-6   /* margin at desktop+ */
.lg\:p-6   /* padding at desktop+ */
.lg\:gap-6 /* gap at desktop+ */
```

### Usage Examples

```html
<!-- Card with consistent spacing -->
<div class="p-6 mb-4">
    <h2 class="mb-3">Card Title</h2>
    <p class="mb-0">Card content</p>
</div>

<!-- Centered container -->
<div class="mx-auto" style="max-width: 1200px;">
    Content
</div>

<!-- Flexbox layout with gap -->
<div class="flex gap-4">
    <div>Item 1</div>
    <div>Item 2</div>
    <div>Item 3</div>
</div>

<!-- Responsive spacing -->
<div class="p-4 md:p-6 lg:p-8">
    Content with responsive padding
</div>

<!-- Negative margin for overlap -->
<div class="-mt-4">
    Overlapping element
</div>
```

### 8-Point Grid System

The spacing scale follows an 8-point grid system for visual consistency:

- **0-6:** Fine-grained spacing (0, 4px, 8px, 12px, 16px, 20px, 24px)
- **8-12:** Medium spacing (32px, 40px, 48px)
- **16-24:** Large spacing (64px, 80px, 96px)

Use multiples of 8 (0, 8, 16, 24) for major layout spacing and intermediate values (1-6, 10, 12) for fine-tuning.

## Grid System

The grid system provides flexible, responsive layout utilities using CSS Grid and Flexbox for creating modern, mobile-first layouts.

### Features

- **CSS Grid Container:** 12-column grid system with responsive variants
- **Column Spanning:** Control how many columns elements span (1-12)
- **Auto-fit/Auto-fill:** Responsive grid patterns that adapt to content
- **Flexbox Utilities:** Direction, alignment, and justification classes
- **Container Classes:** Centered containers with responsive max-widths
- **Responsive Variants:** Tablet (md:) and desktop (lg:) breakpoints
- **Gap Integration:** Works seamlessly with spacing system gap utilities

### Grid Utility Classes

```css
/* Grid Container */
.grid                    /* display: grid */
.grid-cols-1 to .grid-cols-12  /* 1-12 column layouts */
.grid-auto-fit          /* auto-fit pattern with 250px min */
.grid-auto-fill         /* auto-fill pattern with 250px min */

/* Column Spanning */
.col-span-1 to .col-span-12    /* span 1-12 columns */
.col-span-full          /* span all columns (1 / -1) */

/* Flexbox Container */
.flex                   /* display: flex */
.flex-row              /* flex-direction: row */
.flex-col              /* flex-direction: column */
.flex-wrap             /* flex-wrap: wrap */
.flex-nowrap           /* flex-wrap: nowrap */

/* Flex Item Sizing */
.flex-1                /* flex: 1 1 0% (grow and shrink equally) */
.flex-auto             /* flex: 1 1 auto (grow and shrink based on content) */
.flex-none             /* flex: none (don't grow or shrink) */

/* Alignment (Flexbox/Grid) */
.items-start           /* align-items: flex-start */
.items-center          /* align-items: center */
.items-end             /* align-items: flex-end */
.items-stretch         /* align-items: stretch */
.items-baseline        /* align-items: baseline */

/* Justification (Flexbox/Grid) */
.justify-start         /* justify-content: flex-start */
.justify-center        /* justify-content: center */
.justify-end           /* justify-content: flex-end */
.justify-between       /* justify-content: space-between */
.justify-around        /* justify-content: space-around */
.justify-evenly        /* justify-content: space-evenly */

/* Container */
.container             /* Centered container with responsive max-widths */

/* Responsive Variants (768px+) */
.md\:grid-cols-1 to .md\:grid-cols-12
.md\:col-span-1, .md\:col-span-2, .md\:col-span-3, .md\:col-span-4, .md\:col-span-6, .md\:col-span-12
.md\:flex-row, .md\:flex-col

/* Responsive Variants (1024px+) */
.lg\:grid-cols-1 to .lg\:grid-cols-12
.lg\:col-span-1, .lg\:col-span-2, .lg\:col-span-3, .lg\:col-span-4, .lg\:col-span-6, .lg\:col-span-12
.lg\:flex-row, .lg\:flex-col
```

### Usage Examples

```html
<!-- Basic Grid Layout -->
<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
    <div>Item 1</div>
    <div>Item 2</div>
    <div>Item 3</div>
</div>

<!-- Column Spanning -->
<div class="grid grid-cols-12 gap-4">
    <div class="col-span-12 md:col-span-8">Main content (8 cols on tablet+)</div>
    <div class="col-span-12 md:col-span-4">Sidebar (4 cols on tablet+)</div>
</div>

<!-- Auto-fit Grid (responsive without media queries) -->
<div class="grid grid-auto-fit gap-4">
    <div>Card 1</div>
    <div>Card 2</div>
    <div>Card 3</div>
</div>

<!-- Flexbox Layout -->
<div class="flex flex-col md:flex-row items-center justify-between gap-4">
    <div>Left content</div>
    <div>Right content</div>
</div>

<!-- Centered Container -->
<div class="container p-6">
    <h1>Centered content with responsive max-width</h1>
</div>

<!-- Card Grid with Responsive Columns -->
<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
    <div class="p-4">Card 1</div>
    <div class="p-4">Card 2</div>
    <div class="p-4">Card 3</div>
    <div class="p-4">Card 4</div>
</div>

<!-- Flex Item Sizing -->
<div class="flex gap-4">
    <div class="flex-none">Fixed width</div>
    <div class="flex-1">Grows to fill space</div>
    <div class="flex-auto">Grows based on content</div>
</div>
```

### Responsive Breakpoints

The grid system uses the same breakpoints as the spacing system:

- **Mobile:** Base styles (320px-767px)
- **Tablet (md:):** 768px and up
- **Desktop (lg:):** 1024px and up

### Container Max-Widths

The `.container` class automatically adjusts its max-width at different breakpoints:

- **Mobile:** 100% width with padding
- **Tablet (768px+):** max-width: 768px (--container-md)
- **Desktop (1024px+):** max-width: 1024px (--container-lg)

### Auto-fit vs Auto-fill

Both create responsive grids without media queries, but with different behaviors:

- **`.grid-auto-fit`:** Columns expand to fill available space (fewer, wider columns)
- **`.grid-auto-fill`:** Creates as many columns as fit, even if empty (more, narrower columns)

Both use `minmax(250px, 1fr)` for column sizing.

### Integration with Spacing System

The grid system works seamlessly with spacing utilities:

```html
<!-- Grid with gap spacing -->
<div class="grid grid-cols-3 gap-4 md:gap-6 lg:gap-8">
    <div class="p-4">Item with padding</div>
    <div class="p-4">Item with padding</div>
    <div class="p-4">Item with padding</div>
</div>

<!-- Flex with gap spacing -->
<div class="flex gap-3 items-center">
    <button class="px-4 py-2">Button 1</button>
    <button class="px-4 py-2">Button 2</button>
</div>
```

### Mobile-First Approach

Always start with mobile styles and progressively enhance for larger screens:

```html
<!-- Mobile: single column, Tablet: 2 columns, Desktop: 3 columns -->
<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
    <div>Content</div>
</div>

<!-- Mobile: column layout, Tablet+: row layout -->
<div class="flex flex-col md:flex-row gap-4">
    <div>Content</div>
</div>
```

## Integration

To use the design system in your HTML templates:

```html
<link rel="stylesheet" href="/static/css/variables.css">
<link rel="stylesheet" href="/static/css/typography.css">
<link rel="stylesheet" href="/static/css/colors.css">
<link rel="stylesheet" href="/static/css/spacing.css">
<link rel="stylesheet" href="/static/css/grid.css">
```

Load order matters: variables → typography → colors → spacing → grid for proper variable resolution.

## Related Stories

- **Epic:** [07_EPIC_frontend.md](../../docs/00_BACKLOG/07_EPIC_frontend.md)
- **Story 00:** [07_STORY_00_css_variables.md](../../docs/00_BACKLOG/07_STORY_00_css_variables.md)
- **Story 01:** [07_STORY_01_typography.md](../../docs/00_BACKLOG/07_STORY_01_typography.md)
- **Story 02:** [07_STORY_02_color_system.md](../../docs/00_BACKLOG/07_STORY_02_color_system.md)
- **Story 03:** [07_STORY_03_spacing_system.md](../../docs/00_BACKLOG/07_STORY_03_spacing_system.md)
- **Story 04:** [07_STORY_04_responsive_grid.md](../../docs/00_BACKLOG/07_STORY_04_responsive_grid.md)
- **Blocks:** All other frontend stories (05-21)

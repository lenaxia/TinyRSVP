# Layout System Guide

## Overview

This guide explains the consistent layout system used across all pages in TinyRSVP to ensure uniform spacing and alignment.

## Page Container Pattern

All main content pages should follow this pattern for consistent spacing:

### CSS Structure

```css
.page-main-container {
    max-width: 1200px;
    margin: 0 auto;
    width: 100%;
    padding: var(--spacing-4);
}

@media (min-width: 768px) {
    .page-main-container {
        padding: var(--spacing-6) var(--spacing-6);
    }
}

@media (min-width: 1024px) {
    .page-main-container {
        padding: var(--spacing-6) var(--spacing-6);
    }
}
```

### Key Properties

- **max-width: 1200px** - Ensures content doesn't stretch too wide on large screens
- **margin: 0 auto** - Centers the content horizontally
- **width: 100%** - Allows content to fill available space up to max-width
- **Responsive padding**:
  - Mobile (< 768px): `var(--spacing-4)` (16px)
  - Tablet/Desktop (≥ 768px): `var(--spacing-6)` (24px)

## Navigation Consistency

The navigation bar (`app_navigation.css`) uses the same max-width and padding pattern:

```css
.app-nav-header {
    max-width: 1200px;
    margin: 0 auto;
    padding: var(--spacing-4);
}

@media (min-width: 768px) {
    .app-nav-header {
        padding: var(--spacing-4) var(--spacing-6);
    }
}
```

## Existing Page Implementations

### Dashboard (`dashboard.css`)
```css
.dashboard-main {
    flex: 1;
    padding: var(--spacing-4);
    max-width: 1200px;
    margin: 0 auto;
    width: 100%;
    overflow-y: auto;
}
```

### Event List (`event_list.css`)
```css
.event-list {
    display: flex;
    flex-direction: column;
    gap: var(--spacing-6);
    padding: var(--spacing-4);
    max-width: 1200px;
    margin: 0 auto;
    width: 100%;
}
```

## Creating New Pages

When creating a new page, follow these steps:

### 1. HTML Structure

```html
<main class="your-page-name" role="main" aria-label="Main content" id="main-content">
    <!-- Your page content here -->
</main>
```

### 2. CSS for Main Container

```css
.your-page-name {
    padding: var(--spacing-4);
    max-width: 1200px;
    margin: 0 auto;
    width: 100%;
    /* Add any additional properties specific to your page */
}

@media (min-width: 768px) {
    .your-page-name {
        padding: var(--spacing-6) var(--spacing-6);
    }
}

@media (min-width: 1024px) {
    .your-page-name {
        padding: var(--spacing-6) var(--spacing-6);
    }
}
```

### 3. Include Necessary CSS Files

Always include `layout.css` (if using the shared class) or implement the pattern directly:

```html
<link rel="stylesheet" href="/static/css/variables.css">
<link rel="stylesheet" href="/static/css/typography.css">
<link rel="stylesheet" href="/static/css/colors.css">
<link rel="stylesheet" href="/static/css/spacing.css">
<link rel="stylesheet" href="/static/css/grid.css">
<link rel="stylesheet" href="/static/css/buttons.css">
<link rel="stylesheet" href="/static/css/layout.css">  <!-- Optional shared layout -->
<link rel="stylesheet" href="/static/css/app_navigation.css">
<link rel="stylesheet" href="/static/css/your-page.css">
```

## Benefits

1. **Consistency**: All pages have the same spacing and alignment
2. **Responsive**: Automatically adjusts padding for mobile, tablet, and desktop
3. **Maintainability**: Changes to spacing can be made in one place
4. **Alignment**: Content aligns perfectly with the navigation bar

## Common Mistakes to Avoid

❌ **Don't** use different max-width values per page
❌ **Don't** use fixed pixel padding instead of CSS variables
❌ **Don't** forget to add responsive breakpoints
❌ **Don't** nest multiple containers with max-width

✅ **Do** use the standard pattern for all new pages
✅ **Do** use CSS spacing variables (`var(--spacing-4)`, etc.)
✅ **Do** test on mobile, tablet, and desktop viewports
✅ **Do** align with the navigation bar's max-width

## Testing Checklist

When creating or modifying a page, verify:

- [ ] Content aligns with navigation bar on desktop
- [ ] Padding is consistent across breakpoints
- [ ] Content doesn't touch screen edges on mobile
- [ ] Max-width prevents content from being too wide
- [ ] Page looks consistent with other pages (/, /events, /admin)

## Questions?

If you're unsure about implementing the layout pattern, refer to existing pages like `dashboard.html` and `event_list.html` as examples.

# Template System Documentation

## Overview

TinyRSVP uses Go's `html/template` system with a base template and reusable components to ensure consistency across all pages.

## Base Template System

### base.html

The base template provides the foundational HTML structure that all pages extend.

**Usage:**
```go
{{template "base" .}}

{{define "title"}}Your Page Title - TinyRSVP{{end}}

{{define "main-class"}}dashboard-main{{end}}

{{define "content"}}
    <!-- Your page content here -->
{{end}}
```

**Available Blocks:**
- `title` - Page title (default: "TinyRSVP")
- `main-class` - CSS class for main container (default: "dashboard-main")
- `css-extra` - Additional CSS files specific to the page
- `js-extra` - Additional JavaScript files specific to the page
- `content` - Main page content (required)

**Automatically Included:**
- All common CSS files (base, variables, typography, colors, spacing, grid, buttons, navigation, dashboard, mobile, loading, error, keyboard, focus, theme)
- All common JavaScript files (navigation_toggle, loading_states, keyboard_navigation, screen_reader, focus_management, theme_controller)
- Navigation component

### Example: Simple Page

```html
{{template "base" .}}

{{define "title"}}My Page - TinyRSVP{{end}}

{{define "content"}}
    <header class="dashboard-header">
        <h1>My Page</h1>
        <a href="/action" class="btn btn-primary">Action</a>
    </header>

    <section class="stats-grid">
        <!-- Content here -->
    </section>
{{end}}
```

### Example: Page with Extra CSS/JS

```html
{{template "base" .}}

{{define "title"}}Events - TinyRSVP{{end}}

{{define "main-class"}}event-list{{end}}

{{define "css-extra"}}
    <link rel="stylesheet" href="/static/css/forms.css">
    <link rel="stylesheet" href="/static/css/event_list.css">
{{end}}

{{define "content"}}
    <!-- Page content -->
{{end}}
```

## Reusable Components

### components.html

Provides reusable UI components that can be included in any page.

#### stats-card

Displays a statistic with title, value, and optional subtitle.

**Usage:**
```html
{{template "stats-card" (dict "Title" "Total Events" "Value" .Stats.TotalEvents "Subtitle" "All events")}}
```

**Parameters:**
- `Title` (string, required) - Card title
- `Value` (any, required) - Main value to display
- `Subtitle` (string, optional) - Additional context

#### empty-state

Displays an empty state with icon, title, description, and optional action button.

**Usage:**
```html
{{template "empty-state" (dict 
    "Icon" "📅" 
    "Title" "No Events Found" 
    "Description" "Get started by creating your first event."
    "ActionURL" "/events/new"
    "ActionText" "Create Event"
)}}
```

**Parameters:**
- `Icon` (string, required) - Emoji or icon
- `Title` (string, required) - Empty state title
- `Description` (string, required) - Explanation text
- `ActionURL` (string, optional) - URL for action button
- `ActionText` (string, optional) - Button text

#### loading-state

Displays a loading spinner with message.

**Usage:**
```html
{{template "loading-state" (dict "Message" "Loading dashboard...")}}
```

**Parameters:**
- `Message` (string, required) - Loading message

#### error-state

Displays an error message with optional action button.

**Usage:**
```html
{{template "error-state" (dict 
    "Title" "Error Loading Dashboard" 
    "Description" .Error
    "OnClick" "window.location.reload()"
    "ActionText" "Retry"
)}}
```

**Parameters:**
- `Title` (string, required) - Error title
- `Description` (string, required) - Error description
- `ActionURL` (string, optional) - URL for action button
- `OnClick` (string, optional) - JavaScript onclick handler
- `ActionText` (string, optional) - Button text

### page_header.html

Provides a standardized page header with title and optional actions.

**Usage:**
```html
{{template "page-header" (dict "Title" "Dashboard" "Actions" (html `<a href="/events/new" class="btn btn-primary">Create Event</a>`))}}
```

**Parameters:**
- `Title` (string, required) - Page title
- `Actions` (HTML, optional) - Action buttons or other header content

## Standard Page Structure

All pages should follow this structure:

```html
{{template "base" .}}

{{define "title"}}Page Name - TinyRSVP{{end}}

{{define "main-class"}}dashboard-main{{end}}  <!-- or event-list, etc. -->

{{define "css-extra"}}
    <!-- Page-specific CSS if needed -->
{{end}}

{{define "content"}}
    <header class="dashboard-header">
        <h1>Page Title</h1>
        <!-- Optional actions -->
    </header>

    {{if .Error}}
        <!-- Error state -->
    {{else if .Loading}}
        <!-- Loading state -->
    {{else}}
        <!-- Main content -->
    {{end}}
{{end}}

{{define "js-extra"}}
    <!-- Page-specific JavaScript if needed -->
{{end}}
```

## Benefits

1. **DRY Principle**: Common HTML structure defined once
2. **Consistency**: All pages automatically get the same CSS/JS includes
3. **Maintainability**: Update base template to affect all pages
4. **Type Safety**: Go templates provide compile-time checking
5. **Reusability**: Components can be used across multiple pages
6. **Flexibility**: Pages can override or extend base behavior

## Migration Guide

### Converting Existing Pages

1. **Remove boilerplate HTML** (<!DOCTYPE>, <html>, <head>, <body> tags)
2. **Add base template call** at the top: `{{template "base" .}}`
3. **Define required blocks**: `title`, `content`
4. **Move page-specific CSS** to `css-extra` block
5. **Move page-specific JS** to `js-extra` block
6. **Use standardized components** where applicable

### Before (Old Style):
```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Dashboard - TinyRSVP</title>
    <link rel="stylesheet" href="/static/css/base.css">
    <link rel="stylesheet" href="/static/css/variables.css">
    <!-- ... 15 more CSS files ... -->
</head>
<body>
    {{template "navigation" .}}
    <main class="dashboard-main">
        <header class="dashboard-header">
            <h1>Dashboard</h1>
        </header>
        <!-- content -->
    </main>
    <script src="/static/js/navigation_toggle.js"></script>
    <!-- ... 5 more JS files ... -->
</body>
</html>
```

### After (New Style):
```html
{{template "base" .}}

{{define "title"}}Dashboard - TinyRSVP{{end}}

{{define "content"}}
    <header class="dashboard-header">
        <h1>Dashboard</h1>
    </header>
    <!-- content -->
{{end}}
```

## Component Library

The following components are available:

| Component | File | Purpose |
|-----------|------|---------|
| `base` | base.html | Base page structure |
| `navigation` | navigation.html | Top navigation bar |
| `page-header` | page_header.html | Page title with actions |
| `stats-card` | components.html | Statistics display card |
| `empty-state` | components.html | Empty state message |
| `loading-state` | components.html | Loading spinner |
| `error-state` | components.html | Error message |
| `modal-center` | modal_center.html | Centered modal |
| `modal-slide` | modal_slide.html | Slide-in modal |

## Future Enhancements

Consider creating additional reusable components:
- `card` - Generic card container
- `table` - Data table with sorting/filtering
- `list-item` - List item with icon and actions
- `form-field` - Form input with label and validation
- `dropdown` - Dropdown menu component
- `tabs` - Tab navigation component
- `pagination` - Pagination controls
- `breadcrumbs` - Breadcrumb navigation

## Questions?

Refer to existing pages that use the base template system, or consult the LAYOUT_GUIDE.md for CSS patterns.

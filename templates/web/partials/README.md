# Template System Documentation

## Overview

TinyRSVP uses Go's `html/template` with a base template + reusable partials.
Every page **must** compose from the partials rather than reinvent the same
HTML. See [PATTERNS.md](../PATTERNS.md) for a cookbook mapping common design
needs to the partial + CSS class you should reach for.

## Base Template System

### base.html

Provides the foundational HTML structure that all pages extend.

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
- `title` — Page title (default: "TinyRSVP")
- `main-class` — CSS class on `<main>` (default: "dashboard-main")
- `css-extra` — Page-specific CSS imports
- `js-extra` — Page-specific JS imports
- `content` — Main page content (required)

**Automatically Included:**
- All common CSS files, **including `components.css`** which backs every
  partial in `components.html`
- Common JavaScript for nav, loading states, keyboard, screen reader,
  focus, theme
- The `<nav>` navigation

## Reusable Components (`components.html`)

**Every server-side template loader (`cmd/server/main.go`,
`tests/uxserver/server.go`) parses `partials/components.html` alongside
`partials/base.html`.** You don't need to add it per-page — it's ambient.

### stats-card

Statistic display. Can optionally be a link (drilldown) and have a colored
accent border.

**Usage:**
```html
{{template "stats-card" dict
    "Title"    "Total Events"
    "Value"    .Stats.TotalEvents
    "Subtitle" "2 draft, 5 published"
    "Href"     "/events"
    "Accent"   "success"
}}
```

**Parameters:**
- `Title` (string, required)
- `Value` (any, required)
- `Subtitle` (string, optional)
- `Href` (string, optional) — renders as `<a>` when set, else `<div>`
- `Accent` (string, optional) — one of `primary`, `success`, `warning`, `error`
- `Icon` (string, optional) — small emoji/icon

### metric-tile

Denser than stats-card. Compact numeric-first tile for the admin dashboard
"at a glance" strip and the metrics business-summary section.

**Usage:**
```html
{{template "metric-tile" dict
    "Label" "Total Users"
    "Value" 42
    "Sub"   "Active this month"
    "Href"  "/admin/users"
    "Accent" "primary"
}}
```

### section

A titled block with optional description. Prefer this over hand-rolled
`<h2 class="section-title">` markup.

```html
{{template "section" dict
    "Title"       "Server"
    "Description" "Runtime settings loaded at startup"
    "Body"        (safeHTML `<dl class="definition-list">…</dl>`)
}}
```

Or inline (more common):
```html
<section class="ui-section">
    <header class="ui-section-header">
        <h2 class="ui-section-title">Server</h2>
    </header>
    <div class="ui-section-body">
        ...
    </div>
</section>
```

### action-card

Big-tap-target navigation card. Used for the admin dashboard "Quick Actions"
grid.

```html
{{template "action-card" dict
    "Href"        "/admin/users"
    "Icon"        "👥"
    "Title"       "Manage Users"
    "Description" "View and change user roles"
}}
```

Put multiple inside a `<div class="action-grid">…</div>`.

### status-badge

Inline pill for status/state.

```html
{{template "status-badge" dict "Variant" "success" "Label" "Healthy"}}
```

**Variants:** `success`, `warning`, `error`, `info`, `neutral` (default).

### definition-list

`<dl>` with tabular labels for read-only settings / metrics grids.

```html
{{template "definition-list" dict
    "Rows" (list
        (dict "Label" "Host" "Value" "localhost")
        (dict "Label" "Port" "Value" 8080)
    )
}}
```

Or inline `<dl class="definition-list">` if you want template control over the
rows.

### empty-state / loading-state / error-state

Unchanged from earlier revisions — see components.html for signatures.

### page-header

Standardized `<header>` with title + optional actions HTML.

```html
{{template "page-header" dict
    "Title"   "Events"
    "Actions" (safeHTML `<a href="/events/new" class="btn btn-primary">Create</a>`)
}}
```

## CSS Component Backing

Every partial has a matching rule set in `static/css/components.css`. Rules:

- **Design tokens only.** No hardcoded hex/rgb. Guarded by
  `TestComponentsNoHardcodedColors`.
- **Dark-mode-safe by construction.** The tokens (via `variables.css`)
  handle theme switching, so components inherit correct behavior.

## Component Library

| Component | File | Purpose |
|-----------|------|---------|
| `base` | `base.html` | Base page structure |
| `navigation` | `navigation.html` | Top navigation bar |
| `page-header` | `page_header.html` | Page title with actions |
| `stats-card` | `components.html` | Statistics display card (drilldown-capable) |
| `metric-tile` | `components.html` | Compact numeric tile |
| `section` | `components.html` | Titled block with body |
| `action-card` | `components.html` | Nav card with icon + title + description |
| `status-badge` | `components.html` | Inline status pill |
| `definition-list` | `components.html` | `<dl>` grid for read-only key/value pairs |
| `empty-state` | `components.html` | Empty state message |
| `loading-state` | `components.html` | Loading spinner |
| `error-state` | `components.html` | Error message |
| `modal-center` | `modal_center.html` | Centered modal |

## Adding a New Partial

1. Write the test first in `templates/web/components_partials_test.go`.
2. Verify RED with `go test -run TestPartialName ./templates/web/`.
3. Add the `{{define "your-partial"}}` block to `components.html`.
4. If the partial introduces new CSS classes, write CSS tests in
   `static/css/components_test.go` FIRST, then add the rule to
   `static/css/components.css`.
5. Update the table above and (if the pattern is common) [PATTERNS.md](../PATTERNS.md).

## Migration Checklist

Existing pages should be migrated whenever they're touched substantively:

1. Drop the boilerplate `<!DOCTYPE>` / `<html>` / `<head>` / `<body>`.
2. Add `{{template "base" .}}` at the top.
3. Define required blocks: `title`, `content`.
4. Replace inline error/loading/empty markup with the `error-state` /
   `loading-state` / `empty-state` partials.
5. Replace hand-rolled `<h2 class="…-section-title">` + wrapper with
   `<section class="ui-section">` from components.css.
6. Replace hand-rolled `<dl>` grids with `.definition-list`.
7. Replace ad-hoc status pills with `status-badge`.

Not every page has to be migrated in one commit. Do it opportunistically.

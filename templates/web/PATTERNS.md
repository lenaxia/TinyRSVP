# TinyRSVP UI Patterns Cookbook

When adding or modifying a page, look here first. If the pattern you need
already exists, use the referenced partial + CSS class. **Do not reinvent.**

If a needed pattern is missing, add it via TDD (test first, then
partial + CSS), document it here, and open a follow-up to migrate other
pages that hand-rolled the same thing.

---

## "I want a big number on the page"

**Use:** the `metric-tile` partial, wrapped in `.metric-tile-grid`.

Auto-fit responsive grid, no orphans on any card count.

```html
<div class="metric-tile-grid">
    {{template "metric-tile" dict "Label" "Total Users" "Value" 42 "Accent" "primary"}}
    {{template "metric-tile" dict "Label" "Events"      "Value" 7  "Accent" "success"}}
    {{template "metric-tile" dict "Label" "Invites"     "Value" 55 "Accent" "warning"}}
</div>
```

Accents: `primary`, `success`, `warning`, `error`.

## "The number should link somewhere (drilldown)"

Same partial, add `Href`.

```html
{{template "metric-tile" dict "Label" "Users" "Value" 42 "Href" "/admin/users" "Accent" "primary"}}
```

Renders as `<a>` with hover + focus states already styled.

## "I need a stat with a title, value, and a subtitle explaining the value"

**Use:** the `stats-card` partial, wrapped in `.stats-grid`.

```html
<section class="stats-grid">
    {{template "stats-card" dict
        "Title"    "Response Rate"
        "Value"    "78%"
        "Subtitle" "45 of 58 responded"
        "Href"     "/rsvps"
    }}
</section>
```

Use `metric-tile` when Label + Value is enough. Use `stats-card` when you
also need a title-cased heading + prose subtitle.

## "I need a titled section with a body"

**Use:** the `.ui-section` block, either inline or via the `section` partial.

**Inline** (most common — you keep template control over the body):

```html
<section class="ui-section">
    <header class="ui-section-header">
        <h2 class="ui-section-title">Database Pool</h2>
        <p class="ui-section-description">Live connection state.</p>
    </header>
    <div class="ui-section-body">
        <!-- your content -->
    </div>
</section>
```

**Partial** (only when the body is a simple pre-rendered HTML string):

```html
{{template "section" dict "Title" "Foo" "Body" (safeHTML "<p>...</p>")}}
```

## "I need a labeled key/value grid (settings, metrics, etc.)"

**Use:** `<dl class="definition-list">`. Do NOT reinvent `.settings-grid` or
`.metrics-detail-grid`.

```html
<dl class="definition-list">
    <dt>Host</dt><dd>localhost</dd>
    <dt>Port</dt><dd>8080</dd>
</dl>
```

On mobile it renders single-column; on tablet+ it becomes a two-column grid.

## "I need a status pill (Healthy / Failed / Pending)"

**Use:** the `status-badge` partial. Do NOT hand-roll `.status-healthy` or
`.badge-*` classes.

```html
{{template "status-badge" dict "Variant" "success" "Label" "Healthy"}}
```

Variants: `success`, `warning`, `error`, `info`, `neutral`.

## "I need a nav card (icon + title + description → link)"

**Use:** the `action-card` partial, wrapped in `.action-grid`.

```html
<div class="action-grid">
    {{template "action-card" dict
        "Href"        "/admin/users"
        "Icon"        "👥"
        "Title"       "Manage Users"
        "Description" "View and change user roles"
    }}
    {{template "action-card" dict
        "Href"        "/admin/settings"
        "Icon"        "⚙"
        "Title"       "System Settings"
        "Description" "Runtime configuration"
    }}
</div>
```

## "I need a data table with rows and actions"

**Use:** `<table class="data-table">`. See `user_management.html` for a
reference implementation.

```html
<table class="data-table">
    <thead>
        <tr><th>Name</th><th>Email</th><th>Role</th></tr>
    </thead>
    <tbody>
        {{range .Users}}
        <tr><td>{{.Name}}</td><td>{{.Email}}</td><td>{{.Role}}</td></tr>
        {{end}}
    </tbody>
</table>
```

## "I need a panel with a header, body, and footer link"

**Use:** the `.panel` classes. Common for system-health cards where you want
"see more →" at the bottom.

```html
<div class="panel">
    <div class="panel-header">
        <h3 class="panel-header-title">Database</h3>
        {{template "status-badge" dict "Variant" "success" "Label" "Active"}}
    </div>
    <div class="panel-body">
        <dl class="definition-list">
            <dt>Open</dt><dd>5 / 25</dd>
        </dl>
    </div>
    <div class="panel-footer">
        <a href="/admin/metrics" class="panel-footer-link">View details →</a>
    </div>
</div>
```

## "I need an empty state"

```html
{{template "empty-state" dict
    "Icon"        "📅"
    "Title"       "No Events Yet"
    "Description" "Create your first event to get started."
    "ActionURL"   "/events/new"
    "ActionText"  "Create Event"
}}
```

## "I need a loading state"

```html
{{template "loading-state" dict "Message" "Loading dashboard..."}}
```

## "I need an error state with a retry button"

```html
{{template "error-state" dict
    "Title"       "Error Loading Data"
    "Description" .Error
    "OnClick"     "window.location.reload()"
    "ActionText"  "Retry"
}}
```

Or use `ActionURL` for a link instead of an onclick.

## "I need a page-level header with title + action button"

**Use:** the `page-header` partial. The action HTML is escaped by default —
pass through `safeHTML` when you intentionally want raw HTML (e.g. a button).

```html
{{template "page-header" dict
    "Title"   "Events"
    "Actions" (safeHTML `<a href="/events/new" class="btn btn-primary">Create Event</a>`)
}}
```

## "I need to show the current user in the header"

**Use:** the `header-user` classes:

```html
<span class="header-user">
    {{.User.Name}}
    <span class="header-user-role">({{.User.Role}})</span>
</span>
```

Pair with `page-header`'s `Actions` slot.

---

## What NOT to do

**Do NOT:**
- Hardcode hex or rgb colors in any CSS file. Guarded by
  `TestDashboardNoHardcodedColors`, `TestComponentsNoHardcodedColors`, and
  the two new admin-page tests.
- Duplicate `.stats-card` / `.empty-state` markup — use the partials.
- Invent a new "section wrapper" class per page. Use `.ui-section`.
- Invent new `.badge-*` variants. Extend `status-badge` if a new semantic
  variant is needed.
- Load a page-specific CSS file when the design system covers the need.

## Where to add a new pattern

1. Write a partial-render test in
   `templates/web/components_partials_test.go`.
2. If new CSS classes are involved, write class-existence tests in
   `static/css/components_test.go`.
3. Add the `{{define "…"}}` block to `templates/web/partials/components.html`.
4. Add the rules to `static/css/components.css`.
5. Document here.

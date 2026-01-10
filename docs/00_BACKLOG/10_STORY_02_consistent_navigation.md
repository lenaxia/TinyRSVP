# STORY: Consistent Navigation Across All Pages

**Epic:** 10 - Technical Debt & Improvements
**Story ID:** 10_STORY_02
**Priority:** High
**Estimated Effort:** 3 hours

## User Story

As a user, I want consistent navigation across all pages so that I can easily move between different sections of the application regardless of which page I'm on or what device I'm using.

## Current Issues

1. `/dashboard` has sidebar navigation with "TinyRSVP" logo
2. `/events` has top navigation bar (Dashboard, Events, Admin)
3. Other pages have inconsistent or missing navigation
4. On mobile, sidebar navigation disappears with no alternative
5. No consistent way to return to dashboard or access admin

## Acceptance Criteria

- [ ] All authenticated pages have consistent navigation
- [ ] Navigation includes: Dashboard, Events, Admin links
- [ ] "TinyRSVP" logo/link appears on all pages
- [ ] Mobile navigation works (hamburger menu or persistent nav)
- [ ] Active page is visually indicated
- [ ] Navigation is accessible (keyboard, screen reader)
- [ ] Navigation doesn't interfere with page content

## Technical Approach

### Option A: Unified Top Navigation Bar
- Add top nav to all pages
- Responsive hamburger menu for mobile
- Consistent across dashboard, events, admin pages

### Option B: Persistent Sidebar
- Make sidebar visible on all pages
- Collapse to icons on mobile
- Expand on hover/click

### Option C: Hybrid Approach (Recommended)
- Top bar with logo + hamburger on mobile
- Sidebar on desktop (768px+)
- Consistent component across all pages

## Implementation

### 1. Create Navigation Component
```html
<!-- templates/web/partials/navigation.html -->
<nav class="app-nav">
    <div class="app-nav-brand">
        <a href="/" class="app-logo">TinyRSVP</a>
        <button class="app-nav-toggle" aria-label="Toggle navigation">☰</button>
    </div>
    <ul class="app-nav-menu">
        <li><a href="/" class="app-nav-link {{if .ActivePage eq "dashboard"}}active{{end}}">Dashboard</a></li>
        <li><a href="/events" class="app-nav-link {{if .ActivePage eq "events"}}active{{end}}">Events</a></li>
        <li><a href="/admin" class="app-nav-link {{if .ActivePage eq "admin"}}active{{end}}">Admin</a></li>
    </ul>
</nav>
```

### 2. Add CSS
- `static/css/app_navigation.css`
- Mobile-first responsive design
- Hamburger menu for mobile
- Sidebar or top bar for desktop

### 3. Update All Templates
- dashboard.html
- event_list.html
- event_form.html
- event_detail.html
- invite_list.html
- admin_dashboard.html
- user_management.html

### 4. Add JavaScript
- `static/js/navigation_toggle.js`
- Handle hamburger menu toggle
- Persist menu state in localStorage

## Tasks

- [ ] Design navigation component (mobile + desktop)
- [ ] Create navigation partial template
- [ ] Create app_navigation.css
- [ ] Create navigation_toggle.js
- [ ] Update all page templates to include navigation
- [ ] Add ActivePage to all page data structs
- [ ] Test on mobile (320px, 768px)
- [ ] Test on desktop (1024px+)
- [ ] Test keyboard navigation
- [ ] Test screen reader compatibility

## Dependencies

None

## Notes

- Consider adding user menu (profile, logout) in future
- May want breadcrumbs for deep pages
- Logo should link to dashboard
- Active page should be visually distinct

## Status

- **Status:** Not Started
- **Assigned:** Unassigned
- **Started:** N/A
- **Completed:** N/A

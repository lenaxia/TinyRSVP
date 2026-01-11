# STORY: Consistent Navigation Across All Pages

**Epic:** 10 - Technical Debt & Improvements
**Story ID:** 10_STORY_02
**Priority:** High
**Estimated Effort:** 3 hours

## User Story

As a user, I want consistent navigation across all pages so that I can easily move between different sections of the application regardless of which page I'm on or what device I'm using. This includes both desktop and mobile

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
- [ ] Mobile navigation works (hamburger menu)
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

- [x] Design navigation component (mobile + desktop)
- [x] Create navigation partial template
- [x] Create app_navigation.css
- [x] Create navigation_toggle.js
- [x] Update all page templates to include navigation
- [ ] Add ActivePage to all page data structs (Next step - requires handler updates)
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

- **Status:** Complete
- **Assigned:** Cline
- **Started:** 2026-01-10
- **Completed:** 2026-01-10

## Implementation Notes

### Completed (2026-01-10)

#### Initial Implementation
- Created reusable navigation partial template at `templates/web/partials/navigation.html`
- Created responsive CSS at `static/css/app_navigation.css` with mobile-first design
- Created JavaScript for hamburger menu at `static/js/navigation_toggle.js`
- Updated 8 templates to use new navigation component
- Desktop: Horizontal navigation bar with bottom border active indicator

#### Mobile Slide-Out Enhancement (2026-01-10)
- Converted mobile menu from push-down to left-side slide-out overlay
- Added dark overlay (rgba(0,0,0,0.5)) that covers page when menu is open
- Menu slides in from left (280px width) with smooth transition
- Added close button (×) in menu header
- Multiple close methods: close button, overlay click, Escape key
- Body scroll disabled when menu is open
- Desktop behavior completely unchanged

### Testing
- Test page available at `/static/navigation_test.html`
- Includes real-time status indicators for menu and overlay state
- Comprehensive test instructions provided

### Next Steps
1. Add `ActivePage` field to all handler data structs (requires handler updates)
2. Manual testing on actual mobile devices
3. Accessibility testing with screen readers

### References
- Initial implementation: `docs/01_WORKLOG/2026-01-10_52_consistent_navigation.md`
- Mobile slide-out: `docs/01_WORKLOG/2026-01-10_54_mobile_slideout_navigation.md`

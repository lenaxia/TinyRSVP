# Story 10.13: Admin Template Theme Integration

**Epic:** 10 - Technical Debt & Improvements  
**Priority:** Medium  
**Status:** Not Started  
**Identified:** 2026-01-11 (Epic 11 Phase 1 Validation)

---

## User Story

As an **admin user**, I want **light/dark theme switching on all admin pages** so that **I have a consistent experience across the entire application**.

---

## Problem Statement

During Epic 11 Phase 1 validation, integration tests revealed that admin-facing templates do not include theme support (variables.css, theme_toggle.css, theme_controller.js). This creates an inconsistent experience where:

- Guest-facing RSVP pages have full theme support ✅
- Admin pages lack theme toggle functionality ❌
- Admins cannot switch between light/dark modes
- Admin pages don't benefit from CSS variable system

---

## Acceptance Criteria

- [ ] All admin templates include `variables.css`
- [ ] All admin templates include `theme_toggle.css`
- [ ] All admin templates include `theme_controller.js`
- [ ] Navigation partial with theme toggle button included
- [ ] Theme preference persists across admin pages
- [ ] Light/dark mode works on all admin pages
- [ ] Integration tests pass for theme support

---

## Affected Templates

1. `templates/web/admin_dashboard.html`
2. `templates/web/dashboard.html`
3. `templates/web/event_detail.html`
4. `templates/web/event_form.html`
5. `templates/web/event_list.html`
6. `templates/web/invite_list.html`
7. `templates/web/rsvp_summary.html`
8. `templates/web/user_management.html`

---

## Technical Approach

### 1. Update Base Template

Ensure base template includes:
```html
<link rel="stylesheet" href="/static/css/variables.css">
<link rel="stylesheet" href="/static/css/theme_toggle.css">
<script src="/static/js/theme_controller.js"></script>
```

### 2. Verify Navigation Partial

Ensure all templates include navigation partial with theme toggle button.

### 3. Update Integration Tests

Update `TestThemeIntegration` to pass for all admin templates.

---

## Tasks

- [ ] Review base template structure
- [ ] Add CSS includes to base template or individual templates
- [ ] Add JavaScript include to base template or individual templates
- [ ] Verify navigation partial included in all templates
- [ ] Test theme toggle on each admin page
- [ ] Update integration tests
- [ ] Verify theme preference persists
- [ ] Test light/dark mode on all pages

---

## Testing Requirements

### Unit Tests
- [ ] Test base template includes correct CSS/JS
- [ ] Test navigation partial included

### Integration Tests
- [ ] Test theme toggle works on each admin page
- [ ] Test theme preference persists across pages
- [ ] Test light/dark mode renders correctly

### Manual Testing
- [ ] Navigate through all admin pages
- [ ] Toggle theme on each page
- [ ] Verify consistent appearance
- [ ] Check mobile responsiveness

---

## Dependencies

**Prerequisites:**
- ✅ Story 10.12: Light/Dark Theme Switching (complete)
- ✅ Epic 11 Phase 1: Theme system implemented

**Blocks:**
- None (improvement, not blocker)

---

## Notes

- This is a quality-of-life improvement, not a critical bug
- Guest-facing pages already have full theme support
- Admin pages still functional without theme toggle
- Can be completed independently of other stories

---

## Estimated Effort

**Size:** Small (1-2 hours)
- Simple CSS/JS includes
- Template updates
- Test updates

---

**Status:** Ready for Implementation

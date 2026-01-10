# UI Audit and Fixes - 2026-01-10

## Issues Identified

### 1. Missing Route: `/events/{eventId}/invites/new` (404 Error)
**Problem:** The invite list template has "Create Invite" buttons that link to `/events/{eventId}/invites/new`, but this route doesn't exist in the router.

**Root Cause:** No handler or route registered for creating individual invites via UI form.

**Solution Options:**
- **Option A:** Remove the "Create Invite" buttons and rely on API-only invite creation
- **Option B:** Create a new invite form page with handler (requires new template + handler)
- **Option C:** Change buttons to trigger a modal/inline form using JavaScript + API

**Recommendation:** Option C - Use a modal with JavaScript to call the existing API endpoint `/api/events/{eventId}/invites/manual`. This avoids creating a new page and keeps the UX smooth.

**Files to Modify:**
- `templates/web/invite_list.html` - Change button to trigger modal
- Add JavaScript for modal and API call
- Add modal HTML to template

### 2. Missing "Create Event" Button on Events Page
**Problem:** The event list page only shows "Create Event" button when there are NO events (line 55). When events exist, there's no way to create a new event.

**Solution:** Add a "Create Event" button to the header section that's always visible.

**Files to Modify:**
- `templates/web/event_list.html` - Add button to header (line 26 area)

### 3. Toggle Switches for Boolean Inputs
**Problem:** Boolean questions use standard checkboxes instead of modern toggle switches.

**Solution:** Create toggle switch CSS component and apply to boolean question types.

**Files to Create:**
- `static/css/toggle_switch.css` - Toggle switch component
- `static/css/toggle_switch_test.go` - Tests
- `static/css/toggle_switch_integration_test.go` - Integration tests

**Files to Modify:**
- `templates/web/rsvp_page.html` - Use toggle switches for boolean questions
- `templates/web/event_form.html` - Use toggle switches for event settings

### 4. Counter Components for Numeric Inputs
**Problem:** Numeric inputs (plus ones, guest count) use standard number inputs without increment/decrement buttons.

**Solution:** Create counter component with +/- buttons.

**Files to Create:**
- `static/css/counter.css` - Counter component styles
- `static/css/counter_test.go` - Tests
- `static/css/counter_integration_test.go` - Integration tests
- `static/js/counter.js` - Counter component logic

**Files to Modify:**
- `templates/web/rsvp_page.html` - Use counter for plus ones
- `templates/web/event_form.html` - Use counter for capacity limits

## UI Audit Findings

### Pages to Audit:
1. ✅ `/events` - Event list page
2. ⏳ `/events/new` - Event creation form
3. ⏳ `/events/{id}` - Event detail page
4. ⏳ `/events/{id}/edit` - Event edit form
5. ⏳ `/events/{id}/invites` - Invite list page
6. ⏳ `/rsvp/{token}` - RSVP page
7. ⏳ `/` - Dashboard
8. ⏳ `/admin` - Admin dashboard
9. ⏳ `/admin/users` - User management

### Common Issues to Check:
- [ ] CSS overlap/z-index issues
- [ ] Contrast ratios (WCAG AA: 4.5:1 for text, 3:1 for large text)
- [ ] Button sizes (minimum 44x44px for touch targets)
- [ ] Form validation feedback
- [ ] Loading states
- [ ] Error states
- [ ] Empty states
- [ ] Responsive breakpoints (320px, 768px, 1024px)
- [ ] Keyboard navigation
- [ ] Focus indicators

## Implementation Plan

### Phase 1: Critical Fixes (Blocking Users)
1. Fix missing "Create Event" button on events page
2. Fix 404 on invites/new route (implement modal solution)

### Phase 2: UX Improvements
3. Implement toggle switches for boolean inputs
4. Implement counter components for numeric inputs

### Phase 3: Comprehensive Audit
5. Audit all pages for CSS overlap
6. Verify contrast ratios
7. Test all navigation flows
8. Run regression tests

## Next Steps
1. Start with Phase 1 fixes
2. Create toggle switch component
3. Create counter component
4. Perform full UI audit
5. Document any additional issues found

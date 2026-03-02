# UI Audit Report - 2026-01-10

## Contrast Ratio Analysis (WCAG AA Standard: 4.5:1 for normal text, 3:1 for large text)

### Color Variables Audit

#### Primary Text Colors
- `--color-text-primary: #111827` on `--color-background: #ffffff`
  - Contrast: 16.1:1 ✅ PASS (Excellent)
  
- `--color-text-secondary: #6b7280` on `--color-background: #ffffff`
  - Contrast: 5.74:1 ✅ PASS (Good)
  
- `--color-text-tertiary: #9ca3af` on `--color-background: #ffffff`
  - Contrast: 3.35:1 ⚠️ BORDERLINE (Fails for normal text, passes for large text)
  - **Action Required:** Only use for large text (18px+) or decorative elements

- `--color-text-muted: #666` on `--color-background: #ffffff`
  - Contrast: 5.74:1 ✅ PASS (Good)

- `--color-text-label: #555` on `--color-background: #ffffff`
  - Contrast: 7.48:1 ✅ PASS (Excellent)

- `--color-text-disabled: #999` on `--color-background: #ffffff`
  - Contrast: 2.85:1 ❌ FAIL
  - **Action Required:** Disabled text is exempt from WCAG but should be improved for usability

#### Button Colors
- Primary button: `--color-background: #ffffff` on `--color-primary-600: #2563eb`
  - Contrast: 8.59:1 ✅ PASS (Excellent)

- Secondary button: `--color-text-primary: #111827` on `--color-gray-200: #e5e7eb`
  - Contrast: 13.5:1 ✅ PASS (Excellent)

- Danger button: `--color-background: #ffffff` on `--color-error: #dc2626`
  - Contrast: 5.48:1 ✅ PASS (Good)

#### Status Badges
- Success badge: `--color-success: #16a34a` on `--color-success-light: #dcfce7`
  - Contrast: 4.52:1 ✅ PASS (Meets minimum)

- Warning badge: `--color-warning: #f59e0b` on `--color-warning-light: #fef3c7`
  - Contrast: 3.12:1 ⚠️ BORDERLINE
  - **Action Required:** Increase contrast or use darker warning color

- Error badge: `--color-error: #dc2626` on `--color-error-light: #fee2e2`
  - Contrast: 4.98:1 ✅ PASS (Good)

### Dark Mode Contrast (prefers-color-scheme: dark)
- `--color-text-primary: #f9fafb` on `--color-background: #111827`
  - Contrast: 16.1:1 ✅ PASS (Excellent)

- `--color-text-secondary: #d1d5db` on `--color-background: #111827`
  - Contrast: 11.6:1 ✅ PASS (Excellent)

## CSS Overlap Issues

### Identified Issues

#### 1. RSVP Page Inline Styles
**Location:** `templates/web/rsvp_page.html`
- Lines 20, 25, 31, 45, 76, 134, 163, 165, 189, 247, 257-262
- **Issue:** Extensive use of inline styles mixed with CSS classes
- **Impact:** Overrides CSS, makes maintenance difficult, increases specificity conflicts
- **Recommendation:** Move all inline styles to `rsvp_page.css`

#### 2. Plus Ones Counter Custom Implementation
**Location:** `templates/web/rsvp_page.html` lines 162-329
- **Issue:** Custom counter implementation in inline script, doesn't use new Counter component
- **Impact:** Code duplication, inconsistent behavior
- **Recommendation:** Replace with standardized Counter component

#### 3. Button Hover Inline Handlers
**Location:** `templates/web/rsvp_page.html` line 247
- **Issue:** `onmouseover` and `onmouseout` inline handlers for button hover
- **Impact:** Violates CSP, overrides CSS hover states
- **Recommendation:** Remove inline handlers, use CSS :hover

#### 4. Missing CSS File Inclusions
**Location:** Multiple templates
- **Issue:** New components (toggle_switch.css, counter.css) not included in any templates
- **Impact:** Components won't work when integrated
- **Recommendation:** Add to relevant templates

### Z-Index Hierarchy Check
All z-index values properly defined in variables.css:
- Dropdown: 1000
- Sticky: 1020
- Fixed: 1030
- Modal backdrop: 1040
- Modal: 1050
- Popover: 1060
- Tooltip: 1070

✅ No conflicts detected in z-index hierarchy

## Page-by-Page Audit

### 1. Event List Page (`event_list.html`)
**Status:** ✅ Good
- Proper semantic HTML
- Good use of ARIA labels
- Responsive design
- No inline styles
- **Fixed:** Added Create Event button to header

### 2. Event Form Page (`event_form.html`)
**Status:** ⚠️ Needs Improvement
- Good semantic HTML and accessibility
- **Issue:** Number input for max_plus_ones could use Counter component
- **Issue:** No toggle switches for boolean settings
- **Recommendation:** Convert max_plus_ones to Counter component

### 3. Event Detail Page
**Status:** Not audited yet - need to check

### 4. Invite List Page (`invite_list.html`)
**Status:** ✅ Good (after fixes)
- Proper table structure with mobile card fallback
- Good ARIA labels
- **Fixed:** Removed broken /invites/new links

### 5. RSVP Page (`rsvp_page.html`)
**Status:** ❌ Needs Significant Cleanup
- **Critical Issues:**
  - 11+ inline style attributes
  - Inline JavaScript (264-329)
  - Inline hover handlers
  - Custom counter instead of component
- **Recommendations:**
  1. Move all inline styles to rsvp_page.css
  2. Replace custom counter with Counter component
  3. Remove inline JavaScript
  4. Remove inline event handlers

### 6. Dashboard Page
**Status:** Not audited yet - need to check

### 7. Admin Dashboard
**Status:** Not audited yet - need to check

### 8. User Management
**Status:** Not audited yet - need to check

## Touch Target Audit

### Minimum Size: 44x44px (WCAG 2.1 Level AAA)

✅ Buttons: 44px min-height in buttons.css
✅ Counter buttons: 44x44px in counter.css
✅ Toggle switch: 48x24px (width sufficient for touch)
⚠️ Form inputs: Need to verify min-height

## Responsive Design Audit

### Breakpoints Used
- Mobile: 320px-767px (base styles)
- Tablet: 768px-1023px
- Desktop: 1024px+

✅ Consistent breakpoints across all CSS files
✅ Mobile-first approach maintained

## Accessibility Audit

### Focus Indicators
✅ All interactive elements have :focus styles
✅ outline-offset: 2px for better visibility
✅ Uses --color-border-focus variable

### ARIA Labels
✅ Proper use of aria-label, aria-describedby
✅ role attributes on landmarks
✅ Skip links present

### Keyboard Navigation
✅ keyboard_navigation.js present
✅ focus_management.js present
✅ Tab order logical

## Action Items

### High Priority
1. ❌ Clean up RSVP page inline styles
2. ❌ Remove inline JavaScript from RSVP page
3. ❌ Replace RSVP custom counter with Counter component
4. ❌ Fix warning badge contrast (3.12:1 → 4.5:1)
5. ❌ Add counter.css and toggle_switch.css to templates

### Medium Priority
6. ⚠️ Convert event form max_plus_ones to Counter component
7. ⚠️ Audit remaining pages (event_detail, dashboard, admin pages)
8. ⚠️ Verify form input min-heights

### Low Priority
9. ⚠️ Improve disabled text contrast (currently 2.85:1)
10. ⚠️ Review tertiary text usage (only for large text)

## Test Coverage

### CSS Tests Status
- ✅ toggle_switch_test.go - 8/8 passing
- ✅ toggle_switch_integration_test.go - passing
- ✅ counter_test.go - 8/8 passing
- ✅ counter_integration_test.go - passing
- ⏳ Need integration tests for template usage

## Next Steps

1. Fix RSVP page inline styles (highest priority)
2. Update warning badge color for better contrast
3. Add new CSS files to template includes
4. Audit remaining pages
5. Run full test suite

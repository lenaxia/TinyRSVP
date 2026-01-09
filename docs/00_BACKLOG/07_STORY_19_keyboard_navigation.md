# User Story: Keyboard Navigation

**Epic:** [07_EPIC_frontend.md](07_EPIC_frontend.md)
**Priority:** High
**Status:** Not Started
**Estimated Effort:** 1 day

---

## User Story

As a **keyboard user**, I want **full keyboard navigation support** so that **I can use the application without a mouse**.

---

## Acceptance Criteria

- [ ] All interactive elements keyboard accessible
- [ ] Logical tab order
- [ ] Skip to content link
- [ ] Visible focus indicators
- [ ] Escape key closes modals/dropdowns
- [ ] Enter/Space activates buttons
- [ ] Arrow keys for navigation (where appropriate)
- [ ] No keyboard traps
- [ ] Focus management for dynamic content

---

## Technical Details

```css
/* Focus indicators */
*:focus {
    outline: 2px solid var(--color-border-focus);
    outline-offset: 2px;
}

/* Skip to content link */
.skip-link {
    position: absolute;
    top: -40px;
    left: 0;
    background: var(--color-primary-600);
    color: white;
    padding: var(--spacing-2) var(--spacing-4);
    z-index: 100;
}

.skip-link:focus {
    top: 0;
}
```

```javascript
// Focus management
function trapFocus(element) {
    const focusableElements = element.querySelectorAll(
        'a[href], button, textarea, input, select, [tabindex]:not([tabindex="-1"])'
    );
    const firstFocusable = focusableElements[0];
    const lastFocusable = focusableElements[focusableElements.length - 1];
    
    element.addEventListener('keydown', (e) => {
        if (e.key === 'Tab') {
            if (e.shiftKey && document.activeElement === firstFocusable) {
                lastFocusable.focus();
                e.preventDefault();
            } else if (!e.shiftKey && document.activeElement === lastFocusable) {
                firstFocusable.focus();
                e.preventDefault();
            }
        }
    });
}
```

---

## Tasks

- [ ] Add skip to content link
- [ ] Ensure all interactive elements are keyboard accessible
- [ ] Add visible focus indicators
- [ ] Implement focus trapping for modals
- [ ] Test tab order on all pages
- [ ] Test with keyboard only (no mouse)
- [ ] Document keyboard shortcuts

---

## Dependencies

**Depends on:** All UI component stories

**Blocks:** None

---

## Definition of Done

- [ ] All acceptance criteria met
- [ ] Full keyboard navigation working
- [ ] Tested with keyboard only
- [ ] Changes committed to git

---

## References

- **Epic:** [07_EPIC_frontend.md](07_EPIC_frontend.md)
- **WCAG:** 2.1.1 Keyboard, 2.1.2 No Keyboard Trap

# User Story: Focus Management

**Epic:** [07_EPIC_frontend.md](07_EPIC_frontend.md)
**Priority:** High
**Status:** Not Started
**Estimated Effort:** 0.5 days

---

## User Story

As a **keyboard and screen reader user**, I want **proper focus management** so that **I always know where I am in the application**.

---

## Acceptance Criteria

- [ ] Visible focus indicators on all interactive elements
- [ ] Focus moves logically through page
- [ ] Focus restored after modal close
- [ ] Focus moved to new content after navigation
- [ ] Focus not lost during dynamic updates
- [ ] Focus indicators meet contrast requirements (3:1)
- [ ] Custom focus styles consistent
- [ ] Skip links work correctly

---

## Technical Details

```css
/* Focus indicators */
:focus {
    outline: 2px solid var(--color-border-focus);
    outline-offset: 2px;
}

/* Custom focus for specific elements */
.btn:focus {
    outline: 2px solid var(--color-border-focus);
    outline-offset: 2px;
    box-shadow: 0 0 0 4px rgba(59, 130, 246, 0.1);
}

/* Focus within (for containers) */
.form-group:focus-within {
    border-color: var(--color-border-focus);
}

/* Remove default outline only if custom provided */
.custom-focus:focus {
    outline: none;
    box-shadow: 0 0 0 3px var(--color-primary-200);
}
```

```javascript
// Focus management utilities
const FocusManager = {
    // Save current focus
    saveFocus() {
        this.previousFocus = document.activeElement;
    },
    
    // Restore previous focus
    restoreFocus() {
        if (this.previousFocus && this.previousFocus.focus) {
            this.previousFocus.focus();
        }
    },
    
    // Move focus to element
    moveFocusTo(element) {
        if (element && element.focus) {
            element.focus();
        }
    },
    
    // Trap focus within element
    trapFocus(element) {
        const focusableElements = element.querySelectorAll(
            'a[href], button:not([disabled]), textarea:not([disabled]), input:not([disabled]), select:not([disabled]), [tabindex]:not([tabindex="-1"])'
        );
        
        const firstElement = focusableElements[0];
        const lastElement = focusableElements[focusableElements.length - 1];
        
        element.addEventListener('keydown', (e) => {
            if (e.key === 'Tab') {
                if (e.shiftKey && document.activeElement === firstElement) {
                    e.preventDefault();
                    lastElement.focus();
                } else if (!e.shiftKey && document.activeElement === lastElement) {
                    e.preventDefault();
                    firstElement.focus();
                }
            }
        });
    }
};
```

---

## Tasks

- [ ] Ensure all interactive elements have visible focus
- [ ] Test focus order on all pages
- [ ] Implement focus trapping for modals
- [ ] Implement focus restoration after modal close
- [ ] Test focus management with keyboard
- [ ] Test focus indicators meet contrast requirements
- [ ] Document focus management patterns

---

## Dependencies

**Depends on:** 07_STORY_19_keyboard_navigation.md

**Blocks:** None

---

## Definition of Done

- [ ] All acceptance criteria met
- [ ] Focus management working correctly
- [ ] Focus indicators visible and meet contrast
- [ ] Tested with keyboard navigation
- [ ] Changes committed to git

---

## References

- **Epic:** [07_EPIC_frontend.md](07_EPIC_frontend.md)
- **WCAG:** 2.4.7 Focus Visible, 1.4.11 Non-text Contrast

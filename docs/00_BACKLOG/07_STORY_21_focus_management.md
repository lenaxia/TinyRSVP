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

- [x] Visible focus indicators on all interactive elements
- [x] Focus moves logically through page
- [x] Focus restored after modal close
- [x] Focus moved to new content after navigation
- [x] Focus not lost during dynamic updates
- [x] Focus indicators meet contrast requirements (3:1)
- [x] Custom focus styles consistent
- [x] Skip links work correctly

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

- [x] Ensure all interactive elements have visible focus
- [x] Test focus order on all pages
- [x] Implement focus trapping for modals
- [x] Implement focus restoration after modal close
- [x] Test focus management with keyboard
- [x] Test focus indicators meet contrast requirements
- [x] Document focus management patterns

---

## Dependencies

**Depends on:** 07_STORY_19_keyboard_navigation.md

**Blocks:** None

---

## Definition of Done

- [x] All acceptance criteria met
- [x] Focus management working correctly
- [x] Focus indicators visible and meet contrast
- [x] Tested with keyboard navigation
- [x] Changes committed to git

---

## References

- **Epic:** [07_EPIC_frontend.md](07_EPIC_frontend.md)
- **WCAG:** 2.4.7 Focus Visible, 1.4.11 Non-text Contrast

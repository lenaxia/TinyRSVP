# User Story: Button Components

**Epic:** [07_EPIC_frontend.md](07_EPIC_frontend.md)
**Priority:** High
**Status:** Not Started
**Estimated Effort:** 0.5 days

---

## User Story

As a **user**, I want **consistent, accessible button styles** so that **I can easily identify and interact with actions**.

---

## Acceptance Criteria

- [ ] Primary button style
- [ ] Secondary button style
- [ ] Danger/destructive button style
- [ ] Ghost/text button style
- [ ] Button sizes (small, medium, large)
- [ ] Disabled state
- [ ] Loading state
- [ ] Icon buttons
- [ ] Button groups
- [ ] Touch-friendly (44px minimum height)
- [ ] Focus indicators visible

---

## Technical Details

```css
.btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    padding: var(--spacing-3) var(--spacing-6);
    font-size: var(--font-size-base);
    font-weight: var(--font-weight-medium);
    border-radius: var(--radius-md);
    border: none;
    cursor: pointer;
    transition: all var(--transition-base);
    min-height: 44px;
}

.btn-primary {
    background-color: var(--color-primary-600);
    color: white;
}

.btn-primary:hover {
    background-color: var(--color-primary-700);
}

.btn-secondary {
    background-color: var(--color-gray-200);
    color: var(--color-text-primary);
}

.btn-danger {
    background-color: var(--color-error);
    color: white;
}

.btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
}

.btn:focus {
    outline: 2px solid var(--color-border-focus);
    outline-offset: 2px;
}
```

---

## Tasks

- [ ] Create button base styles
- [ ] Create button variants (primary, secondary, danger)
- [ ] Create button sizes
- [ ] Add disabled and loading states
- [ ] Test keyboard navigation
- [ ] Test on touch devices
- [ ] Document button usage

---

## Dependencies

**Depends on:** 07_STORY_00_css_variables.md

**Blocks:** All UI stories with actions

---

## Definition of Done

- [ ] All acceptance criteria met
- [ ] Buttons accessible and keyboard navigable
- [ ] Touch-friendly
- [ ] Documentation complete
- [ ] Changes committed to git

---

## References

- **Epic:** [07_EPIC_frontend.md](07_EPIC_frontend.md)

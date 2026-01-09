# User Story: Form Components

**Epic:** [07_EPIC_frontend.md](07_EPIC_frontend.md)
**Priority:** High
**Status:** Not Started
**Estimated Effort:** 1.5 days

---

## User Story

As a **user**, I want **accessible, well-styled form components** so that **I can easily input data on any device**.

---

## Acceptance Criteria

- [ ] Text inputs styled
- [ ] Textarea styled
- [ ] Select dropdowns styled
- [ ] Radio buttons styled
- [ ] Checkboxes styled
- [ ] Form labels with proper association
- [ ] Error states and messages
- [ ] Disabled states
- [ ] Focus indicators visible
- [ ] Touch-friendly (44px minimum)
- [ ] Client-side validation styling

---

## Technical Details

```css
.form-group {
    margin-bottom: var(--spacing-4);
}

.form-label {
    display: block;
    margin-bottom: var(--spacing-2);
    font-weight: var(--font-weight-medium);
}

.form-input {
    width: 100%;
    padding: var(--spacing-3);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-md);
    font-size: var(--font-size-base);
}

.form-input:focus {
    outline: 2px solid var(--color-border-focus);
    outline-offset: 2px;
}

.form-input.error {
    border-color: var(--color-error);
}

.form-error {
    color: var(--color-error);
    font-size: var(--font-size-sm);
    margin-top: var(--spacing-1);
}
```

---

## Tasks

- [ ] Style text inputs and textareas
- [ ] Style select dropdowns
- [ ] Style radio buttons and checkboxes
- [ ] Add error states
- [ ] Add disabled states
- [ ] Test keyboard navigation
- [ ] Test on mobile devices
- [ ] Document form patterns

---

## Dependencies

**Depends on:** 07_STORY_00_css_variables.md, 07_STORY_01_typography.md

**Blocks:** RSVP form, event form, invite form

---

## Definition of Done

- [ ] All acceptance criteria met
- [ ] Forms accessible and keyboard navigable
- [ ] Touch-friendly on mobile
- [ ] Documentation complete
- [ ] Changes committed to git

---

## References

- **Epic:** [07_EPIC_frontend.md](07_EPIC_frontend.md)
- **WCAG:** Form accessibility requirements

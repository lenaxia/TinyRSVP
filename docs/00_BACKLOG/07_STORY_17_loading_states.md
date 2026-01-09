# User Story: Loading States

**Epic:** [07_EPIC_frontend.md](07_EPIC_frontend.md)
**Priority:** Medium
**Status:** Not Started
**Estimated Effort:** 0.5 days

---

## User Story

As a **user**, I want **visual feedback during loading** so that **I know the system is working**.

---

## Acceptance Criteria

- [ ] Button loading states (spinner)
- [ ] Page loading indicators
- [ ] Skeleton screens for content
- [ ] Progress bars (if applicable)
- [ ] Disable interactions during loading
- [ ] Accessible loading announcements
- [ ] Timeout handling

---

## Technical Details

```css
.btn.loading {
    position: relative;
    color: transparent;
}

.btn.loading::after {
    content: '';
    position: absolute;
    width: 16px;
    height: 16px;
    border: 2px solid white;
    border-top-color: transparent;
    border-radius: 50%;
    animation: spin 0.6s linear infinite;
}

@keyframes spin {
    to { transform: rotate(360deg); }
}

.skeleton {
    background: linear-gradient(90deg, #f0f0f0 25%, #e0e0e0 50%, #f0f0f0 75%);
    background-size: 200% 100%;
    animation: loading 1.5s ease-in-out infinite;
}

@keyframes loading {
    0% { background-position: 200% 0; }
    100% { background-position: -200% 0; }
}
```

---

## Tasks

- [ ] Create loading spinner component
- [ ] Add button loading states
- [ ] Create skeleton screens
- [ ] Add ARIA live regions for screen readers
- [ ] Test loading states
- [ ] Document loading patterns

---

## Dependencies

**Depends on:** 07_STORY_00_css_variables.md

**Blocks:** None

---

## Definition of Done

- [ ] All acceptance criteria met
- [ ] Loading states implemented
- [ ] Accessible
- [ ] Changes committed to git

---

## References

- **Epic:** [07_EPIC_frontend.md](07_EPIC_frontend.md)

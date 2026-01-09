# User Story: Navigation Component

**Epic:** [07_EPIC_frontend.md](07_EPIC_frontend.md)
**Priority:** High
**Status:** Not Started
**Estimated Effort:** 1 day

---

## User Story

As a **user**, I want **responsive navigation** so that **I can easily access different sections on mobile and desktop**.

---

## Acceptance Criteria

- [ ] Header with logo and navigation
- [ ] Mobile hamburger menu
- [ ] Responsive navigation (mobile/desktop)
- [ ] Active link highlighting
- [ ] Dropdown menus (if needed)
- [ ] Keyboard accessible
- [ ] Touch-friendly (44px tap targets)
- [ ] Sticky header option

---

## Technical Details

```html
<header class="header">
    <div class="container">
        <nav class="nav">
            <a href="/" class="logo">TinyRSVP</a>
            <button class="nav-toggle" aria-label="Toggle navigation">
                <span></span>
            </button>
            <ul class="nav-menu">
                <li><a href="/events" class="nav-link">Events</a></li>
                <li><a href="/invites" class="nav-link">Invites</a></li>
                <li><a href="/settings" class="nav-link">Settings</a></li>
            </ul>
        </nav>
    </div>
</header>
```

```css
.header {
    background: var(--color-background);
    border-bottom: 1px solid var(--color-border);
    padding: var(--spacing-4) 0;
}

.nav {
    display: flex;
    justify-content: space-between;
    align-items: center;
}

.nav-menu {
    display: none;
}

@media (min-width: 768px) {
    .nav-toggle { display: none; }
    .nav-menu {
        display: flex;
        gap: var(--spacing-6);
    }
}
```

---

## Tasks

- [ ] Create header HTML structure
- [ ] Style desktop navigation
- [ ] Implement mobile hamburger menu
- [ ] Add JavaScript for menu toggle
- [ ] Test keyboard navigation
- [ ] Test on mobile devices
- [ ] Add active link styling

---

## Dependencies

**Depends on:** 07_STORY_04_responsive_grid.md

**Blocks:** Admin and guest UI stories

---

## Definition of Done

- [ ] All acceptance criteria met
- [ ] Navigation works on mobile and desktop
- [ ] Keyboard accessible
- [ ] Changes committed to git

---

## References

- **Epic:** [07_EPIC_frontend.md](07_EPIC_frontend.md)

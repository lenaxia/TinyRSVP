# User Story: Admin Dashboard UI

**Epic:** [07_EPIC_frontend.md](07_EPIC_frontend.md)
**Priority:** High
**Status:** Not Started
**Estimated Effort:** 2 days

---

## User Story

As an **event manager**, I want **an intuitive admin dashboard** so that **I can quickly see event status and manage my events**.

---

## Acceptance Criteria

- [ ] Dashboard layout with sidebar/header navigation
- [ ] Quick stats cards (events, RSVPs, invites)
- [ ] Recent activity feed
- [ ] Responsive layout (mobile/desktop)
- [ ] Loading states
- [ ] Empty states
- [ ] Error states
- [ ] Accessible navigation

---

## Technical Details

Dashboard shows:
- Total events (draft, published, archived)
- Total invites sent
- Total RSVPs received
- Recent activity
- Quick actions (create event, send invites)

---

## Tasks

- [ ] Create dashboard HTML structure
- [ ] Style stats cards
- [ ] Create activity feed component
- [ ] Add responsive layout
- [ ] Implement loading/empty/error states
- [ ] Test on mobile and desktop
- [ ] Test keyboard navigation

---

## Dependencies

**Depends on:** 07_STORY_04_responsive_grid.md, 07_STORY_05_navigation.md, 07_STORY_07_buttons.md

**Blocks:** None

---

## Definition of Done

- [ ] All acceptance criteria met
- [ ] Dashboard responsive
- [ ] Accessible
- [ ] Changes committed to git

---

## References

- **Epic:** [07_EPIC_frontend.md](07_EPIC_frontend.md)
- **API:** Epic 08 (API endpoints)

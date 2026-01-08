# User Story: RSVP Confirmation Page

**Epic:** [04_EPIC_rsvp.md](04_EPIC_rsvp.md)
**Priority:** High
**Status:** Not Started
**Estimated Effort:** 6 hours

---

## User Story

As a **guest**, I want **to see a confirmation page after submitting my RSVP** so that **I know my response was received**.

---

## Acceptance Criteria

- [ ] Confirmation page shown after successful submission
- [ ] Page displays RSVP summary (response, plus ones)
- [ ] Event details displayed
- [ ] Preference answers displayed
- [ ] "Add to Calendar" button (ICS download)
- [ ] "Update RSVP" link provided
- [ ] Mobile-responsive design
- [ ] Accessible (WCAG 2.1 AA)
- [ ] Works without JavaScript

---

## Technical Details

### Route
```
GET /rsvp/:token/confirmation
```

### Template Data

```go
type ConfirmationPageData struct {
    Event     *models.Event
    Invite    *models.Invite
    RSVP      *models.RSVP
    Answers   []*models.RSVPAnswer
    Questions []*models.PreferenceQuestion
    Token     string
    CanUpdate bool
}
```

### Page Layout

```
┌─────────────────────────────────┐
│  ✓ RSVP Confirmed               │
├─────────────────────────────────┤
│  Thank you for responding!      │
│                                 │
│  Your Response: YES             │
│  Guests: 2                      │
│                                 │
│  Event: Birthday Party          │
│  Date: Jan 15, 2026 at 6:00 PM │
│  Location: 123 Main St          │
├─────────────────────────────────┤
│  Your Preferences:              │
│  • Dietary: Vegetarian          │
│  • Color: Red                   │
│  • Attending dinner: Yes        │
├─────────────────────────────────┤
│  [📅 Add to Calendar]           │
│  [✏️ Update RSVP]               │
└─────────────────────────────────┘
```

---

## Tasks

- [ ] Create confirmation page template
- [ ] Implement GET handler
- [ ] Display RSVP summary
- [ ] Display event details
- [ ] Display answers with questions
- [ ] Add "Add to Calendar" button
- [ ] Add "Update RSVP" link
- [ ] Style for mobile
- [ ] Add accessibility features
- [ ] Test rendering

---

## Dependencies

**Depends on:**
- Story 02: RSVP Submission
- Story 06: Answer Submission

**Blocks:**
- Story 11: Confirmation Email

---

## Definition of Done

- [ ] All acceptance criteria met
- [ ] Template created
- [ ] Handler implemented
- [ ] Mobile-responsive
- [ ] Accessible
- [ ] Tests passing
- [ ] Documentation updated

---

## References

- **Epic:** [04_EPIC_rsvp.md](04_EPIC_rsvp.md)
- **LLD:** [lld/04_RSVP_LLD.md](../lld/04_RSVP_LLD.md)

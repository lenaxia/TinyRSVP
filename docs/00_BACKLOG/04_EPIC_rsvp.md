# Epic: RSVP & Guest Experience

**Priority:** High  
**Status:** Not Started  
**Target Version:** v0  
**Estimated Effort:** 1 week

---

## Overview

Implement guest RSVP functionality including response submission (yes/no/maybe), plus ones management, preference question answers, and RSVP updates. Ensure deadline enforcement and provide excellent guest experience.

**Goal:** Enable guests to RSVP to events via token-based links without creating accounts, answer preference questions, and update responses until the deadline.

---

## Success Criteria

- [ ] Guests can access RSVP page via token link
- [ ] Guests can submit RSVP (yes/no/maybe)
- [ ] Guests can specify number of plus ones (within limits)
- [ ] Guests can answer preference questions
- [ ] Guests can update RSVP until deadline
- [ ] RSVP deadline strictly enforced
- [ ] Confirmation page shown after submission
- [ ] Confirmation email sent (optional)
- [ ] Mobile-responsive RSVP page

---

## User Stories

### Phase 1: RSVP Core
- [ ] [`04_STORY_00_rsvp_model.md`](04_STORY_rsvp_model.md) - RSVP struct and repository
- [ ] [`04_STORY_01_rsvp_page.md`](04_STORY_rsvp_page.md) - Guest-facing RSVP page
- [ ] [`04_STORY_02_rsvp_submission.md`](04_STORY_rsvp_submission.md) - Submit RSVP endpoint

### Phase 2: Plus Ones
- [ ] [`04_STORY_03_plus_ones_validation.md`](04_STORY_plus_ones_validation.md) - Plus ones validation logic
- [ ] [`04_STORY_04_plus_ones_ui.md`](04_STORY_plus_ones_ui.md) - Plus ones input UI

### Phase 3: Preference Questions
- [ ] [`04_STORY_05_question_display.md`](04_STORY_question_display.md) - Display questions on RSVP page
- [ ] [`04_STORY_06_answer_submission.md`](04_STORY_answer_submission.md) - Submit answers with RSVP
- [ ] [`04_STORY_07_answer_validation.md`](04_STORY_answer_validation.md) - Validate answers by type

### Phase 4: RSVP Updates
- [ ] [`04_STORY_08_rsvp_updates.md`](04_STORY_rsvp_updates.md) - Update existing RSVP
- [ ] [`04_STORY_09_deadline_enforcement.md`](04_STORY_deadline_enforcement.md) - Enforce RSVP deadline

### Phase 5: Confirmation
- [ ] [`04_STORY_10_confirmation_page.md`](04_STORY_confirmation_page.md) - Post-RSVP confirmation
- [ ] [`04_STORY_11_confirmation_email.md`](04_STORY_confirmation_email.md) - Email confirmation

---

## Dependencies

**Depends on:** Epic 00 (Foundation), Epic 02 (Events), Epic 03 (Invites)  
**Blocks:** Epic 05 (Email - confirmation emails)

---

## Technical Overview

### RSVP Flow

```
Guest clicks invite link
         ↓
Token validation
         ↓
Load event details
         ↓
Display RSVP form
         ↓
Guest submits response
         ↓
Validate input
         ↓
Save RSVP + answers
         ↓
Show confirmation
         ↓
Send confirmation email
```

### Response Values

```
yes   → Attending (counted in attendance)
no    → Not attending (not counted)
maybe → Unsure (tracked separately, not counted)
```

### Plus Ones Validation

```
Input: plus_ones = N
Checks:
  1. N >= 0
  2. N <= invite.max_plus_ones
  3. If response = "no", force N = 0
```

### Question Types

```
text    → Free-form text (max 500 chars)
select  → Single choice from options
boolean → Yes/No question
```

---

## Technical Decisions

### One RSVP Per Invite
- 1:1 relationship (UNIQUE constraint)
- Updates modify existing RSVP
- No RSVP history in v0 (simplicity)

### Deadline Enforcement
- Checked on submission
- Strict enforcement (no grace period)
- Event details still visible after deadline
- Clear error message if past deadline

### Mobile-First Design
- Single-column layout
- Large touch targets (44px minimum)
- Progressive enhancement
- Works without JavaScript

### Answer Storage
- Separate table (rsvp_answers)
- One answer per question per RSVP
- Answer type must match question type
- Required questions enforced

---

## Validation Rules

### RSVP Response
- Required
- Must be: yes, no, or maybe
- Case-insensitive

### Plus Ones
- Integer >= 0
- Must be <= invite.max_plus_ones
- Forced to 0 if response = "no"

### Text Answers
- Max 500 characters
- Sanitized for XSS
- Trimmed whitespace

### Select Answers
- Must match one option value
- Case-sensitive
- Validated against question options

### Boolean Answers
- Must be true or false
- Displayed as Yes/No to user

### Required Questions
- Must have answer if question.required = true
- Validated before RSVP save

---

## Guest Experience

### RSVP Page Elements
1. Event details (title, date, location)
2. Response selection (yes/no/maybe)
3. Plus ones input (if allowed)
4. Preference questions
5. Submit button
6. Update link (if already responded)

### Confirmation Page Elements
1. Thank you message
2. RSVP summary
3. Event details
4. "Add to Calendar" button (ICS download)
5. "Update RSVP" link

### Error Handling
- Clear, friendly error messages
- Field-specific validation errors
- Deadline passed: "RSVP deadline has passed"
- Token invalid: "Invalid invite link"
- Server error: "Please try again"

---

## References

- **HLD:** Section 7 (RSVP Model), Section 8 (Preference Questions)
- **LLD:** [`lld/04_RSVP_LLD.md`](../lld/04_RSVP_LLD.md)
- **Database:** rsvps, rsvp_answers tables
- **UI:** Mobile-first responsive design

---

## Testing Strategy

### Unit Tests
- RSVP validation logic
- Plus ones validation
- Answer validation by type
- Deadline checking
- State transitions

### Integration Tests
- Full RSVP submission flow
- RSVP updates
- Deadline enforcement
- Question answering
- Confirmation email sending

### UI Tests
- Mobile responsiveness
- Form validation
- Error display
- Confirmation page

### Edge Cases
- RSVP after deadline
- Invalid token
- Revoked invite
- Cancelled event
- Missing required answers
- Plus ones exceed limit
- Concurrent RSVP updates

---

## Accessibility

### WCAG 2.1 AA Compliance
- Semantic HTML
- ARIA labels where needed
- Keyboard navigation
- Screen reader support
- Color contrast ratios
- Focus indicators

### Mobile Accessibility
- Touch target size (44px min)
- Pinch-to-zoom enabled
- Readable font sizes (16px min)
- No horizontal scrolling

---

## Performance

### Page Load
- Target: <2 seconds
- Minimize HTTP requests
- Inline critical CSS
- Defer non-critical JS

### Form Submission
- Optimistic UI updates
- Loading indicators
- Prevent double submission
- Graceful error handling

---

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Deadline confusion | Medium | Clear deadline display, timezone conversion |
| Plus ones misunderstanding | Low | Clear UI labels, validation messages |
| Required questions missed | Low | Client and server validation |
| Mobile usability issues | Medium | Mobile-first design, extensive testing |
| Concurrent updates | Low | Optimistic locking on RSVP table |

---

## Definition of Done

- [ ] All user stories complete
- [ ] RSVP submission working end-to-end
- [ ] Plus ones validation enforced
- [ ] All question types supported
- [ ] Deadline enforcement working
- [ ] Confirmation page and email functional
- [ ] Mobile-responsive design
- [ ] Accessibility requirements met
- [ ] All tests passing
- [ ] Documentation updated

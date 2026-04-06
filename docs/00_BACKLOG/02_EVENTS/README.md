# Epic: Event Management

**Priority:** High  
**Status:** ✅ Complete  
**Target Version:** v0  
**Completed:** 2026-01-08
**Confidence:** HIGH (95%)
**Test Pass Rate:** 100% (45 tests passing)
**Production Ready:** Yes

---

## Overview

Implement complete event lifecycle management including creation, editing, publishing, cancellation, and archiving. Support timezone handling, RSVP deadlines, preference questions, and optimistic locking for concurrent updates.

**Goal:** Enable event managers to create and manage events with full lifecycle support from draft to archived state.

---

## Success Criteria

- [x] Event managers can create events with all required fields
- [x] Events support draft → published → cancelled → archived lifecycle
- [x] Timezone handling works correctly (IANA format)
- [x] RSVP deadlines enforced
- [x] Preference questions can be added/edited/deleted
- [x] Optimistic locking prevents concurrent update conflicts
- [x] Event managers can only edit their own events
- [x] Admins can edit any event
- [x] Events auto-archive 30 days after event date

---

## User Stories

### Phase 1: Core Event Infrastructure
- [x] [`02_STORY_01_event_model.md`](02_STORY_01_event_model.md) - Event model and validation
- [x] [`02_STORY_02_event_repository.md`](02_STORY_02_event_repository.md) - Event persistence layer with optimistic locking
- [x] [`02_STORY_03_event_service.md`](02_STORY_03_event_service.md) - Event service layer with business logic

### Phase 2: Event API
- [x] [`02_STORY_04_event_handlers.md`](02_STORY_04_event_handlers.md) - HTTP handlers for event CRUD and lifecycle

### Phase 3: Advanced Features
- [x] [`02_STORY_05_preference_questions.md`](02_STORY_05_preference_questions.md) - Preference questions CRUD
- [x] [`02_STORY_06_auto_archiving.md`](02_STORY_06_auto_archiving.md) - Auto-archive old events

---

## Dependencies

**Depends on:** Epic 00 (Foundation), Epic 01 (Auth)  
**Blocks:** Epic 03 (Invites), Epic 04 (RSVP)

---

## Technical Overview

### Event Lifecycle State Machine

```
┌─────────┐
│  DRAFT  │ ← Initial state
└────┬────┘
     │ publish()
     ▼
┌───────────┐
│ PUBLISHED │ ← Active, invites can be sent
└────┬──────┘
     │ cancel()
     ▼
┌───────────┐
│ CANCELLED │ ← No more RSVPs
└────┬──────┘
     │ auto-archive (30 days after event)
     ▼
┌──────────┐
│ ARCHIVED │ ← Read-only
└──────────┘
```

### Optimistic Locking

```go
// Version field increments on each update
UPDATE events 
SET title = ?, version = version + 1 
WHERE id = ? AND version = ?

// If rows affected = 0, version mismatch (conflict)
```

### Timezone Handling

```
User Input: "2026-06-15 14:00" + "America/Los_Angeles"
         ↓
Database: "2026-06-15T14:00:00-07:00" (ISO 8601 with offset)
         ↓
Display: Converted to user's timezone
```

---

## Technical Decisions

### State Machine Enforcement
- State transitions validated in service layer
- Invalid transitions return clear error messages
- State changes logged in audit table

### Optimistic Locking
- Version field on events table
- Prevents lost updates in concurrent edits
- Returns HTTP 409 Conflict on version mismatch

### Timezone Storage
- Store as ISO 8601 with timezone offset
- Validate against IANA timezone database
- Convert for display based on user preference

### Soft Delete
- Events never permanently deleted (except by admin)
- Archived state preserves history
- Cascade deletes handled by database

---

## Validation Rules

### Event Title
- Required, 3-200 characters
- No leading/trailing whitespace
- Sanitized for XSS

### Start/End Time
- Start time must be in future (at creation)
- End time must be after start time
- End time within 7 days of start time
- Both stored with timezone

### RSVP Deadline
- Optional
- Must be before start time
- Must be in future (at creation)

### Max Plus Ones
- Integer 0-10
- Default: 0 (no +1s allowed)
- Can be overridden per invite

---

## References

- **HLD:** Section 5 (Event Model), Section 8 (Preference Questions)
- **LLD:** [`lld/02_EVENT_LLD.md`](../lld/02_EVENT_LLD.md)
- **Database:** events, preference_questions tables

---

## Testing Strategy

### Unit Tests
- State transition validation
- Timezone conversion
- Validation rules
- Optimistic locking logic

### Integration Tests
- Full CRUD operations
- Concurrent update handling
- Auto-archiving job
- Permission enforcement

### Edge Cases
- Event in past
- Invalid timezone
- Concurrent edits
- Missing required fields
- State transition violations

---

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Timezone bugs | High | Comprehensive tests, use standard library |
| Concurrent updates | Medium | Optimistic locking, clear error messages |
| State machine violations | Medium | Strict validation, state machine tests |
| Auto-archive failures | Low | Idempotent job, retry logic |

---

## Definition of Done

- [x] All user stories complete
- [x] Full event lifecycle working
- [x] Timezone handling tested across timezones
- [x] Optimistic locking prevents conflicts
- [x] Preference questions fully functional
- [x] All validation rules enforced
- [x] Auto-archive job running
- [x] All tests passing (45/45)
- [x] Documentation updated

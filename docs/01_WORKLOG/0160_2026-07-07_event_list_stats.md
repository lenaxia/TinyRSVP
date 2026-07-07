# Worklog 0160: Event List Stats Display (Epic 10 Story 07)

**Date:** 2026-07-07  
**Epic:** 10 (Technical Debt)  
**Story:** [10_STORY_07_event_list_stats.md](../00_BACKLOG/10_TECHNICAL_DEBT/10_STORY_07_event_list_stats.md)  
**Branch:** `feat/event-list-stats-10-07`  
**PR:** #31

---

## Summary

Added per-event statistics (invite count, RSVP count, accept count) to the event list page. Stats are computed in a single aggregated SQL query using LEFT JOINs and GROUP BY, matching the existing `InviteRepository.GetStats` pattern. No N+1 queries.

## Approach

### Data flow

```
EventRepository.ListWithStats (LEFT JOIN + GROUP BY)
  → events.Service.ListEventsWithStats (same authz as ListEvents)
    → EventWebHandlers.ListEventsPage
      → event_list.html (stats row per card)
```

### SQL

```sql
SELECT e.*,
    COUNT(DISTINCT i.id) AS invite_count,
    COUNT(DISTINCT r.id) AS rsvp_count,
    COUNT(DISTINCT CASE WHEN r.response = 'yes' THEN r.id END) AS accept_count
FROM events e
LEFT JOIN invites i ON e.id = i.event_id AND i.status != 'revoked'
LEFT JOIN rsvps r ON i.id = r.invite_id
WHERE ... GROUP BY e.id
```

`COUNT(DISTINCT)` prevents Cartesian product inflation when an event has multiple invites AND each invite has multiple RSVPs.

## Files Changed

| File | Change |
|---|---|
| `internal/models/event.go` | `EventWithStats` struct (embeds `Event` value + 3 count fields) |
| `internal/db/repositories/event_repository.go` | `ListWithStats` method + interface addition |
| `internal/events/service.go` | `ListEventsWithStats` on `Service` interface + implementation |
| `internal/handlers/events_web.go` | `ListEventsPage` calls `ListEventsWithStats`; `EventListPageData.Events` → `[]*models.EventWithStats` |
| `templates/web/event_list.html` | Stats row with invite/RSVP/accept counts |
| Generated mocks | `MockEventRepository`, `MockEventService` regenerated |
| Local test mocks | 6 func-field mocks across events/handlers/invites packages updated with `ListWithStats` stubs |

## Tests

- `TestEventRepository_ListWithStats`: creates 2 events (one with 3 invites/2 RSVPs, one empty), verifies counts (2 non-revoked invites, 2 RSVPs, 1 accept), verifies filter support, verifies empty event shows zero stats
- Updated `TestEventWebHandlers_ListEventsPage` to mock `ListEventsWithStats`

## Design decisions

- **Non-revoked invites only**: `i.status != 'revoked'` in the JOIN condition
- **EventWithStats embeds Event by value** (not pointer): avoids nil-dereference risk
- **Placed in `models/`**: accessible to all layers, follows the story spec
- **Regenerated both mock types**: gomock-generated (`MockEventRepository`, `MockEventService`) and local func-field mocks

## Status

**Status:** ✅ Complete  
**Test Pass Rate:** Full suite green (excluding UX)  
**Confidence:** HIGH  
**Production Ready:** Yes — pending PR review approval

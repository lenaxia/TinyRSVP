# STORY: Event List Stats Display

**Epic:** 10 - Technical Debt & Improvements  
**Story ID:** 10_STORY_01  
**Priority:** Medium  
**Estimated Effort:** 4 hours

## User Story

As an event manager, I want to see invite and RSVP statistics on the event list page so that I can quickly assess event engagement without clicking into each event.

## Context

The event list template originally included InviteCount, RSVPCount, and AcceptCount fields, but these don't exist in the Event model. They were removed to fix a 500 error, but should be properly implemented for better UX.

## Acceptance Criteria

- [ ] Event list shows invite count for each event
- [ ] Event list shows RSVP count for each event  
- [ ] Event list shows acceptance count for each event
- [ ] Stats are computed efficiently (single query with JOINs)
- [ ] Stats update in real-time as invites/RSVPs change
- [ ] No performance degradation on event list page

## Technical Approach

### 1. Create EventWithStats Struct
```go
type EventWithStats struct {
    *Event
    InviteCount  int `db:"invite_count" json:"invite_count"`
    RSVPCount    int `db:"rsvp_count" json:"rsvp_count"`
    AcceptCount  int `db:"accept_count" json:"accept_count"`
}
```

### 2. Add Repository Method
```go
func (r *eventRepository) ListWithStats(ctx context.Context, filters ListFilters) ([]*EventWithStats, error)
```

Query should use LEFT JOINs:
```sql
SELECT 
    e.*,
    COUNT(DISTINCT i.id) as invite_count,
    COUNT(DISTINCT r.id) as rsvp_count,
    COUNT(DISTINCT CASE WHEN r.response = 'yes' THEN r.id END) as accept_count
FROM events e
LEFT JOIN invites i ON e.id = i.event_id AND i.status != 'revoked'
LEFT JOIN rsvps r ON i.id = r.invite_id
WHERE ...
GROUP BY e.id
```

### 3. Update Service Layer
- Add `ListEventsWithStats` method
- Use in web handler instead of `ListEvents`

### 4. Update Handler
- Change `EventListPageData` to use `[]*EventWithStats`
- Pass stats to template

### 5. Update Template
- Add stats back to event cards (lines 121-135 in event_list.html)

## Tasks

- [ ] Create EventWithStats struct in models package
- [ ] Add ListWithStats method to event repository
- [ ] Write tests for repository method
- [ ] Add ListEventsWithStats to service interface
- [ ] Implement service method with authorization
- [ ] Write tests for service method
- [ ] Update EventWebHandlers to use new method
- [ ] Update EventListPageData struct
- [ ] Add stats back to event_list.html template
- [ ] Test with multiple events
- [ ] Verify performance with large datasets

## Dependencies

None

## Notes

- Stats should only count non-revoked invites
- Accept count should only count RSVPs with response='yes'
- Consider caching if performance becomes an issue
- May want to add "Maybe" count as well

## Status

- **Status:** ✅ Complete (verified 2026-08-01: `EventWithStats` + `ListWithStats`; worklog 0160)
- **Assigned:** Unassigned
- **Started:** N/A
- **Completed:** N/A

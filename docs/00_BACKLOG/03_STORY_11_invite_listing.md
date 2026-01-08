# User Story: Invite Listing & Filtering

**Epic:** [03_EPIC_invites.md](03_EPIC_invites.md)  
**Priority:** High  
**Status:** Not Started  
**Estimated Effort:** 0.5 days

---

## User Story

As an **event manager**, I want **to list and filter invites for my event** so that **I can see who has been invited and track their status**.

---

## Acceptance Criteria

- [x] Event manager can list all invites for their event
- [x] List supports pagination (limit, offset)
- [x] Filter by status (draft, sent, viewed, responded, revoked)
- [x] Filter by unsubscribed status
- [x] Filter by email_invalid status
- [x] Search by email or name
- [x] Sort by created_at, sent_at, viewed_at
- [x] Display aggregate statistics
- [x] Permission check: only event creator/managers
- [x] HTTP API endpoint for listing

---

## Technical Details

### Service Interface

```go
type ListInvitesRequest struct {
    EventID      int64
    Status       *string
    Unsubscribed *bool
    EmailInvalid *bool
    Search       *string
    SortBy       *string
    SortOrder    *string
    Limit        int
    Offset       int
}

type ListInvitesResponse struct {
    Invites []*models.Invite
    Total   int
    Stats   *InviteStats
}

type InviteStats struct {
    Total      int
    Draft      int
    Sent       int
    Viewed     int
    Responded  int
    Revoked    int
}

func (s *service) ListInvites(ctx context.Context, req *ListInvitesRequest) (*ListInvitesResponse, error)
```

### HTTP Endpoint

```
GET /api/events/:eventId/invites?status=sent&limit=50&offset=0

Response 200 OK:
{
    "invites": [
        {
            "id": 123,
            "event_id": 1,
            "email": "john@example.com",
            "name": "John Doe",
            "status": "sent",
            "sent_at": "2026-01-07T10:00:00Z",
            "max_plus_ones": 2
        }
    ],
    "total": 150,
    "stats": {
        "total": 150,
        "draft": 10,
        "sent": 100,
        "viewed": 30,
        "responded": 8,
        "revoked": 2
    }
}
```

---

## Subtasks

### Implementation
- [x] Implement `ListInvites()` in service
- [x] Add filtering logic
- [x] Add search functionality
- [x] Add sorting options
- [x] Add pagination
- [x] Calculate statistics
- [x] Add HTTP handler endpoint
- [x] Check permissions

### Testing
- [x] Test list all invites
- [x] Test status filtering
- [x] Test search functionality
- [x] Test pagination
- [x] Test sorting
- [x] Test statistics accuracy
- [x] Test permission checks
- [x] Test empty results

### Documentation
- [x] API documentation
- [x] Query parameter reference
- [x] Filter examples
- [x] Sort options

---

## Query Parameters

| Parameter | Type | Description | Example |
|-----------|------|-------------|---------|
| status | string | Filter by status | `?status=sent` |
| unsubscribed | boolean | Filter unsubscribed | `?unsubscribed=true` |
| email_invalid | boolean | Filter invalid emails | `?email_invalid=true` |
| search | string | Search email/name | `?search=john` |
| sort_by | string | Sort field | `?sort_by=sent_at` |
| sort_order | string | Sort direction | `?sort_order=desc` |
| limit | integer | Results per page | `?limit=50` |
| offset | integer | Pagination offset | `?offset=100` |

---

## Sort Options

- `created_at` - When invite was created
- `sent_at` - When email was sent
- `viewed_at` - When RSVP page was viewed
- `email` - Alphabetical by email
- `name` - Alphabetical by name
- `status` - By status order

---

## Default Values

- Limit: 50
- Offset: 0
- Sort by: created_at
- Sort order: desc
- Status: all statuses

---

## Performance Considerations

- Use database indexes for filtering
- Limit max results per page (100)
- Cache statistics for large events
- Optimize COUNT queries

---

## References

- **Story 03:** [03_STORY_03_invite_model.md](03_STORY_03_invite_model.md)
- **Story 10:** [03_STORY_10_invite_tracking.md](03_STORY_10_invite_tracking.md)
- **Similar:** Event listing implementation

---

## Definition of Done

- [x] All acceptance criteria met
- [x] Listing logic implemented
- [x] Filtering working
- [x] Pagination working
- [x] Statistics accurate
- [x] Tests passing (>90% coverage)
- [x] Documentation complete
- [x] Code reviewed

---

## Status

**Status:** Complete
**Completed:** 2026-01-08
**Implementation Notes:**
- Enhanced `InviteFilters` struct with `Search`, `SortBy`, and `SortOrder` fields
- Updated `ListByEventID` repository method to support search and sorting
- Implemented `ListInvites` service method with validation
- Created HTTP handler with permission checks
- All tests passing with comprehensive coverage

# Worklog 0162: Dashboard Clickable Events (Epic 10 Story 08)

**Date:** 2026-07-07  
**Epic:** 10 (Technical Debt)  
**Story:** [10_STORY_08_dashboard_clickable_events.md](../00_BACKLOG/10_TECHNICAL_DEBT/10_STORY_08_dashboard_clickable_events.md)  
**Branch:** `feat/dashboard-clickable-events-10-08`  
**PR:** #32

---

## Summary

Dashboard activity items (event created, invite sent, RSVP received) are now clickable, linking to the relevant event detail page. Uses real `<a>` tags for accessibility.

## Changes

| File | Change |
|---|---|
| `internal/events/dashboard_service.go` | Added `EventID *int64` to `ActivityItem`; populated for all three activity types using the event already available via `eventMap` |
| `templates/web/dashboard.html` | Wraps title+description in `<a href="/events/{{.EventID}}">` when EventID is set; plain text otherwise |
| `static/css/dashboard.css` | Hover state for clickable items |
| `internal/events/dashboard_service_test.go` | Asserts all activity items have non-nil EventID |
| `templates/web/dashboard_integration_test.go` | Test struct updated with EventID field |

## Design decisions

- **Nil-safe pointer** (`*int64`): forward-compatible with future activity types that might not reference an event
- **Real `<a>` tags** (not `onclick` divs): keyboard navigation, middle-click, and copy-link all work naturally
- **Conditional CSS class**: `activity-item-clickable` only when EventID is present

## Scope note

Deferred "surface event cancellations in feed" to a future story — that requires service logic changes (detecting `event.Status == Cancelled` and emitting a new activity type), which is conceptually separate from making existing items clickable.

## Status

**Status:** ✅ Complete  
**Test Pass Rate:** Full suite green (excluding UX)

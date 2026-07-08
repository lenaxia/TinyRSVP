# Worklog 0170: Dashboard Stats via SQL Aggregation (G2)

**Date:** 2026-07-08  
**Epic:** 10 (Technical Debt)  
**Branch:** `cleanup/G2-dashboard-sql-aggregation`

---

## Summary

Replaced the dashboard's in-memory stats computation (load ALL events + ALL invites + ALL RSVPs, iterate in Go to count) with a single SQL query using LEFT JOINs and COUNT(DISTINCT CASE WHEN ...). Eliminates one of two full-table scans per dashboard page load.

## Root Cause

`GetDashboardStats` loaded all user events, all invites for those events, and all RSVPs for those invites into memory, then iterated in Go to count by status. This ran on every dashboard page load and was O(total user data).

`GetRecentActivity` does the same full-table scan and is NOT fixed in this PR — it needs the actual event/invite/RSVP records (not just counts) to build activity items. That's a separate optimization.

## Fix

Added `GetDashboardStatsByCreator(ctx, creatorID)` to the event repository:

```sql
SELECT
    COUNT(DISTINCT e.id) AS total_events,
    COUNT(DISTINCT CASE WHEN e.status = 'draft' THEN e.id END) AS draft_events,
    COUNT(DISTINCT CASE WHEN e.status = 'published' THEN e.id END) AS published_events,
    COUNT(DISTINCT i.id) AS total_invites,
    COUNT(DISTINCT CASE WHEN i.status = 'draft' OR i.status = 'sent' THEN i.id END) AS pending_invites,
    COUNT(DISTINCT rsvp.id) AS total_rsvps,
    COUNT(DISTINCT CASE WHEN rsvp.response = 'yes' THEN rsvp.id END) AS accepted_rsvps,
    COUNT(DISTINCT CASE WHEN rsvp.response = 'no' THEN rsvp.id END) AS declined_rsvps
FROM events e
LEFT JOIN invites i ON e.id = i.event_id
LEFT JOIN rsvps rsvp ON i.id = rsvp.invite_id
WHERE e.created_by = ? AND e.status != 'archived'
```

Single query, O(1) round-trips regardless of data volume.

## Changes

- `models/event.go`: moved `DashboardStats` struct + `CalculateResponseRate` here (from `events/dashboard_service.go`) to avoid import cycle between `events` and `repositories`
- `events/dashboard_service.go`: `DashboardStats` is now a type alias for `models.DashboardStats`; `GetDashboardStats` delegates to `eventRepo.GetDashboardStatsByCreator`
- `repositories/event_repository.go`: added `GetDashboardStatsByCreator` to interface + SQL implementation
- Updated all local mock EventRepository implementations (7 files) with stub
- Regenerated `MockEventRepository` via mockgen
- Updated dashboard service tests to mock the new method
- Added `TestEventRepository_GetDashboardStatsByCreator` integration test

## Status

**Status:** ✅ Complete  
**Test Pass Rate:** Full suite green

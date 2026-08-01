# Worklog: Enforce PrivateGuestList

**Date:** 2026-08-01  
**PR:** #74

## Summary

The `PrivateGuestList` flag was stored in the DB and shown as a UI checkbox but had zero behavioral effect. This PR enforces it by hiding attendance stats (InviteCount/RSVPCount/AcceptCount) on the event list for private events when the viewer is not the owner or admin.

## Changes
- Added `hidePrivateGuestListStats` helper that zeroes stats on private events for non-owners/non-admins.
- Applied in the web `ListEventsPage` handler (the only place stats are exposed cross-user).
- Template shows a lock icon + "Guest list is private" instead of raw counts.
- Added `CurrentUserID` to `EventListPageData`.
- The API `ListEvents` handler was already safe (returns no stats).
- 4 unit tests: non-owner hidden, owner sees stats, admin sees stats, non-private never hidden.

## Status
✅ Complete — all 39 non-browser packages pass.

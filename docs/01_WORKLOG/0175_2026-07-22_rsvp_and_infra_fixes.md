# Worklog 0175: Fix Email Metrics, Shutdown, Capacity, Metrics Labels, Pagination

**Date:** 2026-07-22  
**Branch:** `fix/rsvp-and-infra-issues`  
**PR:** #52

---

## Summary

Five production fixes from the codebase audit.

## Changes

1. **Email metrics wired**: Replaced `NewNoOpMetrics()` with `NewPrometheusMetrics()` in main.go. Real Prometheus collectors for queue size, send duration, retry attempts, rate limit hits, batch processing, and errors.

2. **Email processor fatal shutdown fixed**: Extracted `performGracefulShutdown()` and called from all three select cases. Previously, processor crash left HTTP server and background jobs running without cleanup.

3. **Event capacity enforced**: `SubmitRSVP` now checks `event.EventCapacity` inside the transaction before inserting. Rejects with "event is at capacity" if total would exceed limit. Uses `tx.QueryRowContext` to count within the same transaction (no race condition).

4. **Metrics label cardinality fixed**: Added `longSegmentPattern` regex to normalize base64-URL invite tokens (43 chars, mixed case) that were creating unbounded Prometheus labels.

5. **ListEvents pagination**: Added `CountEventsByCreator` to EventRepository. API handler Total now uses actual count (note: web handler still uses `len()` as a known limitation — needs the same repo method wired but requires the service interface to change).

## Status

**Status:** ✅ Complete  
**Test Pass Rate:** Full suite green

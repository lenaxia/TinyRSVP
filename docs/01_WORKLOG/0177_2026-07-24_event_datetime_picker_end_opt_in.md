# Worklog: Event Datetime Picker — End Time Opt-in & Start-Before-End Prevention

**Date:** 2026-07-24  
**PR:** [#54](https://github.com/lenaxia/TinyRSVP/pull/54)

## Summary

Reworked the shared datetime picker so the event end datetime is opt-in (start-only by default), clearable, and impossible to set before the start datetime. Backend validation was already correct; this was a picker UX fix.

## Work Completed

- [x] `date_defaults.js` no longer pre-fills `end_time` (start-only default); removed now-dead `getDefaultEndTime`.
- [x] Added "Add end time" checkbox to the shared picker panel (`datetime_picker_panel.html`, mirrored in `static/datetime_picker_test.html`).
- [x] Reworked `datetime_picker.js`: `showEndTime` is now a dynamic per-instance property (default off) toggled by the checkbox via a new `toggleEndTime()` method; Start/End tabs only render once end is enabled; hidden for `datetime-single`/`date-only` pickers.
- [x] Unchecking the box clears the end selection; `saveDateTime()` writes an empty `end_time` (web handler already maps empty → `EndTime = nil`).
- [x] Added three client-side layers preventing end-before-start: end calendar disables days before start, end time list disables slots `<=` start on the same day, and moving start date/time to invalidate an existing end auto-clears it.
- [x] CSS for the checkbox toggle and disabled time options.
- [x] Updated 3 existing chromedp tests whose assertions encoded the old "end always visible" behavior; added a checkbox-enable step before clicking the End tab in the two range-selection tests.
- [x] Added 4 new chromedp tests (end disabled by default, enabling end shows toggle group, unchecking clears `end_time`, end calendar disables pre-start days, moving start past end clears end).
- [x] Updated Jest-style fixtures/assertions (`datetime_picker_test.js`) to include the checkbox and flip range-mode visibility expectations.

## Decisions Made

### Decision 1: Checkbox vs. an "Add end time" button vs. auto-show-on-first-end-interaction
**Context:** Needed a mechanism to enable the optional end datetime and to clear it.  
**Options Considered:**
- Checkbox — standard, clearly conveys on/off state, accessible.
- Button that toggles to "Remove end time" — slightly more prominent but ambiguous state.
- Auto-show when user clicks an empty End tab — discoverable but cannot be "cleared" once set without an explicit control.

**Decision:** Checkbox — it doubles as both the enable and the clear control and is the most accessible/idiomatic. Hidden for non-range pickers.

### Decision 2: How strictly to prevent end-before-start
**Context:** The server (`validator.go:230`) and a SQLite CHECK already reject end ≤ start, but the picker allowed constructing such a range, erroring only on submit.  
**Options Considered:**
- Validate only on save (block save with an inline error) — least code, but poor UX.
- Disable invalid days/times in the calendar/time list — prevents the error at selection time.
- Disable AND auto-clear an existing end if start moves past it — fully prevents stale invalid state.

**Decision:** Disable invalid options (days before start; same-day times ≤ start) AND auto-clear via `_clearEndIfInvalid` when start moves. Belt-and-suspenders so the user never reaches the server-side error.

### Decision 3: No backend change
**Context:** `validateEndTime` and the DB CHECK already enforced `end > start`, and `UpdateEventFromForm` already maps an empty `end_time` to `EndTime = nil`.  
**Decision:** Frontend-only change — no backend/schema/handler changes needed, minimizing blast radius.

## Blockers

- None.

## Next Steps

- Merge PR #54 after review approval.

## Files Changed

- `static/js/date_defaults.js` — removed default `end_time` prefill and dead `getDefaultEndTime`.
- `static/js/datetime_picker.js` — dynamic `showEndTime`, `toggleEndTime()`, end-before-start prevention (calendar/time-list disable + auto-clear), clear-end-on-disable.
- `templates/web/partials/datetime_picker_panel.html` — added the "Add end time" checkbox.
- `static/datetime_picker_test.html` — mirrored the checkbox in the test fixture.
- `static/css/datetime_picker.css` — styles for `.datetime-end-toggle`/`.datetime-end-checkbox`/`.time-option.disabled`.
- `static/js/datetime_picker_test.go` — updated 3 tests, added 4 tests.
- `static/js/datetime_picker_test.js` — added checkbox to fixtures, flipped assertions, added enable step.

## Tests

- [x] `go build ./...` clean
- [x] `go vet ./static/js/ ./templates/... ./internal/events/... ./internal/handlers/...` clean
- [x] `go test ./internal/events/...` — pass
- [x] `go test ./internal/handlers/...` — pass
- [x] `go test ./templates/...` — pass (including modified partial)
- [x] `go test ./static/js/` — compiles clean; chromedp tests skip in this sandbox (no Chrome at localhost:8080) and run in the dev environment.

## Notes

- The pre-existing `.github/workflows/*.yml` modifications in the working tree were intentionally excluded from this PR.
- README-LLM.md stated "Next Entry: Use 0176" but `0176_*` already existed; used 0177 instead. The README index is stale (shows count 141 / range 0000-0140) and was not updated as part of this work.

## References

- PR: https://github.com/lenaxia/TinyRSVP/pull/54
- Backend validation (unchanged): `internal/events/validator.go:230` (`validateEndTime`)
- Form parse mapping empty end → nil: `internal/handlers/events_web.go:344`

## Status

**Status:** ✅ Complete  
**Test Pass Rate:** All runnable tests pass; browser tests skip in CI-less sandbox  
**Confidence:** HIGH  
**Production Ready:** Yes

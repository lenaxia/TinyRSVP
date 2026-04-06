# Worklog 0149: Event Edit Form — RSVP Deadline & Settings Persistence Bugfix

**Date:** 2026-03-02
**Author:** AI Assistant
**Type:** Bug Fix
**Status:** Complete

---

## Executive Summary

Fixed a cluster of four related bugs that prevented RSVP deadline and RSVP settings from persisting when saving the event edit form. The root cause was architectural: the RSVP settings panel is rendered outside the `<form>` tag, so its inputs were never submitted. Additional server-side bugs caused false 400 validation errors and silent data loss.

**Test result:** 19/19 internal packages pass, 0 regressions.

---

## Bugs Fixed

### Bug 1 — Wrong datetime picker pane shown for RSVP deadline

**Symptom:** After using the event datetime picker in range mode (start + end time), opening the RSVP deadline picker showed the wrong content pane. Selecting a date/time wrote to `end_time` instead of `rsvp_deadline_input`.

**Root cause:** `openPanel()` in `datetime_picker.js` only called `switchMode(currentMode)` when `showEndTime = true`. When the event picker left the panel in `end` mode, the RSVP deadline picker (single mode) inherited that state and its clicks were bound to the event instance's `selectedEndDate`.

**Fix:** `openPanel()` now unconditionally calls `switchMode(this.currentMode)` regardless of `showEndTime`.

**Files:** `static/js/datetime_picker.js`

---

### Bug 2 — RSVP settings panel values never submitted with the form

**Symptom:** RSVP deadline, allow maybe RSVP, private guest list, family headcount, allow RSVP after deadline, max plus ones, and event capacity changes were silently discarded on every save.

**Root cause:** The RSVP settings panel is rendered in `{{define "modals"}}` in `event_form.html`, which is outside the `<form>` tag. HTML only submits inputs that are descendants of the form element. None of the panel's inputs (`rsvp_deadline_input`, `allow_maybe_rsvp`, etc.) were ever serialized into the POST body.

**Fix:** Added `injectRSVPSettingsIntoForm()` as a standalone module-level function in `rsvp_settings.js`. It is registered directly on `form.event-form`'s `submit` event inside `initRSVPSettings()`, independently of the `RSVPSettingsPanel` class. On every form submission it:
- Injects all panel values as `type="hidden"` inputs with `data-rsvp-injected` attribute (creates once, updates on repeat submissions)
- Injects a `rsvp_settings_saved=1` sentinel field
- Handles checkboxes correctly (`'on'` when checked, `''` when unchecked)

`RSVPSettingsPanel.handleSave()` also delegates to this function so values are injected immediately when "Save Settings" is clicked.

The submit listener is intentionally placed outside the class so it fires unconditionally even if `RSVPSettingsPanel` fails to initialize (e.g. missing elements).

**Files:** `static/js/rsvp_settings.js`

---

### Bug 3 — Server handler overwrote RSVP settings with empty/false values on every save

**Symptom:** Every event save silently reset `AllowMaybeRSVP`, `PrivateGuestList`, `FamilyHeadcount`, and `AllowRSVPAfterDeadline` to `false`, even when the user never touched RSVP settings.

**Root cause:** `UpdateEventFromForm` in `events_web.go` unconditionally read `allow_maybe_rsvp`, `private_guest_list`, etc. from the form and assigned them directly. Since these were never submitted (Bug 2), `r.FormValue()` returned `""`, evaluating to `!= "on"` = `false`.

**Fix:** The entire RSVP settings block is now gated behind `r.Form.Has("rsvp_settings_saved")`. When the sentinel is absent, existing DB values for all RSVP settings fields are preserved untouched. When present, all submitted values are applied — including clearing `RSVPDeadline` to `nil` when `rsvp_deadline` is empty.

**Files:** `internal/handlers/events_web.go`

---

### Bug 4 — Validator rejected past RSVP deadlines on update (400 error)

**Symptom:** Editing any field on an event whose RSVP deadline had already passed (e.g. a test event created with a near-future deadline) returned `400: validation error on rsvp_deadline: RSVP deadline must be in the future`, even when `rsvp_deadline` was not in the POST body at all.

**Root cause:** `ValidateUpdate` called `validateRSVPDeadline` which required the deadline to be in the future. The handler passed `existing.RSVPDeadline` (from DB) unchanged into validation. An expired deadline that was legitimately set in the past always failed this check.

**Fix:** Added `validateRSVPDeadlineForUpdate` which only validates that the deadline is before the event start time — no future requirement. `ValidateUpdate` now uses this relaxed method. `ValidateCreate` retains the strict version (future + before start time).

This is semantically correct: a published event with an expired RSVP deadline is in a valid state and should remain editable.

**Files:** `internal/events/validator.go`, `internal/events/validator_test.go`

---

## Debugging Notes

### Submit listener not firing (Browser cache)

Initial fix placed the submit listener inside `RSVPSettingsPanel.initRSVPSettings()`. Even after refactoring to a standalone function, the browser (Firefox) was serving a cached version of `rsvp_settings.js` and ignoring the updated file.

Temporary cache-bust `?v=3` added to the script tag in `event_form.html`, confirmed the fix worked, then removed. The browser cache was the only reason the fix appeared non-functional for two iterations.

### Why the listener was inside the class initially

The first attempt attached the submit listener inside `initRSVPSettings()` on `this` (the class instance). If the class guard (`!this.panel || !this.overlay || !this.triggerBtn`) fired early, `initRSVPSettings()` was skipped entirely and no listener was registered. Moving the listener to the module-level `initRSVPSettings()` function made it unconditional.

---

## Files Changed

| File | Change |
|---|---|
| `static/js/datetime_picker.js` | `openPanel()` calls `switchMode(this.currentMode)` unconditionally |
| `static/js/rsvp_settings.js` | `injectRSVPSettingsIntoForm()` standalone function; registered on form submit |
| `internal/handlers/events_web.go` | RSVP settings block gated behind `rsvp_settings_saved` sentinel |
| `internal/events/validator.go` | `validateRSVPDeadlineForUpdate()` — no future requirement for updates |
| `internal/events/validator_test.go` | 3 new test cases for update-path RSVP deadline validation |

---

## Test Results

```
ok  github.com/lenaxia/tinyrsvp/internal/admin
ok  github.com/lenaxia/tinyrsvp/internal/assets
ok  github.com/lenaxia/tinyrsvp/internal/auth
ok  github.com/lenaxia/tinyrsvp/internal/config
ok  github.com/lenaxia/tinyrsvp/internal/db
ok  github.com/lenaxia/tinyrsvp/internal/db/repositories
ok  github.com/lenaxia/tinyrsvp/internal/email
ok  github.com/lenaxia/tinyrsvp/internal/events
ok  github.com/lenaxia/tinyrsvp/internal/handlers
ok  github.com/lenaxia/tinyrsvp/internal/invites
ok  github.com/lenaxia/tinyrsvp/internal/jobs
ok  github.com/lenaxia/tinyrsvp/internal/middleware
ok  github.com/lenaxia/tinyrsvp/internal/models
ok  github.com/lenaxia/tinyrsvp/internal/rsvp
ok  github.com/lenaxia/tinyrsvp/internal/storage
ok  github.com/lenaxia/tinyrsvp/internal/templates
ok  github.com/lenaxia/tinyrsvp/internal/templates/defaults
ok  github.com/lenaxia/tinyrsvp/internal/testutil
ok  github.com/lenaxia/tinyrsvp/internal/testutil/builders
```

19/19 packages pass. 0 regressions.

---

## Known Pre-existing Test Failures (not caused by this work)

- `TestDateTimePickerToggleGroupHiddenOnOpen`
- `TestDateTimePickerRangeModeShowsEndDateInDisplay`
- `TestDateTimePickerSingleModeNoEndDateAfterRangeMode`

All three use `WaitNotVisible('.datetime-picker-panel.open')` which times out because the timezone panel also carries the `.datetime-picker-panel` class. These failures pre-date this session.

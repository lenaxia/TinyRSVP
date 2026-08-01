# Worklog: Wire Up Preference Questions UI

**Date:** 2026-08-01  
**Branch:** `feat/preference-questions-ui`  
**Issue:** #11

## Summary

The preference-questions backend (table, repository, service, API at `/api/events/{id}/questions`) was 100% complete, but the event form UI was gated behind a 🚧 "Coming Soon" overlay with disabled inputs. Wired the UI end-to-end.

## Changes

- **`internal/handlers/events_web.go`**
  - Added `questionService events.QuestionService` to `EventWebHandlers`; threaded through `NewEventWebHandlers`.
  - `EditEventForm` now loads existing questions via `GetQuestions` (was hardcoded to empty).
  - `CreateEventFromForm` creates the submitted questions after event creation (event ID + display order assigned).
  - `UpdateEventFromForm` calls a new `syncQuestions` helper that reconciles the submitted list with persisted questions (update existing, create new, delete removed, apply order). Only runs while the event is draft, matching the service's "cannot modify questions on published event" rule.
  - Added `parseQuestionsFromForm` helper that extracts `questions[N][{id,text,type,required,options}]` form fields.
- **`templates/web/event_form.html`** — removed the "Coming Soon" overlay, enabled all fields, fixed the answer-type select to use the real enum values (`text`/`single_choice`/`multiple_choice` instead of the stale `text`/`choice`), and added hidden `id`, `required` checkbox, and `options` (one-per-line) fields.
- **`static/js/questions.js`** — add/remove question rows, renumbering of field names, and show/hide the options textarea based on selected answer type.
- **`cmd/server/main.go` + `tests/uxserver/server.go`** — pass `questionService` to `NewEventWebHandlers`.

## Tests

- New `internal/handlers/questions_form_test.go`: `parseQuestionsFromForm` happy path + empty-row skipping, and create-flow that verifies questions get the created event's ID.
- Updated `NewEventWebHandlers` callers in tests to pass the new arg.
- Full suite: all 39 non-browser packages pass; `go build`/`go vet` clean.

## Notes

- The JSON API (`/api/events/{id}/questions`) is unchanged; the form now writes through the same `QuestionService`.
- `help_text` (wired in an earlier PR) is not surfaced in this UI yet; options are entered one-per-line.

## Status

**Status:** ✅ Complete  
**Test Pass Rate:** All non-browser tests pass  
**Confidence:** HIGH  
**Production Ready:** Yes

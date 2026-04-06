# Worklog 0154 — Epic 14 Stories 2 & 3: Email Bug Fixes

**Date:** 2026-04-06  
**Session type:** Bug fix  
**Packages changed:** `internal/invites`, `internal/email`, `internal/templates`, `internal/handlers`, `cmd/server`

---

## Context

Following the 2026-04-06 code validation pass (worklog 0153 and the subsequent state-of-repo assessment), Epic 14 was created to track all confirmed bugs. This session implements Stories 02 and 03.

---

## Story 02 — Invite emails render DB template instead of hardcoded plaintext

**File:** `internal/invites/service.go`  
**Root cause:** `SendInvite` unconditionally built a plain-text body via `fmt.Sprintf`. The seeded `TemplateTypeInviteEmail` DB template was never used.

**Investigation:**
- Two separate template systems exist: file-based (`templates/email/`) and DB-backed (`internal/templates/`).
- The file-based renderer only has `rsvp_confirmation` — no invite email.
- The DB system (`templates.Service.GetTemplateForEvent`) is used for the RSVP page and supports per-event customisation. This is the correct long-term path for invite emails too.
- Chose DB path to enable future per-event invite email customisation without rework.

**Changes:**
- Added `RenderEmailTemplate(ctx, eventID, templateType, data)` to `templates.Service` interface and `*service` implementation. Uses `GetTemplateForEvent` (per-event with system default fallback) + `Engine.Parse/ExecuteToString` for both HTML and text.
- `inviteService` gains `templateService templates.Service` field.
- `NewInviteServiceWithTemplates` constructor for production; `NewInviteService` unchanged (nil templateService = plaintext fallback, backward compatible).
- `SendInvite` calls `templateService.RenderEmailTemplate` with `{Event, Invite, RSVPURL, MaxPlusOnes}` matching `defaults/invite_email.html/.txt` variables. Falls back to plaintext silently on nil service or render error.
- Subject updated to `"You're Invited: {EventTitle}"` when event is provided.
- `cmd/server/main.go` uses `NewInviteServiceWithTemplates(tokenGenerator, inviteRepo, templateService)`.
- Generated mock (`mock_template_service.go`) regenerated to include `RenderEmailTemplate`.
- `internal/handlers/templates_test.go` local stub updated with no-op `RenderEmailTemplate`.
- `internal/handlers/invites_send.go` passes `event` in `SendInviteRequest`.
- 3 new tests in `service_send_test.go`: template path, fallback on render error, no-service path.

---

## Story 03 — Confirmation emails show real question text instead of being omitted

**File:** `internal/email/confirmation_service.go`  
**Root cause:** `prepareTemplateData` used `fmt.Sprintf("Question %d", answer.QuestionID)` — question text was never looked up. The test `TestSendConfirmationEmail_WithAnswers` asserted this broken behaviour and passed.

**Investigation and corrections (two rounds):**

**Round 1:**
- Added `questionRepo repositories.QuestionRepository` to `confirmationService`.
- `NewConfirmationServiceWithQuestions` constructor for production.
- `SendConfirmationEmail` looked up question text per-answer via `GetByID` (N queries).
- `prepareTemplateData` received `questionTexts map[int64]string`.
- Initial fallback was "Question N" — corrected to silently omit unlabelled answers instead.

**Round 2 (revalidation caught):**
- N+1 query: replaced N `GetByID` calls with a single `GetByEventID(ctx, event.ID)` call. `event` is always non-nil at both `rsvp/service.go` callsites (fetched from DB, early return on error).
- `mockQuestionRepository.GetByEventID` in tests was returning `nil, nil` regardless of input — fixed to filter by `EventID`, matching production behaviour.
- Two stale comments referencing "Question N" fallback removed.

**Behaviour:**
- Questions for the event fetched in one query and mapped `ID -> QuestionText`.
- Answers with no matching question (deleted after RSVP) are silently omitted.
- If no repo configured or `GetByEventID` errors: all answers omitted, email still sends.
- `{{if .Answers}}` in both email templates handles nil/empty slice correctly.

**Changes:**
- `NewConfirmationServiceWithQuestions` constructor; `NewConfirmationService` unchanged.
- `cmd/server/main.go` uses `NewConfirmationServiceWithQuestions(..., questionRepo)`.
- `TestSendConfirmationEmail_WithAnswers` updated: mock injects real question labels via `GetByEventID`, asserts `"Dietary requirements?"` and `"T-shirt size?"`.
- 2 new tests: `_FallsBackWhenQuestionNotFound` (partial repo hit → 1 answer), `_NoQuestionRepo_OmitsAnswers` (no repo → 0 answers).

---

## Other

- Added `server` binary to `.gitignore` (was previously untracked).
- Epic 14 directory and 6 story files created.
- Epic backlogs (00, 05, 06, 07, 08, 09, 10, 11, 12) updated with code-verified status.

---

## Test results

32/32 non-browser packages pass (`go test -count=1 ./...` excluding `tests/ux`).

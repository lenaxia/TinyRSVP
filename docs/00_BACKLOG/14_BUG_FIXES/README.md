# Epic 14: Bug Fixes & Code Gaps

**Priority:** High  
**Status:** Not Started  
**Target Version:** v0.1  
**Estimated Effort:** 1 week  
**Source:** Code-level validation performed 2026-04-06 — all findings are verified against actual source files and test output, not inferred from documentation.

---

## Overview

This epic tracks all gaps and bugs identified during the 2026-04-06 code validation pass. Every item here has a confirmed file:line reference and a demonstrated failure mode. Nothing in this epic is speculative.

The items are ordered by severity:

1. **Critical** — production security issue (must fix before any public deployment)
2. **High** — visible functional breakage affecting end users
3. **Medium** — dead code / incomplete feature that is silently unreachable
4. **Low** — code quality / structural issues with no user-visible impact

---

## Stories

### Critical

| Story | Title | File | Status |
|-------|-------|------|--------|
| [14_STORY_01](14_STORY_01_remove_test_auth_bypass.md) | Remove X-Test-User-ID auth bypass from production middleware | `internal/middleware/rbac.go:16` | Not Started |

### High — User-Visible Functional Bugs

| Story | Title | File | Status |
|-------|-------|------|--------|
| [14_STORY_02](14_STORY_02_invite_email_template.md) | Render invite email template instead of hardcoded plaintext | `internal/invites/service.go:275` | ✅ Complete |
| [14_STORY_03](14_STORY_03_confirmation_question_names.md) | Look up question text in confirmation emails (not "Question N") | `internal/email/confirmation_service.go:157` | ✅ Complete |
| [14_STORY_04](14_STORY_04_unsubscribe_page.md) | Fix unsubscribe page — add template to rsvpPageTemplates | `cmd/server/main.go:410` | ✅ Complete |

### Medium — Dead Code / Unreachable Features

| Story | Title | File | Status |
|-------|-------|------|--------|
| [14_STORY_05](14_STORY_05_wire_template_editor.md) | Wire template editor routes into router (or delete) | `cmd/server/main.go`, `internal/handlers/router.go` | ✅ Complete |

### Low — Code Quality

| Story | Title | File | Status |
|-------|-------|------|--------|
| [14_STORY_06](14_STORY_06_mock_service_in_production_file.md) | Move MockService out of production package | `internal/email/service.go:13` | Not Started |

---

## Acceptance Criteria (Epic-Level)

- [ ] All 6 stories complete with tests passing
- [ ] Each story updates the relevant epic README to reflect the fix
- [ ] `00_BACKLOG_SUMMARY.md` updated after each story completion
- [ ] No regressions — all 32 packages continue to pass
- [ ] Epic 05 (Email) status updated to ✅ Complete after stories 02, 03, 04
- [ ] Epic 08 (API) status updated after story 01 and 05

---

## Dependencies

**Depends on:** Epics 00–08, 11 (bugs exist in those implementations)  
**Blocks:** Epic 09 (Security audit — story 01 must be done first)

---

## Definition of Done

- [ ] All stories complete
- [ ] All 32 non-browser packages pass
- [ ] Each fixed epic's README updated to ✅ Complete where appropriate
- [ ] `00_BACKLOG_SUMMARY.md` table updated
- [ ] No new `X-Test-User-ID` references in non-test files

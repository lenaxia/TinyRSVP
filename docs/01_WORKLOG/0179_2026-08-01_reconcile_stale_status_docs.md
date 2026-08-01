# Worklog: Reconcile Stale Status Docs (Epics 10, 12, 14, Backlog, README-LLM)

**Date:** 2026-08-01  
**Branch:** `docs/reconcile-stale-status`

## Summary

Code-verified reconciliation of project status docs that had drifted significantly behind the codebase. The backlog summary listed Epic 14 as "Not Started" with X-Test-User-ID as a critical blocker (resolved months ago in worklog 0174); Epic 10's README listed 5 shipped stories as "not implemented"; Epic 12's "0 manual mocks" goal was unachievable due to Go import cycles.

## Work Completed

- [x] **Epic 10 README** — rewrote to reflect code-verified reality: 21/22 stories complete. Moved stories 07, 08, 10, 11, 17 from "Planned/Deferred" to "Completed" (all verified done). Corrected Story 09 to "In Progress (partial)" — status-filter backend works but the frontend controls mis-wire to `/not-implemented`, and search/sort are unimplemented (the only genuine open Epic 10 work). Fixed shifted line refs (G6 router refactor). Added a standalone tech-debt worklog table (G7/G2/G6/G12).
- [x] **Epic 10 story files** — updated status headers in 07, 08, 10, 11 (Not Started → Complete) and 09 (Not Started → In Progress/partial).
- [x] **Epic 12 README** — corrected Story 05's false "`//go:generate mockgen` directives" claim (generation is via `scripts/generate_mocks.sh`); corrected Story 11 count (12 handler files use mocks, 26 still local); reframed Stories 12 & 14 around the Go import-cycle constraint; added an "Architectural Constraint" section explaining why in-package service tests cannot use generated mocks and endorsing the func-field pattern; rescoped the "0 manual mocks" Definition of Done.
- [x] **Backlog summary** (`00_BACKLOG_SUMMARY.md`) — updated the epic status table: Epics 05/08/10/12/14 all corrected; X-Test-User-ID noted as resolved; `/metrics` protection noted. Updated the Epic 14 section from "Not Started" to "✅ Complete" with per-story status. Updated "Next Action".
- [x] **Status assessment** (`PROJECT_STATUS_ASSESSMENT.md`) — added a 2026-08-01 reconciliation banner pointing to the code-verified backlog summary.
- [x] **README-LLM.md** — updated worklog count (176→178), next-entry (0176→0178), Last Updated date, merged-branches table (added #48/#54/#61/#62), and version history (2.3).

## Findings (verified against code)

- **Epic 10:** 21/22 done. Only Story 09 genuinely open.
- **Epic 12:** 16/20 stories done; remaining work is ~10h of handler-mock migration + helper consolidation, plus a documented exception for cycle-constrained packages (not a defect).
- **Epic 14:** fully complete (all 6 stories, worklogs 0174-0176).
- **Production blocker status:** the previously-cited blocker (X-Test-User-ID) is resolved. Public-beta gate is now purely Epic 09 (Security audit).

## Files Changed

- `docs/00_BACKLOG/10_TECHNICAL_DEBT/README.md` — full rewrite
- `docs/00_BACKLOG/10_TECHNICAL_DEBT/10_STORY_{07,08,09,10,11}_*.md` — status header corrections
- `docs/00_BACKLOG/12_TEST_INFRASTRUCTURE/README.md` — corrected stories + import-cycle note
- `docs/00_BACKLOG/00_BACKLOG_SUMMARY.md` — epic table + Epic 14 section + Next Action
- `docs/04_SUMMARIES/PROJECT_STATUS_ASSESSMENT.md` — reconciliation banner
- `README-LLM.md` — worklog count, branches, version history

## Tests

Docs-only changes; `go build ./...` clean; template/static tests pass.

## Status

**Status:** ✅ Complete  
**Confidence:** HIGH — all claims code-verified, not doc assertions

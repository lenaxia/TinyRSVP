# Worklog 0158: AI Workflow Migration to rathena-client Pattern

**Date:** 2026-07-07
**Epic:** N/A (CI/CD infrastructure)
**PR:** #24
**Branch:** `feat/ai-workflow-migration`

---

## Summary

Migrated the GitHub Actions AI-assistant workflow from a monolithic `issue-opened.yml` to a modular, three-workflow architecture adapted from the rathena-client pattern. The new setup separates concerns: issue triage, AI command routing (12 slash commands), and automated PR review.

## Motivation

The previous `issue-opened.yml` attempted to handle issue triage, AI command execution, and prompt assembly all in one file. This made it hard to extend, review, and trust. The rathena-client project had already solved this with a clean separation:
- `issue-opened.yml` — fires only on `issues: [opened]`, does read-only triage
- `ai-comment.yml` — fires on issue/PR comments, routes 12 slash commands via a single shell script
- `pr-review.yml` — fires on PR open/synchronize, runs automated code review

## Changes

### New Workflows (3)

| File | Trigger | Purpose |
|------|---------|---------|
| `.github/workflows/issue-opened.yml` | `issues: [opened]` | Read-only triage; categorize, prioritize, suggest labels |
| `.github/workflows/ai-comment.yml` | `issue_comment`, `pull_request_review_comment` | Route 12 slash commands to command-specific prompts |
| `.github/workflows/pr-review.yml` | `pull_request: [opened, synchronize]` | Automated code review; excludes `renovate[bot]` |

### New Scripts (1)

- **`.github/scripts/route-command.sh`** — Single source of truth for command routing logic. Sourced by `ai-comment.yml`. Detects command token (prefix or inline position), extracts the trailing NOTE text, handles the `--no-merge` hold flag, and assembles the full prompt (context + core-rules + command-specific file). Has a companion test suite (`test_route_command.sh`, 99 assertions).

### New Prompts (17)

All under `.github/prompts/`, customized for TinyRSVP's Go/Chi/SQLite/RSVP domain (not rathena's packet decode):

| File | Role |
|------|------|
| `context.md` | Repository overview, key directories, architecture |
| `core-rules.md` | TDD, type safety, error handling, no-debt rules |
| `code-change-workflow.md` | Branch → TDD → PR → review → merge lifecycle |
| `pr-review.md` | Reviewer rubric (correctness, architecture, tests, security) |
| `issue-responder.md` | Default issue analysis prompt |
| `commands-footer.md` | Slash-command reference table (auto-posted) |
| `analyze.md` | Read-only deep analysis |
| `design.md` | Design doc iteration (always holds merge) |
| `explain.md` | Code/architecture explanation |
| `fix.md` | Bug fix (TDD, auto-merge unless `--no-merge`) |
| `help.md` | Command listing |
| `implement.md` | Feature/story implementation (TDD, auto-merge) |
| `merge.md` | Explicit merge of an approved PR |
| `review.md` | Explicit code review |
| `security.md` | Security-focused review |
| `test.md` | Write/improve tests |
| `triage.md` | Issue triage |

## Authorization

All AI commands are restricted to `OWNER`, `MEMBER`, and `COLLABORATOR` roles (`ai-comment.yml:36-38`). No unauthenticated or public access to AI-driven code changes.

## `--no-merge` Semantics

The `--no-merge` flag is recognized **only** when it is the trailing non-whitespace token of the comment. This eliminates false positives where the literal token appears mid-description. It is stripped from NOTE for every command (so it never pollutes the description), but only acts on the four auto-merging code-change commands: `/fix`, `/implement`, `/test`, `/security`. `/design` always holds via its own prompt; `/merge` is the explicit release and ignores the flag.

## Testing

A comprehensive bash test suite (`test_route_command.sh`, 99 assertions) covers:

- **Command detection (24 cases):** all 12 commands in both prefix and inline positions; bare commands with no arguments; false-positive guards (`/testing` is not `/test`).
- **NOTE extraction (8 cases):** command prefix stripping, whitespace trimming, slashes, question marks, empty notes, inline command notes.
- **`--no-merge` handling (25 cases):** trailing flag sets `HOLD_MERGE=1` on the four code-change commands; stripped from NOTE on all commands; mid-position flag does NOT trigger hold; `/design` holds via its own prompt.
- **Prompt assembly (13 cases):** context.md + core-rules.md always present; command-specific file included; code-change-workflow.md included only for code-change commands; `--no-merge` adds MERGE HOLD directive; `/ai` routes to issue-responder or pr-review based on context.
- **Edge cases (6 cases):** unrecognized text defaults to `/ai`; empty body; unicode preservation.

Run locally:
```bash
.github/scripts/test_route_command.sh
```

## Acceptance Criteria

- [x] Three-workflow separation (issues, AI commands, PR review)
- [x] 12 slash commands routed correctly
- [x] Authorization restricted to OWNER/MEMBER/COLLABORATOR
- [x] `--no-merge` hold flag with trailing-only detection
- [x] `renovate[bot]` excluded from automated PR review
- [x] Prompt files customized for TinyRSVP domain
- [x] route-command.sh has a test suite (99 assertions, all passing)
- [x] Worklog created (this document)

## Related

- **PR:** #24
- **Pattern source:** rathena-client `.github/` directory
- **OpenCode action:** `anomalyco/opencode/github@0cf0294787322664c6d668fa5ab0a9ce26796f78`

## Status

**Status:** ✅ Complete
**Test Pass Rate:** 99/99 (100%)
**Confidence:** HIGH
**Production Ready:** Yes — pending PR review approval

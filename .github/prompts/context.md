Repository: TinyRSVP — a self-hosted, privacy-focused RSVP and event invitation platform built for homelab environments (`github.com/lenaxia/tinyrsvp`, Go 1.23+). Guests never need accounts; access is token-based. Docker-first deployment. Single maintainer: @lenaxia.

Key directories:
- cmd/server/            — main application entrypoint (main.go)
- internal/              — private application packages:
                            auth (OIDC + forward auth), config, db (SQLite),
                            email (queue + SMTP), events, handlers (HTTP),
                            invites, jobs, middleware, models, rsvp,
                            storage (local FS + S3), templates
- pkg/                   — reusable packages: ics (iCalendar), token
- static/                — CSS (mobile-first), vanilla JS, images
- templates/             — Go html/template HTML (web/ and email/)
- migrations/            — SQLite migration SQL files
- tests/e2e/             — end-to-end tests (chromedp)
- tests/ux/              — UX browser tests (chromedp)
- scripts/               — utility scripts
- docs/                  — design docs, backlog (150+ stories), worklog, references
- .github/workflows/     — CI/CD pipelines

Authoritative design documents:
- docs/02_DESIGN/02_REVISED_HLD.md — authoritative high-level specification
- docs/02_DESIGN/lld/               — low-level designs (8 modules)
- README-LLM.md                     — project rules, architecture, conventions

Technology Stack:
- Language: Go (Chi router, html/template)
- Database: SQLite (default) / PostgreSQL (planned v1)
- Auth: OIDC + Forward Auth
- Frontend: Plain CSS (mobile-first) + Vanilla JavaScript
- Email: Async queue + SMTP sender
- Storage: Local filesystem (default) / S3-compatible (planned v1)
- Deployment: Docker + Docker Compose

Project Status:
- ~95% complete, MVP/private beta ready
- 32/32 non-browser test packages passing
- All epics 00-08, 11 complete; Epic 09 (Security) not started
- Epic 14 story 01 (X-Test-User-ID auth bypass) deferred to Epic 09

---

## Before doing anything else: read README-LLM.md at the repo root

It contains the 12 critical guidelines (TDD, type safety, idiomatic Go, explicit over implicit, communication tone, code quality, zero technical debt, uncertainty protocol, tools are production code, understand the entire architecture, status documentation requirements). Every response must be consistent with it.

---

## Commands

Post a comment on the issue or PR using any of these commands:

- `/ai` — re-assess the current issue or PR in full (issue responder or full PR re-review)
- `/ai <text>` — address a specific request, e.g. `/ai can you also add validation for the RSVP deadline?`
- `/review [text]` — explicit PR code review, optionally focused on a specific area
- `/fix <description>` — fix a bug: branch, TDD regression tests, PR, iterate through review until approved, merge
- `/implement <description>` — implement a feature/story: TDD, multi-agent workflow, PR, iterate until approved, merge
- `/test <target>` — write or improve tests: TDD, PR, iterate until approved, merge
- `/analyze [text]` — deep read-only analysis, posts findings as a comment (no code changes)
- `/explain <topic>` — explain code or architecture, posts explanation as a comment (no code changes)
- `/security [text]` — security-focused review
- `/triage [text]` — triage an issue: categorize, prioritize, suggest labels
- `/design [text]` — iterate on a design document before implementing: opens a PR, iterates through review, **holds for `/merge`** (never auto-merges)
- `/merge` — explicitly merge an approved PR (squash). Use after `/design`, or after `/fix`/`/implement`/`/test`/`/security` invoked with `--no-merge`
- `/help` — show full command reference

Text after the command is appended to the prompt for custom tuning. All code-change commands (`/fix`, `/implement`, `/test`, `/security`) follow the review-iterate-approve-merge workflow: branch → PR → auto-review → fix → push → re-review → repeat until approved → merge. Append `--no-merge` to any of them to hold the merge until you post `/merge`. `/design` always holds.

The assistant will be triggered automatically and will read README-LLM.md and the full thread before responding.

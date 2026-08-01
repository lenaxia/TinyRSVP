# Epic 10: Technical Debt & Improvements

**Status:** Active — 21 of 22 stories complete  
**Last Verified:** 2026-08-01 (code-level, not doc assertions)  
**Priority:** Medium  
**Owner:** LLM Implementation Team

---

## Overview

Epic 10 is reserved for technical debt, improvements, and issues that don't fit cleanly into other feature epics. This includes:

- Gaps identified during validation of completed stories
- Performance optimizations
- Refactoring opportunities
- Cross-cutting concerns
- Non-blocking improvements to existing functionality

---

## Stories

### Completed ✅ (21)

- [10_STORY_01](10_STORY_01_consistent_navigation.md) — Consistent navigation across all pages
- [10_STORY_02](10_STORY_02_system_user_exclusion.md) — Exclude system user from user management
- [10_STORY_03](10_STORY_03_invite_management_ui.md) — Invite management UI with modals
- [10_STORY_04](10_STORY_04_reduce_ui_padding.md) — Reduce excessive padding and spacing
- [10_STORY_05](10_STORY_05_reusable_datetime_picker.md) — Reusable datetime picker component _(enhanced 2026-07-24: end-time opt-in checkbox, end-before-start prevention — worklog 0177)_
- [10_STORY_06](10_STORY_06_oidc_return_url.md) — Return URL preservation in OIDC flow _(short-lived cookie; 8 tests)_
- [10_STORY_07](10_STORY_07_event_list_stats.md) — Event list stats display _(verified: `EventWithStats` struct + `ListWithStats` repo; rendered in `event_list.html`; worklog 0160)_
- [10_STORY_08](10_STORY_08_dashboard_clickable_events.md) — Dashboard clickable activity items _(verified: `ActivityItem.EventID`; worklog 0160)_
- [10_STORY_10](10_STORY_10_admin_settings_page.md) — Admin settings page _(verified: `/admin/settings` route + template + redaction; worklog 0175a)_
- [10_STORY_11](10_STORY_11_admin_metrics_dashboard.md) — Admin metrics dashboard _(verified: `/admin/metrics` + `AdminMetricsStats`; worklog 0175a)_
- [10_STORY_12](10_STORY_12_theme_switching.md) — Light/dark theme switching
- [10_STORY_13](10_STORY_13_admin_theme_integration.md) — Admin template theme integration
- [10_STORY_14](10_STORY_14_rsvp_summary_template_fix.md) — RSVP summary template structure fix
- [10_STORY_15](10_STORY_15_auth_test_expectations.md) — Auth test expectations (303 redirect)
- [10_STORY_16](10_STORY_16_auth_test_compilation.md) — Auth test compilation fix
- **10_STORY_17** — `X-Test-User-ID` auth bypass removed from production `RequireAuth`; moved to test-only `TestRequireAuth` wrapper (`internal/middleware/rbac_test_bypass.go`); 5 regression tests prove production rejects the header _(worklog 0174; was previously listed as "deferred to Epic 09" — that note was stale)_
- **10_STORY_18** — `MockService` removed from `internal/email/service.go`; generated `MockEmailService` at `internal/testutil/mocks/services/mock_email_service.go`
- **10_STORY_19** — Template editor wired into router via `TemplateEditorHandlers.RegisterRoutes` (`internal/handlers/router.go:270`); constructed in `cmd/server/main.go:599,712` _(line numbers shifted after the G6 router refactor, worklog 0171)_
- **10_STORY_20** — Invite email renders `models.TemplateTypeInviteEmail` (`internal/invites/service.go:308`)
- **10_STORY_21** — Confirmation email uses `questionTexts map[int64]string` (`internal/email/confirmation_service.go:152`)
- **10_STORY_22** — `unsubscribe.html` registered in `rsvpPageTemplates` (`cmd/server/main.go:423`); rendered by `internal/handlers/rsvp.go:955`

### In Progress ⚠️ (1)

- [10_STORY_09](10_STORY_09_event_filtering_sorting.md) — Event list filtering and sorting. **Backend `?status=` works** (`events_web.go:106` parses status → `ListFilters.Status`), but the frontend controls are mis-wired: the status `<select>`, search input, and sort `<select>` all navigate to `/not-implemented` (`event_list.html:48,64,70`). **Search and sort are entirely unimplemented** (`ListFilters` has no `Search`/`SortBy`/`SortOrder` fields). This is the **only genuinely open Epic 10 work**.

### Analysis Documents
- [10_ANALYSIS_image_templates_wysiwyg.md](10_ANALYSIS_image_templates_wysiwyg.md) — Image templates & WYSIWYG editor analysis

---

## Standalone Tech-Debt Work (not formal stories)

The following audit-driven cleanups (tagged `G#`) were done as standalone PRs and are tracked here for visibility. They have no `10_STORY_XX` file.

| Worklog | Tag | What | Status |
|---|---|---|---|
| 0169 | G7 | Replace hand-rolled `/api/users` routing with chi routes | ✅ |
| 0170 | G2 | Dashboard stats via single SQL aggregation | ✅ |
| 0171 | G6 | Split `NewRouter` monolith into `router_setup.go` (shifted line refs for stories 19/22) | ✅ |
| 0172 | G12 | Remove dead HMAC methods from `ConfigRepository` | ✅ |

---

## Principles

1. **Non-Blocking**: Stories in Epic 10 should not block other epic completion
2. **Quality Focus**: Improvements that enhance code quality, maintainability, or performance
3. **User Impact**: Consider user experience improvements even if not critical
4. **Technical Excellence**: Address architectural concerns and technical debt

---

## Adding Stories to Epic 10

When validation identifies gaps or improvements:

1. Create story file: `10_STORY_XX_description.md`
2. Follow standard user story template
3. Link to the epic/story where gap was identified
4. Add to "In Progress" list above
5. Prioritize based on impact and effort

---

## References

- **HLD:** [docs/02_REVISED_HLD.md](../02_REVISED_HLD.md)
- **Backlog:** [docs/00_BACKLOG/](.)

**Note:** Statuses above were verified against actual source on 2026-08-01, not against prior documentation. Prior versions of this README listed stories 07/08/10/11 as "not implemented" and 17 as "deferred" — all were already complete in code.

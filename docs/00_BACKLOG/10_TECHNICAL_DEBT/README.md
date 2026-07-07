# Epic 10: Technical Debt & Improvements

**Status:** Active  
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

### Completed ✅
- [10_STORY_01_consistent_navigation.md](10_STORY_01_consistent_navigation.md) - Consistent navigation across all pages
- [10_STORY_02_system_user_exclusion.md](10_STORY_02_system_user_exclusion.md) - Exclude system user from user management
- [10_STORY_03_invite_management_ui.md](10_STORY_03_invite_management_ui.md) - Invite management UI with modals
- [10_STORY_04_reduce_ui_padding.md](10_STORY_04_reduce_ui_padding.md) - Reduce excessive padding and spacing
- [10_STORY_05_reusable_datetime_picker.md](10_STORY_05_reusable_datetime_picker.md) - Reusable datetime picker component
- [10_STORY_06_oidc_return_url.md](10_STORY_06_oidc_return_url.md) - Return URL preservation in OIDC flow _(implemented via short-lived cookie; 8 tests added)_
- [10_STORY_09_event_filtering_sorting.md](10_STORY_09_event_filtering_sorting.md) - Event list filtering and sorting _(code-verified: handler + template both support status filter)_
- [10_STORY_12_theme_switching.md](10_STORY_12_theme_switching.md) - Light/dark theme switching _(code-verified: `theme_controller.js` + `theme_toggle.css` + toggle button implemented)_
- **10_STORY_18** - `MockService` removed from production `internal/email/service.go`; generated `MockEmailService` lives at `internal/testutil/mocks/services/mock_email_service.go` _(verified 2026-07-07; README example updated)_
- **10_STORY_19** - Template editor wired into router via `TemplateEditorHandlers.RegisterRoutes` at `internal/handlers/router.go:559`; constructed and templated in `cmd/server/main.go:571-572` _(verified 2026-07-07)_
- **10_STORY_20** - Invite email renders `models.TemplateTypeInviteEmail` at `internal/invites/service.go:308` _(verified 2026-07-07)_
- **10_STORY_21** - Confirmation email uses `questionTexts map[int64]string` resolved from question IDs at `internal/email/confirmation_service.go:152` _(verified 2026-07-07)_
- **10_STORY_22** - `templates/web/unsubscribe.html` registered in `rsvpPageTemplates` at `cmd/server/main.go:422`; rendered by `internal/handlers/rsvp.go:955` _(verified 2026-07-07)_

### In Progress
- _(none)_

### Planned
- [10_STORY_07_event_list_stats.md](10_STORY_07_event_list_stats.md) - Event list stats display (invite/RSVP counts) _(not implemented: event list template has no InviteCount/RSVPCount fields)_
- [10_STORY_08_dashboard_clickable_events.md](10_STORY_08_dashboard_clickable_events.md) - Dashboard recent events clickable links _(not implemented: `ActivityItem` struct has no EventID or URL field; dashboard activity items are plain text)_
- [10_STORY_10_admin_settings_page.md](10_STORY_10_admin_settings_page.md) - Admin settings page _(not implemented: no template, no route in router.go)_
- [10_STORY_11_admin_metrics_dashboard.md](10_STORY_11_admin_metrics_dashboard.md) - Admin metrics dashboard _(not implemented)_
- [10_STORY_13_admin_theme_integration.md](10_STORY_13_admin_theme_integration.md) - Admin template theme integration
- [10_STORY_14_rsvp_summary_template_fix.md](10_STORY_14_rsvp_summary_template_fix.md) - RSVP summary template structure fix
- [10_STORY_15_auth_test_expectations.md](10_STORY_15_auth_test_expectations.md) - Auth test expectations fix
- [10_STORY_16_auth_test_compilation.md](10_STORY_16_auth_test_compilation.md) - Auth test compilation fix

### Deferred to Epic 09 (Security)
- **10_STORY_17**: `X-Test-User-ID` auth bypass in `internal/middleware/rbac.go:16` — deferred per Epic 09 pre-finding; intentionally left in place for now (UX tests depend on it)

### Analysis Documents
- [10_ANALYSIS_image_templates_wysiwyg.md](10_ANALYSIS_image_templates_wysiwyg.md) - Image templates & WYSIWYG editor analysis

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
4. Add to "Planned" list above
5. Prioritize based on impact and effort

---

## References

- **HLD:** [docs/02_REVISED_HLD.md](../02_REVISED_HLD.md)
- **Backlog:** [docs/00_BACKLOG/](.)

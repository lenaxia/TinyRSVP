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

### In Progress
- _(none)_

### Planned
- [10_STORY_06_oidc_return_url.md](10_STORY_06_oidc_return_url.md) - Return URL preservation in OIDC flow
- [10_STORY_07_event_list_stats.md](10_STORY_07_event_list_stats.md) - Event list stats display (invite/RSVP counts)
- [10_STORY_08_dashboard_clickable_events.md](10_STORY_08_dashboard_clickable_events.md) - Dashboard recent events clickable links
- [10_STORY_09_event_filtering_sorting.md](10_STORY_09_event_filtering_sorting.md) - Event list filtering and sorting
- [10_STORY_10_admin_settings_page.md](10_STORY_10_admin_settings_page.md) - Admin settings page
- [10_STORY_11_admin_metrics_dashboard.md](10_STORY_11_admin_metrics_dashboard.md) - Admin metrics dashboard
- [10_STORY_12_theme_switching.md](10_STORY_12_theme_switching.md) - Light/dark theme switching with real-time toggle

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

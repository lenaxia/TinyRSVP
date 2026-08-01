# Epic 10: Technical Debt & Improvements
## Story 02: Admin Settings Page

### User Story
As an admin, I want to access a settings page from the admin dashboard so that I can configure system-wide settings.

### Problem
The admin dashboard has a link to `/admin/settings` but this route returns a 404 error. There is no settings page or handler implemented.

### Acceptance Criteria
- [ ] Create `/admin/settings` route in router
- [ ] Create settings handler with appropriate admin authorization
- [ ] Create settings page template
- [ ] Implement basic settings display (read-only initially)
- [ ] Add tests for settings route and handler
- [ ] Settings page should be accessible only to admins

### Technical Notes
- Route should be added in `internal/handlers/router.go`
- Handler should be created in `internal/handlers/` (e.g., `settings.go`)
- Template should be created in `templates/web/`
- Initial implementation can be read-only display of current configuration
- Future enhancements can add edit capabilities

### Status
- Status: ✅ Complete (verified 2026-08-01: `/admin/settings` route + template + secret redaction; worklog 0175a)
- Priority: Medium
- Assigned: Unassigned
- Created: 2026-01-10

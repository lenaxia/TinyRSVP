# Epic 10: Technical Debt & Improvements
## Story 03: Exclude System User from User Management and Stats

### User Story
As an admin, I want the system user to be excluded from user management pages and user count statistics so that I only see actual human users.

### Problem
1. The system user (if it exists) appears in the user management page where admins can potentially modify it
2. The system user counts towards the total user count in admin statistics
3. System users should be internal-only and not exposed to admin UI

### Acceptance Criteria
- [ ] Identify if a system user exists in the codebase
- [ ] Exclude system user from `ListUsers` queries in user management
- [ ] Exclude system user from `CountUsers` in admin stats
- [ ] Add filter to user repository/service to exclude system users
- [ ] Update tests to verify system user exclusion
- [ ] Document what constitutes a "system user" (e.g., specific email pattern, role, or flag)

### Technical Notes
- Need to determine how system users are identified (email pattern like `system@*`, special role, or dedicated flag)
- May need to add a `is_system` boolean column to users table if not already present
- Update `internal/handlers/users.go` ListUsers and CountUsers methods
- Update `internal/admin/service.go` GetAdminStats method
- Consider if system user should be completely hidden or just marked as non-editable

### Investigation Needed
- Check if system user actually exists in current implementation
- Determine identification mechanism for system users
- Review if any seeding or initialization creates system users

### Status
- Status: Not Started
- Priority: Medium
- Assigned: Unassigned
- Created: 2026-01-10

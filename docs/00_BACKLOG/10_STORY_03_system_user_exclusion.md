# Epic 10: Technical Debt & Improvements
## Story 03: Exclude System User from User Management and Stats

### User Story
As an admin, I want the system user to be excluded from user management pages and user count statistics so that I only see actual human users.

### Problem
1. The system user (if it exists) appears in the user management page where admins can potentially modify it
2. The system user counts towards the total user count in admin statistics
3. System users should be internal-only and not exposed to admin UI

### Acceptance Criteria
- [x] Identify if a system user exists in the codebase
- [x] Exclude system user from `ListUsers` queries in user management
- [x] Exclude system user from `CountUsers` in admin stats
- [x] Add filter to user repository/service to exclude system users
- [x] Update tests to verify system user exclusion
- [x] Document what constitutes a "system user" (e.g., specific email pattern, role, or flag)

### Implementation Summary

**System User Identification:**
- System user is identified by email: `system@tinyrsvp.local`
- Created during application bootstrap in `cmd/server/main.go`
- Used for seeding default templates and background jobs

**Changes Made:**
1. Added `SystemUserEmail` constant in `internal/models/user.go`
2. Added `User.IsSystem()` helper method for identification
3. Updated `UserRepository.List()` to exclude system user via WHERE clause
4. Updated `UserRepository.Count()` to exclude system user via WHERE clause
5. Added comprehensive tests for system user exclusion

**Impact:**
- Admin dashboard stats now exclude system user from total count
- User management UI no longer displays system user
- System user remains in database for internal operations
- All existing tests pass with new exclusion logic

### Status
- Status: Complete
- Priority: Medium
- Assigned: LLM
- Created: 2026-01-10
- Completed: 2026-01-10

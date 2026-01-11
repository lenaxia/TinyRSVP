# Story 10.15: Auth Test Expectations Fix

**Epic:** 10 - Technical Debt & Improvements  
**Priority:** Low  
**Status:** Not Started  
**Identified:** 2026-01-11 (Epic 11 Phase 1 Validation)

---

## User Story

As a **developer**, I want **auth tests to have correct expectations** so that **the test suite passes cleanly**.

---

## Problem Statement

During Epic 11 Phase 1 validation, several auth-related tests fail because they expect HTTP 401 Unauthorized but the application returns HTTP 303 See Other (redirect to login). Both behaviors are valid, but tests need to match actual implementation.

---

## Failing Tests

1. `TestMain_RouterIntegration/events_list_requires_auth`
   - Expected: 401
   - Actual: 303 redirect to `/login?return=/api/events`

2. `TestMain_RouterIntegration_AuthenticatedRoutes/unauthenticated_events_list`
   - Expected: 401
   - Actual: 303 redirect

3. `TestMain_RouterIntegration_AuthenticatedRoutes/unauthenticated_users_list`
   - Expected: 401
   - Actual: 303 redirect

4. `TestMain_RouterIntegration_AuthenticatedRoutes/unauthenticated_templates_list`
   - Expected: 401
   - Actual: 303 redirect

5. `TestMain_RouterIntegration_MiddlewareChain`
   - Expected: 401
   - Actual: 303 redirect

6. `TestForwardAuthFlow/protected_endpoint_without_auth`
   - Expected: 401
   - Actual: 303 redirect

7. `TestForwardAuthFlow/protected_endpoint_with_invalid_session`
   - Expected: 401
   - Actual: 303 redirect

8. `TestInviteEndpointExists/unauthorized_request_fails`
   - Expected: 401
   - Actual: 303 redirect

---

## Acceptance Criteria

- [ ] All auth tests pass
- [ ] Test expectations match actual behavior
- [ ] Decision documented: 303 redirect vs 401 response
- [ ] Consistent behavior across all endpoints

---

## Technical Approach

### Option 1: Update Tests (Recommended)

Update tests to accept 303 redirects:
```go
if status != http.StatusSeeOther {
    t.Errorf("Expected redirect to login, got %d", status)
}
if !strings.Contains(location, "/login?return=") {
    t.Errorf("Expected redirect to login with return URL")
}
```

**Pros:**
- No code changes needed
- Current behavior is user-friendly (redirects to login)
- Preserves return URL for post-login redirect

**Cons:**
- None

### Option 2: Change Auth Middleware

Change middleware to return 401 for API endpoints:
```go
if isAPIEndpoint(r.URL.Path) {
    http.Error(w, "Unauthorized", http.StatusUnauthorized)
    return
}
http.Redirect(w, r, "/login?return="+r.URL.Path, http.StatusSeeOther)
```

**Pros:**
- RESTful API behavior (401 for API calls)
- Tests pass without changes

**Cons:**
- Code changes required
- Need to define what's an API endpoint
- May break existing behavior

---

## Recommendation

**Use Option 1: Update Tests**

The current 303 redirect behavior is correct and user-friendly. Tests should be updated to match actual implementation.

---

## Tasks

- [ ] Review all failing auth tests
- [ ] Update test expectations to accept 303
- [ ] Verify redirect URLs are correct
- [ ] Verify return URL preservation works
- [ ] Run all auth tests
- [ ] Document decision in test comments

---

## Testing Requirements

### Unit Tests
- [ ] All auth middleware tests pass
- [ ] All router integration tests pass
- [ ] All forward auth flow tests pass

### Integration Tests
- [ ] Auth flow works end-to-end
- [ ] Return URL preserved after login
- [ ] Unauthorized access properly handled

---

## Dependencies

**Prerequisites:**
- None (can be done independently)

**Blocks:**
- Clean test suite

---

## Notes

- This is a test-only fix, no functional changes
- Current behavior is correct
- Tests were written with wrong expectations
- Low priority since functionality works

---

## Estimated Effort

**Size:** Small (30 minutes)
- Update test expectations
- Verify redirects work
- Run tests

---

**Status:** Ready for Implementation

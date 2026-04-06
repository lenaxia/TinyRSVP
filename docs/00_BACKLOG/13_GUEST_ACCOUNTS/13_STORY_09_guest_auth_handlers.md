# User Story: Guest Auth HTTP Handlers

**Epic:** [Epic 13 - Guest Accounts & Encryption at Rest](README.md)  
**Priority:** High  
**Status:** Not Started  
**Estimated Effort:** 6 hours  

---

## User Story

As a **guest**, I want **to request an OTP, verify it, view my linked invitations, and delete my account via HTTP endpoints** so that **I can manage my optional guest account through a browser**.

---

## Acceptance Criteria

- [ ] `POST /guest/auth/request-otp` accepts `identifier` form field, calls `GuestAccountService.RequestOTP`, returns 200 or error
- [ ] `POST /guest/auth/verify-otp` accepts `identifier` and `code` form fields, calls `GuestAccountService.VerifyOTP`, sets `tinyrsvp_guest` cookie on success
- [ ] `POST /guest/auth/logout` deletes the guest session and clears the cookie
- [ ] `GET /guest/account` renders the guest's linked invitations (requires `RequireGuestAuth` middleware)
- [ ] `DELETE /guest/account` deletes the guest account and clears the cookie
- [ ] All routes registered under `/guest/` prefix in the router, completely separate from staff routes
- [ ] `ErrRateLimitExceeded` mapped to HTTP 429
- [ ] `ErrInvalidOTP` / `ErrOTPExpired` mapped to HTTP 422 with user-friendly message
- [ ] All responses use `HandleError(w, r, err)` for error handling (no custom error writers)
- [ ] Handler tests use generated mocks for `GuestAccountService` and `GuestSessionManager`
- [ ] All tests pass with timeout

---

## Technical Details

### Handler File

```
internal/handlers/guest_auth.go
internal/handlers/guest_auth_test.go
```

### Handler Struct

```go
type GuestAuthHandler struct {
    service    guestauth.GuestAccountService
    sessionMgr guestauth.GuestSessionManager
}

func NewGuestAuthHandler(
    service guestauth.GuestAccountService,
    sessionMgr guestauth.GuestSessionManager,
) *GuestAuthHandler
```

### Routes (registered in `internal/handlers/router.go` or equivalent)

```
POST   /guest/auth/request-otp   → GuestAuthHandler.HandleRequestOTP
POST   /guest/auth/verify-otp    → GuestAuthHandler.HandleVerifyOTP
POST   /guest/auth/logout        → GuestAuthHandler.HandleLogout        [RequireGuestAuth]
GET    /guest/account            → GuestAuthHandler.HandleAccount        [RequireGuestAuth]
DELETE /guest/account            → GuestAuthHandler.HandleDeleteAccount  [RequireGuestAuth]
```

### HandleRequestOTP

```go
func (h *GuestAuthHandler) HandleRequestOTP(w http.ResponseWriter, r *http.Request) {
    identifier := strings.TrimSpace(r.FormValue("identifier"))
    if identifier == "" {
        HandleError(w, r, &ValidationError{Field: "identifier", Message: "email or phone is required"})
        return
    }

    if err := h.service.RequestOTP(r.Context(), identifier); err != nil {
        HandleError(w, r, err)
        return
    }

    // Render "check your email/phone" confirmation page or 200 JSON
}
```

### HandleVerifyOTP

```go
func (h *GuestAuthHandler) HandleVerifyOTP(w http.ResponseWriter, r *http.Request) {
    identifier := strings.TrimSpace(r.FormValue("identifier"))
    code := strings.TrimSpace(r.FormValue("code"))

    session, err := h.service.VerifyOTP(r.Context(), identifier, code)
    if err != nil {
        HandleError(w, r, err)
        return
    }

    if err := h.sessionMgr.SetSessionCookie(w, session.ID); err != nil {
        HandleError(w, r, err)
        return
    }

    http.Redirect(w, r, "/guest/account", http.StatusSeeOther)
}
```

### Error Mapping

`ErrRateLimitExceeded` must be wrapped as an HTTP 429 type error. `ErrInvalidOTP` and `ErrOTPExpired` must map to 422. Implement custom error types in `internal/guestauth/errors.go` that satisfy the existing error-typing conventions used by `HandleError`.

```go
type RateLimitError struct{ Message string }
func (e *RateLimitError) Error() string { return e.Message }
func (e *RateLimitError) StatusCode() int { return http.StatusTooManyRequests }

type OTPValidationError struct{ Message string }
func (e *OTPValidationError) Error() string { return e.Message }
func (e *OTPValidationError) StatusCode() int { return http.StatusUnprocessableEntity }
```

---

## Tasks

### Phase 1: HandleRequestOTP (TDD)
- [ ] Write test: `TestHandleRequestOTP_MissingIdentifier` — returns 400
- [ ] Write test: `TestHandleRequestOTP_RateLimited` — service returns rate limit error, handler returns 429
- [ ] Write test: `TestHandleRequestOTP_Success` — returns 200
- [ ] Run tests (should fail)
- [ ] Implement `HandleRequestOTP`
- [ ] Run tests (should pass)

### Phase 2: HandleVerifyOTP (TDD)
- [ ] Write test: `TestHandleVerifyOTP_InvalidCode` — returns 422
- [ ] Write test: `TestHandleVerifyOTP_ExpiredCode` — returns 422
- [ ] Write test: `TestHandleVerifyOTP_Success` — sets cookie, redirects to `/guest/account`
- [ ] Run tests (should fail)
- [ ] Implement `HandleVerifyOTP`
- [ ] Run tests (should pass)

### Phase 3: HandleLogout and Account (TDD)
- [ ] Write test: `TestHandleLogout_DeletesSession`
- [ ] Write test: `TestHandleAccount_ListsInvites`
- [ ] Write test: `TestHandleDeleteAccount_RemovesAccount`
- [ ] Run tests (should fail)
- [ ] Implement `HandleLogout`, `HandleAccount`, `HandleDeleteAccount`
- [ ] Run tests (should pass)

### Phase 4: Router Registration
- [ ] Register all guest routes in the router
- [ ] Confirm `/guest/*` routes do not interfere with any existing staff routes
- [ ] Run full test suite

---

## Testing Requirements

```go
func TestHandleRequestOTP_RateLimited(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockSvc := mocksvcs.NewMockGuestAccountService(ctrl)
    mockSvc.EXPECT().
        RequestOTP(gomock.Any(), "alice@example.com").
        Return(&guestauth.RateLimitError{Message: "too many requests"})

    handler := handlers.NewGuestAuthHandler(mockSvc, nil)

    r := httptest.NewRequest(http.MethodPost, "/guest/auth/request-otp",
        strings.NewReader("identifier=alice%40example.com"))
    r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    r.Header.Set("Accept", "application/json")
    w := httptest.NewRecorder()

    handler.HandleRequestOTP(w, r)

    if w.Code != http.StatusTooManyRequests {
        t.Errorf("expected 429, got %d", w.Code)
    }
}

func TestHandleVerifyOTP_Success(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    session := &models.GuestSession{
        ID:        "sess-abc",
        ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
    }

    mockSvc := mocksvcs.NewMockGuestAccountService(ctrl)
    mockSvc.EXPECT().
        VerifyOTP(gomock.Any(), "alice@example.com", "123456").
        Return(session, nil)

    mockMgr := mocksvcs.NewMockGuestSessionManager(ctrl)
    mockMgr.EXPECT().
        SetSessionCookie(gomock.Any(), "sess-abc").
        Return(nil)

    handler := handlers.NewGuestAuthHandler(mockSvc, mockMgr)

    body := strings.NewReader("identifier=alice%40example.com&code=123456")
    r := httptest.NewRequest(http.MethodPost, "/guest/auth/verify-otp", body)
    r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    w := httptest.NewRecorder()

    handler.HandleVerifyOTP(w, r)

    if w.Code != http.StatusSeeOther {
        t.Errorf("expected 303 redirect, got %d", w.Code)
    }
    if loc := w.Header().Get("Location"); loc != "/guest/account" {
        t.Errorf("expected redirect to /guest/account, got %q", loc)
    }
}
```

---

## Dependencies

**Depends on:** Story 08 (guestauth package), Story 10 (RequireGuestAuth middleware — can be developed in parallel, wired after)  
**Blocks:** Story 11 (RSVP opt-in prompt references guest auth routes)

---

## Definition of Done

- [ ] All acceptance criteria met
- [ ] All tasks completed
- [ ] All tests pass: `go test -timeout 30s -race ./internal/handlers/...`
- [ ] All error responses go through `HandleError`
- [ ] No direct `w.WriteHeader` or `json.NewEncoder(w).Encode` calls bypassing `HandleError`
- [ ] `go vet ./...` clean

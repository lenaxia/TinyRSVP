# User Story: `RequireGuestAuth` Middleware

**Epic:** [Epic 13 - Guest Accounts & Encryption at Rest](README.md)  
**Priority:** High  
**Status:** Not Started  
**Estimated Effort:** 3 hours  

---

## User Story

As a **developer**, I want **a `RequireGuestAuth` middleware that validates the `tinyrsvp_guest` session cookie and injects the authenticated `*models.GuestAccount` into the request context** so that **guest-only routes can identify the caller without duplicating session lookup logic**.

---

## Acceptance Criteria

- [ ] `RequireGuestAuth` middleware implemented in `internal/middleware/guest.go`
- [ ] Reads `tinyrsvp_guest` cookie from request; returns 401 if absent
- [ ] Looks up session via `GuestSessionManager.GetSession`; returns 401 if not found or expired
- [ ] Injects `*models.GuestAccountView` into context via a typed context key distinct from the staff `userContextKey`
- [ ] `GuestAccountFromContext(ctx)` helper retrieves the injected account
- [ ] A missing or expired session on a non-JSON request redirects to `/guest/auth/login` instead of returning 401
- [ ] `RequireGuestAuth` never satisfies `RequireAuth` — the two middleware use completely separate context keys and cannot be confused
- [ ] Middleware is only applied to `/guest/account` routes (not to request-otp or verify-otp)
- [ ] All tests pass with timeout

---

## Technical Details

### File

```
internal/middleware/guest.go
internal/middleware/guest_test.go
```

### Context Key

```go
type guestContextKey string

const guestAccountContextKey guestContextKey = "guest_account"

func WithGuestAccount(ctx context.Context, account *models.GuestAccountView) context.Context {
    return context.WithValue(ctx, guestAccountContextKey, account)
}

func GuestAccountFromContext(ctx context.Context) (*models.GuestAccountView, bool) {
    account, ok := ctx.Value(guestAccountContextKey).(*models.GuestAccountView)
    return account, ok
}
```

This key type is unexported and distinct from the `contextKey` type used in `internal/auth/auth.go`, making cross-contamination a compile-time impossibility.

### Middleware Signature

```go
func RequireGuestAuth(
    sessionMgr guestauth.GuestSessionManager,
    accountSvc guestauth.GuestAccountService,
) func(http.Handler) http.Handler
```

### Logic

```go
func RequireGuestAuth(sessionMgr guestauth.GuestSessionManager, accountSvc guestauth.GuestAccountService) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            sessionID, err := sessionMgr.GetSessionFromRequest(r)
            if err != nil {
                handleUnauthenticated(w, r)
                return
            }

            session, err := sessionMgr.GetSession(r.Context(), sessionID)
            if err != nil || session.IsExpired() {
                sessionMgr.ClearSessionCookie(w)
                handleUnauthenticated(w, r)
                return
            }

            account, err := accountSvc.GetAccount(r.Context(), session.GuestAccountID)
            if err != nil {
                handleUnauthenticated(w, r)
                return
            }

            ctx := WithGuestAccount(r.Context(), account)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

func handleUnauthenticated(w http.ResponseWriter, r *http.Request) {
    if r.Header.Get("Accept") == "application/json" {
        HandleError(w, r, &UnauthorizedError{})
        return
    }
    http.Redirect(w, r, "/guest/auth/login", http.StatusSeeOther)
}
```

---

## Tasks

### Phase 1: Middleware (TDD)
- [ ] Write test: `TestRequireGuestAuth_NoCookie` — returns 303 redirect (HTML) or 401 (JSON)
- [ ] Write test: `TestRequireGuestAuth_InvalidSession` — session not found, returns redirect
- [ ] Write test: `TestRequireGuestAuth_ExpiredSession` — expired session, clears cookie, returns redirect
- [ ] Write test: `TestRequireGuestAuth_Valid` — injects account into context, calls next handler
- [ ] Write test: `TestRequireGuestAuth_DoesNotSatisfyRequireAuth` — staff middleware context key not set
- [ ] Run tests (should fail)
- [ ] Implement `internal/middleware/guest.go`
- [ ] Run tests (should pass)

### Phase 2: Context Helpers (TDD)
- [ ] Write test: `TestGuestAccountFromContext_Present` — returns account and true
- [ ] Write test: `TestGuestAccountFromContext_Absent` — returns nil and false
- [ ] Run tests (should fail)
- [ ] Implement `WithGuestAccount` and `GuestAccountFromContext`
- [ ] Run tests (should pass)

### Phase 3: Wire into Router
- [ ] Apply `RequireGuestAuth` to `GET /guest/account`, `DELETE /guest/account`, `POST /guest/auth/logout`
- [ ] Confirm `POST /guest/auth/request-otp` and `POST /guest/auth/verify-otp` do NOT have `RequireGuestAuth`
- [ ] Run full test suite

---

## Testing Requirements

```go
func TestRequireGuestAuth_Valid(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    account := &models.GuestAccountView{
        GuestAccount: models.GuestAccount{ID: 42},
        Email:        testutil.StringPtr("alice@example.com"),
    }

    mockMgr := mocksvcs.NewMockGuestSessionManager(ctrl)
    mockMgr.EXPECT().GetSessionFromRequest(gomock.Any()).Return("sess-xyz", nil)
    mockMgr.EXPECT().GetSession(gomock.Any(), "sess-xyz").Return(&models.GuestSession{
        ID:             "sess-xyz",
        GuestAccountID: 42,
        ExpiresAt:      time.Now().Add(time.Hour),
    }, nil)

    mockSvc := mocksvcs.NewMockGuestAccountService(ctrl)
    mockSvc.EXPECT().GetAccount(gomock.Any(), int64(42)).Return(account, nil)

    var capturedAccount *models.GuestAccountView
    next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        capturedAccount, _ = middleware.GuestAccountFromContext(r.Context())
        w.WriteHeader(http.StatusOK)
    })

    mw := middleware.RequireGuestAuth(mockMgr, mockSvc)
    r := httptest.NewRequest(http.MethodGet, "/guest/account", nil)
    r.AddCookie(&http.Cookie{Name: "tinyrsvp_guest", Value: "sess-xyz"})
    w := httptest.NewRecorder()

    mw(next).ServeHTTP(w, r)

    if w.Code != http.StatusOK {
        t.Errorf("expected 200, got %d", w.Code)
    }
    if capturedAccount == nil || capturedAccount.ID != 42 {
        t.Errorf("expected account ID 42 in context, got %v", capturedAccount)
    }
}

func TestRequireGuestAuth_DoesNotSatisfyRequireAuth(t *testing.T) {
    // A request that has passed RequireGuestAuth must NOT have a staff user in context
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockMgr := mocksvcs.NewMockGuestSessionManager(ctrl)
    mockMgr.EXPECT().GetSessionFromRequest(gomock.Any()).Return("sess-xyz", nil)
    mockMgr.EXPECT().GetSession(gomock.Any(), "sess-xyz").Return(&models.GuestSession{
        ID: "sess-xyz", GuestAccountID: 1, ExpiresAt: time.Now().Add(time.Hour),
    }, nil)

    mockSvc := mocksvcs.NewMockGuestAccountService(ctrl)
    mockSvc.EXPECT().GetAccount(gomock.Any(), int64(1)).Return(
        &models.GuestAccountView{GuestAccount: models.GuestAccount{ID: 1}}, nil,
    )

    next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        _, staffPresent := auth.UserFromContext(r.Context())
        if staffPresent {
            t.Error("staff user context key must not be set by RequireGuestAuth")
        }
        w.WriteHeader(http.StatusOK)
    })

    mw := middleware.RequireGuestAuth(mockMgr, mockSvc)
    r := httptest.NewRequest(http.MethodGet, "/guest/account", nil)
    r.AddCookie(&http.Cookie{Name: "tinyrsvp_guest", Value: "sess-xyz"})
    mw(next).ServeHTTP(httptest.NewRecorder(), r)
}
```

---

## Dependencies

**Depends on:** Story 08 (guestauth package — GuestSessionManager, GuestAccountService interfaces)  
**Blocks:** Story 09 (handlers wire RequireGuestAuth), Story 11 (confirmation page checks for guest session)

---

## Definition of Done

- [ ] All acceptance criteria met
- [ ] All tasks completed
- [ ] All tests pass: `go test -timeout 30s -race ./internal/middleware/...`
- [ ] `RequireGuestAuth` and `RequireAuth` context keys confirmed distinct (by test)
- [ ] `go vet ./...` clean

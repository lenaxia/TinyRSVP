# TinyRSVP Testing Guide

## Testing Philosophy

TinyRSVP follows Test-Driven Development (TDD). Tests are written **before** implementation code.

**Workflow:**
1. Write a failing test
2. Write the minimal code to make it pass
3. Refactor if needed
4. Repeat

All tests must pass before committing. The pre-commit hook enforces this automatically.

---

## Test Categories

### Unit Tests
- **Purpose:** Test a single component in isolation
- **Speed:** < 10ms per test
- **Dependencies:** All external dependencies mocked
- **Pattern:** Use generated gomock mocks (see below)
- **Location:** `*_test.go` alongside the source file

```go
func TestGetInvite_Success(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockService := services.NewMockInviteService(ctrl)
    mockService.EXPECT().
        GetInviteByID(gomock.Any(), int64(1)).
        Return(&models.Invite{ID: 1}, nil)

    handler := NewGetInviteHandlers(mockService, nil)
    // ...
}
```

### Integration Tests
- **Purpose:** Test components working together with a real database
- **Speed:** < 500ms per test
- **Dependencies:** Real SQLite in-memory DB; external services mocked
- **Pattern:** Use `testutil.SetupTestDBWithMigrations`, real repositories

```go
func TestEventRepository_Create(t *testing.T) {
    db := testutil.SetupTestDBWithMigrations(t, "../../migrations/sqlite")
    repo := repositories.NewEventRepository(db)
    // test real SQL operations...
}
```

### E2E Tests
- **Purpose:** Test complete HTTP request/response cycles
- **Speed:** < 1s per test
- **Dependencies:** Full handler stack with mocked services
- **Location:** `tests/e2e/` or `*_integration_test.go` files

### UX Tests (Browser)
- **Purpose:** Test real browser flows (form submit, navigation, copy-link) against an in-process test server
- **Speed:** ~1-2 minutes total (browser startup overhead)
- **Dependencies:** Headless Chrome via `chromedp`; in-process `httptest.NewServer` with real SQLite DB and all real handlers wired
- **Location:** `tests/ux/`
- **Auth bypass:** Uses the `X-Test-User-ID` header (supported by production `RequireAuth` middleware) injected via `chromedp` network headers — no separate test auth path
- **Seed:** `SeedDefaults` (not `SeedThemes`) so the RSVP page renders as a legacy HTML form

```bash
go test -timeout 180s -v ./tests/ux/...
```

> **Note:** Headless Chrome must be installed. See `tests/ux/server_test.go` for the shared `setupUXTestServer` fixture and per-flow test files.

---

## When to Mock

| Situation | Decision | Reason |
|---|---|---|
| External HTTP services (OIDC, SMTP) | **Always mock** | Slow, non-deterministic, side effects |
| File system / network calls | **Always mock** | Slow, environment-dependent |
| Repositories in service tests | **Mock** | Isolate service logic from SQL |
| Services in handler tests | **Mock** | Isolate handler logic |
| Real DB in repository tests | **Use real** | Repository tests *are* the DB integration |
| Simple validators / pure functions | **Never mock** | No dependencies to isolate |
| The code under test | **Never mock** | That defeats the purpose |

---

## Using Generated Mocks

All major interfaces have generated mocks in `internal/testutil/mocks/`. Regenerate after interface changes:

```bash
./scripts/generate_mocks.sh
```

### Available Mock Packages

| Import alias | Path | Contents |
|---|---|---|
| `mocksvcs` | `internal/testutil/mocks/services` | `MockInviteService`, `MockRSVPService`, `MockDashboardService`, `MockEventService`, `MockEmailService`, `MockTemplateService`, `MockUserService`, `MockAdminDashboardService`, `MockUserListService` |
| `mockrepos` | `internal/testutil/mocks/repositories` | `MockEventRepository`, `MockInviteRepository`, `MockRSVPRepository`, `MockQuestionRepository`, `MockAnswerRepository`, `MockEmailQueueRepository`, `MockSessionRepository`, `MockUserRepository`, `MockTemplateRepository`, `MockConfigRepository` |
| `mockother` | `internal/testutil/mocks/other` | `MockDatabase`, `MockUserService` (auth), `MockSessionManager`, `MockAuthenticator`, `MockProvider`, `MockAuthorizationChecker`, `MockEventValidator`, `MockTemplateValidator`, `MockUserCounter`, `MockEventCounter`, `MockInviteCounter`, `MockJobsEventService`, `MockSMTPSender`, `MockRateLimiter`, `MockTemplateRenderer`, `MockEmailMetrics` |

### Basic Setup

```go
import (
    mocksvcs "github.com/lenaxia/tinyrsvp/internal/testutil/mocks/services"
    mockrepos "github.com/lenaxia/tinyrsvp/internal/testutil/mocks/repositories"
    "go.uber.org/mock/gomock"
)

func TestMyHandler(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockSvc := mocksvcs.NewMockInviteService(ctrl)
    mockRepo := mockrepos.NewMockEventRepository(ctrl)
    // ...
}
```

### Setting Expectations

```go
// Single call returning a value
mockSvc.EXPECT().
    GetInviteByID(gomock.Any(), int64(1)).
    Return(&models.Invite{ID: 1}, nil)

// Call that returns an error
mockSvc.EXPECT().
    GetInviteByID(gomock.Any(), int64(99)).
    Return(nil, &models.NotFoundError{Resource: "invite"})

// Call expected exactly N times
mockSvc.EXPECT().
    SendInvite(gomock.Any(), gomock.Any()).
    Return(nil).
    Times(3)

// Call expected zero or more times
mockSvc.EXPECT().
    GetInviteByID(gomock.Any(), gomock.Any()).
    Return(nil, nil).
    AnyTimes()
```

### Argument Matchers

```go
// Match any argument
gomock.Any()

// Match a specific value
gomock.Eq(int64(42))

// Match with a custom function
gomock.Cond(func(x any) bool {
    invite := x.(*models.Invite)
    return invite.Status == models.InviteStatusSent
})
```

### Ordered Calls

```go
first := mockRepo.EXPECT().GetByID(gomock.Any(), int64(1)).Return(event, nil)
second := mockSvc.EXPECT().ProcessEvent(gomock.Any(), event).Return(nil)
gomock.InOrder(first, second)
```

---

## Common Patterns

### Handler Test Pattern

```go
func TestMyHandler_Success(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    // 1. Create mocks
    mockSvc := mocksvcs.NewMockInviteService(ctrl)

    // 2. Set expectations
    mockSvc.EXPECT().
        GetInviteByID(gomock.Any(), int64(1)).
        Return(&models.Invite{ID: 1, EventID: 10}, nil)

    // 3. Build handler
    handler := NewMyHandler(mockSvc)

    // 4. Create request with auth context
    ctx := testutil.CreateAdminContext()
    req := httptest.NewRequest("GET", "/invites/1", nil).WithContext(ctx)
    // add chi URL params if needed:
    rctx := chi.NewRouteContext()
    rctx.URLParams.Add("id", "1")
    req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

    // 5. Execute and assert
    w := httptest.NewRecorder()
    handler.ServeHTTP(w, req)

    if w.Code != http.StatusOK {
        t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
    }
}
```

### Service Test Pattern (func-field mocks)

Service tests in packages like `internal/events/`, `internal/invites/`, `internal/auth/` use the **func-field pattern** — local mock structs with optional function fields. This is intentional: it provides useful defaults and works within packages that cannot import `testutil` (import cycle).

```go
type mockEventRepository struct {
    GetByIDFunc func(ctx context.Context, id int64) (*models.Event, error)
    // ...other methods with func fields...
}

func (m *mockEventRepository) GetByID(ctx context.Context, id int64) (*models.Event, error) {
    if m.GetByIDFunc != nil {
        return m.GetByIDFunc(ctx, id)
    }
    return nil, nil  // sensible default
}

func TestMyService(t *testing.T) {
    repo := &mockEventRepository{
        GetByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
            return &models.Event{ID: id, Title: "Test"}, nil
        },
    }
    svc := NewEventService(repo, nil)
    // ...
}
```

### Repository Test Pattern

Repository tests use a **real in-memory SQLite database**. Never use mocks in repository tests.

```go
func TestEventRepository_Create(t *testing.T) {
    database := setupEventTestDB(t)  // local helper: runs migrations, seeds user
    defer database.Close()

    repo := NewEventRepository(database)
    event := &models.Event{
        Title:     "Test Event",
        CreatedBy: 1,
        Status:    models.EventStatusDraft,
    }

    err := repo.Create(context.Background(), event)
    if err != nil {
        t.Fatalf("Create failed: %v", err)
    }
    if event.ID == 0 {
        t.Error("expected non-zero ID after create")
    }
}
```

### Error Scenario Testing

```go
func TestMyHandler_NotFound(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockSvc := mocksvcs.NewMockInviteService(ctrl)
    mockSvc.EXPECT().
        GetInviteByID(gomock.Any(), int64(99)).
        Return(nil, &models.NotFoundError{Resource: "invite"})

    handler := NewGetInviteHandlers(mockSvc, nil)
    // ...assert 404 response
}
```

---

## Test Utilities (`internal/testutil`)

### Pointer Helpers

```go
import "github.com/lenaxia/tinyrsvp/internal/testutil"

email := testutil.StringPtr("user@example.com")
capacity := testutil.IntPtr(100)
eventID := testutil.Int64Ptr(42)
active := testutil.BoolPtr(true)
deadline := testutil.TimePtr(time.Now().Add(24 * time.Hour))
score := testutil.Float64Ptr(4.5)
```

> **Note:** These helpers cannot be used in `internal/db/...`, `internal/auth/...`, or `internal/models/...` packages due to import cycles. Use local helpers in those packages.

### Database Helpers

```go
// Setup test DB with migrations (path relative to test file)
db := testutil.SetupTestDBWithMigrations(t, "../../migrations/sqlite")

// Create test data
user := testutil.CreateTestUser(t, db, models.RoleAdmin)
eventID := testutil.CreateTestEvent(t, db, user.ID)
inviteID := testutil.CreateTestInvite(t, db, eventID, "unique-token-hash")
```

### Auth Context Helpers

```go
// Admin context (for handler tests)
ctx := testutil.CreateAdminContext()

// Custom user context
user := &models.User{ID: 5, Role: models.RoleEventManager}
ctx := testutil.CreateTestContext(user)
```

---

## Running Tests

```bash
# All non-UX tests (excludes browser tests; fast feedback loop)
go test -timeout 60s $(go list ./... | grep -v '/tests/ux')

# All tests including UX (requires headless Chrome; slow)
go test -timeout 180s ./...

# UX tests only
go test -timeout 180s -v ./tests/ux/...

# Specific package
go test -timeout 30s ./internal/handlers/...

# Single test function
go test -timeout 30s ./internal/handlers/ -run TestGetInvite_Success

# With coverage
go test -timeout 30s ./... -cover

# With race detector (recommended before merging)
go test -timeout 60s -race ./...

# With verbose output
go test -timeout 30s -v ./internal/auth/...
```

> **Always pass `-timeout`** — without it, a hung test runs forever. The pre-commit hook enforces `-timeout 30s`.

---

## Keep-As-Is Exceptions

These test files intentionally keep manual mock structs and are **not** migrated to gomock:

| Package | Reason |
|---|---|
| `internal/auth/*_test.go` | Import cycle: `testutil/mocks/other` imports `internal/auth` |
| `internal/events/*_test.go` (service tests) | Func-field pattern provides useful defaults |
| `internal/invites/*_test.go` (service tests) | Func-field pattern |
| `internal/db/repositories/*_test.go` | Use real DB; pointer helpers can't use testutil (import cycle) |
| `internal/handlers/auth_test.go` | `AuthResult` type incompatibility between packages |
| Handler tests with func-field patterns | Architectural mock types shared across multiple test files |

---

## Troubleshooting

### "unexpected call to X"
A mock method was called that had no `.EXPECT()` set up. Add the expectation or use `.AnyTimes()` for optional calls.

### "expected call to X that was not received"
The code under test did not call the mocked method. Check the logic path being exercised.

### Test flakiness
- Use `gomock.Any()` for timestamps and non-deterministic values
- Avoid `time.Sleep` — use channels or polling helpers instead
- Ensure test DB cleanup with `defer database.Close()`

### Import cycle when adding testutil
If you get `import cycle not allowed in test`, the package you're testing is already imported by `testutil`. Use local pointer helpers instead.

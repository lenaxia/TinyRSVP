# User Story: Generate Priority 1 Repository Mocks

**Epic:** [12_EPIC_test_infrastructure.md](../12_EPIC_test_infrastructure.md)
**Priority:** Critical
**Status:** Not Started
**Estimated Effort:** 1 hour
**Phase:** 2 - Mock Generation Setup

---

## User Story

As a **developer**, I want **generated mocks for core repository interfaces** so that **I can test services and handlers without manual mock definitions**.

---

## Acceptance Criteria

- [ ] Mock for db.Database generated
- [ ] Mock for EventRepository generated  
- [ ] Mock for InviteRepository generated
- [ ] Mock for UserRepository generated
- [ ] Mock for AuthorizationChecker generated
- [ ] All mocks compile successfully
- [ ] Mocks can be used in tests

---

## Interfaces to Mock

1. **db.Database** - Foundation for all repositories
2. **repositories.EventRepository** - 17 methods, widely used
3. **repositories.InviteRepository** - 13 methods
4. **repositories.UserRepository** - 12 methods
5. **auth.AuthorizationChecker** - Permission testing

---

## Mock Generation Commands

```bash
# Database interface
mockgen -source=internal/db/database.go \
    -destination=internal/testutil/mocks/mock_database.go \
    -package=mocks

# EventRepository
mockgen -source=internal/db/repositories/event_repository.go \
    -destination=internal/testutil/mocks/mock_event_repository.go \
    -package=mocks

# InviteRepository
mockgen -source=internal/db/repositories/invite_repository.go \
    -destination=internal/testutil/mocks/mock_invite_repository.go \
    -package=mocks

# UserRepository
mockgen -source=internal/db/repositories/user_repository.go \
    -destination=internal/testutil/mocks/mock_user_repository.go \
    -package=mocks

# AuthorizationChecker
mockgen -source=internal/auth/authorization.go \
    -destination=internal/testutil/mocks/mock_authorization.go \
    -package=mocks
```

---

## Tasks

- [ ] Add commands to `scripts/generate_mocks.sh`
- [ ] Run script to generate mocks
- [ ] Verify all mocks compile: `go build ./internal/testutil/mocks/...`
- [ ] Test import in a sample test file
- [ ] Commit generated mocks

---

## Dependencies

**Depends on:** Story 05 (mockgen setup)  
**Blocks:** Story 09 (validate pattern)

---

## Validation

```bash
./scripts/generate_mocks.sh
go build ./internal/testutil/mocks/...
go test ./internal/testutil/mocks/...
```
